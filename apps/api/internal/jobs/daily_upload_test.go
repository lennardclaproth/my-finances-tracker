package jobs

import (
	"context"
	"io"
	"log/slog"
	"strings"
	"sync"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/lennardclaproth/my-finances-tracker/internal/logging"
	"github.com/lennardclaproth/my-finances-tracker/internal/marketdata"
	marketdataParsers "github.com/lennardclaproth/my-finances-tracker/internal/marketdata/parsers"
)

type fakeDailyUploadStore struct {
	mu      sync.Mutex
	uploads map[uuid.UUID]*marketdata.DailyUpload
}

func (s *fakeDailyUploadStore) FetchByID(_ context.Context, id uuid.UUID) (*marketdata.DailyUpload, error) {
	s.mu.Lock()
	defer s.mu.Unlock()
	up, ok := s.uploads[id]
	if !ok {
		return nil, marketdata.ErrDailyUploadNotFound
	}
	cpy := *up
	cpy.RowErrors = append([]marketdata.DailyUploadRowError{}, up.RowErrors...)
	return &cpy, nil
}

func (s *fakeDailyUploadStore) ListPending(_ context.Context, limit int) ([]*marketdata.DailyUpload, error) {
	s.mu.Lock()
	defer s.mu.Unlock()
	out := make([]*marketdata.DailyUpload, 0)
	for _, up := range s.uploads {
		if up.Status == marketdata.DailyUploadStatusPending {
			cpy := *up
			out = append(out, &cpy)
		}
	}
	if limit > 0 && len(out) > limit {
		out = out[:limit]
	}
	return out, nil
}

func (s *fakeDailyUploadStore) TryMarkProcessing(_ context.Context, id uuid.UUID) (bool, error) {
	s.mu.Lock()
	defer s.mu.Unlock()
	up, ok := s.uploads[id]
	if !ok {
		return false, nil
	}
	if up.Status != marketdata.DailyUploadStatusPending {
		return false, nil
	}
	up.MarkProcessing()
	return true, nil
}

func (s *fakeDailyUploadStore) UpdateState(_ context.Context, upload *marketdata.DailyUpload) error {
	s.mu.Lock()
	defer s.mu.Unlock()
	cpy := *upload
	cpy.RowErrors = append([]marketdata.DailyUploadRowError{}, upload.RowErrors...)
	s.uploads[upload.ID] = &cpy
	return nil
}

func (s *fakeDailyUploadStore) get(id uuid.UUID) *marketdata.DailyUpload {
	s.mu.Lock()
	defer s.mu.Unlock()
	return s.uploads[id]
}

type fakeDailyUploadListingStore struct {
	byID map[uuid.UUID]*marketdata.Listing
}

func (s *fakeDailyUploadListingStore) FetchByID(_ context.Context, id uuid.UUID) (*marketdata.Listing, error) {
	v, ok := s.byID[id]
	if !ok {
		return nil, nil
	}
	return v, nil
}

type fakeDailyPersistStore struct {
	mu       sync.Mutex
	inserted map[string]bool
}

func (s *fakeDailyPersistStore) CreateWithInsertStatus(_ context.Context, daily *marketdata.Daily) (bool, error) {
	s.mu.Lock()
	defer s.mu.Unlock()
	if s.inserted == nil {
		s.inserted = map[string]bool{}
	}
	key := daily.ListingID.String() + "|" + daily.Date.Format("2006-01-02")
	if _, exists := s.inserted[key]; exists {
		return false, nil
	}
	s.inserted[key] = true
	return true, nil
}

type fakeDailyUploadReader struct{}

func (r *fakeDailyUploadReader) ReadCsv(_ string) (io.ReadCloser, error) {
	return io.NopCloser(strings.NewReader("dummy")), nil
}

type stubDailyParser struct {
	result marketdataParsers.ParseResult
}

func (p *stubDailyParser) ParseAll(_ io.ReadCloser) (marketdataParsers.ParseResult, error) {
	return p.result, nil
}

