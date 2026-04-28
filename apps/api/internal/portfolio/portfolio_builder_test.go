// package portfolio

// import (
// 	"context"
// 	"errors"
// 	"math"
// 	"strings"
// 	"testing"
// 	"time"

// 	"github.com/google/uuid"
// 	"github.com/lennardclaproth/my-finances-tracker/internal/date"
// )

// type fakePositionFetcher struct {
// 	getForAccountFn        func(ctx context.Context, accID uuid.UUID) ([]*Position, error)
// 	getSnapshotsForAcctFn  func(ctx context.Context, accID uuid.UUID) ([]*PositionSnapshot, error)
// 	getForAccountCalls     int
// 	getSnapshotsCalls      int
// 	getForAccountArg       uuid.UUID
// 	getSnapshotsForAcctArg uuid.UUID
// }

// func (f *fakePositionFetcher) GetForAccount(ctx context.Context, accID uuid.UUID) ([]*Position, error) {
// 	f.getForAccountCalls++
// 	f.getForAccountArg = accID
// 	if f.getForAccountFn != nil {
// 		return f.getForAccountFn(ctx, accID)
// 	}
// 	return []*Position{}, nil
// }

// func (f *fakePositionFetcher) GetSnapshotsForAccount(ctx context.Context, accID uuid.UUID) ([]*PositionSnapshot, error) {
// 	f.getSnapshotsCalls++
// 	f.getSnapshotsForAcctArg = accID
// 	if f.getSnapshotsForAcctFn != nil {
// 		return f.getSnapshotsForAcctFn(ctx, accID)
// 	}
// 	return []*PositionSnapshot{}, nil
// }

// type fakeAccountStore struct {
// 	tryAcquireFn    func(ctx context.Context, id uuid.UUID) (bool, error)
// 	releaseFn       func(ctx context.Context, id uuid.UUID) error
// 	tryAcquireArg   uuid.UUID
// 	releaseArg      uuid.UUID
// 	tryAcquireCalls int
// 	releaseCalls    int
// }

// func (f *fakeAccountStore) TryAcquireBuildLock(ctx context.Context, id uuid.UUID) (bool, error) {
// 	f.tryAcquireCalls++
// 	f.tryAcquireArg = id
// 	if f.tryAcquireFn != nil {
// 		return f.tryAcquireFn(ctx, id)
// 	}
// 	return true, nil
// }

// func (f *fakeAccountStore) ReleaseBuildLock(ctx context.Context, id uuid.UUID) error {
// 	f.releaseCalls++
// 	f.releaseArg = id
// 	if f.releaseFn != nil {
// 		return f.releaseFn(ctx, id)
// 	}
// 	return nil
// }

// type fakePortfolioStore struct {
// 	createSnapshotFn func(ctx context.Context, snapshot *PortfolioSnapshot) error
// 	cleanFn          func(ctx context.Context, accID uuid.UUID) error
// 	createCalls      int
// 	created          []*PortfolioSnapshot
// 	cleanCalls       int
// 	cleanArg         uuid.UUID
// }

// func (f *fakePortfolioStore) CreateSnapshot(ctx context.Context, snapshot *PortfolioSnapshot) error {
// 	f.createCalls++
// 	f.created = append(f.created, snapshot)
// 	if f.createSnapshotFn != nil {
// 		return f.createSnapshotFn(ctx, snapshot)
// 	}
// 	return nil
// }

// func (f *fakePortfolioStore) Clean(ctx context.Context, accID uuid.UUID) error {
// 	f.cleanCalls++
// 	f.cleanArg = accID
// 	if f.cleanFn != nil {
// 		return f.cleanFn(ctx, accID)
// 	}
// 	return nil
// }

// func newTestBuilder(ps *fakePositionStore, ts *fakeTransactionStore, ls *fakeListingStore, mds *fakeMarketDataService, pf *fakePositionFetcher, as *fakeAccountStore, pst *fakePortfolioStore) PortfolioBuilder {
// 	if ps == nil {
// 		ps = &fakePositionStore{}
// 	}
// 	if ts == nil {
// 		ts = &fakeTransactionStore{}
// 	}
// 	if ls == nil {
// 		ls = &fakeListingStore{}
// 	}
// 	if mds == nil {
// 		mds = &fakeMarketDataService{}
// 	}
// 	if pf == nil {
// 		pf = &fakePositionFetcher{}
// 	}
// 	if as == nil {
// 		as = &fakeAccountStore{}
// 	}
// 	if pst == nil {
// 		pst = &fakePortfolioStore{}
// 	}
// 	posBuilder := NewPositionBuilder(ps, ts, ls, mds)
// 	return NewPortfolioBuilder(*posBuilder, pf, ts, as, pst)
// }

