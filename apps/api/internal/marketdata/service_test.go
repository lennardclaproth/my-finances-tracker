package marketdata

import (
	"context"
	"errors"
	"iter"
	"strings"
	"sync"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/lennardclaproth/my-finances-tracker/internal/date"
	"github.com/lennardclaproth/my-finances-tracker/internal/money"
)

type fakeListingStore struct {
	mu sync.Mutex

	createFn                 func(ctx context.Context, listing *Listing) error
	updateFieldsFn           func(ctx context.Context, listing *Listing) error
	listFn                   func(ctx context.Context) ([]*Listing, error)
	fetchBySymbolFn          func(ctx context.Context, symbol string) (*Listing, error)
	fetchByIDFn              func(ctx context.Context, id uuid.UUID) (*Listing, error)
	tryAcquireSyncLockFn     func(ctx context.Context, id uuid.UUID) (bool, error)
	releaseSyncLockFn        func(ctx context.Context, id uuid.UUID) error
	updateShouldAccumulateFn func(ctx context.Context, id uuid.UUID, shouldAccumulate bool) error
	updateAccumulatedRangeFn func(ctx context.Context, id uuid.UUID, accumulatedStart, accumulatedEnd *time.Time) error

	createCalls                 int
	fetchBySymbolCalls          int
	fetchByIDCalls              int
	tryAcquireSyncLockCalls     int
	releaseSyncLockCalls        int
	updateShouldAccumulateCalls int
	lastUpdateShouldAccumulate  bool
}

func (f *fakeListingStore) Create(ctx context.Context, listing *Listing) error {
	f.mu.Lock()
	f.createCalls++
	f.mu.Unlock()
	if f.createFn != nil {
		return f.createFn(ctx, listing)
	}
	return nil
}

func (f *fakeListingStore) List(ctx context.Context) ([]*Listing, error) {
	if f.listFn != nil {
		return f.listFn(ctx)
	}
	return []*Listing{}, nil
}

func (f *fakeListingStore) UpdateFields(ctx context.Context, listing *Listing) error {
	if f.updateFieldsFn != nil {
		return f.updateFieldsFn(ctx, listing)
	}
	return nil
}

func (f *fakeListingStore) FetchBySymbol(ctx context.Context, symbol string) (*Listing, error) {
	f.mu.Lock()
	f.fetchBySymbolCalls++
	f.mu.Unlock()
	if f.fetchBySymbolFn != nil {
		return f.fetchBySymbolFn(ctx, symbol)
	}
	return nil, nil
}

func (f *fakeListingStore) FetchByID(ctx context.Context, id uuid.UUID) (*Listing, error) {
	f.mu.Lock()
	f.fetchByIDCalls++
	f.mu.Unlock()
	if f.fetchByIDFn != nil {
		return f.fetchByIDFn(ctx, id)
	}
	return nil, nil
}

func (f *fakeListingStore) TryAcquireSyncLock(ctx context.Context, id uuid.UUID) (bool, error) {
	f.mu.Lock()
	f.tryAcquireSyncLockCalls++
	f.mu.Unlock()
	if f.tryAcquireSyncLockFn != nil {
		return f.tryAcquireSyncLockFn(ctx, id)
	}
	return true, nil
}

func (f *fakeListingStore) ReleaseSyncLock(ctx context.Context, id uuid.UUID) error {
	f.mu.Lock()
	f.releaseSyncLockCalls++
	f.mu.Unlock()
	if f.releaseSyncLockFn != nil {
		return f.releaseSyncLockFn(ctx, id)
	}
	return nil
}

func (f *fakeListingStore) UpdateShouldAccumulate(ctx context.Context, id uuid.UUID, shouldAccumulate bool) error {
	f.mu.Lock()
	f.updateShouldAccumulateCalls++
	f.lastUpdateShouldAccumulate = shouldAccumulate
	f.mu.Unlock()
	if f.updateShouldAccumulateFn != nil {
		return f.updateShouldAccumulateFn(ctx, id, shouldAccumulate)
	}
	return nil
}

