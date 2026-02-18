package jobs

import (
	"context"
	"log/slog"
	"sync"
	"testing"
	"time"

	"github.com/lennardclaproth/my-finances-tracker/internal/logging"
	"github.com/lennardclaproth/my-finances-tracker/internal/storage"
)

type fakeBulkTagStore struct {
	mu      sync.Mutex
	calls   int
	updated int
}

func (s *fakeBulkTagStore) UpdateTagByQuery(_ context.Context, _ storage.CashflowTransactionQuery, _ string) (int, error) {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.calls++
	return s.updated, nil
}

func (s *fakeBulkTagStore) callCount() int {
	s.mu.Lock()
	defer s.mu.Unlock()
	return s.calls
}

func TestNewBulkTagJob_BoundsWorkerPool(t *testing.T) {
	t.Parallel()

	job := NewBulkTagJob(&fakeBulkTagStore{}, logging.NewSlogLogger(slog.LevelError), 10, 32)
	if got := job.WorkerCount(); got != maxBulkTagWorkers {
		t.Fatalf("expected worker count %d, got %d", maxBulkTagWorkers, got)
	}
	if got := job.QueueSize(); got != 32 {
		t.Fatalf("expected queue size 32, got %d", got)
	}
}

func TestBulkTagJob_EnqueueFilter_QueueFull(t *testing.T) {
	t.Parallel()

	job := NewBulkTagJob(&fakeBulkTagStore{}, logging.NewSlogLogger(slog.LevelError), 1, 1)
	ctx := context.Background()
	if err := job.EnqueueFilter(ctx, storage.CashflowTransactionQuery{}, "food"); err != nil {
		t.Fatalf("unexpected first enqueue error: %v", err)
	}
	if err := job.EnqueueFilter(ctx, storage.CashflowTransactionQuery{}, "travel"); err == nil {
		t.Fatalf("expected queue full error")
	}
}

func TestBulkTagJob_Start_ProcessesQueuedWork(t *testing.T) {
	t.Parallel()

	store := &fakeBulkTagStore{updated: 3}
	job := NewBulkTagJob(store, logging.NewSlogLogger(slog.LevelError), 1, 4)
	ctx, cancel := context.WithTimeout(context.Background(), time.Second)
	defer cancel()

	if err := job.EnqueueFilter(context.Background(), storage.CashflowTransactionQuery{}, "food"); err != nil {
		t.Fatalf("enqueue failed: %v", err)
	}

	done := make(chan error, 1)
	go func() {
		done <- job.Start(ctx)
	}()

	deadline := time.Now().Add(500 * time.Millisecond)
	for time.Now().Before(deadline) {
		if store.callCount() > 0 {
			break
		}
		time.Sleep(10 * time.Millisecond)
	}
	if store.callCount() == 0 {
		t.Fatalf("expected queued work to be processed")
	}

	cancel()
	select {
	case <-time.After(500 * time.Millisecond):
		t.Fatalf("bulk tag job did not stop after cancel")
	case <-done:
	}
}