// func TestPortfolioBuilder_Build_TryAcquireLockError(t *testing.T) {
// 	accID := uuid.New()
// 	as := &fakeAccountStore{
// 		tryAcquireFn: func(ctx context.Context, id uuid.UUID) (bool, error) {
// 			if id != accID {
// 				t.Fatalf("expected accID %s, got %s", accID, id)
// 			}
// 			return false, errors.New("lock down")
// 		},
// 	}
// 	pb := newTestBuilder(nil, nil, nil, nil, nil, as, nil)

// 	err := pb.Build(context.Background(), accID)
// 	if err == nil || !strings.Contains(err.Error(), "lock down") {
// 		t.Fatalf("expected lock error, got %v", err)
// 	}
// 	if as.releaseCalls != 0 {
// 		t.Fatalf("expected release not to be called, got %d", as.releaseCalls)
// 	}
// }

// func TestPortfolioBuilder_Build_LockNotAcquired(t *testing.T) {
// 	accID := uuid.New()
// 	as := &fakeAccountStore{
// 		tryAcquireFn: func(ctx context.Context, id uuid.UUID) (bool, error) {
// 			return false, nil
// 		},
// 	}
// 	pb := newTestBuilder(nil, nil, nil, nil, nil, as, nil)

// 	err := pb.Build(context.Background(), accID)
// 	if !errors.Is(err, ErrBuildInProgress) {
// 		t.Fatalf("expected ErrBuildInProgress, got %v", err)
// 	}
// 	if as.releaseCalls != 0 {
// 		t.Fatalf("expected release not to be called, got %d", as.releaseCalls)
// 	}
// }

// func TestPortfolioBuilder_Build_BuildPositionsError_PropagatesAndReleases(t *testing.T) {
// 	accID := uuid.New()
// 	ps := &fakePositionStore{
// 		cleanFn: func(ctx context.Context, gotAccID uuid.UUID) error {
// 			return errors.New("clean failed")
// 		},
// 	}
// 	as := &fakeAccountStore{}
// 	pb := newTestBuilder(ps, &fakeTransactionStore{}, &fakeListingStore{}, &fakeMarketDataService{}, nil, as, nil)

// 	err := pb.Build(context.Background(), accID)
// 	if err == nil || !strings.Contains(err.Error(), "clean failed") {
// 		t.Fatalf("expected clean error, got %v", err)
// 	}
// 	if as.releaseCalls != 1 {
// 		t.Fatalf("expected release called once, got %d", as.releaseCalls)
// 	}
// }

// func TestPortfolioBuilder_Build_GetForAccountError_PropagatesAndReleases(t *testing.T) {
// 	accID := uuid.New()
// 	pf := &fakePositionFetcher{
// 		getForAccountFn: func(ctx context.Context, id uuid.UUID) ([]*Position, error) {
// 			return nil, errors.New("get positions failed")
// 		},
// 	}
// 	as := &fakeAccountStore{}
// 	pb := newTestBuilder(&fakePositionStore{}, &fakeTransactionStore{}, &fakeListingStore{}, &fakeMarketDataService{}, pf, as, nil)

// 	err := pb.Build(context.Background(), accID)
// 	if err == nil || !strings.Contains(err.Error(), "get positions failed") {
// 		t.Fatalf("expected GetForAccount error, got %v", err)
// 	}
// 	if as.releaseCalls != 1 {
// 		t.Fatalf("expected release called once, got %d", as.releaseCalls)
// 	}
// }

// func TestPortfolioBuilder_Build_BuildPositionSnapshotsError_Propagates(t *testing.T) {
// 	accID := uuid.New()
// 	posID := uuid.New()
// 	ps := &fakePositionStore{
// 		getByIDFn: func(ctx context.Context, id uuid.UUID) (*Position, error) {
// 			if id == posID {
// 				return nil, errors.New("get position failed")
// 			}
// 			return &Position{ID: id}, nil
// 		},
// 	}
// 	pf := &fakePositionFetcher{
// 		getForAccountFn: func(ctx context.Context, id uuid.UUID) ([]*Position, error) {
// 			return []*Position{{ID: posID}}, nil
// 		},
// 		getSnapshotsForAcctFn: func(ctx context.Context, id uuid.UUID) ([]*PositionSnapshot, error) {
// 			return []*PositionSnapshot{{OccurredAt: date.StartOfDayUTC(time.Now().AddDate(0, 0, -1))}}, nil
// 		},
// 	}
// 	as := &fakeAccountStore{}
// 	pb := newTestBuilder(ps, &fakeTransactionStore{}, &fakeListingStore{}, &fakeMarketDataService{}, pf, as, &fakePortfolioStore{})