func (f *fakeListingStore) UpdateAccumulatedRange(ctx context.Context, id uuid.UUID, accumulatedStart, accumulatedEnd *time.Time) error {
	if f.updateAccumulatedRangeFn != nil {
		return f.updateAccumulatedRangeFn(ctx, id, accumulatedStart, accumulatedEnd)
	}
	return nil
}

type fakeDailyStore struct {
	mu sync.Mutex

	createFn func(ctx context.Context, daily *Daily) error
	fetchFn  func(ctx context.Context, listingID uuid.UUID, from, to *time.Time, limit, offset int) (*[]Daily, error)

	createCalls int
	fetchCalls  int
}

func (f *fakeDailyStore) Create(ctx context.Context, daily *Daily) error {
	f.mu.Lock()
	f.createCalls++
	f.mu.Unlock()
	if f.createFn != nil {
		return f.createFn(ctx, daily)
	}
	return nil
}

func (f *fakeDailyStore) FetchByListingID(ctx context.Context, listingID uuid.UUID, from, to *time.Time, limit, offset int) (*[]Daily, error) {
	f.mu.Lock()
	f.fetchCalls++
	f.mu.Unlock()
	if f.fetchFn != nil {
		return f.fetchFn(ctx, listingID, from, to, limit, offset)
	}
	out := []Daily{}
	return &out, nil
}

type fakeEODClient struct {
	mu sync.Mutex

	getEODFn func(ctx context.Context, symbols []string, from, to *time.Time) iter.Seq2[Daily, error]

	calls       int
	lastSymbols []string
	lastFrom    *time.Time
	lastTo      *time.Time
}

func (f *fakeEODClient) GetEOD(ctx context.Context, symbols []string, from, to *time.Time) iter.Seq2[Daily, error] {
	f.mu.Lock()
	f.calls++
	f.lastSymbols = append([]string{}, symbols...)
	if from != nil {
		v := *from
		f.lastFrom = &v
	} else {
		f.lastFrom = nil
	}
	if to != nil {
		v := *to
		f.lastTo = &v
	} else {
		f.lastTo = nil
	}
	fn := f.getEODFn
	f.mu.Unlock()

	if fn != nil {
		return fn(ctx, symbols, from, to)
	}
	return func(yield func(Daily, error) bool) {}
}

type fakeLogger struct{}

func (l *fakeLogger) Info(ctx context.Context, msg string, fields ...any) {}

func (l *fakeLogger) Error(ctx context.Context, msg string, err error, fields ...any) {}

func TestLatestBusinessDate(t *testing.T) {
	loc := time.UTC

	mon := time.Date(2026, 2, 9, 15, 0, 0, 0, loc) // Monday
	got := date.LatestBusinessDate(mon, loc)
	want := time.Date(2026, 2, 6, 0, 0, 0, 0, loc) // Friday
	if !got.Equal(want) {
		t.Fatalf("expected %s, got %s", want, got)
	}

	wed := time.Date(2026, 2, 11, 15, 0, 0, 0, loc) // Wednesday
	got = date.LatestBusinessDate(wed, loc)
	want = time.Date(2026, 2, 10, 0, 0, 0, 0, loc) // Tuesday
	if !got.Equal(want) {
		t.Fatalf("expected %s, got %s", want, got)
	}

	sun := time.Date(2026, 2, 8, 15, 0, 0, 0, loc) // Sunday
	got = date.LatestBusinessDate(sun, loc)
	want = time.Date(2026, 2, 6, 0, 0, 0, 0, loc) // Friday
	if !got.Equal(want) {
		t.Fatalf("expected %s, got %s", want, got)
	}
}

func TestDateOnly(t *testing.T) {
	loc := time.FixedZone("UTC+2", 2*3600)
	input := time.Date(2026, 2, 9, 18, 45, 30, 0, time.UTC)
	got := date.DateOnly(input, loc)
	want := time.Date(2026, 2, 9, 0, 0, 0, 0, loc)
	if !got.Equal(want) {
		t.Fatalf("expected %s, got %s", want, got)
	}
}

