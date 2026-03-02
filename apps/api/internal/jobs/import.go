package jobs

import (
	"context"
	"errors"
	"fmt"
	"io"
	"math"
	"sync"
	"time"

	"github.com/google/uuid"
	"github.com/lennardclaproth/my-finances-tracker/api"
	"github.com/lennardclaproth/my-finances-tracker/internal/bus"
	"github.com/lennardclaproth/my-finances-tracker/internal/cashflow"
	cashflowParsers "github.com/lennardclaproth/my-finances-tracker/internal/cashflow/parsers"
	"github.com/lennardclaproth/my-finances-tracker/internal/importer"
	"github.com/lennardclaproth/my-finances-tracker/internal/logging"
	"github.com/lennardclaproth/my-finances-tracker/internal/portfolio"
	portfolioParsers "github.com/lennardclaproth/my-finances-tracker/internal/portfolio/parsers"
	"github.com/lennardclaproth/my-finances-tracker/internal/vendor"
	"go.elastic.co/apm/v2"
)

const (
	defaultImportQueueSize  = 256
	defaultSyncBatchSize    = 512
	defaultReconcileTimeout = 5 * time.Second
)

var (
	ErrImportQueueFull = errors.New("import queue is full")
)

type importVendorStore interface {
	FetchById(ctx context.Context, id uuid.UUID) (*vendor.Vendor, error)
}

type importStore interface {
	FetchByID(ctx context.Context, id uuid.UUID) (*importer.Import, error)
	ListPending(ctx context.Context, limit int) ([]*importer.Import, error)
	TryMarkInProgress(ctx context.Context, id uuid.UUID) (bool, error)
	UpdateState(ctx context.Context, imp *importer.Import) error
}

type cashflowTransactionStore interface {
	Create(ctx context.Context, tx *cashflow.Transaction) error
}

type portfolioTransactionStore interface {
	Create(ctx context.Context, tx *portfolio.Transaction) error
}

type importFileReader interface {
	ReadCsv(path string) (io.ReadCloser, error)
}

type cashflowParserFactory func(id vendor.VendorID) (cashflow.CsvParser, error)
type portfolioParserFactory func(id vendor.VendorID) (portfolio.CsvParser, error)

// ImportJob consumes import IDs from a bounded queue and reconciles with DB pending imports.
type ImportJob struct {
	vendorStore     importVendorStore
	importStore     importStore
	cashflowStore   cashflowTransactionStore
	portfolioStore  portfolioTransactionStore
	fileReader      importFileReader
	log             logging.Logger
	reconcileEvery  time.Duration
	syncBatchSize   int
	queue           chan uuid.UUID
	inQueue         map[uuid.UUID]struct{}
	inFlight        map[uuid.UUID]struct{}
	mu              sync.Mutex
	cashflowParser  cashflowParserFactory
	portfolioParser portfolioParserFactory
	b               bus.Bus
}

func NewImportJob(
	vendorStore importVendorStore,
	importStore importStore,
	cashflowStore cashflowTransactionStore,
	portfolioStore portfolioTransactionStore,
	fileReader importFileReader,
	log logging.Logger,
	reconcileEvery time.Duration,
	queueSize int,
	b bus.Bus,
) *ImportJob {
	if queueSize <= 0 {
		queueSize = defaultImportQueueSize
	}
	if reconcileEvery <= 0 {
		reconcileEvery = defaultReconcileTimeout
	}
	return &ImportJob{
		vendorStore:     vendorStore,
		importStore:     importStore,
		cashflowStore:   cashflowStore,
		portfolioStore:  portfolioStore,
		fileReader:      fileReader,
		log:             log,
		reconcileEvery:  reconcileEvery,
		syncBatchSize:   defaultSyncBatchSize,
		queue:           make(chan uuid.UUID, queueSize),
		inQueue:         make(map[uuid.UUID]struct{}),
		inFlight:        make(map[uuid.UUID]struct{}),
		cashflowParser:  cashflowParsers.CreateCsvParser,
		portfolioParser: portfolioParsers.CreateCsvParser,
		b:               b,
	}
}

func (j *ImportJob) Name() string {
	return "ImportJob"
}

