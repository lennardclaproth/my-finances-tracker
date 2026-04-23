package assets

import (
	"context"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/lennardclaproth/my-finances-tracker/internal/account"
	"github.com/lennardclaproth/my-finances-tracker/internal/money"
	"github.com/lennardclaproth/my-finances-tracker/internal/portfolio"
)

func TestParseEffectiveDate_FutureDateRejected(t *testing.T) {
	now := func() time.Time {
		return time.Date(2026, 3, 10, 8, 0, 0, 0, time.UTC)
	}

	_, err := parseEffectiveDate("2026-03-11", now)
	if err == nil {
		t.Fatal("expected future-date error, got nil")
	}
	if err != ErrAssetEffectiveDateFuture {
		t.Fatalf("expected ErrAssetEffectiveDateFuture, got %v", err)
	}
}

func TestParseDirection_InvalidDirectionRejected(t *testing.T) {
	_, err := parseDirection("sideways")
	if err == nil {
		t.Fatal("expected direction validation error, got nil")
	}
	if err != ErrAssetDirectionInvalid {
		t.Fatalf("expected ErrAssetDirectionInvalid, got %v", err)
	}
}

func TestParseSignedWorth_AllowsNegative(t *testing.T) {
	value, err := parseSignedWorth("-125000.50")
	if err != nil {
		t.Fatalf("expected successful parse, got %v", err)
	}

	expected, err := money.NewPrice(125000.50)
	if err != nil {
		t.Fatalf("failed to build expected price: %v", err)
	}
	if value != -expected {
		t.Fatalf("expected %d, got %d", -expected, value)
	}
}

func TestToGrowthPoints_UsesLatestValuePerDay(t *testing.T) {
	classID := uuid.New()
	itemID := uuid.New()
	accountID := uuid.New()
	entries := []*HistoryEntry{
		{
			ID:              uuid.New(),
			AccountID:       accountID,
			ClassID:         classID,
			ItemID:          itemID,
			ClassTotalWorth: 100_000_000_000,
			EffectiveDate:   time.Date(2026, 3, 1, 9, 0, 0, 0, time.UTC),
		},
		{
			ID:              uuid.New(),
			AccountID:       accountID,
			ClassID:         classID,
			ItemID:          itemID,
			ClassTotalWorth: 120_000_000_000,
			EffectiveDate:   time.Date(2026, 3, 1, 14, 0, 0, 0, time.UTC),
		},
		{
			ID:              uuid.New(),
			AccountID:       accountID,
			ClassID:         classID,
			ItemID:          itemID,
			ClassTotalWorth: 150_000_000_000,
			EffectiveDate:   time.Date(2026, 3, 2, 10, 0, 0, 0, time.UTC),
		},
	}

	points := toGrowthPoints(entries)
	if len(points) != 2 {
		t.Fatalf("expected 2 growth points, got %d", len(points))
	}

	if points[0].Date.Format("2006-01-02") != "2026-03-01" {
		t.Fatalf("expected first date 2026-03-01, got %s", points[0].Date.Format("2006-01-02"))
	}
	if points[0].TotalWorth != 120_000_000_000 {
		t.Fatalf("expected latest same-day worth 120000000000, got %d", points[0].TotalWorth)
	}
	if points[1].Date.Format("2006-01-02") != "2026-03-02" {
		t.Fatalf("expected second date 2026-03-02, got %s", points[1].Date.Format("2006-01-02"))
	}
	if points[1].TotalWorth != 150_000_000_000 {
		t.Fatalf("expected second-day worth 150000000000, got %d", points[1].TotalWorth)
	}
}

func TestGrowthPctFromInception_ComputesFromFirstAndLatestWorth(t *testing.T) {
	result := growthPctFromInception(100_000_000_000, 125_000_000_000)
	if result == nil {
		t.Fatal("expected growth percentage, got nil")
	}
	if *result != 25 {
		t.Fatalf("expected growth percentage 25, got %v", *result)
	}
}

func TestGrowthPctFromInception_InceptionZeroReturnsNil(t *testing.T) {
	result := growthPctFromInception(0, 125_000_000_000)
	if result != nil {
		t.Fatalf("expected nil growth percentage, got %v", *result)
	}
}

