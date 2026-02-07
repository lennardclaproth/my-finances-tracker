package jobs

import (
	"context"
	"fmt"
	"time"

	"github.com/lennardclaproth/my-finances-tracker/internal/importer"
	"github.com/lennardclaproth/my-finances-tracker/internal/logging"
	"github.com/lennardclaproth/my-finances-tracker/internal/parser"
	"github.com/lennardclaproth/my-finances-tracker/internal/storage"
	"github.com/lennardclaproth/my-finances-tracker/internal/transaction"
	"go.elastic.co/apm/v2"
)

// ImportJob is responsible for processing imported csv files and
// and creating transactions from them.
type ImportJob struct {
	vendorStore      *storage.SQLXVendorStore
	importStore      *storage.SQLXImportStore
	transactionStore *storage.SQLXTransactionStore
	dh               *storage.Disk
	log              logging.Logger
	interval         time.Duration
}

func NewImportJob(
	vendorStore *storage.SQLXVendorStore,
	importStore *storage.SQLXImportStore,
	transactionStore *storage.SQLXTransactionStore,
	dh *storage.Disk,
	log logging.Logger,
	interval time.Duration,
) *ImportJob {
	return &ImportJob{
		vendorStore:      vendorStore,
		importStore:      importStore,
		transactionStore: transactionStore,
		dh:               dh,
		log:              log,
		interval:         interval,
	}
}

func (j *ImportJob) Name() string {
	return "ImportJob"
}

func (j *ImportJob) Start(ctx context.Context) error {
	ticker := time.NewTicker(j.interval)
	defer ticker.Stop()
	for {
		select {
		case <-ctx.Done():
			return ctx.Err()
		case <-ticker.C:
			if err := j.process(ctx); err != nil {
				j.log.Error(ctx, "Error processing import job: %v", err)
			}
		}
	}
}

func (j *ImportJob) process(ctx context.Context) error {
	tx := apm.DefaultTracer().StartTransaction("ImportJob.process", "job")
	defer tx.End()
	ctx = apm.ContextWithTransaction(ctx, tx)
	imp, err := j.importStore.OldestPending()
	if err != nil {
		if err == importer.ErrNoImportsPending {
			return nil // No pending imports, just return
		}
		return err
	}
	v, err := j.vendorStore.FetchById(ctx, imp.VendorID)
	if err != nil {
		j.handleError(ctx, imp, err)
		return err
	}
	p, err := parser.CreateCsvParser(v.Name)
	if err != nil {
		j.handleError(ctx, imp, err)
		return err
	}
	rc, err := j.dh.ReadCsv(imp.Path)
	if err != nil {
		j.handleError(ctx, imp, err)
		return err
	}
	txds, err := p.ParseAll(rc)
	if err != nil {
		j.handleError(ctx, imp, err)
		return err
	}
	// maybe bad?
	defer rc.Close()
	// txs := []*transaction.Transaction{}
	for i, txd := range txds {
		tx, err := transaction.NewTransaction(txd.Description, txd.Note, string(v.Name), transaction.CashFlowDirection(txd.Direction), txd.Amount, txd.Date, i, imp.ID)
		if err != nil {
			j.handleError(ctx, imp, err)
			return err
		}
		// txs = append(txs, tx)
		if err := j.transactionStore.Create(ctx, tx); err != nil {
			j.handleError(ctx, imp, err)
			return err
		}
	}
	imp.MarkCompleted(0, 0, 0, 0)
	if err := j.importStore.UpdateState(ctx, imp); err != nil {
		j.log.Error(ctx, "Error marking import with id %s as completed: %v", err, imp.ID)
	}
	return err
}

func (j *ImportJob) handleError(ctx context.Context, imp *importer.Import, err error) {
	j.log.Error(ctx, "Error processing import with id %s: %v", err, imp.ID)
	imp.MarkFailed(fmt.Errorf("error processing import: %v", err).Error())
	if err := j.importStore.UpdateState(ctx, imp); err != nil {
		j.log.Error(ctx, "Error marking import with id %s as failed: %v", err, imp.ID)
	}
}