func (j *ImportJob) Enqueue(ctx context.Context, importID uuid.UUID) error {
	if importID == uuid.Nil {
		return fmt.Errorf("cannot enqueue nil import id")
	}

	j.mu.Lock()
	if _, exists := j.inQueue[importID]; exists {
		j.mu.Unlock()
		return nil
	}
	if _, exists := j.inFlight[importID]; exists {
		j.mu.Unlock()
		return nil
	}
	j.inQueue[importID] = struct{}{}
	j.mu.Unlock()

	select {
	case <-ctx.Done():
		j.mu.Lock()
		delete(j.inQueue, importID)
		j.mu.Unlock()
		return ctx.Err()
	case j.queue <- importID:
		return nil
	default:
		j.mu.Lock()
		delete(j.inQueue, importID)
		j.mu.Unlock()
		return ErrImportQueueFull
	}
}

func (j *ImportJob) Start(ctx context.Context) error {
	if err := j.syncQueueFromDB(ctx); err != nil {
		j.log.Error(ctx, "failed initial import queue reconciliation", err)
	}

	ticker := time.NewTicker(j.reconcileEvery)
	defer ticker.Stop()

	for {
		select {
		case <-ctx.Done():
			return ctx.Err()
		case importID := <-j.queue:
			j.markDequeued(importID)
			if err := j.processByID(ctx, importID); err != nil {
				j.log.Error(ctx, "failed processing import", err, "import_id", importID)
			}
			j.markDone(importID)
			if err := j.syncQueueFromDB(ctx); err != nil {
				j.log.Error(ctx, "failed import queue reconciliation after processing", err)
			}
		case <-ticker.C:
			if err := j.syncQueueFromDB(ctx); err != nil {
				j.log.Error(ctx, "failed periodic import queue reconciliation", err)
			}
		}
	}
}

func (j *ImportJob) syncQueueFromDB(ctx context.Context) error {
	pending, err := j.importStore.ListPending(ctx, j.syncBatchSize)
	if err != nil {
		return err
	}
	for _, imp := range pending {
		if imp == nil {
			continue
		}
		err := j.Enqueue(ctx, imp.ID)
		if err == nil {
			continue
		}
		if errors.Is(err, ErrImportQueueFull) {
			// Stop filling on first full signal; periodic reconciliation will continue later.
			return nil
		}
		if errors.Is(err, context.Canceled) || errors.Is(err, context.DeadlineExceeded) {
			return err
		}
		j.log.Error(ctx, "failed to enqueue pending import", err, "import_id", imp.ID)
	}
	return nil
}

func (j *ImportJob) markDequeued(importID uuid.UUID) {
	j.mu.Lock()
	defer j.mu.Unlock()
	delete(j.inQueue, importID)
	j.inFlight[importID] = struct{}{}
}

func (j *ImportJob) markDone(importID uuid.UUID) {
	j.mu.Lock()
	defer j.mu.Unlock()
	delete(j.inQueue, importID)
	delete(j.inFlight, importID)
}

func (j *ImportJob) processByID(ctx context.Context, importID uuid.UUID) error {
	apmTx := apm.DefaultTracer().StartTransaction("ImportJob.processByID", "job")
	defer apmTx.End()
	ctx = apm.ContextWithTransaction(ctx, apmTx)

	imp, err := j.importStore.FetchByID(ctx, importID)
	if err != nil {
		if errors.Is(err, importer.ErrNoImportsPending) {
			return nil
		}
		return err
	}
	claimed, err := j.importStore.TryMarkInProgress(ctx, imp.ID)
	if err != nil {
		return err
	}
	if !claimed {
		return nil
	}

	v, err := j.vendorStore.FetchById(ctx, imp.VendorID)
	if err != nil {
		j.markImportFailed(ctx, imp, fmt.Errorf("fetch vendor: %w", err))
		return err
	}

	totalRows, duplicates, importedCount, failedCount, err := j.processCashflow(ctx, imp, v)
	if err != nil {
		j.markImportFailed(ctx, imp, fmt.Errorf("process cashflow rows: %w", err))
		return err
	}

	if v.Type == vendor.VendorTypeBrokerage {
		pDuplicates, pFailed, pErr := j.processPortfolio(ctx, imp, v)
		duplicates += pDuplicates
		failedCount += pFailed
		if pErr != nil {
			// Keep import as completed with failures so users can recover manually.
			failedCount += totalRows
			j.log.Error(ctx, "portfolio processing failed; import will be completed with failures", pErr, "import_id", imp.ID)
		}
	}

	imp.MarkCompleted(duplicates, totalRows, importedCount, failedCount)
	if err := j.importStore.UpdateState(ctx, imp); err != nil {
		j.log.Error(ctx, "failed to mark import as completed", err, "import_id", imp.ID)
		return err
	}
	return nil
}

