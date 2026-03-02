package jobs

import (
	"context"
	"io"
	"iter"
	"log/slog"
	"sort"
	"strings"
	"sync"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/lennardclaproth/my-finances-tracker/internal/cashflow"
	"github.com/lennardclaproth/my-finances-tracker/internal/importer"
	"github.com/lennardclaproth/my-finances-tracker/internal/logging"
	"github.com/lennardclaproth/my-finances-tracker/internal/portfolio"
	"github.com/lennardclaproth/my-finances-tracker/internal/vendor"
)

type fakeVendorStore struct {
	byID map[uuid.UUID]*vendor.Vendor
}

func (s *fakeVendorStore) FetchById(_ context.Context, id uuid.UUID) (*vendor.Vendor, error) {
	v, ok := s.byID[id]
	if !ok {
		return nil, vendor.ErrVendorNotFound
	}
	return v, nil
}

type fakeImportStore struct {
	mu      sync.Mutex
	imports map[uuid.UUID]*importer.Import
}

func (s *fakeImportStore) FetchByID(_ context.Context, id uuid.UUID) (*importer.Import, error) {
	s.mu.Lock()
	defer s.mu.Unlock()
	imp, ok := s.imports[id]
	if !ok {
		return nil, importer.ErrNoImportsPending
	}
	cpy := *imp
	return &cpy, nil
}

func (s *fakeImportStore) ListPending(_ context.Context, limit int) ([]*importer.Import, error) {
	s.mu.Lock()
	defer s.mu.Unlock()
	var out []*importer.Import
	for _, imp := range s.imports {
		if imp.Status == importer.ImportStatusPending {
			cpy := *imp
			out = append(out, &cpy)
		}
	}
	sort.Slice(out, func(i, j int) bool {
		return out[i].CreatedAt.Before(out[j].CreatedAt)
	})
	if limit > 0 && len(out) > limit {
		out = out[:limit]
	}
	return out, nil
}

func (s *fakeImportStore) TryMarkInProgress(_ context.Context, id uuid.UUID) (bool, error) {
	s.mu.Lock()
	defer s.mu.Unlock()
	imp, ok := s.imports[id]
	if !ok {
		return false, nil
	}
	if imp.Status != importer.ImportStatusPending {
		return false, nil
	}
	imp.Status = importer.ImportStatusInProgress
	imp.UpdatedAt = time.Now().UTC()
	return true, nil
}

func (s *fakeImportStore) UpdateState(_ context.Context, imp *importer.Import) error {
	s.mu.Lock()
	defer s.mu.Unlock()
	if _, ok := s.imports[imp.ID]; !ok {
		return nil
	}
	cpy := *imp
	s.imports[imp.ID] = &cpy
	return nil
}

func (s *fakeImportStore) status(id uuid.UUID) importer.ImportStatus {
	s.mu.Lock()
	defer s.mu.Unlock()
	return s.imports[id].Status
}

type fakeCashflowStore struct {
	mu      sync.Mutex
	created int
}

func (s *fakeCashflowStore) Create(_ context.Context, _ *cashflow.Transaction) error {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.created++
	return nil
}

func (s *fakeCashflowStore) count() int {
	s.mu.Lock()
	defer s.mu.Unlock()
	return s.created
}

type fakePortfolioStore struct{}

func (s *fakePortfolioStore) Create(_ context.Context, _ *portfolio.Transaction) error {
	return nil
}

type fakeReader struct{}

func (r *fakeReader) ReadCsv(_ string) (io.ReadCloser, error) {
	return io.NopCloser(strings.NewReader("dummy")), nil
}

type stubCashflowParser struct{}

func (p *stubCashflowParser) ParseAll(rc io.ReadCloser) (iter.Seq2[int, cashflow.TransactionData], error) {
	_ = rc.Close()
	seq := func(yield func(int, cashflow.TransactionData) bool) {
		yield(1, cashflow.TransactionData{
			Description: "test",
			Note:        "",
			Source:      "ING",
			Direction:   cashflow.CashIn,
			Amount:      10,
			Date:        time.Date(2026, 2, 10, 0, 0, 0, 0, time.UTC),
		})
	}
	return seq, nil
}

func TestImportJob_Enqueue_Deduplicates(t *testing.T) {
	t.Parallel()

	vID := uuid.New()
	job := NewImportJob(
		&fakeVendorStore{byID: map[uuid.UUID]*vendor.Vendor{vID: &vendor.Vendor{ID: vID, Name: vendor.VendorING, Type: vendor.VendorTypeBank}}},
		&fakeImportStore{imports: map[uuid.UUID]*importer.Import{}},
		&fakeCashflowStore{},
		&fakePortfolioStore{},
		&fakeReader{},
		logging.NewSlogLogger(slog.LevelError),
		time.Second,
		2,
		nil,
	)

	id := uuid.New()
	if err := job.Enqueue(context.Background(), id); err != nil {
		t.Fatalf("unexpected enqueue error: %v", err)
	}
	if err := job.Enqueue(context.Background(), id); err != nil {
		t.Fatalf("unexpected enqueue duplicate error: %v", err)
	}
	if got := len(job.queue); got != 1 {
		t.Fatalf("expected queue size 1 after duplicate enqueue, got %d", got)
	}
}

