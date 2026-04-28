package importer

import (
	"context"
	"errors"
	"fmt"
	"io"
	"math"

	"github.com/google/uuid"
	"github.com/lennardclaproth/my-finances-tracker/api"
	"github.com/lennardclaproth/my-finances-tracker/internal/bus"
	"github.com/lennardclaproth/my-finances-tracker/internal/cashflow"
	"github.com/lennardclaproth/my-finances-tracker/internal/logging"
	"github.com/lennardclaproth/my-finances-tracker/internal/observability"
	"github.com/lennardclaproth/my-finances-tracker/internal/portfolio"
	"github.com/lennardclaproth/my-finances-tracker/internal/vendor"
	"go.elastic.co/apm/v2"
)

// CashflowParserFactory resolves a CSV parser for cashflow imports by vendor.
type CashflowParserFactory func(id vendor.VendorID) (cashflow.CsvParser, error)

// PortfolioParserFactory resolves a CSV parser for portfolio imports by vendor.
type PortfolioParserFactory func(id vendor.VendorID) (portfolio.CsvParser, error)

type asyncImportVendorStore interface {
	FetchById(ctx context.Context, id uuid.UUID) (*vendor.Vendor, error)
}

type asyncImportStore interface {
	FetchByID(ctx context.Context, id uuid.UUID) (*Import, error)
	TryMarkInProgress(ctx context.Context, id uuid.UUID) (bool, error)
	UpdateState(ctx context.Context, imp *Import) error
}

type asyncImportCashflowStore interface {
	Create(ctx context.Context, tx *cashflow.Transaction) error
}

type asyncImportPortfolioStore interface {
	Create(ctx context.Context, tx *portfolio.Transaction) error
}

type asyncImportFileReader interface {
	ReadCsv(path string) (io.ReadCloser, error)
}

// AsyncProcessor executes import domain/application workflow for one import ID.
type AsyncProcessor struct {
	vendorStore     asyncImportVendorStore
	importStore     asyncImportStore
	cashflowStore   asyncImportCashflowStore
	portfolioStore  asyncImportPortfolioStore
	fileReader      asyncImportFileReader
	cashflowParser  CashflowParserFactory
	portfolioParser PortfolioParserFactory
	log             logging.Logger
	b               bus.Bus
}

// NewAsyncProcessor constructs the import processing orchestrator.
func NewAsyncProcessor(
	vendorStore asyncImportVendorStore,
	importStore asyncImportStore,
	cashflowStore asyncImportCashflowStore,
	portfolioStore asyncImportPortfolioStore,
	fileReader asyncImportFileReader,
	cashflowParser CashflowParserFactory,
	portfolioParser PortfolioParserFactory,
	log logging.Logger,
	b bus.Bus,
) *AsyncProcessor {
	return &AsyncProcessor{
		vendorStore:     vendorStore,
		importStore:     importStore,
		cashflowStore:   cashflowStore,
		portfolioStore:  portfolioStore,
		fileReader:      fileReader,
		cashflowParser:  cashflowParser,
		portfolioParser: portfolioParser,
		log:             log,
		b:               b,
	}
}

// ProcessByID processes one persisted import and updates terminal state/events.
func (p *AsyncProcessor) ProcessByID(ctx context.Context, importID uuid.UUID, headers map[string]string) (err error) {
	ctx = observability.ContextWithPropagationHeaders(ctx, headers)

	apmTx, txCtx, txErr := observability.StartTransactionFromHeaders(
		ctx,
		observability.JobOperation("import"),
		"job",
		headers,
	)
	if txErr != nil {
		p.log.Error(ctx, "failed to parse incoming trace headers for import job", txErr, "import_id", importID)
	}
	ctx = txCtx
	observability.SetSafeTransactionLabels(apmTx, map[string]any{
		"operation": observability.JobOperation("import"),
		"component": "job",
		"import_id": importID.String(),
		"stage":     "process",
	})
	apmTx.Result = "success"
	apmTx.Outcome = "success"
	defer func() {
		if err != nil {
			apmTx.Result = "error"
			apmTx.Outcome = "failure"
			apm.CaptureError(ctx, err).Send()
		}
		apmTx.End()
	}()

	imp, err := p.importStore.FetchByID(ctx, importID)
	if err != nil {
		if errors.Is(err, ErrNoImportsPending) {
			return nil
		}
		return err
	}
	claimed, err := p.importStore.TryMarkInProgress(ctx, imp.ID)
	if err != nil {
		return err
	}
	if !claimed {
		return nil
	}

	v, err := p.vendorStore.FetchById(ctx, imp.VendorID)
	if err != nil {
		p.markImportFailed(ctx, imp, fmt.Errorf("fetch vendor: %w", err))
		return err
	}

	totalRows, duplicates, importedCount, failedCount, err := p.processCashflow(ctx, imp, v)
	if err != nil {
		p.markImportFailed(ctx, imp, fmt.Errorf("process cashflow rows: %w", err))
		return err
	}

	if v.Type == vendor.VendorTypeBrokerage {
		pDuplicates, pFailed, pErr := p.processPortfolio(ctx, imp, v)
		duplicates += pDuplicates
		failedCount += pFailed
		if pErr != nil {
			// Keep import as completed with failures so users can recover manually.
			failedCount += totalRows
			p.log.Error(ctx, "portfolio processing failed; import will be completed with failures", pErr, "import_id", imp.ID)
		}
	}

	imp.MarkCompleted(duplicates, totalRows, importedCount, failedCount)
	if err := p.importStore.UpdateState(ctx, imp); err != nil {
		p.log.Error(ctx, "failed to mark import as completed", err, "import_id", imp.ID)
		return err
	}
	p.publishImportCompleted(ctx, imp)
	return nil
}