func TestSyncDailyData_AcquireLockError(t *testing.T) {
	ls := &fakeListingStore{
		tryAcquireSyncLockFn: func(ctx context.Context, id uuid.UUID) (bool, error) {
			return false, errors.New("db down")
		},
	}
	ds := &fakeDailyStore{}
	client := &fakeEODClient{}
	svc := NewService(ls, ds, client, &fakeLogger{})

	err := svc.syncDailyData(context.Background(), uuid.New(), nil, nil)
	if err == nil || !strings.Contains(err.Error(), "failed to acquire sync lock") {
		t.Fatalf("expected sync lock error, got %v", err)
	}
	if ls.releaseSyncLockCalls != 0 {
		t.Fatalf("expected no release call, got %d", ls.releaseSyncLockCalls)
	}
}

func TestSyncDailyData_AlreadyInProgress(t *testing.T) {
	ls := &fakeListingStore{
		tryAcquireSyncLockFn: func(ctx context.Context, id uuid.UUID) (bool, error) {
			return false, nil
		},
	}
	ds := &fakeDailyStore{}
	client := &fakeEODClient{}
	svc := NewService(ls, ds, client, &fakeLogger{})

	err := svc.syncDailyData(context.Background(), uuid.New(), nil, nil)
	if !errors.Is(err, ErrSyncInProgress) {
		t.Fatalf("expected ErrSyncInProgress, got %v", err)
	}
	if ls.releaseSyncLockCalls != 0 {
		t.Fatalf("expected no release call, got %d", ls.releaseSyncLockCalls)
	}
}

func TestSyncDailyData_ReleasesLockWhenFetchByIDFails(t *testing.T) {
	ls := &fakeListingStore{
		tryAcquireSyncLockFn: func(ctx context.Context, id uuid.UUID) (bool, error) {
			return true, nil
		},
		fetchByIDFn: func(ctx context.Context, id uuid.UUID) (*Listing, error) {
			return nil, errors.New("not found")
		},
	}
	ds := &fakeDailyStore{}
	client := &fakeEODClient{}
	svc := NewService(ls, ds, client, &fakeLogger{})

	err := svc.syncDailyData(context.Background(), uuid.New(), nil, nil)
	if err == nil || !strings.Contains(err.Error(), "failed to fetch listing") {
		t.Fatalf("expected fetch listing error, got %v", err)
	}
	if ls.releaseSyncLockCalls != 1 {
		t.Fatalf("expected release call once, got %d", ls.releaseSyncLockCalls)
	}
}

func TestSyncDailyData_DefaultsFromAndTo(t *testing.T) {
	accEnd := time.Date(2026, 1, 2, 0, 0, 0, 0, time.UTC)
	id := uuid.New()
	ls := &fakeListingStore{
		tryAcquireSyncLockFn: func(ctx context.Context, listingID uuid.UUID) (bool, error) {
			return true, nil
		},
		fetchByIDFn: func(ctx context.Context, listingID uuid.UUID) (*Listing, error) {
			return &Listing{
				ID:             id,
				Symbol:         "AAA",
				Active:         true,
				Source:         SourceAlphaVantage,
				AccumulatedEnd: &accEnd,
			}, nil
		},
	}
	ds := &fakeDailyStore{}
	client := &fakeEODClient{}
	svc := NewService(ls, ds, client, &fakeLogger{})

	err := svc.syncDailyData(context.Background(), id, nil, nil)
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}
	if client.calls != 1 {
		t.Fatalf("expected client call once, got %d", client.calls)
	}
	if client.lastFrom == nil || !client.lastFrom.Equal(accEnd) {
		t.Fatalf("expected from=%s, got %v", accEnd, client.lastFrom)
	}
	if client.lastTo == nil {
		t.Fatalf("expected non-nil to")
	}
	if ls.releaseSyncLockCalls != 1 {
		t.Fatalf("expected release call once, got %d", ls.releaseSyncLockCalls)
	}
}