func TestImportJob_Start_ReconcilesAndProcessesPending(t *testing.T) {
	t.Parallel()

	vID := uuid.New()
	v := &vendor.Vendor{ID: vID, Name: vendor.VendorING, Type: vendor.VendorTypeBank}
	firstID := uuid.New()
	secondID := uuid.New()
	now := time.Now().UTC()

	store := &fakeImportStore{
		imports: map[uuid.UUID]*importer.Import{
			firstID: {
				ID:        firstID,
				VendorID:  vID,
				Path:      "first.csv",
				Status:    importer.ImportStatusPending,
				CreatedAt: now,
				UpdatedAt: now,
			},
			secondID: {
				ID:        secondID,
				VendorID:  vID,
				Path:      "second.csv",
				Status:    importer.ImportStatusPending,
				CreatedAt: now.Add(time.Second),
				UpdatedAt: now.Add(time.Second),
			},
		},
	}
	cashflowStore := &fakeCashflowStore{}
	job := NewImportJob(
		&fakeVendorStore{byID: map[uuid.UUID]*vendor.Vendor{vID: v}},
		store,
		cashflowStore,
		&fakePortfolioStore{},
		&fakeReader{},
		logging.NewSlogLogger(slog.LevelError),
		20*time.Millisecond,
		1,
		nil,
	)
	job.cashflowParser = func(_ vendor.VendorID) (cashflow.CsvParser, error) {
		return &stubCashflowParser{}, nil
	}

	ctx, cancel := context.WithTimeout(context.Background(), 2*time.Second)
	defer cancel()

	done := make(chan error, 1)
	go func() {
		done <- job.Start(ctx)
	}()

	waitUntil := func(id uuid.UUID, status importer.ImportStatus) {
		t.Helper()
		deadline := time.Now().Add(1500 * time.Millisecond)
		for time.Now().Before(deadline) {
			if store.status(id) == status {
				return
			}
			time.Sleep(10 * time.Millisecond)
		}
		t.Fatalf("timed out waiting for import %s to reach status %s", id, status)
	}

	waitUntil(firstID, importer.ImportStatusCompleted)
	waitUntil(secondID, importer.ImportStatusCompleted)

	cancel()
	select {
	case <-time.After(time.Second):
		t.Fatalf("job did not stop after cancel")
	case err := <-done:
		if err == nil {
			t.Fatalf("expected context cancellation error, got nil")
		}
	}

	if got := cashflowStore.count(); got != 2 {
		t.Fatalf("expected 2 cashflow rows persisted, got %d", got)
	}
}

func TestImportJob_SyncQueueFromDB_CatchesUpAfterQueueFull(t *testing.T) {
	t.Parallel()

	vID := uuid.New()
	firstID := uuid.New()
	secondID := uuid.New()
	now := time.Now().UTC()
	store := &fakeImportStore{
		imports: map[uuid.UUID]*importer.Import{
			firstID: {
				ID:        firstID,
				VendorID:  vID,
				Path:      "first.csv",
				Status:    importer.ImportStatusPending,
				CreatedAt: now,
				UpdatedAt: now,
			},
			secondID: {
				ID:        secondID,
				VendorID:  vID,
				Path:      "second.csv",
				Status:    importer.ImportStatusPending,
				CreatedAt: now.Add(time.Second),
				UpdatedAt: now.Add(time.Second),
			},
		},
	}

	job := NewImportJob(
		&fakeVendorStore{byID: map[uuid.UUID]*vendor.Vendor{vID: &vendor.Vendor{ID: vID, Name: vendor.VendorING, Type: vendor.VendorTypeBank}}},
		store,
		&fakeCashflowStore{},
		&fakePortfolioStore{},
		&fakeReader{},
		logging.NewSlogLogger(slog.LevelError),
		time.Second,
		1,
		nil,
	)

	if err := job.syncQueueFromDB(context.Background()); err != nil {
		t.Fatalf("initial sync failed: %v", err)
	}
	if got := len(job.queue); got != 1 {
		t.Fatalf("expected queue to be full with 1 item, got %d", got)
	}

	id := <-job.queue
	job.markDequeued(id)
	job.markDone(id)

	s := store.imports[id]
	s.Status = importer.ImportStatusCompleted
	store.imports[id] = s

	if err := job.syncQueueFromDB(context.Background()); err != nil {
		t.Fatalf("second sync failed: %v", err)
	}
	if got := len(job.queue); got != 1 {
		t.Fatalf("expected second pending import to be queued, got queue size %d", got)
	}
	next := <-job.queue
	if next == id {
		t.Fatalf("expected different import ID to be queued after catch-up")
	}
}
