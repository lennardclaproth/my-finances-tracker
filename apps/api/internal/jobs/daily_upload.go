package jobs

import (
	"context"
	"errors"
	"fmt"
	"io"
	"sync"
	"time"

	"github.com/google/uuid"
	"github.com/lennardclaproth/my-finances-tracker/internal/logging"
	"github.com/lennardclaproth/my-finances-tracker/internal/marketdata"
	marketdataParsers "github.com/lennardclaproth/my-finances-tracker/internal/marketdata/parsers"
	"github.com/lennardclaproth/my-finances-tracker/internal/observability"
	"go.elastic.co/apm/v2"
)

const (
	defaultDailyUploadQueueSize  = 256
	defaultDailyUploadSyncBatch  = 512
	defaultDailyUploadReconcile  = 5 * time.Second
	defaultDailyUploadErrMaxRows = 50
)

var (
	ErrDailyUploadQueueFull = errors.New("daily upload queue is full")
)

type dailyUploadStore interface {
	FetchByID(ctx context.Context, id uuid.UUID) (*marketdata.DailyUpload, error)
	ListPending(ctx context.Context, limit int) ([]*marketdata.DailyUpload, error)
	TryMarkProcessing(ctx context.Context, id uuid.UUID) (bool, error)
	UpdateState(ctx context.Context, upload *marketdata.DailyUpload) error
}

type dailyUploadListingStore interface {
	FetchByID(ctx context.Context, id uuid.UUID) (*marketdata.Listing, error)
}

type dailyUploadDailyStore interface {
	CreateWithInsertStatus(ctx context.Context, daily *marketdata.Daily) (bool, error)
}

type dailyUploadFileReader interface {
	ReadCsv(path string) (io.ReadCloser, error)
}

type dailyUploadParserFactory func(source marketdata.Source) (marketdataParsers.DailyParser, error)

type DailyUploadEnqueuer interface {
	Enqueue(ctx context.Context, uploadID uuid.UUID) error
}

type DailyUploadJob struct {
	uploadStore    dailyUploadStore
	listingStore   dailyUploadListingStore
	dailyStore     dailyUploadDailyStore
	fileReader     dailyUploadFileReader
	log            logging.Logger
	reconcileEvery time.Duration
	syncBatchSize  int
	queue          chan uuid.UUID
	inQueue        map[uuid.UUID]struct{}
	inFlight       map[uuid.UUID]struct{}
	queueHeaders   map[uuid.UUID]map[string]string
	mu             sync.Mutex
	parserFactory  dailyUploadParserFactory
}

func NewDailyUploadJob(
	uploadStore dailyUploadStore,
	listingStore dailyUploadListingStore,
	dailyStore dailyUploadDailyStore,
	fileReader dailyUploadFileReader,
	log logging.Logger,
	reconcileEvery time.Duration,
	queueSize int,
) *DailyUploadJob {
	if queueSize <= 0 {
		queueSize = defaultDailyUploadQueueSize
	}
	if reconcileEvery <= 0 {
		reconcileEvery = defaultDailyUploadReconcile
	}
	return &DailyUploadJob{
		uploadStore:    uploadStore,
		listingStore:   listingStore,
		dailyStore:     dailyStore,
		fileReader:     fileReader,
		log:            log,
		reconcileEvery: reconcileEvery,
		syncBatchSize:  defaultDailyUploadSyncBatch,
		queue:          make(chan uuid.UUID, queueSize),
		inQueue:        make(map[uuid.UUID]struct{}),
		inFlight:       make(map[uuid.UUID]struct{}),
		queueHeaders:   make(map[uuid.UUID]map[string]string),
		parserFactory:  marketdataParsers.CreateDailyParser,
	}
}

func (j *DailyUploadJob) Name() string {
	return "DailyUploadJob"
}