func TestDailyUploadJob_Enqueue_Deduplicates(t *testing.T) {
	t.Parallel()

	job := NewDailyUploadJob(
		&fakeDailyUploadStore{uploads: map[uuid.UUID]*marketdata.DailyUpload{}},
		&fakeDailyUploadListingStore{byID: map[uuid.UUID]*marketdata.Listing{}},
		&fakeDailyPersistStore{},
		&fakeDailyUploadReader{},
		logging.NewSlogLogger(slog.LevelError),
		time.Second,
		2,
	)

	id := uuid.New()
	if err := job.Enqueue(context.Background(), id); err != nil {
		t.Fatalf("unexpected enqueue error: %v", err)
	}
	if err := job.Enqueue(context.Background(), id); err != nil {
		t.Fatalf("unexpected duplicate enqueue error: %v", err)
	}
	if got := len(job.queue); got != 1 {
		t.Fatalf("expected queue size 1, got %d", got)
	}
}

func TestDailyUploadJob_Start_ReconcilesAndProcessesPending(t *testing.T) {
	t.Parallel()

	listingID := uuid.New()
	upload, err := marketdata.NewDailyUpload(listingID, marketdata.SourceBrandNewDay, "test.csv", "test.csv")
	if err != nil {
		t.Fatalf("failed to create upload fixture: %v", err)
	}
	store := &fakeDailyUploadStore{
		uploads: map[uuid.UUID]*marketdata.DailyUpload{
			upload.ID: upload,
		},
	}
	listing := &marketdata.Listing{
		ID:     listingID,
		Symbol: "BND.AS",
		Source: marketdata.SourceBrandNewDay,
		Active: true,
	}

	job := NewDailyUploadJob(
		store,
		&fakeDailyUploadListingStore{byID: map[uuid.UUID]*marketdata.Listing{listingID: listing}},
		&fakeDailyPersistStore{},
		&fakeDailyUploadReader{},
		logging.NewSlogLogger(slog.LevelError),
		20*time.Millisecond,
		1,
	)
	job.parserFactory = func(source marketdata.Source) (marketdataParsers.DailyParser, error) {
		return &stubDailyParser{
			result: marketdataParsers.ParseResult{
				TotalRows: 2,
				Rows: []marketdataParsers.DailyRow{
					{
						RowNumber: 1,
						Date:      time.Date(2026, 2, 28, 0, 0, 0, 0, time.UTC),
						Open:      38.286191,
						High:      38.286191,
						Low:       38.286191,
						Close:     38.286191,
					},
					{
						RowNumber: 2,
						Date:      time.Date(2026, 2, 28, 0, 0, 0, 0, time.UTC), // duplicate DB insert
						Open:      38.286191,
						High:      38.286191,
						Low:       38.286191,
						Close:     38.286191,
					},
				},
			},
		}, nil
	}

	ctx, cancel := context.WithTimeout(context.Background(), 2*time.Second)
	defer cancel()

	done := make(chan error, 1)
	go func() {
		done <- job.Start(ctx)
	}()

	deadline := time.Now().Add(1500 * time.Millisecond)
	for time.Now().Before(deadline) {
		got := store.get(upload.ID)
		if got != nil && (got.Status == marketdata.DailyUploadStatusSucceeded || got.Status == marketdata.DailyUploadStatusPartial) {
			if got.InsertedRows != 1 {
				t.Fatalf("expected inserted_rows=1, got %d", got.InsertedRows)
			}
			if got.DuplicateRows != 1 {
				t.Fatalf("expected duplicate_rows=1, got %d", got.DuplicateRows)
			}
			if got.ErrorRows != 0 {
				t.Fatalf("expected error_rows=0, got %d", got.ErrorRows)
			}
			cancel()
			select {
			case <-time.After(time.Second):
				t.Fatalf("job did not stop after cancel")
			case err := <-done:
				if err == nil {
					t.Fatalf("expected context cancellation error, got nil")
				}
			}
			return
		}
		time.Sleep(10 * time.Millisecond)
	}
	t.Fatalf("timed out waiting for daily upload to complete")
}