func TestSyncDailyData_UsesExplicitFromTo(t *testing.T) {
	id := uuid.New()
	from := time.Date(2025, 1, 1, 0, 0, 0, 0, time.UTC)
	to := time.Date(2025, 2, 1, 0, 0, 0, 0, time.UTC)
	ls := &fakeListingStore{
		tryAcquireSyncLockFn: func(ctx context.Context, listingID uuid.UUID) (bool, error) {
			return true, nil
		},
		fetchByIDFn: func(ctx context.Context, listingID uuid.UUID) (*Listing, error) {
			return &Listing{
				ID:     id,
				Symbol: "BBB",
				Active: true,
				Source: SourceAlphaVantage,
			}, nil
		},
	}
	ds := &fakeDailyStore{}
	client := &fakeEODClient{}
	svc := NewService(ls, ds, client, &fakeLogger{})

	err := svc.syncDailyData(context.Background(), id, &from, &to)
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}
	if client.lastFrom == nil || !client.lastFrom.Equal(from) {
		t.Fatalf("expected explicit from=%s, got %v", from, client.lastFrom)
	}
	if client.lastTo == nil || !client.lastTo.Equal(to) {
		t.Fatalf("expected explicit to=%s, got %v", to, client.lastTo)
	}
}

func TestSyncDailyData_ContinuesOnRowAndPersistErrors(t *testing.T) {
	id := uuid.New()
	ls := &fakeListingStore{
		tryAcquireSyncLockFn: func(ctx context.Context, listingID uuid.UUID) (bool, error) {
			return true, nil
		},
		fetchByIDFn: func(ctx context.Context, listingID uuid.UUID) (*Listing, error) {
			return &Listing{
				ID:     id,
				Symbol: "CCC",
				Active: true,
				Source: SourceAlphaVantage,
			}, nil
		},
	}
	createCalls := 0
	ds := &fakeDailyStore{
		createFn: func(ctx context.Context, d *Daily) error {
			createCalls++
			if createCalls == 1 {
				return errors.New("insert failed")
			}
			return nil
		},
	}
	d1 := Daily{ID: uuid.New(), Symbol: "CCC", Date: time.Now()}
	d2 := Daily{ID: uuid.New(), Symbol: "CCC", Date: time.Now().Add(24 * time.Hour)}
	client := &fakeEODClient{
		getEODFn: func(ctx context.Context, symbols []string, from, to *time.Time) iter.Seq2[Daily, error] {
			return func(yield func(Daily, error) bool) {
				yield(d1, nil)
				yield(Daily{}, errors.New("bad row"))
				yield(d2, nil)
			}
		},
	}
	svc := NewService(ls, ds, client, &fakeLogger{})

	err := svc.syncDailyData(context.Background(), id, nil, nil)
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}
	if ds.createCalls != 2 {
		t.Fatalf("expected create to be called for two valid rows, got %d", ds.createCalls)
	}
	if ls.releaseSyncLockCalls != 1 {
		t.Fatalf("expected release call once, got %d", ls.releaseSyncLockCalls)
	}
}

func TestSyncDailyData_ReleaseLockErrorDoesNotBubbleUp(t *testing.T) {
	id := uuid.New()
	ls := &fakeListingStore{
		tryAcquireSyncLockFn: func(ctx context.Context, listingID uuid.UUID) (bool, error) {
			return true, nil
		},
		fetchByIDFn: func(ctx context.Context, listingID uuid.UUID) (*Listing, error) {
			return &Listing{
				ID:     id,
				Symbol: "DDD",
				Active: true,
				Source: SourceAlphaVantage,
			}, nil
		},
		releaseSyncLockFn: func(ctx context.Context, listingID uuid.UUID) error {
			return errors.New("unlock failed")
		},
	}
	ds := &fakeDailyStore{}
	client := &fakeEODClient{}
	svc := NewService(ls, ds, client, &fakeLogger{})

	err := svc.syncDailyData(context.Background(), id, nil, nil)
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}
}

