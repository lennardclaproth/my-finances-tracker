package eod

import (
	"context"
	"errors"
	"fmt"
	"io"
	"time"

	"github.com/google/uuid"
	"github.com/lennardclaproth/my-finances-tracker/internal/logging"
	"github.com/lennardclaproth/my-finances-tracker/internal/marketdata"
	"github.com/lennardclaproth/my-finances-tracker/internal/observability"
	"go.elastic.co/apm/v2"
)

type dailyUploadProcessStore interface {
	FetchByID(ctx context.Context, id uuid.UUID) (*DailyUpload, error)
	TryMarkProcessing(ctx context.Context, id uuid.UUID) (bool, error)
	UpdateState(ctx context.Context, upload *DailyUpload) error
}

type dailyUploadProcessListingStore interface {
	FetchByID(ctx context.Context, id uuid.UUID) (*Listing, error)
}

type dailyUploadProcessDailyStore interface {
	CreateWithInsertStatus(ctx context.Context, daily *EOD) (bool, error)
}

type dailyUploadProcessFileReader interface {
	ReadCsv(path string) (io.ReadCloser, error)
}

// DailyUploadParsedRow is one parsed daily row from an uploaded file.
type DailyUploadParsedRow struct {
	RowNumber int
	Date      time.Time
	Open      float64
	High      float64
	Low       float64
	Close     float64
	Volume    int64
}

// DailyUploadParseResult captures parser output for async daily uploads.
type DailyUploadParseResult struct {
	Rows      []DailyUploadParsedRow
	RowErrors []DailyUploadRowError
	TotalRows int
}

// DailyUploadParser parses one upload file into normalized rows and row errors.
type DailyUploadParser interface {
	ParseAll(rc io.ReadCloser) (DailyUploadParseResult, error)
}

// DailyUploadParserFactory resolves a parser for a listing source.
type DailyUploadParserFactory func(source marketdata.Source) (DailyUploadParser, error)

// DailyUploadProcessor executes async daily upload processing for one upload aggregate.
type DailyUploadProcessor struct {
	uploadStore    dailyUploadProcessStore
	listingStore   dailyUploadProcessListingStore
	dailyStore     dailyUploadProcessDailyStore
	fileReader     dailyUploadProcessFileReader
	parserFactory  DailyUploadParserFactory
	log            logging.Logger
	rowErrorSample int
}

// NewDailyUploadProcessor constructs the daily upload processing orchestrator.
func NewDailyUploadProcessor(
	uploadStore dailyUploadProcessStore,
	listingStore dailyUploadProcessListingStore,
	dailyStore dailyUploadProcessDailyStore,
	fileReader dailyUploadProcessFileReader,
	parserFactory DailyUploadParserFactory,
	log logging.Logger,
	rowErrorSample int,
) *DailyUploadProcessor {
	if rowErrorSample <= 0 {
		rowErrorSample = 50
	}
	return &DailyUploadProcessor{
		uploadStore:    uploadStore,
		listingStore:   listingStore,
		dailyStore:     dailyStore,
		fileReader:     fileReader,
		parserFactory:  parserFactory,
		log:            log,
		rowErrorSample: rowErrorSample,
	}
}

// ProcessByID processes one daily upload aggregate and persists terminal state.
func (p *DailyUploadProcessor) ProcessByID(ctx context.Context, uploadID uuid.UUID, headers map[string]string) (err error) {
	ctx = observability.ContextWithPropagationHeaders(ctx, headers)
	apmTx, txCtx, txErr := observability.StartTransactionFromHeaders(
		ctx,
		observability.JobOperation("daily_upload"),
		"job",
		headers,
	)
	if txErr != nil {
		p.log.Error(ctx, "failed to parse incoming trace headers for daily upload job", txErr, "upload_id", uploadID)
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

	upload, err := p.uploadStore.FetchByID(ctx, uploadID)
	if err != nil {
		if errors.Is(err, ErrDailyUploadNotFound) {
			return nil
		}
		return err
	}

	claimed, err := p.uploadStore.TryMarkProcessing(ctx, upload.ID)
	if err != nil {
		return err
	}
	if !claimed {
		return nil
	}
	upload.MarkProcessing()

	listing, err := p.listingStore.FetchByID(ctx, upload.ListingID)
	if err != nil {
		return p.markFailed(ctx, upload, fmt.Sprintf("failed to fetch listing: %v", err))
	}
	if listing == nil {
		return p.markFailed(ctx, upload, "listing not found")
	}

	parser, err := p.parserFactory(listing.Source)
	if err != nil {
		return p.markFailed(ctx, upload, fmt.Sprintf("unsupported parser for source %s", listing.Source))
	}

	rc, err := p.fileReader.ReadCsv(upload.StoredFilename)
	if err != nil {
		return p.markFailed(ctx, upload, fmt.Sprintf("failed to open uploaded file: %v", err))
	}

	parseSpan, parseCtx := apm.StartSpan(ctx, "parse", "job")
	parsed, err := parser.ParseAll(rc)
	parseSpan.End()
	if err != nil {
		return p.markFailed(ctx, upload, fmt.Sprintf("failed parsing file: %v", err))
	}

	allErrors := make([]DailyUploadRowError, 0, len(parsed.RowErrors))
	allErrors = append(allErrors, parsed.RowErrors...)
	insertedRows := 0
	duplicateRows := 0

	persistSpan, persistCtx := apm.StartSpan(parseCtx, "persist", "job")
	defer persistSpan.End()

	for _, row := range parsed.Rows {
		daily, err := NewEOD(
			listing.Symbol,
			row.Date,
			row.Open,
			row.Close,
			row.High,
			row.Low,
			row.Volume,
		)
		if err != nil {
			allErrors = append(allErrors, DailyUploadRowError{
				RowNumber: row.RowNumber,
				Reason:    "invalid price payload",
			})
			continue
		}
		daily.ListingID = listing.ID
		inserted, err := p.dailyStore.CreateWithInsertStatus(persistCtx, &daily)
		if err != nil {
			allErrors = append(allErrors, DailyUploadRowError{
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
	if err := upload.SetRowErrors(upload.RowErrors, p.rowErrorSample); err != nil {
		return err
	}
	if err := p.uploadStore.UpdateState(persistCtx, upload); err != nil {
		return err
	}
	return nil
}

func (p *DailyUploadProcessor) markFailed(ctx context.Context, upload *DailyUpload, message string) error {
	if upload == nil {
		return nil
	}
	upload.StatusMessage = message
	if err := upload.MarkFailed(message); err != nil {
		return err
	}
	if err := upload.SetRowErrors(upload.RowErrors, p.rowErrorSample); err != nil {
		return err
	}
	if err := p.uploadStore.UpdateState(ctx, upload); err != nil {
		p.log.Error(ctx, "failed to mark daily upload as failed", err, "upload_id", upload.ID)
		return err
	}
	return nil
}