// 	err := pb.Build(context.Background(), accID)
// 	if err == nil || !strings.Contains(err.Error(), "get position failed") {
// 		t.Fatalf("expected BuildPositionSnapshots error, got %v", err)
// 	}
// }

// func TestPortfolioBuilder_Build_CreateSnapshotError_Propagates(t *testing.T) {
// 	accID := uuid.New()
// 	pf := &fakePositionFetcher{
// 		getForAccountFn: func(ctx context.Context, id uuid.UUID) ([]*Position, error) {
// 			return []*Position{}, nil
// 		},
// 		getSnapshotsForAcctFn: func(ctx context.Context, id uuid.UUID) ([]*PositionSnapshot, error) {
// 			return []*PositionSnapshot{{OccurredAt: date.StartOfDayUTC(time.Now().AddDate(0, 0, -1))}}, nil
// 		},
// 	}
// 	pst := &fakePortfolioStore{
// 		createSnapshotFn: func(ctx context.Context, snap *PortfolioSnapshot) error {
// 			return errors.New("create failed")
// 		},
// 	}
// 	pb := newTestBuilder(&fakePositionStore{}, &fakeTransactionStore{}, &fakeListingStore{}, &fakeMarketDataService{}, pf, &fakeAccountStore{}, pst)

// 	err := pb.Build(context.Background(), accID)
// 	if err == nil || !strings.Contains(err.Error(), "create failed") {
// 		t.Fatalf("expected CreateSnapshot error, got %v", err)
// 	}
// }

// func TestPortfolioBuilder_Build_NoPositionSnapshots_NoPanic_NoCreate(t *testing.T) {
// 	accID := uuid.New()
// 	pf := &fakePositionFetcher{
// 		getForAccountFn: func(ctx context.Context, id uuid.UUID) ([]*Position, error) {
// 			return []*Position{}, nil
// 		},
// 		getSnapshotsForAcctFn: func(ctx context.Context, id uuid.UUID) ([]*PositionSnapshot, error) {
// 			return []*PositionSnapshot{}, nil
// 		},
// 	}
// 	pst := &fakePortfolioStore{}
// 	pb := newTestBuilder(&fakePositionStore{}, &fakeTransactionStore{}, &fakeListingStore{}, &fakeMarketDataService{}, pf, &fakeAccountStore{}, pst)

// 	defer func() {
// 		if r := recover(); r != nil {
// 			t.Fatalf("unexpected panic on empty snapshots: %v", r)
// 		}
// 	}()

// 	err := pb.Build(context.Background(), accID)
// 	if !errors.Is(err, ErrPortfolioNoSnapshots) {
// 		t.Fatalf("expected ErrPortfolioNoSnapshots, got %v", err)
// 	}
// 	if pst.createCalls != 0 {
// 		t.Fatalf("expected no portfolio snapshots created, got %d", pst.createCalls)
// 	}
// }

// func TestPortfolioBuilder_Build_DailyPnL_UsesPrevTotalPnL(t *testing.T) {
// 	accID := uuid.New()
// 	today := date.StartOfDayUTC(time.Now())
// 	day1 := today.AddDate(0, 0, -2)
// 	day2 := today.AddDate(0, 0, -1)
// 	pf := &fakePositionFetcher{
// 		getForAccountFn: func(ctx context.Context, id uuid.UUID) ([]*Position, error) {
// 			return []*Position{}, nil
// 		},
// 		getSnapshotsForAcctFn: func(ctx context.Context, id uuid.UUID) ([]*PositionSnapshot, error) {
// 			return []*PositionSnapshot{
// 				{OccurredAt: day1, MarketValue: mustPrice(t, 100), TotalPnL: mustPrice(t, 100)},
// 				{OccurredAt: day2, MarketValue: mustPrice(t, 110), TotalPnL: mustPrice(t, 110)},
// 			}, nil
// 		},
// 	}
// 	pst := &fakePortfolioStore{}
// 	pb := newTestBuilder(&fakePositionStore{}, &fakeTransactionStore{}, &fakeListingStore{}, &fakeMarketDataService{}, pf, &fakeAccountStore{}, pst)