func TestGetDailies_FetchListingError(t *testing.T) {
	ls := &fakeListingStore{
		fetchBySymbolFn: func(ctx context.Context, symbol string) (*Listing, error) {
			return nil, errors.New("store down")
		},
	}
	ds := &fakeDailyStore{}
	client := &fakeEODClient{}
	svc := NewService(ls, ds, client, &fakeLogger{})

	_, err := svc.GetDailies(context.Background(), "AAA", nil, nil, 10, 0)
	if err == nil || !strings.Contains(err.Error(), "failed to fetch listing") {
		t.Fatalf("expected fetch listing error, got %v", err)
	}
}

func TestGetDailies_ListingNotFound(t *testing.T) {
	ls := &fakeListingStore{
		fetchBySymbolFn: func(ctx context.Context, symbol string) (*Listing, error) {
			return nil, nil
		},
	}
	ds := &fakeDailyStore{}
	client := &fakeEODClient{}
	svc := NewService(ls, ds, client, &fakeLogger{})

	_, err := svc.GetDailies(context.Background(), "AAA", nil, nil, 10, 0)
	if err == nil || !strings.Contains(err.Error(), "not found") {
		t.Fatalf("expected not found error, got %v", err)
	}
}

func TestGetDailies_ListingInactive(t *testing.T) {
	ls := &fakeListingStore{
		fetchBySymbolFn: func(ctx context.Context, symbol string) (*Listing, error) {
			return &Listing{ID: uuid.New(), Symbol: symbol, Active: false, Source: SourceAlphaVantage}, nil
		},
	}
	ds := &fakeDailyStore{}
	client := &fakeEODClient{}
	svc := NewService(ls, ds, client, &fakeLogger{})

	_, err := svc.GetDailies(context.Background(), "AAA", nil, nil, 10, 0)
	if err == nil || !strings.Contains(err.Error(), "not active") {
		t.Fatalf("expected inactive error, got %v", err)
	}
}

func TestGetDailies_WhenSyncingReturnsStaleMessage(t *testing.T) {
	dailies := []Daily{{ID: uuid.New(), Symbol: "AAA"}}
	ls := &fakeListingStore{
		fetchBySymbolFn: func(ctx context.Context, symbol string) (*Listing, error) {
			return &Listing{ID: uuid.New(), Symbol: symbol, Active: true, Syncing: true, Source: SourceAlphaVantage}, nil
		},
	}
	ds := &fakeDailyStore{
		fetchFn: func(ctx context.Context, listingID uuid.UUID, from, to *time.Time, limit, offset int) (*[]Daily, error) {
			return &dailies, nil
		},
	}
	client := &fakeEODClient{}
	svc := NewService(ls, ds, client, &fakeLogger{})

	resp, err := svc.GetDailies(context.Background(), "AAA", nil, nil, 10, 0)
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}
	if resp.Metadata.Message != "Data may be stale, listing is currently syncing" {
		t.Fatalf("unexpected metadata message: %s", resp.Metadata.Message)
	}
	if len(resp.Data) != 1 {
		t.Fatalf("expected 1 daily row, got %d", len(resp.Data))
	}
	if ls.updateShouldAccumulateCalls != 0 {
		t.Fatalf("expected no update should accumulate call while syncing")
	}
}

func TestGetDailies_WhenSyncingDailyFetchError(t *testing.T) {
	ls := &fakeListingStore{
		fetchBySymbolFn: func(ctx context.Context, symbol string) (*Listing, error) {
			return &Listing{ID: uuid.New(), Symbol: symbol, Active: true, Syncing: true, Source: SourceAlphaVantage}, nil
		},
	}
	ds := &fakeDailyStore{
		fetchFn: func(ctx context.Context, listingID uuid.UUID, from, to *time.Time, limit, offset int) (*[]Daily, error) {
			return nil, errors.New("fetch failed")
		},
	}
	client := &fakeEODClient{}
	svc := NewService(ls, ds, client, &fakeLogger{})

	_, err := svc.GetDailies(context.Background(), "AAA", nil, nil, 10, 0)
	if err == nil || !strings.Contains(err.Error(), "failed to fetch daily data") {
		t.Fatalf("expected daily fetch error, got %v", err)
	}
}

