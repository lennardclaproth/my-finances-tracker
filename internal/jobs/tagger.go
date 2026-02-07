package jobs

import (
	"context"
	"time"

	"github.com/lennardclaproth/my-finances-tracker/internal/agent"
	"github.com/lennardclaproth/my-finances-tracker/internal/logging"
	"github.com/lennardclaproth/my-finances-tracker/internal/storage"
	"github.com/lennardclaproth/my-finances-tracker/internal/transaction"
	"go.elastic.co/apm/v2"
)

// TaggerJob is responsible for automatically tagging transactions based on predefined rules.
// when there are no untagged transactions, it should sleep with exponential backoff until new transactions are imported.
type TaggerJob struct {
	ar  *agent.Runner
	ts  *storage.SQLXTransactionStore
	df  time.Duration
	log logging.Logger
}

func NewTaggerJob(ar *agent.Runner, ts *storage.SQLXTransactionStore, df time.Duration, log logging.Logger) *TaggerJob {
	return &TaggerJob{ar: ar, ts: ts, df: df, log: log}
}

func (j *TaggerJob) Name() string {
	return "TaggerJob"
}

func (j *TaggerJob) Start(ctx context.Context) error {
	interval := j.df
	ticker := time.NewTicker(interval)
	defer ticker.Stop()

	for {
		select {
		case <-ctx.Done():
			return ctx.Err()
		case <-ticker.C:
			untagged, err := j.ts.FetchUntagged(ctx, 1, 1)
			if err != nil {
				j.log.Error(ctx, "failed to fetch untagged transactions", err)
			}
			if len(untagged) == 0 {
				j.log.Info(ctx, "no untagged transactions found, increasing interval with exponential backoff")
				// No untagged transactions, increase the interval with exponential backoff
				interval *= 2
				if interval > time.Minute {
					interval = time.Minute
				}
				ticker.Reset(interval)
				continue
			}
			interval = j.df // reset interval to default when we find untagged transactions
			ticker.Reset(interval)
			tx := untagged[0]
			if err := j.process(ctx, tx); err != nil {
				// If tagging fails, log the error and tag the transaction as "unk" to avoid blocking the queue.
				j.log.Error(ctx, "failed to process tagging for transaction %d: %v", err, tx.ID)
				j.ts.Tag(ctx, tx.ID, "unk")
			}
		}
	}
}

func (j *TaggerJob) process(ctx context.Context, tx *transaction.Transaction) error {
	apmTx := apm.DefaultTracer().StartTransaction("TaggerJob.process", "job")
	defer apmTx.End()

	ctx = apm.ContextWithTransaction(ctx, apmTx)

	span, ctx := apm.StartSpan(ctx, "RunTagAgent", "app")
	defer span.End()

	err := j.ar.RunTagAgent(ctx, tx)
	if err != nil {
		apmTx.Result = "error"
		apm.CaptureError(ctx, err).Send()
	}
	return err
}