// 	err := pb.Build(context.Background(), accID)
// 	if err != nil {
// 		t.Fatalf("expected no error, got %v", err)
// 	}
// 	if pst.createCalls != 2 {
// 		t.Fatalf("expected 2 snapshots, got %d", pst.createCalls)
// 	}
// 	if pst.created[1].DailyDeltaPnL != mustPrice(t, 10) {
// 		t.Fatalf("expected day2 DailyPnL 10.000000, got %s", pst.created[1].DailyDeltaPnL)
// 	}
// 	if math.Abs(pst.created[1].DailyDeltaPnLPct-10) > 1e-9 {
// 		t.Fatalf("expected day2 DailyDeltaPnLPct 10%%, got %f", pst.created[1].DailyDeltaPnLPct)
// 	}
// }

// func TestPortfolioBuilder_Build_CarriesForwardAfterLastSnapshotDay(t *testing.T) {
// 	accID := uuid.New()
// 	today := date.StartOfDayUTC(time.Now())
// 	day1 := today.AddDate(0, 0, -2)
// 	pf := &fakePositionFetcher{
// 		getForAccountFn: func(ctx context.Context, id uuid.UUID) ([]*Position, error) {
// 			return []*Position{}, nil
// 		},
// 		getSnapshotsForAcctFn: func(ctx context.Context, id uuid.UUID) ([]*PositionSnapshot, error) {
// 			return []*PositionSnapshot{{OccurredAt: day1, MarketValue: mustPrice(t, 100)}}, nil
// 		},
// 	}
// 	pst := &fakePortfolioStore{}
// 	pb := newTestBuilder(&fakePositionStore{}, &fakeTransactionStore{}, &fakeListingStore{}, &fakeMarketDataService{}, pf, &fakeAccountStore{}, pst)

// 	err := pb.Build(context.Background(), accID)
// 	if err != nil {
// 		t.Fatalf("expected no error, got %v", err)
// 	}
// 	if pst.createCalls != 2 {
// 		t.Fatalf("expected 2 snapshots (day1 + yesterday), got %d", pst.createCalls)
// 	}
// 	if pst.created[1].MarketValue != mustPrice(t, 100) {
// 		t.Fatalf("expected day2 carried MarketValue 100.000000, got %s", pst.created[1].MarketValue)
// 	}
// }

// func TestPortfolioBuilder_Build_CashFlowNeutralReturn_UsesSnapshotDerivedNetCashflow(t *testing.T) {
// 	accID := uuid.New()
// 	today := date.StartOfDayUTC(time.Now())
// 	day1 := today.AddDate(0, 0, -2)
// 	day2 := today.AddDate(0, 0, -1)
// 	pf := &fakePositionFetcher{
// 		getForAccountFn: func(ctx context.Context, id uuid.UUID) ([]*Position, error) {
// 			return []*Position{}, nil
// 		},
// 		getSnapshotsForAcctFn: func(ctx context.Context, id uuid.UUID) ([]*PositionSnapshot, error) {
// 			return []*PositionSnapshot{
// 				{OccurredAt: day1, MarketValue: mustPrice(t, 100), TotalPnL: mustPrice(t, 20)},
// 				{OccurredAt: day2, MarketValue: mustPrice(t, 150), TotalPnL: mustPrice(t, 20)},
// 			}, nil
// 		},
// 	}
// 	pst := &fakePortfolioStore{}
// 	pb := newTestBuilder(&fakePositionStore{}, &fakeTransactionStore{}, &fakeListingStore{}, &fakeMarketDataService{}, pf, &fakeAccountStore{}, pst)

//		err := pb.Build(context.Background(), accID)
//		if err != nil {
//			t.Fatalf("expected no error, got %v", err)
//		}
//		if pst.createCalls != 2 {
//			t.Fatalf("expected 2 snapshots, got %d", pst.createCalls)
//		}
//		day2Snap := pst.created[1]
//		if day2Snap.NetCashflow != mustPrice(t, 50) {
//			t.Fatalf("expected day2 net cashflow 50.000000, got %s", day2Snap.NetCashflow)
//		}
//		if math.Abs(day2Snap.DailyDeltaPnLPct) > 1e-9 {
//			t.Fatalf("expected day2 DailyDeltaPnLPct 0%%, got %f", day2Snap.DailyDeltaPnLPct)
//		}
//		if math.Abs(day2Snap.TimeWeightedReturnPct) > 1e-9 {
//			t.Fatalf("expected day2 flow-neutral return 0%%, got %f", day2Snap.TimeWeightedReturnPct)
//		}
//	}
package portfolio