func (j *ImportJob) processCashflow(ctx context.Context, imp *importer.Import, v *vendor.Vendor) (totalRows, duplicates, importedCount, failedCount int, err error) {
	parser, err := j.cashflowParser(v.Name)
	if err != nil {
		return 0, 0, 0, 0, err
	}
	rc, err := j.fileReader.ReadCsv(imp.Path)
	if err != nil {
		return 0, 0, 0, 0, err
	}
	seq, err := parser.ParseAll(rc)
	if err != nil {
		return 0, 0, 0, 0, err
	}

	for rowNumber, txd := range seq {
		totalRows++
		amount := math.Abs(txd.Amount)
		tx, err := cashflow.NewTransaction(
			txd.Description,
			txd.Note,
			string(v.Name),
			txd.Direction,
			amount,
			txd.Date,
			rowNumber,
			imp.ID,
			txd.AccountType,
			imp.AccountID,
		)
		if err != nil {
			failedCount++
			j.log.Error(ctx, "failed creating cashflow transaction", err, "import_id", imp.ID, "row_number", rowNumber)
			continue
		}

		if err := j.cashflowStore.Create(ctx, tx); err != nil {
			if errors.Is(err, cashflow.ErrDuplicateTransaction) {
				duplicates++
				continue
			}
			failedCount++
			j.log.Error(ctx, "failed persisting cashflow transaction", err, "import_id", imp.ID, "row_number", rowNumber)
			continue
		}
		importedCount++
	}

	return totalRows, duplicates, importedCount, failedCount, nil
}

func (j *ImportJob) processPortfolio(ctx context.Context, imp *importer.Import, v *vendor.Vendor) (duplicates, failedCount int, err error) {
	parser, err := j.portfolioParser(v.Name)
	if err != nil {
		return 0, 0, err
	}
	rc, err := j.fileReader.ReadCsv(imp.Path)
	if err != nil {
		return 0, 0, err
	}
	seq, err := parser.ParseAll(rc)
	if err != nil {
		return 0, 0, err
	}

	for rowNumber, txd := range seq {
		ptx, err := portfolio.NewTransaction(txd, rowNumber, imp.ID, imp.AccountID, nil)
		if err != nil {
			failedCount++
			j.log.Error(ctx, "failed creating portfolio transaction", err, "import_id", imp.ID, "row_number", rowNumber)
			continue
		}
		if err := j.portfolioStore.Create(ctx, ptx); err != nil {
			if errors.Is(err, portfolio.ErrDuplicateTransaction) {
				duplicates++
				continue
			}
			failedCount++
			j.log.Error(ctx, "failed persisting portfolio transaction", err, "import_id", imp.ID, "row_number", rowNumber)
			continue
		}
	}
	// publish TransactionsCreated event
	if imp.AccountID == nil {
		j.log.Info(ctx, "skip transactions created event: import has no account id", "import_id", imp.ID)
		return duplicates, failedCount, nil
	}
	if j.b == nil {
		j.log.Info(ctx, "skip transactions created event: bus not configured", "import_id", imp.ID, "account_id", imp.AccountID.String())
		return duplicates, failedCount, nil
	}
	msg, err := bus.NewJSONEnvelope(api.TransactionsCreated{AccID: *imp.AccountID})
	if err != nil {
		j.log.Error(ctx, "failed to encode transactions created event", err, "import_id", imp.ID, "account_id", imp.AccountID.String())
		return duplicates, failedCount, nil
	}
	if err := j.b.Publish(ctx, msg); err != nil {
		j.log.Error(ctx, "failed to publish transactions created event", err, "import_id", imp.ID, "account_id", imp.AccountID.String())
	}
	return duplicates, failedCount, nil
}

func (j *ImportJob) markImportFailed(ctx context.Context, imp *importer.Import, reason error) {
	if imp == nil {
		return
	}
	imp.MarkFailed(reason.Error())
	if err := j.importStore.UpdateState(ctx, imp); err != nil {
		j.log.Error(ctx, "failed to mark import as failed", err, "import_id", imp.ID)
	}
}
