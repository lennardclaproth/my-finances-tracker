package jobs

import (
	"context"
	"time"

	"github.com/lennardclaproth/my-finances-tracker/internal/agent"
	"github.com/lennardclaproth/my-finances-tracker/internal/cashflow"
	"github.com/lennardclaproth/my-finances-tracker/internal/logging"
	"github.com/lennardclaproth/my-finances-tracker/internal/observability"
	"github.com/lennardclaproth/my-finances-tracker/internal/storage"
	"go.elastic.co/apm/v2"
)

// TaggerJob is responsible for automatically tagging transactions based on predefined rules.
// when there are no untagged transactions, it should sleep with exponential backoff until new transactions are imported.
type TaggerJob struct {
	ar  *agent.Runner
	ts  *storage.SQLXBankTransactionStore
	df  time.Duration
	log logging.Logger
}

const (
	initialTaggerBackoff = time.Second
	maxTaggerBackoff     = time.Minute
)

func NewTaggerJob(ar *agent.Runner, ts *storage.SQLXBankTransactionStore, df time.Duration, log logging.Logger) *TaggerJob {
	return &TaggerJob{ar: ar, ts: ts, df: df, log: log}
}

func (j *TaggerJob) Name() string {
	return "TaggerJob"
}

func (j *TaggerJob) Start(ctx context.Context) error {
	interval := j.defaultInterval()
	backoff := initialTaggerBackoff
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
				interval, backoff = nextBackoff(backoff)
				ticker.Reset(interval)
				continue
			}
			if len(untagged) == 0 {
				j.log.Info(ctx, "no untagged transactions found, applying shared backoff")
				interval, backoff = nextBackoff(backoff)
				ticker.Reset(interval)
				continue
			}
			tx := untagged[0]
			if err := j.process(ctx, tx); err != nil {
				if agent.IsClientError(err) {
					j.log.Error(ctx, "agent returned client error; applying fallback tag", err, "transaction_id", tx.ID)
					if tagErr := j.ts.Tag(ctx, tx.ID, "unk"); tagErr != nil {
						j.log.Error(ctx, "failed to apply fallback tag", tagErr, "transaction_id", tx.ID)
						interval, backoff = nextBackoff(backoff)
						ticker.Reset(interval)
						continue
					}
					interval = j.defaultInterval()
					backoff = initialTaggerBackoff
					ticker.Reset(interval)
					continue
				}

				j.log.Error(ctx, "agent unavailable or server failed; skipping fallback tag and retrying with backoff", err, "transaction_id", tx.ID)
				interval, backoff = nextBackoff(backoff)
				ticker.Reset(interval)
				continue
			}

			interval = j.defaultInterval()
			backoff = initialTaggerBackoff
			ticker.Reset(interval)
		}
	}
}

func (j *TaggerJob) process(ctx context.Context, tx *cashflow.Transaction) error {
	apmTx := apm.DefaultTracer().StartTransaction(observability.JobOperation("tagger"), "job")
	defer apmTx.End()

	ctx = apm.ContextWithTransaction(ctx, apmTx)
	apmTx.Result = "success"
	apmTx.Outcome = "success"
	observability.SetSafeTransactionLabels(apmTx, map[string]any{
		"operation":      observability.JobOperation("tagger"),
		"component":      "job",
		"transaction_id": tx.ID.String(),
		"stage":          "agent_call",
	})

	span, ctx := apm.StartSpan(ctx, "agent_call", "job")
	defer span.End()

	err := j.ar.RunTagAgent(ctx, tx)
	if err != nil {
		apmTx.Result = "error"
		apmTx.Outcome = "failure"
		apm.CaptureError(ctx, err).Send()
	}
	return err
}

func (j *TaggerJob) defaultInterval() time.Duration {
	if j.df <= 0 {
		return initialTaggerBackoff
	}
	return j.df
}

func nextBackoff(current time.Duration) (interval time.Duration, next time.Duration) {
	if current < initialTaggerBackoff {
		current = initialTaggerBackoff
	}
	interval = current
	next = current * 2
	if next > maxTaggerBackoff {
		next = maxTaggerBackoff
	}
	return interval, next
}
