package jobs

import (
	"context"
	"errors"
	"fmt"
	"sync"

	"github.com/lennardclaproth/my-finances-tracker/internal/logging"
	"github.com/lennardclaproth/my-finances-tracker/internal/observability"
	"github.com/lennardclaproth/my-finances-tracker/internal/storage"
	"go.elastic.co/apm/v2"
)

const (
	defaultBulkTagQueueSize = 256
	maxBulkTagWorkers       = 4
)

var (
	ErrBulkTagQueueFull = errors.New("bulk tag queue is full")
)

type BulkTagEnqueuer interface {
	EnqueueFilter(ctx context.Context, query storage.CashflowTransactionQuery, tag string) error
}

type bulkTagStore interface {
	UpdateTagByQuery(ctx context.Context, query storage.CashflowTransactionQuery, tag string) (int, error)
}

type bulkTagTask struct {
	query   storage.CashflowTransactionQuery
	tag     string
	headers map[string]string
}

type BulkTagJob struct {
	store   bulkTagStore
	log     logging.Logger
	queue   chan bulkTagTask
	workers int
}

func NewBulkTagJob(store bulkTagStore, log logging.Logger, workers, queueSize int) *BulkTagJob {
	if workers <= 0 {
		workers = 1
	}
	if workers > maxBulkTagWorkers {
		workers = maxBulkTagWorkers
	}
	if queueSize <= 0 {
		queueSize = defaultBulkTagQueueSize
	}
	return &BulkTagJob{
		store:   store,
		log:     log,
		queue:   make(chan bulkTagTask, queueSize),
		workers: workers,
	}
}

func (j *BulkTagJob) Name() string {
	return "BulkTagJob"
}

func (j *BulkTagJob) EnqueueFilter(ctx context.Context, query storage.CashflowTransactionQuery, tag string) error {
	task := bulkTagTask{
		query:   query,
		tag:     tag,
		headers: observability.PropagationHeadersFromContext(ctx),
	}
	select {
	case <-ctx.Done():
		return ctx.Err()
	case j.queue <- task:
		return nil
	default:
		return ErrBulkTagQueueFull
	}
}

func (j *BulkTagJob) Start(ctx context.Context) error {
	var wg sync.WaitGroup
	for i := 0; i < j.workers; i++ {
		wg.Add(1)
		go func(workerID int) {
			defer wg.Done()
			j.runWorker(ctx, workerID)
		}(i + 1)
	}
	<-ctx.Done()
	wg.Wait()
	return ctx.Err()
}

func (j *BulkTagJob) runWorker(ctx context.Context, workerID int) {
	for {
		select {
		case <-ctx.Done():
			return
		case task := <-j.queue:
			workerCtx := observability.ContextWithPropagationHeaders(ctx, task.headers)
			apmTx, txCtx, txErr := observability.StartTransactionFromHeaders(
				workerCtx,
				observability.JobOperation("bulk_tag"),
				"job",
				task.headers,
			)
			if txErr != nil {
				j.log.Error(workerCtx, "failed to parse incoming trace headers for bulk tag job", txErr, "worker_id", workerID)
			}
			apmTx.Result = "success"
			apmTx.Outcome = "success"
			observability.SetSafeTransactionLabels(apmTx, map[string]any{
				"operation": observability.JobOperation("bulk_tag"),
				"component": "job",
				"worker_id": workerID,
				"stage":     "persist",
			})

			persistSpan, persistCtx := apm.StartSpan(txCtx, "persist", "job")
			updated, err := j.store.UpdateTagByQuery(persistCtx, task.query, task.tag)
			persistSpan.End()
			if err != nil {
				apmTx.Result = "error"
				apmTx.Outcome = "failure"
				apm.CaptureError(txCtx, err).Send()
				j.log.Error(txCtx, "bulk tag job failed", err, "worker_id", workerID)
				apmTx.End()
				continue
			}
			j.log.Info(txCtx, "bulk tag job completed", "worker_id", workerID, "updated_count", updated)
			apmTx.End()
		}
	}
}

func (j *BulkTagJob) WorkerCount() int {
	return j.workers
}

func (j *BulkTagJob) QueueSize() int {
	return cap(j.queue)
}

func (j *BulkTagJob) QueueLength() int {
	return len(j.queue)
}

func (j *BulkTagJob) String() string {
	return fmt.Sprintf("%s(workers=%d,queue=%d)", j.Name(), j.workers, cap(j.queue))
}
