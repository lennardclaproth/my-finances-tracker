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
	parserFactory := func(source marketdata.Source) (marketdata.DailyUploadParser, error) {
		parser, err := j.parserFactory(source)
		if err != nil {
			return nil, err
		}
		return dailyUploadParserAdapter{parser: parser}, nil
	}
	processor := marketdata.NewDailyUploadProcessor(
		j.uploadStore,
		j.listingStore,
		j.dailyStore,
		j.fileReader,
		parserFactory,
		j.log,
		defaultDailyUploadErrMaxRows,
	)
	return processor.ProcessByID(ctx, uploadID, headers)
}

type dailyUploadParserAdapter struct {
	parser marketdataParsers.DailyParser
}

func (a dailyUploadParserAdapter) ParseAll(rc io.ReadCloser) (marketdata.DailyUploadParseResult, error) {
	parsed, err := a.parser.ParseAll(rc)
	if err != nil {
		return marketdata.DailyUploadParseResult{}, err
	}
	rows := make([]marketdata.DailyUploadParsedRow, 0, len(parsed.Rows))
	for _, row := range parsed.Rows {
		rows = append(rows, marketdata.DailyUploadParsedRow{
			RowNumber: row.RowNumber,
			Date:      row.Date,
			Open:      row.Open,
			High:      row.High,
			Low:       row.Low,
			Close:     row.Close,
			Volume:    row.Volume,
		})
	}
	return marketdata.DailyUploadParseResult{
		Rows:      rows,
		RowErrors: parsed.RowErrors,
		TotalRows: parsed.TotalRows,
	}, nil
}