func TestBuildSnapshotsFromHistory_DenseCarryForwardIncludesToday(t *testing.T) {
	accountID := uuid.New()
	classA := uuid.New()
	classB := uuid.New()
	itemA := uuid.New()
	itemB := uuid.New()

	p90 := mustPrice(t, 90)
	p100 := mustPrice(t, 100)
	p50 := mustPrice(t, 50)
	p120 := mustPrice(t, 120)
	p80 := mustPrice(t, 80)

	history := []*HistoryEntry{
		{
			ID:              uuid.New(),
			AccountID:       accountID,
			ClassID:         classA,
			ItemID:          itemA,
			ClassTotalWorth: p90,
			EffectiveDate:   time.Date(2026, 3, 1, 9, 0, 0, 0, time.UTC),
		},
		{
			ID:              uuid.New(),
			AccountID:       accountID,
			ClassID:         classA,
			ItemID:          itemA,
			ClassTotalWorth: p100,
			EffectiveDate:   time.Date(2026, 3, 1, 12, 0, 0, 0, time.UTC),
		},
		{
			ID:              uuid.New(),
			AccountID:       accountID,
			ClassID:         classB,
			ItemID:          itemB,
			ClassTotalWorth: p50,
			EffectiveDate:   time.Date(2026, 3, 2, 11, 0, 0, 0, time.UTC),
		},
		{
			ID:              uuid.New(),
			AccountID:       accountID,
			ClassID:         classA,
			ItemID:          itemA,
			ClassTotalWorth: p120,
			EffectiveDate:   time.Date(2026, 3, 3, 10, 0, 0, 0, time.UTC),
		},
		{
			ID:              uuid.New(),
			AccountID:       accountID,
			ClassID:         classB,
			ItemID:          itemB,
			ClassTotalWorth: p80,
			EffectiveDate:   time.Date(2026, 3, 4, 10, 0, 0, 0, time.UTC),
		},
	}

	today := time.Date(2026, 3, 5, 0, 0, 0, 0, time.UTC)
	snapshots := buildSnapshotsFromHistory(accountID, history, today)
	if len(snapshots) != 5 {
		t.Fatalf("expected dense 5-day snapshot series, got %d", len(snapshots))
	}

	expectedTotals := []money.Price{
		mustPrice(t, 100), // latest same-day class A value should win
		mustPrice(t, 150), // class A carry-forward + class B first value
		mustPrice(t, 170), // class A updated, class B carry-forward
		mustPrice(t, 200), // class B updated
		mustPrice(t, 200), // include today with carry-forward
	}
	expectedDates := []string{
		"2026-03-01",
		"2026-03-02",
		"2026-03-03",
		"2026-03-04",
		"2026-03-05",
	}
	for i, row := range snapshots {
		if row == nil {
			t.Fatalf("snapshot at index %d is nil", i)
		}
		if row.AccountID != accountID {
			t.Fatalf("expected account id %s, got %s", accountID, row.AccountID)
		}
		if rowDate := row.OccurredAt.Format("2006-01-02"); rowDate != expectedDates[i] {
			t.Fatalf("expected snapshot date %s, got %s", expectedDates[i], rowDate)
		}
		if row.TotalWorth != expectedTotals[i] {
			t.Fatalf("expected total %d at index %d, got %d", expectedTotals[i], i, row.TotalWorth)
		}
	}
}

func TestRebuildTotalSnapshots_NoHistoryFallsBackToCurrentWorth(t *testing.T) {
	accountID := uuid.New()
	store := &fallbackSnapshotStore{
		sumAccountWorth: mustPrice(t, 24000),
	}
	svc := NewService(
		fallbackAccountFetcher{},
		fallbackPortfolioSnapshotLister{},
		store,
		nil,
	)
	svc.now = func() time.Time {
		return time.Date(2026, 3, 10, 15, 4, 0, 0, time.UTC)
	}

	if err := svc.RebuildTotalSnapshots(context.Background(), accountID); err != nil {
		t.Fatalf("unexpected rebuild error: %v", err)
	}
	if !store.deletedSnapshots {
		t.Fatal("expected snapshots to be deleted before upsert")
	}
	if len(store.upsertedSnapshots) != 1 {
		t.Fatalf("expected exactly one fallback snapshot, got %d", len(store.upsertedSnapshots))
	}
	row := store.upsertedSnapshots[0]
	if row == nil {
		t.Fatal("expected fallback snapshot row, got nil")
	}
	if row.AccountID != accountID {
		t.Fatalf("expected fallback snapshot account id %s, got %s", accountID, row.AccountID)
	}
	if row.TotalWorth != store.sumAccountWorth {
		t.Fatalf("expected fallback total %d, got %d", store.sumAccountWorth, row.TotalWorth)
	}
	if rowDate := row.OccurredAt.Format("2006-01-02"); rowDate != "2026-03-10" {
		t.Fatalf("expected fallback date 2026-03-10, got %s", rowDate)
	}
}