func TestGetDailies_StaleListingUpdateFlagErrorStillReturnsData(t *testing.T) {
	id := uuid.New()
	dailies := []Daily{{ID: uuid.New(), Symbol: "AAA"}}
	ls := &fakeListingStore{
		fetchBySymbolFn: func(ctx context.Context, symbol string) (*Listing, error) {
			return &Listing{ID: id, Symbol: symbol, Active: true, Source: SourceAlphaVantage}, nil
		},
		updateShouldAccumulateFn: func(ctx context.Context, listingID uuid.UUID, shouldAccumulate bool) error {
			return errors.New("cannot update")
		},
	}
	ds := &fakeDailyStore{
		fetchFn: func(ctx context.Context, listingID uuid.UUID, from, to *time.Time, limit, offset int) (*[]Daily, error) {
			return &dailies, nil
		},
	}
	client := &fakeEODClient{}
	svc := NewService(ls, ds, client, &fakeLogger{})

	resp, err := svc.GetDailies(context.Background(), "AAA", nil, nil, 10, 0)
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}
	if len(resp.Data) != 1 {
		t.Fatalf("expected 1 daily row, got %d", len(resp.Data))
	}
	if ls.updateShouldAccumulateCalls != 1 {
		t.Fatalf("expected update should accumulate to be called once, got %d", ls.updateShouldAccumulateCalls)
	}
}

func TestGetDailies_StaleListingStartsAsyncSync(t *testing.T) {
	id := uuid.New()
	started := make(chan struct{}, 1)
	dailies := []Daily{{ID: uuid.New(), Symbol: "AAA"}}
	ls := &fakeListingStore{
		fetchBySymbolFn: func(ctx context.Context, symbol string) (*Listing, error) {
			return &Listing{ID: id, Symbol: symbol, Active: true, Source: SourceAlphaVantage}, nil
		},
		updateShouldAccumulateFn: func(ctx context.Context, listingID uuid.UUID, shouldAccumulate bool) error {
			return nil
		},
		tryAcquireSyncLockFn: func(ctx context.Context, listingID uuid.UUID) (bool, error) {
			select {
			case started <- struct{}{}:
			default:
			}
			return false, nil
		},
	}
	ds := &fakeDailyStore{
		fetchFn: func(ctx context.Context, listingID uuid.UUID, from, to *time.Time, limit, offset int) (*[]Daily, error) {
			return &dailies, nil
		},
	}
	client := &fakeEODClient{}
	svc := NewService(ls, ds, client, &fakeLogger{})

	_, err := svc.GetDailies(context.Background(), "AAA", nil, nil, 10, 0)
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}

	select {
	case <-started:
	case <-time.After(500 * time.Millisecond):
		t.Fatalf("expected async sync to start")
	}
}

func TestCreateListing_DuplicateSameSource(t *testing.T) {
	ls := &fakeListingStore{
		fetchBySymbolFn: func(ctx context.Context, symbol string) (*Listing, error) {
			return &Listing{ID: uuid.New(), Symbol: symbol, Active: true, Source: SourceAlphaVantage}, nil
		},
	}
	ds := &fakeDailyStore{}
	client := &fakeEODClient{}
	svc := NewService(ls, ds, client, &fakeLogger{})

	_, err := svc.CreateListing(context.Background(), "AAA", "Name", SourceAlphaVantage)
	if err == nil || !strings.Contains(err.Error(), "already exists") {
		t.Fatalf("expected duplicate error, got %v", err)
	}
	if ls.createCalls != 0 {
		t.Fatalf("expected no create call, got %d", ls.createCalls)
	}
}