func (p *AsyncProcessor) processCashflow(ctx context.Context, imp *Import, v *vendor.Vendor) (totalRows, duplicates, importedCount, failedCount int, err error) {
	parser, err := p.cashflowParser(v.Name)
	if err != nil {
		return 0, 0, 0, 0, err
	}
	rc, err := p.fileReader.ReadCsv(imp.Path)
	if err != nil {
		return 0, 0, 0, 0, err
	}
	parseSpan, parseCtx := apm.StartSpan(ctx, "parse", "job")
	seq, err := parser.ParseAll(rc)
	parseSpan.End()
	if err != nil {
		return 0, 0, 0, 0, err
	}

	persistSpan, persistCtx := apm.StartSpan(parseCtx, "persist", "job")
	defer persistSpan.End()

	for rowNumber, txd := range seq {
		totalRows++
		amount := math.Abs(txd.Amount)
		tx, err := cashflow.NewTransaction(
			txd.Description,
			txd.Note,
			string(v.Name),
			"",
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
			p.log.Error(persistCtx, "failed creating cashflow transaction", err, "import_id", imp.ID, "row_number", rowNumber)
			continue
		}
		if err := p.cashflowStore.Create(persistCtx, tx); err != nil {
			if errors.Is(err, cashflow.ErrDuplicateTransaction) {
				duplicates++
				continue
			}
			failedCount++
			p.log.Error(persistCtx, "failed persisting cashflow transaction", err, "import_id", imp.ID, "row_number", rowNumber)
			continue
		}
		importedCount++
	}

	return totalRows, duplicates, importedCount, failedCount, nil
}

func (p *AsyncProcessor) processPortfolio(ctx context.Context, imp *Import, v *vendor.Vendor) (duplicates, failedCount int, err error) {
	parser, err := p.portfolioParser(v.Name)
	if err != nil {
		return 0, 0, err
	}
	rc, err := p.fileReader.ReadCsv(imp.Path)
	if err != nil {
		return 0, 0, err
	}
	parseSpan, parseCtx := apm.StartSpan(ctx, "parse", "job")
	seq, err := parser.ParseAll(rc)
	parseSpan.End()
	if err != nil {
		return 0, 0, err
	}

	persistSpan, persistCtx := apm.StartSpan(parseCtx, "persist", "job")
	defer persistSpan.End()

	for rowNumber, txd := range seq {
		ptx, err := portfolio.NewTransaction(txd, rowNumber, imp.ID, imp.AccountID, nil)
		if err != nil {
			failedCount++
			p.log.Error(persistCtx, "failed creating portfolio transaction", err, "import_id", imp.ID, "row_number", rowNumber)
			continue
		}
		if err := p.portfolioStore.Create(persistCtx, ptx); err != nil {
			if errors.Is(err, portfolio.ErrDuplicateTransaction) {
				duplicates++
				continue
			}
			failedCount++
			p.log.Error(persistCtx, "failed persisting portfolio transaction", err, "import_id", imp.ID, "row_number", rowNumber)
			continue
		}
	}

	if imp.AccountID == nil {
		p.log.Info(ctx, "skip transactions created event: import has no account id", "import_id", imp.ID)
		return duplicates, failedCount, nil
	}
	if p.b == nil {
		p.log.Info(ctx, "skip transactions created event: bus not configured", "import_id", imp.ID, "account_id", imp.AccountID.String())
		return duplicates, failedCount, nil
	}
	publishSpan, publishCtx := apm.StartSpan(ctx, "publish", "job")
	defer publishSpan.End()

	msg, err := bus.NewJSONEnvelopeFromContext(publishCtx, api.TransactionsCreated{AccID: *imp.AccountID})
	if err != nil {
		p.log.Error(ctx, "failed to encode transactions created event", err, "import_id", imp.ID, "account_id", imp.AccountID.String())
		return duplicates, failedCount, nil
	}
	if err := p.b.Publish(publishCtx, msg); err != nil {
		p.log.Error(ctx, "failed to publish transactions created event", err, "import_id", imp.ID, "account_id", imp.AccountID.String())
	}
	return duplicates, failedCount, nil
}

func (p *AsyncProcessor) markImportFailed(ctx context.Context, imp *Import, reason error) {
	if imp == nil {
		return
	}
	imp.MarkFailed(reason.Error())
	if err := p.importStore.UpdateState(ctx, imp); err != nil {
		p.log.Error(ctx, "failed to mark import as failed", err, "import_id", imp.ID)
	}
}

func (p *AsyncProcessor) publishImportCompleted(ctx context.Context, imp *Import) {
	if imp == nil || imp.AccountID == nil || p.b == nil {
		return
	}
	msg, err := bus.NewJSONEnvelopeFromContext(ctx, api.ImportCompleted{
		AccID:    *imp.AccountID,
		ImportID: imp.ID,
	})
	if err != nil {
		p.log.Error(ctx, "failed to encode import completed event", err, "import_id", imp.ID, "account_id", imp.AccountID.String())
		return
	}
	if err := p.b.Publish(ctx, msg); err != nil {
		p.log.Error(ctx, "failed to publish import completed event", err, "import_id", imp.ID, "account_id", imp.AccountID.String())
	}
}
