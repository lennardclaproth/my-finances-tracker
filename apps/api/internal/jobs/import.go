package jobs

import (
	"context"
	"errors"
	"fmt"
	"io"
	"sync"
	"time"

	"github.com/google/uuid"
	"github.com/lennardclaproth/my-finances-tracker/internal/bus"
	"github.com/lennardclaproth/my-finances-tracker/internal/cashflow"
	cashflowParsers "github.com/lennardclaproth/my-finances-tracker/internal/cashflow/parsers"
	"github.com/lennardclaproth/my-finances-tracker/internal/importer"
	"github.com/lennardclaproth/my-finances-tracker/internal/logging"
	"github.com/lennardclaproth/my-finances-tracker/internal/observability"
	"github.com/lennardclaproth/my-finances-tracker/internal/portfolio"
	portfolioParsers "github.com/lennardclaproth/my-finances-tracker/internal/portfolio/parsers"
	"github.com/lennardclaproth/my-finances-tracker/internal/vendor"
)

const (
	defaultImportQueueSize  = 256
	defaultSyncBatchSize    = 512
	defaultReconcileTimeout = 5 * time.Second
)

var (
	// ErrImportQueueFull indicates the in-memory import queue is at capacity.
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
	queueHeaders    map[uuid.UUID]map[string]string
	mu              sync.Mutex
	cashflowParser  importer.CashflowParserFactory
	portfolioParser importer.PortfolioParserFactory
	b               bus.Bus
}

// NewImportJob constructs the asynchronous CSV import worker with queue reconciliation.
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
		queueHeaders:    make(map[uuid.UUID]map[string]string),
		cashflowParser:  cashflowParsers.CreateCsvParser,
		portfolioParser: portfolioParsers.CreateCsvParser,
		b:               b,
	}
}

// Name returns the worker name used by the job manager.
func (j *ImportJob) Name() string {
	return "ImportJob"
}

// Enqueue schedules an import ID for processing when it is not already queued or in-flight.
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
	j.queueHeaders[importID] = observability.PropagationHeadersFromContext(ctx)
	j.mu.Unlock()

	select {
	case <-ctx.Done():
		j.mu.Lock()
		delete(j.inQueue, importID)
		delete(j.queueHeaders, importID)
		j.mu.Unlock()
		return ctx.Err()
	case j.queue <- importID:
		return nil
	default:
		j.mu.Lock()
		delete(j.inQueue, importID)
		delete(j.queueHeaders, importID)
		j.mu.Unlock()
		return ErrImportQueueFull
	}
}

// Start runs the import worker loop until the context is canceled.
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
			headers := j.markDequeued(importID)
			if err := j.processByID(ctx, importID, headers); err != nil {
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

func (j *ImportJob) markDequeued(importID uuid.UUID) map[string]string {
	j.mu.Lock()
	defer j.mu.Unlock()
	headers := j.queueHeaders[importID]
	delete(j.inQueue, importID)
	delete(j.queueHeaders, importID)
	j.inFlight[importID] = struct{}{}
	return headers
}

func (j *ImportJob) markDone(importID uuid.UUID) {
	j.mu.Lock()
	defer j.mu.Unlock()
	delete(j.inQueue, importID)
	delete(j.queueHeaders, importID)
	delete(j.inFlight, importID)
}

func (j *ImportJob) processByID(ctx context.Context, importID uuid.UUID, headers map[string]string) (err error) {
	processor := importer.NewAsyncProcessor(
		j.vendorStore,
		j.importStore,
		j.cashflowStore,
		j.portfolioStore,
		j.fileReader,
		j.cashflowParser,
		j.portfolioParser,
		j.log,
		j.b,
	)
	return processor.ProcessByID(ctx, importID, headers)
}