func (j *DailyUploadJob) Enqueue(ctx context.Context, uploadID uuid.UUID) error {
	if uploadID == uuid.Nil {
		return fmt.Errorf("cannot enqueue nil daily upload id")
	}

	j.mu.Lock()
	if _, exists := j.inQueue[uploadID]; exists {
		j.mu.Unlock()
		return nil
	}
	if _, exists := j.inFlight[uploadID]; exists {
		j.mu.Unlock()
		return nil
	}
	j.inQueue[uploadID] = struct{}{}
	j.queueHeaders[uploadID] = observability.PropagationHeadersFromContext(ctx)
	j.mu.Unlock()

	select {
	case <-ctx.Done():
		j.mu.Lock()
		delete(j.inQueue, uploadID)
		delete(j.queueHeaders, uploadID)
		j.mu.Unlock()
		return ctx.Err()
	case j.queue <- uploadID:
		return nil
	default:
		j.mu.Lock()
		delete(j.inQueue, uploadID)
		delete(j.queueHeaders, uploadID)
		j.mu.Unlock()
		return ErrDailyUploadQueueFull
	}
}

func (j *DailyUploadJob) Start(ctx context.Context) error {
	if err := j.syncQueueFromDB(ctx); err != nil {
		j.log.Error(ctx, "failed initial daily upload queue reconciliation", err)
	}

	ticker := time.NewTicker(j.reconcileEvery)
	defer ticker.Stop()

	for {
		select {
		case <-ctx.Done():
			return ctx.Err()
		case uploadID := <-j.queue:
			headers := j.markDequeued(uploadID)
			if err := j.processByID(ctx, uploadID, headers); err != nil {
				j.log.Error(ctx, "failed processing daily upload", err, "upload_id", uploadID)
			}
			j.markDone(uploadID)
			if err := j.syncQueueFromDB(ctx); err != nil {
				j.log.Error(ctx, "failed daily upload reconciliation after processing", err)
			}
		case <-ticker.C:
			if err := j.syncQueueFromDB(ctx); err != nil {
				j.log.Error(ctx, "failed periodic daily upload queue reconciliation", err)
			}
		}
	}
}

func (j *DailyUploadJob) syncQueueFromDB(ctx context.Context) error {
	pending, err := j.uploadStore.ListPending(ctx, j.syncBatchSize)
	if err != nil {
		return err
	}
	for _, upload := range pending {
		if upload == nil {
			continue
		}
		err := j.Enqueue(ctx, upload.ID)
		if err == nil {
			continue
		}
		if errors.Is(err, ErrDailyUploadQueueFull) {
			return nil
		}
		if errors.Is(err, context.Canceled) || errors.Is(err, context.DeadlineExceeded) {
			return err
		}
		j.log.Error(ctx, "failed to enqueue pending daily upload", err, "upload_id", upload.ID)
	}
	return nil
}

func (j *DailyUploadJob) markDequeued(uploadID uuid.UUID) map[string]string {
	j.mu.Lock()
	defer j.mu.Unlock()
	headers := j.queueHeaders[uploadID]
	delete(j.inQueue, uploadID)
	delete(j.queueHeaders, uploadID)
	j.inFlight[uploadID] = struct{}{}
	return headers
}

func (j *DailyUploadJob) markDone(uploadID uuid.UUID) {
	j.mu.Lock()
	defer j.mu.Unlock()
	delete(j.inQueue, uploadID)
	delete(j.queueHeaders, uploadID)
	delete(j.inFlight, uploadID)
}