func TestCreateListing_NewListingValidationError(t *testing.T) {
	ls := &fakeListingStore{}
	ds := &fakeDailyStore{}
	client := &fakeEODClient{}
	svc := NewService(ls, ds, client, &fakeLogger{})

	_, err := svc.CreateListing(context.Background(), "", "Name", SourceAlphaVantage)
	if err == nil || !strings.Contains(err.Error(), "failed to create listing") {
		t.Fatalf("expected NewListing validation error, got %v", err)
	}
}

func TestCreateListing_PersistError(t *testing.T) {
	ls := &fakeListingStore{
		createFn: func(ctx context.Context, listing *Listing) error {
			return errors.New("insert fail")
		},
	}
	ds := &fakeDailyStore{}
	client := &fakeEODClient{}
	svc := NewService(ls, ds, client, &fakeLogger{})

	_, err := svc.CreateListing(context.Background(), "AAA", "Name", SourceAlphaVantage)
	if err == nil || !strings.Contains(err.Error(), "failed to persist listing") {
		t.Fatalf("expected persist error, got %v", err)
	}
}

func TestCreateListing_SuccessStartsAsyncSync(t *testing.T) {
	started := make(chan struct{}, 1)
	ls := &fakeListingStore{
		tryAcquireSyncLockFn: func(ctx context.Context, listingID uuid.UUID) (bool, error) {
			select {
			case started <- struct{}{}:
			default:
			}
			return false, nil
		},
		fetchByIDFn: func(ctx context.Context, listingID uuid.UUID) (*Listing, error) {
			return &Listing{ID: listingID, Symbol: "AAA", Active: true, Source: SourceAlphaVantage}, nil
		},
	}
	ds := &fakeDailyStore{}
	client := &fakeEODClient{}
	svc := NewService(ls, ds, client, &fakeLogger{})

	listing, err := svc.CreateListing(context.Background(), "AAA", "Name", SourceAlphaVantage)
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}
	if listing == nil || listing.Symbol != "AAA" {
		t.Fatalf("expected listing to be returned")
	}
	if ls.createCalls != 1 {
		t.Fatalf("expected create call once, got %d", ls.createCalls)
	}

	select {
	case <-started:
	case <-time.After(500 * time.Millisecond):
		t.Fatalf("expected async sync to start")
	}
}

func TestUpdateListingFields_UpdatesOnlyProvidedFields(t *testing.T) {
	id := uuid.New()
	desc := "updated description"
	typ := "ETF"
	existingCurrency := "EUR"
	existing := &Listing{
		ID:          id,
		Symbol:      "AAA",
		Name:        "Original",
		Type:        &typ,
		Description: &typ,
		Currency:    (*money.Currency)(&existingCurrency),
		Source:      SourceAlphaVantage,
		Active:      true,
	}

	var updated *Listing
	ls := &fakeListingStore{
		fetchByIDFn: func(ctx context.Context, gotID uuid.UUID) (*Listing, error) {
			if gotID != id {
				t.Fatalf("expected id %s, got %s", id, gotID)
			}
			copy := *existing
			return &copy, nil
		},
		updateFieldsFn: func(ctx context.Context, listing *Listing) error {
			updated = listing
			return nil
		},
	}
	svc := NewService(ls, &fakeDailyStore{}, &fakeEODClient{}, &fakeLogger{})

	out, err := svc.UpdateListingFields(context.Background(), id, &desc, nil, nil, nil, nil, nil, nil)
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}
	if out.Description == nil || *out.Description != desc {
		t.Fatalf("expected updated description, got %+v", out.Description)
	}
	if out.Type == nil || *out.Type != typ {
		t.Fatalf("expected type unchanged, got %+v", out.Type)
	}
	if out.Currency == nil || *out.Currency != money.Currency(existingCurrency) {
		t.Fatalf("expected currency unchanged, got %+v", out.Currency)
	}
	if updated == nil {
		t.Fatalf("expected updateFields call")
	}
}