func mustPrice(t *testing.T, value float64) money.Price {
	t.Helper()
	price, err := money.NewPrice(value)
	if err != nil {
		t.Fatalf("failed building price from %f: %v", value, err)
	}
	return price
}

type fallbackAccountFetcher struct{}

func (fallbackAccountFetcher) FetchByID(_ context.Context, id uuid.UUID) (*account.Account, error) {
	return &account.Account{ID: id}, nil
}

type fallbackPortfolioSnapshotLister struct{}

func (fallbackPortfolioSnapshotLister) ListForAccount(_ context.Context, _ uuid.UUID, _, _ *time.Time) ([]*portfolio.PortfolioSnapshot, error) {
	return nil, nil
}

type fallbackSnapshotStore struct {
	sumAccountWorth   money.Price
	deletedSnapshots  bool
	upsertedSnapshots []*Snapshot
}

func (s *fallbackSnapshotStore) WithTx(ctx context.Context, fn func(ctx context.Context) error) error {
	return fn(ctx)
}

func (s *fallbackSnapshotStore) EnsureAccount(_ context.Context, _ *Account) error { return nil }
func (s *fallbackSnapshotStore) CreateClass(_ context.Context, _ *Class) error     { return nil }
func (s *fallbackSnapshotStore) FetchClassByID(_ context.Context, _, _ uuid.UUID) (*Class, error) {
	return nil, nil
}
func (s *fallbackSnapshotStore) FetchClassBySource(_ context.Context, _ uuid.UUID, _ ClassSource) (*Class, error) {
	return nil, nil
}
func (s *fallbackSnapshotStore) ListClassesForAccount(_ context.Context, _ uuid.UUID, _ bool) ([]*Class, error) {
	return nil, nil
}
func (s *fallbackSnapshotStore) UpdateClass(_ context.Context, _, _ uuid.UUID, _ *string, _ *bool) error {
	return nil
}
func (s *fallbackSnapshotStore) DeleteClass(_ context.Context, _, _ uuid.UUID) (int64, error) {
	return 0, nil
}
func (s *fallbackSnapshotStore) CreateItem(_ context.Context, _ *Item) error { return nil }
func (s *fallbackSnapshotStore) FetchItemByID(_ context.Context, _, _, _ uuid.UUID) (*Item, error) {
	return nil, nil
}
func (s *fallbackSnapshotStore) FetchItemByClassAndName(_ context.Context, _, _ uuid.UUID, _ string) (*Item, error) {
	return nil, nil
}
func (s *fallbackSnapshotStore) ListItemsByClass(_ context.Context, _, _ uuid.UUID, _ bool) ([]*Item, error) {
	return nil, nil
}
func (s *fallbackSnapshotStore) UpdateItemWorth(_ context.Context, _, _, _ uuid.UUID, _ money.Price) error {
	return nil
}
func (s *fallbackSnapshotStore) SumClassWorth(_ context.Context, _, _ uuid.UUID) (money.Price, error) {
	return 0, nil
}
func (s *fallbackSnapshotStore) CreateHistory(_ context.Context, _ *HistoryEntry) error { return nil }
func (s *fallbackSnapshotStore) DeleteHistoryByClass(_ context.Context, _, _ uuid.UUID) error {
	return nil
}
func (s *fallbackSnapshotStore) ListHistoryByClass(_ context.Context, _, _ uuid.UUID, _ int, _ bool) ([]*HistoryEntry, error) {
	return nil, nil
}
func (s *fallbackSnapshotStore) ListHistoryForAccount(_ context.Context, _ uuid.UUID, _ int, _ bool) ([]*HistoryEntry, error) {
	return nil, nil
}
func (s *fallbackSnapshotStore) DeleteSnapshotsByAccount(_ context.Context, _ uuid.UUID) error {
	s.deletedSnapshots = true
	return nil
}
func (s *fallbackSnapshotStore) UpsertSnapshots(_ context.Context, snapshots []*Snapshot) error {
	s.upsertedSnapshots = snapshots
	return nil
}
func (s *fallbackSnapshotStore) ListSnapshotsForAccount(_ context.Context, _ uuid.UUID, _, _ *time.Time) ([]*Snapshot, error) {
	return nil, nil
}
func (s *fallbackSnapshotStore) SumAccountWorth(_ context.Context, _ uuid.UUID, _ bool) (money.Price, error) {
	return s.sumAccountWorth, nil
}