func (j *DailyUploadJob) processByID(ctx context.Context, uploadID uuid.UUID, headers map[string]string) (err error) {
	ctx = observability.ContextWithPropagationHeaders(ctx, headers)
	apmTx, txCtx, txErr := observability.StartTransactionFromHeaders(
		ctx,
		observability.JobOperation("daily_upload"),
		"job",
		headers,
	)
	if txErr != nil {
		j.log.Error(ctx, "failed to parse incoming trace headers for daily upload job", txErr, "upload_id", uploadID)
	}
	ctx = txCtx
	apmTx.Result = "success"
	apmTx.Outcome = "success"
	observability.SetSafeTransactionLabels(apmTx, map[string]any{
		"operation": observability.JobOperation("daily_upload"),
		"component": "job",
		"upload_id": uploadID.String(),
		"stage":     "process",
	})
	defer func() {
		if err != nil {
			apmTx.Result = "error"
			apmTx.Outcome = "failure"
			apm.CaptureError(ctx, err).Send()
		}
		apmTx.End()
	}()

	upload, err := j.uploadStore.FetchByID(ctx, uploadID)
	if err != nil {
		if errors.Is(err, marketdata.ErrDailyUploadNotFound) {
			return nil
		}
		return err
	}

	claimed, err := j.uploadStore.TryMarkProcessing(ctx, upload.ID)
	if err != nil {
		return err
	}
	if !claimed {
		return nil
	}
	upload.MarkProcessing()

	listing, err := j.listingStore.FetchByID(ctx, upload.ListingID)
	if err != nil {
		return j.markFailed(ctx, upload, fmt.Sprintf("failed to fetch listing: %v", err))
	}
	if listing == nil {
		return j.markFailed(ctx, upload, "listing not found")
	}

	parser, err := j.parserFactory(listing.Source)
	if err != nil {
		return j.markFailed(ctx, upload, fmt.Sprintf("unsupported parser for source %s", listing.Source))
	}

	rc, err := j.fileReader.ReadCsv(upload.StoredFilename)
	if err != nil {
		return j.markFailed(ctx, upload, fmt.Sprintf("failed to open uploaded file: %v", err))
	}

	parseSpan, parseCtx := apm.StartSpan(ctx, "parse", "job")
	parsed, err := parser.ParseAll(rc)
	parseSpan.End()
	if err != nil {
		return j.markFailed(ctx, upload, fmt.Sprintf("failed parsing file: %v", err))
	}

	allErrors := make([]marketdata.DailyUploadRowError, 0, len(parsed.RowErrors))
	allErrors = append(allErrors, parsed.RowErrors...)
	insertedRows := 0
	duplicateRows := 0

	persistSpan, persistCtx := apm.StartSpan(parseCtx, "persist", "job")
	defer persistSpan.End()

	for _, row := range parsed.Rows {
		daily, err := marketdata.NewDaily(
			listing.Symbol,
			row.Date,
			row.Open,
			row.Close,
			row.High,
			row.Low,
			row.Volume,
		)
		if err != nil {
			allErrors = append(allErrors, marketdata.DailyUploadRowError{
				RowNumber: row.RowNumber,
				Reason:    "invalid price payload",
			})
			continue
		}
		daily.ListingID = listing.ID
		inserted, err := j.dailyStore.CreateWithInsertStatus(persistCtx, &daily)
		if err != nil {
			allErrors = append(allErrors, marketdata.DailyUploadRowError{
				RowNumber: row.RowNumber,
				Reason:    "failed to persist row",
			})
			continue
		}
		if inserted {
			insertedRows++
		} else {
			duplicateRows++
		}
	}

	statusMsg := ""
	if len(allErrors) > 0 {
		statusMsg = "completed with row errors"
	}
	if err := upload.MarkCompleted(parsed.TotalRows, insertedRows, duplicateRows, len(allErrors), allErrors, statusMsg); err != nil {
		return err
	}
	if len(upload.RowErrors) > defaultDailyUploadErrMaxRows {
		upload.RowErrors = upload.RowErrors[:defaultDailyUploadErrMaxRows]
		if err := upload.SetRowErrors(upload.RowErrors, defaultDailyUploadErrMaxRows); err != nil {
			return err
		}
	}
	if err := j.uploadStore.UpdateState(persistCtx, upload); err != nil {
		return err
	}
	return nil
}

func (j *DailyUploadJob) markFailed(ctx context.Context, upload *marketdata.DailyUpload, message string) error {
	if upload == nil {
		return nil
	}
	upload.StatusMessage = message
	if err := upload.MarkFailed(message); err != nil {
		return err
	}
	if err := j.uploadStore.UpdateState(ctx, upload); err != nil {
		j.log.Error(ctx, "failed to mark daily upload as failed", err, "upload_id", upload.ID)
		return err
	}
	return nil
}