func TestUpdateListingFields_RejectsEmptyPatch(t *testing.T) {
	ls := &fakeListingStore{}
	svc := NewService(ls, &fakeDailyStore{}, &fakeEODClient{}, &fakeLogger{})

	_, err := svc.UpdateListingFields(context.Background(), uuid.New(), nil, nil, nil, nil, nil, nil, nil)
	if !errors.Is(err, ErrNoListingFieldsToUpdate) {
		t.Fatalf("expected ErrNoListingFieldsToUpdate, got %v", err)
	}
}

func TestUpdateListingFields_RejectsInvalidCurrency(t *testing.T) {
	id := uuid.New()
	invalid := "XYZ"
	ls := &fakeListingStore{
		fetchByIDFn: func(ctx context.Context, id uuid.UUID) (*Listing, error) {
			return &Listing{ID: id, Symbol: "AAA", Name: "A", Active: true, Source: SourceAlphaVantage}, nil
		},
	}
	svc := NewService(ls, &fakeDailyStore{}, &fakeEODClient{}, &fakeLogger{})

	_, err := svc.UpdateListingFields(context.Background(), id, nil, nil, nil, &invalid, nil, nil, nil)
	if !errors.Is(err, ErrInvalidListingCurrency) {
		t.Fatalf("expected ErrInvalidListingCurrency, got %v", err)
	}
}

func TestUpdateListingFields_ReturnsNotFound(t *testing.T) {
	id := uuid.New()
	desc := "d"
	ls := &fakeListingStore{
		fetchByIDFn: func(ctx context.Context, gotID uuid.UUID) (*Listing, error) {
			return nil, nil
		},
	}
	svc := NewService(ls, &fakeDailyStore{}, &fakeEODClient{}, &fakeLogger{})

	_, err := svc.UpdateListingFields(context.Background(), id, &desc, nil, nil, nil, nil, nil, nil)
	if !errors.Is(err, ErrListingNotFound) {
		t.Fatalf("expected ErrListingNotFound, got %v", err)
	}
}

func TestUpdateListingFields_PersistError(t *testing.T) {
	id := uuid.New()
	desc := "d"
	ls := &fakeListingStore{
		fetchByIDFn: func(ctx context.Context, gotID uuid.UUID) (*Listing, error) {
			return &Listing{ID: gotID, Symbol: "AAA", Name: "A", Active: true, Source: SourceAlphaVantage}, nil
		},
		updateFieldsFn: func(ctx context.Context, listing *Listing) error {
			return errors.New("persist failed")
		},
	}
	svc := NewService(ls, &fakeDailyStore{}, &fakeEODClient{}, &fakeLogger{})

	_, err := svc.UpdateListingFields(context.Background(), id, &desc, nil, nil, nil, nil, nil, nil)
	if err == nil || !strings.Contains(err.Error(), "failed to persist listing") {
		t.Fatalf("expected persist error, got %v", err)
	}
}

func TestListListings_ReturnsListings(t *testing.T) {
	id := uuid.New()
	ls := &fakeListingStore{
		listFn: func(ctx context.Context) ([]*Listing, error) {
			return []*Listing{
				{
					ID:     id,
					Symbol: "AAA",
					Name:   "Test",
					Source: SourceAlphaVantage,
					Active: true,
				},
			}, nil
		},
	}
	svc := NewService(ls, &fakeDailyStore{}, &fakeEODClient{}, &fakeLogger{})

	listings, err := svc.ListListings(context.Background())
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}
	if len(listings) != 1 {
		t.Fatalf("expected one listing, got %d", len(listings))
	}
	if listings[0].ID != id {
		t.Fatalf("expected listing id %s, got %s", id, listings[0].ID)
	}
}

func TestListListings_StoreError(t *testing.T) {
	ls := &fakeListingStore{
		listFn: func(ctx context.Context) ([]*Listing, error) {
			return nil, errors.New("store unavailable")
		},
	}
	svc := NewService(ls, &fakeDailyStore{}, &fakeEODClient{}, &fakeLogger{})

	_, err := svc.ListListings(context.Background())
	if err == nil || !strings.Contains(err.Error(), "ListListings failed to fetch listings") {
		t.Fatalf("expected list error, got %v", err)
	}
}
