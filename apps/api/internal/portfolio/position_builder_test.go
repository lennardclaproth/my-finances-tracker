// package portfolio

// import (
// 	"context"
// 	"errors"
// 	"strings"
// 	"testing"
// 	"time"

// 	"github.com/google/uuid"
// 	"github.com/lennardclaproth/my-finances-tracker/internal/date"
// 	"github.com/lennardclaproth/my-finances-tracker/internal/marketdata"
// 	"github.com/lennardclaproth/my-finances-tracker/internal/money"
// )

// type fakePositionStore struct {
// 	cleanFn           func(ctx context.Context, accID uuid.UUID) error
// 	createManyFn      func(ctx context.Context, positions []*Position) error
// 	getByIDFn         func(ctx context.Context, id uuid.UUID) (*Position, error)
// 	getLastSnapshotFn func(ctx context.Context, positionID uuid.UUID) (*PositionSnapshot, error)
// 	createSnapshotFn  func(ctx context.Context, snap *PositionSnapshot) error

// 	cleanCalls           int
// 	createManyCalls      int
// 	getByIDCalls         int
// 	getLastSnapshotCalls int
// 	createSnapshotCalls  int

// 	createManyArg      []*Position
// 	getByIDArg         uuid.UUID
// 	getLastSnapshotArg uuid.UUID
// 	createdSnapshots   []*PositionSnapshot
// }

// func (f *fakePositionStore) Clean(ctx context.Context, accID uuid.UUID) error {
// 	f.cleanCalls++
// 	if f.cleanFn != nil {
// 		return f.cleanFn(ctx, accID)
// 	}
// 	return nil
// }

// func (f *fakePositionStore) CreateMany(ctx context.Context, positions []*Position) error {
// 	f.createManyCalls++
// 	f.createManyArg = append([]*Position{}, positions...)
// 	if f.createManyFn != nil {
// 		return f.createManyFn(ctx, positions)
// 	}
// 	return nil
// }

// func (f *fakePositionStore) GetByID(ctx context.Context, id uuid.UUID) (*Position, error) {
// 	f.getByIDCalls++
// 	f.getByIDArg = id
// 	if f.getByIDFn != nil {
// 		return f.getByIDFn(ctx, id)
// 	}
// 	return nil, nil
// }

// func (f *fakePositionStore) GetLastSnapshot(ctx context.Context, positionID uuid.UUID) (*PositionSnapshot, error) {
// 	f.getLastSnapshotCalls++
// 	f.getLastSnapshotArg = positionID
// 	if f.getLastSnapshotFn != nil {
// 		return f.getLastSnapshotFn(ctx, positionID)
// 	}
// 	return nil, nil
// }

// func (f *fakePositionStore) CreateSnapshot(ctx context.Context, snap *PositionSnapshot) error {
// 	f.createSnapshotCalls++
// 	f.createdSnapshots = append(f.createdSnapshots, snap)
// 	if f.createSnapshotFn != nil {
// 		return f.createSnapshotFn(ctx, snap)
// 	}
// 	return nil
// }

// type fakeTransactionStore struct {
// 	getASCFn          func(ctx context.Context, accID uuid.UUID) ([]Transaction, error)
// 	updatePositionsFn func(ctx context.Context, transactions []Transaction) error
// 	getByPositionIDFn func(ctx context.Context, positionID uuid.UUID, from *time.Time) ([]Transaction, error)

// 	getASCCalls          int
// 	updatePositionsCalls int
// 	getByPositionIDCalls int

// 	updatePositionsArg []Transaction
// 	getByPositionIDArg struct {
// 		positionID uuid.UUID
// 		from       *time.Time
// 	}
// }

// func (f *fakeTransactionStore) GetASC(ctx context.Context, accID uuid.UUID) ([]Transaction, error) {
// 	f.getASCCalls++
// 	if f.getASCFn != nil {
// 		return f.getASCFn(ctx, accID)
// 	}
// 	return nil, nil
// }

// func (f *fakeTransactionStore) UpdatePositions(ctx context.Context, transactions []Transaction) error {
// 	f.updatePositionsCalls++
// 	f.updatePositionsArg = append([]Transaction{}, transactions...)
// 	if f.updatePositionsFn != nil {
// 		return f.updatePositionsFn(ctx, transactions)
// 	}
// 	return nil
// }

// func (f *fakeTransactionStore) GetByPositionID(ctx context.Context, positionID uuid.UUID, from *time.Time) ([]Transaction, error) {
// 	f.getByPositionIDCalls++
// 	f.getByPositionIDArg.positionID = positionID
// 	if from != nil {
// 		v := *from
// 		f.getByPositionIDArg.from = &v
// 	}
// 	if f.getByPositionIDFn != nil {
// 		return f.getByPositionIDFn(ctx, positionID, from)
// 	}
// 	return nil, nil
// }

// type fakeListingStore struct {
// 	fetchBySymbolOrISINFn func(ctx context.Context, val string) (*marketdata.Listing, error)
// 	fetchByIDFn           func(ctx context.Context, id uuid.UUID) (*marketdata.Listing, error)

// 	fetchBySymbolOrISINCalls int
// 	fetchByIDCalls           int
// 	fetchBySymbolOrISINArgs  []string
// 	fetchByIDArg             uuid.UUID
// }

// func (f *fakeListingStore) FetchBySymbolOrISIN(ctx context.Context, val string) (*marketdata.Listing, error) {
// 	f.fetchBySymbolOrISINCalls++
// 	f.fetchBySymbolOrISINArgs = append(f.fetchBySymbolOrISINArgs, val)
// 	if f.fetchBySymbolOrISINFn != nil {
// 		return f.fetchBySymbolOrISINFn(ctx, val)
// 	}
// 	return nil, nil
// }

// func (f *fakeListingStore) FetchByID(ctx context.Context, id uuid.UUID) (*marketdata.Listing, error) {
// 	f.fetchByIDCalls++
// 	f.fetchByIDArg = id
// 	if f.fetchByIDFn != nil {
// 		return f.fetchByIDFn(ctx, id)
// 	}
// 	return nil, nil
// }

// type fakeMarketDataService struct {
// 	getDailiesFn func(ctx context.Context, symbol string, from, to *time.Time, limit, offset int) (*marketdata.DailyResponse, error)

// 	getDailiesCalls int
// 	getDailiesArg   struct {
// 		symbol string
// 		from   *time.Time
// 		to     *time.Time
// 		limit  int
// 		offset int
// 	}
// }

// func (f *fakeMarketDataService) GetDailies(ctx context.Context, symbol string, from, to *time.Time, limit, offset int) (*marketdata.DailyResponse, error) {
// 	f.getDailiesCalls++
// 	f.getDailiesArg.symbol = symbol
// 	if from != nil {
// 		v := *from
// 		f.getDailiesArg.from = &v
// 	}
// 	if to != nil {
// 		v := *to
// 		f.getDailiesArg.to = &v
// 	}
// 	f.getDailiesArg.limit = limit
// 	f.getDailiesArg.offset = offset
// 	if f.getDailiesFn != nil {
// 		return f.getDailiesFn(ctx, symbol, from, to, limit, offset)
// 	}
// 	return &marketdata.DailyResponse{}, nil
// }

// func mustPrice(t *testing.T, amount float64) money.Price {
// 	t.Helper()
// 	p, err := money.NewPrice(amount)
// 	if err != nil {
// 		t.Fatalf("expected valid price for %f, got error: %v", amount, err)
// 	}
// 	return p
// }

// func mustTx(t *testing.T, tt TransactionType, day time.Time, isin, symbol *string, qty, price, amount float64) Transaction {
// 	t.Helper()
// 	return Transaction{
// 		ID:          uuid.New(),
// 		OccurredAt:  day,
// 		Type:        tt,
// 		ISIN:        isin,
// 		Symbol:      symbol,
// 		Quantity:    qty,
// 		UnitPrice:   mustPrice(t, price),
// 		AmountCents: mustPrice(t, amount),
// 	}
// }

// func TestBuildPositions_CleanError(t *testing.T) {
// 	accID := uuid.New()
// 	ps := &fakePositionStore{
// 		cleanFn: func(ctx context.Context, gotAccID uuid.UUID) error {
// 			if gotAccID != accID {
// 				t.Fatalf("expected account id %s, got %s", accID, gotAccID)
// 			}
// 			return errors.New("clean failed")
// 		},
// 	}
// 	ts := &fakeTransactionStore{}
// 	ls := &fakeListingStore{}
// 	svc := NewPositionBuilder(ps, ts, ls, &fakeMarketDataService{})

// 	err := svc.BuildPositions(context.Background(), accID)
// 	if err == nil || !strings.Contains(err.Error(), "clean failed") {
// 		t.Fatalf("expected clean error, got %v", err)
// 	}
// 	if ts.getASCCalls != 0 {
// 		t.Fatalf("expected GetASC not to be called, got %d", ts.getASCCalls)
// 	}
// }

// func TestBuildPositions_GetASCError(t *testing.T) {
// 	accID := uuid.New()
// 	ps := &fakePositionStore{}
// 	ts := &fakeTransactionStore{
// 		getASCFn: func(ctx context.Context, gotAccID uuid.UUID) ([]Transaction, error) {
// 			if gotAccID != accID {
// 				t.Fatalf("expected account id %s, got %s", accID, gotAccID)
// 			}
// 			return nil, errors.New("get asc failed")
// 		},
// 	}
// 	ls := &fakeListingStore{}
// 	svc := NewPositionBuilder(ps, ts, ls, &fakeMarketDataService{})

// 	err := svc.BuildPositions(context.Background(), accID)
// 	if err == nil || !strings.Contains(err.Error(), "get asc failed") {
// 		t.Fatalf("expected GetASC error, got %v", err)
// 	}
// 	if ps.createManyCalls != 0 {
// 		t.Fatalf("expected CreateMany not to be called, got %d", ps.createManyCalls)
// 	}
// }

// func TestBuildPositions_SkipsCashTransactions(t *testing.T) {
// 	accID := uuid.New()
// 	isin := "NL0000000001"
// 	day := time.Now().UTC().Add(-2 * time.Hour)
// 	ps := &fakePositionStore{}
// 	ts := &fakeTransactionStore{
// 		getASCFn: func(ctx context.Context, gotAccID uuid.UUID) ([]Transaction, error) {
// 			return []Transaction{
// 				mustTx(t, TxCash, day, &isin, nil, 0, 0, 123.45),
// 				mustTx(t, TxBuy, day.Add(time.Minute), &isin, nil, 1, 10, 0),
// 			}, nil
// 		},
// 	}
// 	listing := &marketdata.Listing{ID: uuid.New(), Symbol: "AAA"}
// 	ls := &fakeListingStore{
// 		fetchBySymbolOrISINFn: func(ctx context.Context, val string) (*marketdata.Listing, error) {
// 			return listing, nil
// 		},
// 	}
// 	svc := NewPositionBuilder(ps, ts, ls, &fakeMarketDataService{})

// 	if err := svc.BuildPositions(context.Background(), accID); err != nil {
// 		t.Fatalf("expected no error, got %v", err)
// 	}
// 	if ls.fetchBySymbolOrISINCalls != 1 {
// 		t.Fatalf("expected one listing lookup, got %d", ls.fetchBySymbolOrISINCalls)
// 	}
// 	if len(ps.createManyArg) != 1 {
// 		t.Fatalf("expected one position persisted, got %d", len(ps.createManyArg))
// 	}
// }

// func TestBuildPositions_MapsTransactionsToSinglePositionByID(t *testing.T) {
// 	accID := uuid.New()
// 	isin := "NL0000000002"
// 	openDay := time.Now().UTC().AddDate(0, 0, -3)
// 	ps := &fakePositionStore{}
// 	ts := &fakeTransactionStore{
// 		getASCFn: func(ctx context.Context, gotAccID uuid.UUID) ([]Transaction, error) {
// 			return []Transaction{
// 				mustTx(t, TxBuy, openDay, &isin, nil, 2, 10, 0),
// 				mustTx(t, TxDividend, openDay.Add(12*time.Hour), &isin, nil, 0, 0, 2),
// 				mustTx(t, TxSell, openDay.Add(24*time.Hour), &isin, nil, 1, 11, 0),
// 			}, nil
// 		},
// 	}
// 	ls := &fakeListingStore{
// 		fetchBySymbolOrISINFn: func(ctx context.Context, val string) (*marketdata.Listing, error) {
// 			return &marketdata.Listing{ID: uuid.New(), Symbol: "AAA"}, nil
// 		},
// 	}
// 	svc := NewPositionBuilder(ps, ts, ls, &fakeMarketDataService{})

// 	if err := svc.BuildPositions(context.Background(), accID); err != nil {
// 		t.Fatalf("expected no error, got %v", err)
// 	}
// 	if len(ps.createManyArg) != 1 {
// 		t.Fatalf("expected one position for repeated id, got %d", len(ps.createManyArg))
// 	}
// 	pos := ps.createManyArg[0]
// 	if pos.Quantity != 1 {
// 		t.Fatalf("expected quantity 1 after flow, got %f", pos.Quantity)
// 	}
// }

// func TestBuildPositions_PersistsPositionsWhenListingLookupErrors(t *testing.T) {
// 	accID := uuid.New()
// 	isin := "NL0000000003"
// 	ps := &fakePositionStore{}
// 	ts := &fakeTransactionStore{
// 		getASCFn: func(ctx context.Context, gotAccID uuid.UUID) ([]Transaction, error) {
// 			return []Transaction{
// 				mustTx(t, TxBuy, time.Now().UTC(), &isin, nil, 1, 10, 0),
// 			}, nil
// 		},
// 	}
// 	ls := &fakeListingStore{
// 		fetchBySymbolOrISINFn: func(ctx context.Context, val string) (*marketdata.Listing, error) {
// 			return nil, errors.New("lookup down")
// 		},
// 	}
// 	svc := NewPositionBuilder(ps, ts, ls, &fakeMarketDataService{})

// 	if err := svc.BuildPositions(context.Background(), accID); err != nil {
// 		t.Fatalf("expected no error even if lookup fails, got %v", err)
// 	}
// 	if len(ps.createManyArg) != 1 {
// 		t.Fatalf("expected unresolved position to still be persisted, got %d positions", len(ps.createManyArg))
// 	}
// 	if ps.createManyArg[0].ListingID != nil {
// 		t.Fatalf("expected unresolved position ListingID to remain nil")
// 	}
// }

// func TestBuildPositions_PersistsPositionsWhenListingLookupReturnsNil(t *testing.T) {
// 	accID := uuid.New()
// 	isin := "NL0000000004"
// 	ps := &fakePositionStore{}
// 	ts := &fakeTransactionStore{
// 		getASCFn: func(ctx context.Context, gotAccID uuid.UUID) ([]Transaction, error) {
// 			return []Transaction{
// 				mustTx(t, TxBuy, time.Now().UTC(), &isin, nil, 1, 10, 0),
// 			}, nil
// 		},
// 	}
// 	ls := &fakeListingStore{
// 		fetchBySymbolOrISINFn: func(ctx context.Context, val string) (*marketdata.Listing, error) {
// 			return nil, nil
// 		},
// 	}
// 	svc := NewPositionBuilder(ps, ts, ls, &fakeMarketDataService{})

// 	defer func() {
// 		if r := recover(); r != nil {
// 			t.Fatalf("BuildPositions should not panic on nil listing result; panic: %v", r)
// 		}
// 	}()

// 	if err := svc.BuildPositions(context.Background(), accID); err != nil {
// 		t.Fatalf("expected no error, got %v", err)
// 	}
// 	if len(ps.createManyArg) != 1 {
// 		t.Fatalf("expected unresolved position to still be persisted, got %d positions", len(ps.createManyArg))
// 	}
// }

// func TestBuildPositions_MixedResolvedAndUnresolvedPositions(t *testing.T) {
// 	accID := uuid.New()
// 	isin1 := "NL0000000005"
// 	isin2 := "NL0000000006"
// 	ps := &fakePositionStore{}
// 	ts := &fakeTransactionStore{
// 		getASCFn: func(ctx context.Context, gotAccID uuid.UUID) ([]Transaction, error) {
// 			return []Transaction{
// 				mustTx(t, TxBuy, time.Now().UTC(), &isin1, nil, 1, 10, 0),
// 				mustTx(t, TxBuy, time.Now().UTC().Add(time.Minute), &isin2, nil, 1, 20, 0),
// 			}, nil
// 		},
// 	}
// 	resolvedListing := &marketdata.Listing{ID: uuid.New(), Symbol: "AAA"}
// 	ls := &fakeListingStore{
// 		fetchBySymbolOrISINFn: func(ctx context.Context, val string) (*marketdata.Listing, error) {
// 			if val == isin1 {
// 				return resolvedListing, nil
// 			}
// 			return nil, errors.New("lookup failed")
// 		},
// 	}
// 	svc := NewPositionBuilder(ps, ts, ls, &fakeMarketDataService{})

// 	if err := svc.BuildPositions(context.Background(), accID); err != nil {
// 		t.Fatalf("expected no error, got %v", err)
// 	}
// 	if len(ps.createManyArg) != 2 {
// 		t.Fatalf("expected both positions to be persisted, got %d", len(ps.createManyArg))
// 	}
// 	var withListing, withoutListing int
// 	for _, pos := range ps.createManyArg {
// 		if pos.ListingID != nil {
// 			withListing++
// 		} else {
// 			withoutListing++
// 		}
// 	}
// 	if withListing != 1 || withoutListing != 1 {
// 		t.Fatalf("expected one resolved and one unresolved position, got with=%d without=%d", withListing, withoutListing)
// 	}
// }

// func TestBuildPositions_UpdatePositionsReceivesMappedPositionIDs(t *testing.T) {
// 	accID := uuid.New()
// 	isin := "NL0000000007"
// 	ps := &fakePositionStore{}
// 	ts := &fakeTransactionStore{
// 		getASCFn: func(ctx context.Context, gotAccID uuid.UUID) ([]Transaction, error) {
// 			day := time.Now().UTC()
// 			return []Transaction{
// 				mustTx(t, TxBuy, day, &isin, nil, 1, 10, 0),
// 				mustTx(t, TxSell, day.Add(time.Hour), &isin, nil, 1, 11, 0),
// 			}, nil
// 		},
// 	}
// 	ls := &fakeListingStore{
// 		fetchBySymbolOrISINFn: func(ctx context.Context, val string) (*marketdata.Listing, error) {
// 			return &marketdata.Listing{ID: uuid.New(), Symbol: "AAA"}, nil
// 		},
// 	}
// 	svc := NewPositionBuilder(ps, ts, ls, &fakeMarketDataService{})

// 	if err := svc.BuildPositions(context.Background(), accID); err != nil {
// 		t.Fatalf("expected no error, got %v", err)
// 	}
// 	if ts.updatePositionsCalls != 1 {
// 		t.Fatalf("expected UpdatePositions to be called once, got %d", ts.updatePositionsCalls)
// 	}
// 	for i, tx := range ts.updatePositionsArg {
// 		if tx.PositionID == nil {
// 			t.Fatalf("expected transaction %d to have PositionID set before UpdatePositions", i)
// 		}
// 	}
// }

// func TestBuildPositions_CreateManyErrorStopsBeforeUpdatePositions(t *testing.T) {
// 	accID := uuid.New()
// 	isin := "NL0000000008"
// 	ps := &fakePositionStore{
// 		createManyFn: func(ctx context.Context, positions []*Position) error {
// 			return errors.New("create many failed")
// 		},
// 	}
// 	ts := &fakeTransactionStore{
// 		getASCFn: func(ctx context.Context, gotAccID uuid.UUID) ([]Transaction, error) {
// 			return []Transaction{
// 				mustTx(t, TxBuy, time.Now().UTC(), &isin, nil, 1, 10, 0),
// 			}, nil
// 		},
// 	}
// 	ls := &fakeListingStore{
// 		fetchBySymbolOrISINFn: func(ctx context.Context, val string) (*marketdata.Listing, error) {
// 			return &marketdata.Listing{ID: uuid.New(), Symbol: "AAA"}, nil
// 		},
// 	}
// 	svc := NewPositionBuilder(ps, ts, ls, &fakeMarketDataService{})

// 	err := svc.BuildPositions(context.Background(), accID)
// 	if err == nil || !strings.Contains(err.Error(), "create many failed") {
// 		t.Fatalf("expected CreateMany error, got %v", err)
// 	}
// 	if ts.updatePositionsCalls != 0 {
// 		t.Fatalf("expected UpdatePositions not to be called, got %d", ts.updatePositionsCalls)
// 	}
// }

// func TestBuildPositions_UpdatePositionsErrorPropagates(t *testing.T) {
// 	accID := uuid.New()
// 	isin := "NL0000000009"
// 	ps := &fakePositionStore{}
// 	ts := &fakeTransactionStore{
// 		getASCFn: func(ctx context.Context, gotAccID uuid.UUID) ([]Transaction, error) {
// 			return []Transaction{
// 				mustTx(t, TxBuy, time.Now().UTC(), &isin, nil, 1, 10, 0),
// 			}, nil
// 		},
// 		updatePositionsFn: func(ctx context.Context, transactions []Transaction) error {
// 			return errors.New("update positions failed")
// 		},
// 	}
// 	ls := &fakeListingStore{
// 		fetchBySymbolOrISINFn: func(ctx context.Context, val string) (*marketdata.Listing, error) {
// 			return &marketdata.Listing{ID: uuid.New(), Symbol: "AAA"}, nil
// 		},
// 	}
// 	svc := NewPositionBuilder(ps, ts, ls, &fakeMarketDataService{})

// 	err := svc.BuildPositions(context.Background(), accID)
// 	if err == nil || !strings.Contains(err.Error(), "update positions failed") {
// 		t.Fatalf("expected UpdatePositions error, got %v", err)
// 	}
// }

// func TestBuildPositions_CreatesNewCycleAfterEachFullClose(t *testing.T) {
// 	accID := uuid.New()
// 	isin := "NL0000000010"
// 	base := time.Now().UTC().AddDate(0, 0, -2)
// 	ps := &fakePositionStore{}
// 	ts := &fakeTransactionStore{
// 		getASCFn: func(ctx context.Context, gotAccID uuid.UUID) ([]Transaction, error) {
// 			return []Transaction{
// 				mustTx(t, TxBuy, base, &isin, nil, 1, 10, 0),
// 				mustTx(t, TxSell, base.Add(time.Hour), &isin, nil, 1, 10, 0),
// 				mustTx(t, TxBuy, base.Add(2*time.Hour), &isin, nil, 2, 12, 0),
// 				mustTx(t, TxSell, base.Add(3*time.Hour), &isin, nil, 2, 12, 0),
// 				mustTx(t, TxBuy, base.Add(4*time.Hour), &isin, nil, 3, 9, 0),
// 			}, nil
// 		},
// 	}
// 	ls := &fakeListingStore{
// 		fetchBySymbolOrISINFn: func(ctx context.Context, val string) (*marketdata.Listing, error) {
// 			return &marketdata.Listing{ID: uuid.New(), Symbol: "AAA"}, nil
// 		},
// 	}
// 	svc := NewPositionBuilder(ps, ts, ls, &fakeMarketDataService{})

// 	if err := svc.BuildPositions(context.Background(), accID); err != nil {
// 		t.Fatalf("expected no error, got %v", err)
// 	}
// 	if len(ps.createManyArg) != 3 {
// 		t.Fatalf("expected 3 lifecycle cycles, got %d", len(ps.createManyArg))
// 	}
// 	first := ps.createManyArg[0]
// 	second := ps.createManyArg[1]
// 	third := ps.createManyArg[2]
// 	if first.CloseDate == nil || second.CloseDate == nil {
// 		t.Fatalf("expected first two cycles to be closed")
// 	}
// 	if third.CloseDate != nil {
// 		t.Fatalf("expected third cycle to stay open, got CloseDate=%v", third.CloseDate)
// 	}
// 	if third.Quantity != 3 {
// 		t.Fatalf("expected open third cycle quantity 3, got %f", third.Quantity)
// 	}
// 	if len(ts.updatePositionsArg) != 5 {
// 		t.Fatalf("expected all 5 transactions to be forwarded, got %d", len(ts.updatePositionsArg))
// 	}
// 	for i, tx := range ts.updatePositionsArg {
// 		if tx.PositionID == nil {
// 			t.Fatalf("expected transaction %d to have position mapping", i)
// 		}
// 	}
// 	if *ts.updatePositionsArg[0].PositionID != first.ID || *ts.updatePositionsArg[1].PositionID != first.ID {
// 		t.Fatalf("expected first two transactions to map to first cycle")
// 	}
// 	if *ts.updatePositionsArg[2].PositionID != second.ID || *ts.updatePositionsArg[3].PositionID != second.ID {
// 		t.Fatalf("expected third and fourth transactions to map to second cycle")
// 	}
// 	if *ts.updatePositionsArg[4].PositionID != third.ID {
// 		t.Fatalf("expected fifth transaction to map to third cycle")
// 	}
// }

// func TestBuildPositions_PartialSellUsesAverageCost(t *testing.T) {
// 	accID := uuid.New()
// 	isin := "NL0000000011"
// 	base := time.Now().UTC().AddDate(0, 0, -2)
// 	ps := &fakePositionStore{}
// 	ts := &fakeTransactionStore{
// 		getASCFn: func(ctx context.Context, gotAccID uuid.UUID) ([]Transaction, error) {
// 			return []Transaction{
// 				mustTx(t, TxBuy, base, &isin, nil, 2, 10, 0),
// 				mustTx(t, TxBuy, base.Add(time.Hour), &isin, nil, 2, 20, 0),
// 				mustTx(t, TxSell, base.Add(2*time.Hour), &isin, nil, 1, 30, 0),
// 			}, nil
// 		},
// 	}
// 	ls := &fakeListingStore{
// 		fetchBySymbolOrISINFn: func(ctx context.Context, val string) (*marketdata.Listing, error) {
// 			return &marketdata.Listing{ID: uuid.New(), Symbol: "AAA"}, nil
// 		},
// 	}
// 	svc := NewPositionBuilder(ps, ts, ls, &fakeMarketDataService{})

// 	if err := svc.BuildPositions(context.Background(), accID); err != nil {
// 		t.Fatalf("expected no error, got %v", err)
// 	}
// 	if len(ps.createManyArg) != 1 {
// 		t.Fatalf("expected single cycle, got %d", len(ps.createManyArg))
// 	}
// 	pos := ps.createManyArg[0]
// 	if pos.Quantity != 3 {
// 		t.Fatalf("expected quantity 3 after partial sell, got %f", pos.Quantity)
// 	}
// 	if pos.CostBasis != mustPrice(t, 45) {
// 		t.Fatalf("expected average-cost basis 45.000000, got %s", pos.CostBasis)
// 	}
// 	if pos.CloseDate != nil {
// 		t.Fatalf("expected position to remain open after partial sell")
// 	}
// }

// func TestBuildPositions_OversellClampsToZeroAndClosesCycle(t *testing.T) {
// 	accID := uuid.New()
// 	isin := "NL0000000012"
// 	base := time.Now().UTC().AddDate(0, 0, -2)
// 	ps := &fakePositionStore{}
// 	ts := &fakeTransactionStore{
// 		getASCFn: func(ctx context.Context, gotAccID uuid.UUID) ([]Transaction, error) {
// 			return []Transaction{
// 				mustTx(t, TxBuy, base, &isin, nil, 1, 10, 0),
// 				mustTx(t, TxSell, base.Add(time.Hour), &isin, nil, 3, 12, 0),
// 				mustTx(t, TxSell, base.Add(2*time.Hour), &isin, nil, 1, 13, 0),
// 			}, nil
// 		},
// 	}
// 	ls := &fakeListingStore{
// 		fetchBySymbolOrISINFn: func(ctx context.Context, val string) (*marketdata.Listing, error) {
// 			return &marketdata.Listing{ID: uuid.New(), Symbol: "AAA"}, nil
// 		},
// 	}
// 	svc := NewPositionBuilder(ps, ts, ls, &fakeMarketDataService{})

// 	if err := svc.BuildPositions(context.Background(), accID); err != nil {
// 		t.Fatalf("expected no error, got %v", err)
// 	}
// 	if len(ps.createManyArg) != 1 {
// 		t.Fatalf("expected one closed cycle, got %d", len(ps.createManyArg))
// 	}
// 	pos := ps.createManyArg[0]
// 	if pos.Quantity != 0 {
// 		t.Fatalf("expected oversell clamp to zero quantity, got %f", pos.Quantity)
// 	}
// 	if pos.CostBasis != 0 {
// 		t.Fatalf("expected oversell clamp to zero cost basis, got %s", pos.CostBasis)
// 	}
// 	if pos.CloseDate == nil {
// 		t.Fatalf("expected cycle to be closed after oversell")
// 	}
// 	if len(ts.updatePositionsArg) != 3 {
// 		t.Fatalf("expected 3 transactions in UpdatePositions, got %d", len(ts.updatePositionsArg))
// 	}
// 	if ts.updatePositionsArg[2].PositionID != nil {
// 		t.Fatalf("expected sell after close to remain unmapped")
// 	}
// }

// func TestBuildPositions_SymbolOnlyCyclePromotesToISIN(t *testing.T) {
// 	accID := uuid.New()
// 	isin := "NL0000000013"
// 	symbol := "PROMO"
// 	base := time.Now().UTC().AddDate(0, 0, -2)
// 	ps := &fakePositionStore{}
// 	ts := &fakeTransactionStore{
// 		getASCFn: func(ctx context.Context, gotAccID uuid.UUID) ([]Transaction, error) {
// 			return []Transaction{
// 				mustTx(t, TxBuy, base, nil, &symbol, 1, 10, 0),
// 				mustTx(t, TxBuy, base.Add(time.Hour), &isin, &symbol, 1, 11, 0),
// 			}, nil
// 		},
// 	}
// 	ls := &fakeListingStore{
// 		fetchBySymbolOrISINFn: func(ctx context.Context, val string) (*marketdata.Listing, error) {
// 			if val != isin {
// 				t.Fatalf("expected listing lookup by promoted isin key %s, got %s", isin, val)
// 			}
// 			return &marketdata.Listing{ID: uuid.New(), Symbol: symbol}, nil
// 		},
// 	}
// 	svc := NewPositionBuilder(ps, ts, ls, &fakeMarketDataService{})

// 	if err := svc.BuildPositions(context.Background(), accID); err != nil {
// 		t.Fatalf("expected no error, got %v", err)
// 	}
// 	if len(ps.createManyArg) != 1 {
// 		t.Fatalf("expected one promoted cycle, got %d", len(ps.createManyArg))
// 	}
// 	pos := ps.createManyArg[0]
// 	if pos.ISIN == nil || *pos.ISIN != isin {
// 		t.Fatalf("expected cycle to be promoted to isin %s, got %+v", isin, pos.ISIN)
// 	}
// 	if ts.updatePositionsArg[0].PositionID == nil || ts.updatePositionsArg[1].PositionID == nil {
// 		t.Fatalf("expected both transactions to be mapped")
// 	}
// 	if *ts.updatePositionsArg[0].PositionID != pos.ID || *ts.updatePositionsArg[1].PositionID != pos.ID {
// 		t.Fatalf("expected both transactions to map to the same promoted cycle")
// 	}
// }

// func TestBuildPositionSnapshots_GetByIDError(t *testing.T) {
// 	ps := &fakePositionStore{
// 		getByIDFn: func(ctx context.Context, id uuid.UUID) (*Position, error) {
// 			return nil, errors.New("position fetch failed")
// 		},
// 	}
// 	svc := NewPositionBuilder(ps, &fakeTransactionStore{}, &fakeListingStore{}, &fakeMarketDataService{})

// 	err := svc.BuildPositionSnapshots(context.Background(), uuid.New(), uuid.New())
// 	if err == nil || !strings.Contains(err.Error(), "position fetch failed") {
// 		t.Fatalf("expected GetByID error, got %v", err)
// 	}
// }

// func TestBuildPositionSnapshots_EarlyReturnWhenListingIDMissing(t *testing.T) {
// 	posID := uuid.New()
// 	ps := &fakePositionStore{
// 		getByIDFn: func(ctx context.Context, id uuid.UUID) (*Position, error) {
// 			return &Position{ID: posID, OpenDate: time.Now().UTC().AddDate(0, 0, -2)}, nil
// 		},
// 	}
// 	ls := &fakeListingStore{}
// 	ts := &fakeTransactionStore{}
// 	mds := &fakeMarketDataService{}
// 	svc := NewPositionBuilder(ps, ts, ls, mds)

// 	if err := svc.BuildPositionSnapshots(context.Background(), uuid.New(), posID); err != nil {
// 		t.Fatalf("expected no error, got %v", err)
// 	}
// 	if ls.fetchByIDCalls != 0 || ts.getByPositionIDCalls != 0 || mds.getDailiesCalls != 0 {
// 		t.Fatalf("expected early return before listing/tx/marketdata calls")
// 	}
// }

// func TestBuildPositionSnapshots_ListingFetchError(t *testing.T) {
// 	listingID := uuid.New()
// 	posID := uuid.New()
// 	ps := &fakePositionStore{
// 		getByIDFn: func(ctx context.Context, id uuid.UUID) (*Position, error) {
// 			return &Position{ID: posID, ListingID: &listingID, OpenDate: time.Now().UTC().AddDate(0, 0, -2)}, nil
// 		},
// 	}
// 	ls := &fakeListingStore{
// 		fetchByIDFn: func(ctx context.Context, id uuid.UUID) (*marketdata.Listing, error) {
// 			return nil, errors.New("listing fetch failed")
// 		},
// 	}
// 	svc := NewPositionBuilder(ps, &fakeTransactionStore{}, ls, &fakeMarketDataService{})

// 	err := svc.BuildPositionSnapshots(context.Background(), uuid.New(), posID)
// 	if err == nil || !strings.Contains(err.Error(), "listing fetch failed") {
// 		t.Fatalf("expected listing fetch error, got %v", err)
// 	}
// }

// func TestBuildPositionSnapshots_EarlyReturnWhenListingMissing(t *testing.T) {
// 	listingID := uuid.New()
// 	posID := uuid.New()
// 	ps := &fakePositionStore{
// 		getByIDFn: func(ctx context.Context, id uuid.UUID) (*Position, error) {
// 			return &Position{ID: posID, ListingID: &listingID, OpenDate: time.Now().UTC().AddDate(0, 0, -2)}, nil
// 		},
// 	}
// 	ls := &fakeListingStore{
// 		fetchByIDFn: func(ctx context.Context, id uuid.UUID) (*marketdata.Listing, error) {
// 			return nil, nil
// 		},
// 	}
// 	ts := &fakeTransactionStore{}
// 	mds := &fakeMarketDataService{}
// 	svc := NewPositionBuilder(ps, ts, ls, mds)

// 	if err := svc.BuildPositionSnapshots(context.Background(), uuid.New(), posID); err != nil {
// 		t.Fatalf("expected no error, got %v", err)
// 	}
// 	if ts.getByPositionIDCalls != 0 || mds.getDailiesCalls != 0 {
// 		t.Fatalf("expected early return before tx/marketdata calls")
// 	}
// }

// func TestBuildPositionSnapshots_EarlyReturnWhenListingSymbolEmpty(t *testing.T) {
// 	listingID := uuid.New()
// 	posID := uuid.New()
// 	ps := &fakePositionStore{
// 		getByIDFn: func(ctx context.Context, id uuid.UUID) (*Position, error) {
// 			return &Position{ID: posID, ListingID: &listingID, OpenDate: time.Now().UTC().AddDate(0, 0, -2)}, nil
// 		},
// 	}
// 	ls := &fakeListingStore{
// 		fetchByIDFn: func(ctx context.Context, id uuid.UUID) (*marketdata.Listing, error) {
// 			return &marketdata.Listing{ID: listingID, Symbol: ""}, nil
// 		},
// 	}
// 	ts := &fakeTransactionStore{}
// 	mds := &fakeMarketDataService{}
// 	svc := NewPositionBuilder(ps, ts, ls, mds)

// 	if err := svc.BuildPositionSnapshots(context.Background(), uuid.New(), posID); err != nil {
// 		t.Fatalf("expected no error, got %v", err)
// 	}
// 	if ts.getByPositionIDCalls != 0 || mds.getDailiesCalls != 0 {
// 		t.Fatalf("expected early return before tx/marketdata calls")
// 	}
// }

// func TestBuildPositionSnapshots_GetLastSnapshotError(t *testing.T) {
// 	listingID := uuid.New()
// 	posID := uuid.New()
// 	ps := &fakePositionStore{
// 		getByIDFn: func(ctx context.Context, id uuid.UUID) (*Position, error) {
// 			return &Position{ID: posID, ListingID: &listingID, OpenDate: time.Now().UTC().AddDate(0, 0, -2)}, nil
// 		},
// 		getLastSnapshotFn: func(ctx context.Context, positionID uuid.UUID) (*PositionSnapshot, error) {
// 			return nil, errors.New("last snapshot failed")
// 		},
// 	}
// 	ls := &fakeListingStore{
// 		fetchByIDFn: func(ctx context.Context, id uuid.UUID) (*marketdata.Listing, error) {
// 			return &marketdata.Listing{ID: listingID, Symbol: "AAA"}, nil
// 		},
// 	}
// 	svc := NewPositionBuilder(ps, &fakeTransactionStore{}, ls, &fakeMarketDataService{})

// 	err := svc.BuildPositionSnapshots(context.Background(), uuid.New(), posID)
// 	if err == nil || !strings.Contains(err.Error(), "last snapshot failed") {
// 		t.Fatalf("expected GetLastSnapshot error, got %v", err)
// 	}
// }

// func TestBuildPositionSnapshots_GetByPositionIDError(t *testing.T) {
// 	listingID := uuid.New()
// 	posID := uuid.New()
// 	ps := &fakePositionStore{
// 		getByIDFn: func(ctx context.Context, id uuid.UUID) (*Position, error) {
// 			return &Position{ID: posID, ListingID: &listingID, OpenDate: time.Now().UTC().AddDate(0, 0, -2)}, nil
// 		},
// 	}
// 	ts := &fakeTransactionStore{
// 		getByPositionIDFn: func(ctx context.Context, positionID uuid.UUID, from *time.Time) ([]Transaction, error) {
// 			return nil, errors.New("transactions failed")
// 		},
// 	}
// 	ls := &fakeListingStore{
// 		fetchByIDFn: func(ctx context.Context, id uuid.UUID) (*marketdata.Listing, error) {
// 			return &marketdata.Listing{ID: listingID, Symbol: "AAA"}, nil
// 		},
// 	}
// 	svc := NewPositionBuilder(ps, ts, ls, &fakeMarketDataService{})

// 	err := svc.BuildPositionSnapshots(context.Background(), uuid.New(), posID)
// 	if err == nil || !strings.Contains(err.Error(), "transactions failed") {
// 		t.Fatalf("expected GetByPositionID error, got %v", err)
// 	}
// }

// func TestBuildPositionSnapshots_GetDailiesError(t *testing.T) {
// 	listingID := uuid.New()
// 	posID := uuid.New()
// 	ps := &fakePositionStore{
// 		getByIDFn: func(ctx context.Context, id uuid.UUID) (*Position, error) {
// 			return &Position{ID: posID, ListingID: &listingID, OpenDate: time.Now().UTC().AddDate(0, 0, -2)}, nil
// 		},
// 	}
// 	ts := &fakeTransactionStore{
// 		getByPositionIDFn: func(ctx context.Context, positionID uuid.UUID, from *time.Time) ([]Transaction, error) {
// 			return []Transaction{}, nil
// 		},
// 	}
// 	ls := &fakeListingStore{
// 		fetchByIDFn: func(ctx context.Context, id uuid.UUID) (*marketdata.Listing, error) {
// 			return &marketdata.Listing{ID: listingID, Symbol: "AAA"}, nil
// 		},
// 	}
// 	mds := &fakeMarketDataService{
// 		getDailiesFn: func(ctx context.Context, symbol string, from, to *time.Time, limit, offset int) (*marketdata.DailyResponse, error) {
// 			return nil, errors.New("dailies failed")
// 		},
// 	}
// 	svc := NewPositionBuilder(ps, ts, ls, mds)

// 	err := svc.BuildPositionSnapshots(context.Background(), uuid.New(), posID)
// 	if err == nil || !strings.Contains(err.Error(), "dailies failed") {
// 		t.Fatalf("expected GetDailies error, got %v", err)
// 	}
// }

// func TestBuildPositionSnapshots_CreateSnapshotError(t *testing.T) {
// 	today := date.StartOfDayUTC(time.Now())
// 	listingID := uuid.New()
// 	posID := uuid.New()
// 	ps := &fakePositionStore{
// 		getByIDFn: func(ctx context.Context, id uuid.UUID) (*Position, error) {
// 			return &Position{ID: posID, ListingID: &listingID, OpenDate: today.AddDate(0, 0, -1)}, nil
// 		},
// 		createSnapshotFn: func(ctx context.Context, snap *PositionSnapshot) error {
// 			return errors.New("insert failed")
// 		},
// 	}
// 	ts := &fakeTransactionStore{
// 		getByPositionIDFn: func(ctx context.Context, positionID uuid.UUID, from *time.Time) ([]Transaction, error) {
// 			return []Transaction{}, nil
// 		},
// 	}
// 	ls := &fakeListingStore{
// 		fetchByIDFn: func(ctx context.Context, id uuid.UUID) (*marketdata.Listing, error) {
// 			return &marketdata.Listing{ID: listingID, Symbol: "AAA"}, nil
// 		},
// 	}
// 	mds := &fakeMarketDataService{
// 		getDailiesFn: func(ctx context.Context, symbol string, from, to *time.Time, limit, offset int) (*marketdata.DailyResponse, error) {
// 			return &marketdata.DailyResponse{}, nil
// 		},
// 	}
// 	svc := NewPositionBuilder(ps, ts, ls, mds)

// 	err := svc.BuildPositionSnapshots(context.Background(), uuid.New(), posID)
// 	if err == nil || !strings.Contains(err.Error(), "insert failed") {
// 		t.Fatalf("expected CreateSnapshot error, got %v", err)
// 	}
// }

// func TestBuildPositionSnapshots_NoSnapshotsWhenStartDateIsToday(t *testing.T) {
// 	today := date.StartOfDayUTC(time.Now())
// 	listingID := uuid.New()
// 	posID := uuid.New()
// 	ps := &fakePositionStore{
// 		getByIDFn: func(ctx context.Context, id uuid.UUID) (*Position, error) {
// 			return &Position{ID: posID, ListingID: &listingID, OpenDate: today}, nil
// 		},
// 	}
// 	ts := &fakeTransactionStore{
// 		getByPositionIDFn: func(ctx context.Context, positionID uuid.UUID, from *time.Time) ([]Transaction, error) {
// 			return []Transaction{}, nil
// 		},
// 	}
// 	ls := &fakeListingStore{
// 		fetchByIDFn: func(ctx context.Context, id uuid.UUID) (*marketdata.Listing, error) {
// 			return &marketdata.Listing{ID: listingID, Symbol: "AAA"}, nil
// 		},
// 	}
// 	mds := &fakeMarketDataService{
// 		getDailiesFn: func(ctx context.Context, symbol string, from, to *time.Time, limit, offset int) (*marketdata.DailyResponse, error) {
// 			return &marketdata.DailyResponse{}, nil
// 		},
// 	}
// 	svc := NewPositionBuilder(ps, ts, ls, mds)

// 	if err := svc.BuildPositionSnapshots(context.Background(), uuid.New(), posID); err != nil {
// 		t.Fatalf("expected no error, got %v", err)
// 	}
// 	if ps.createSnapshotCalls != 0 {
// 		t.Fatalf("expected no snapshots for start date today, got %d", ps.createSnapshotCalls)
// 	}
// }

// func TestBuildPositionSnapshots_BuildsOneSnapshotPerDayFromOpenDate(t *testing.T) {
// 	today := date.StartOfDayUTC(time.Now())
// 	open := today.AddDate(0, 0, -3)
// 	listingID := uuid.New()
// 	posID := uuid.New()
// 	ps := &fakePositionStore{
// 		getByIDFn: func(ctx context.Context, id uuid.UUID) (*Position, error) {
// 			return &Position{ID: posID, ListingID: &listingID, OpenDate: open}, nil
// 		},
// 	}
// 	ts := &fakeTransactionStore{
// 		getByPositionIDFn: func(ctx context.Context, positionID uuid.UUID, from *time.Time) ([]Transaction, error) {
// 			return []Transaction{}, nil
// 		},
// 	}
// 	ls := &fakeListingStore{
// 		fetchByIDFn: func(ctx context.Context, id uuid.UUID) (*marketdata.Listing, error) {
// 			return &marketdata.Listing{ID: listingID, Symbol: "AAA"}, nil
// 		},
// 	}
// 	mds := &fakeMarketDataService{
// 		getDailiesFn: func(ctx context.Context, symbol string, from, to *time.Time, limit, offset int) (*marketdata.DailyResponse, error) {
// 			return &marketdata.DailyResponse{Data: []marketdata.EOD{}}, nil
// 		},
// 	}
// 	svc := NewPositionBuilder(ps, ts, ls, mds)

// 	if err := svc.BuildPositionSnapshots(context.Background(), uuid.New(), posID); err != nil {
// 		t.Fatalf("expected no error, got %v", err)
// 	}
// 	if ps.createSnapshotCalls != 3 {
// 		t.Fatalf("expected 3 snapshots, got %d", ps.createSnapshotCalls)
// 	}
// 	if !ps.createdSnapshots[0].OccurredAt.Equal(open) {
// 		t.Fatalf("expected first snapshot at %s, got %s", open, ps.createdSnapshots[0].OccurredAt)
// 	}
// 	if !ps.createdSnapshots[2].OccurredAt.Equal(today.AddDate(0, 0, -1)) {
// 		t.Fatalf("expected last snapshot at %s, got %s", today.AddDate(0, 0, -1), ps.createdSnapshots[2].OccurredAt)
// 	}
// }

// func TestBuildPositionSnapshots_ClosedCycleStopsAtCloseDate(t *testing.T) {
// 	today := date.StartOfDayUTC(time.Now())
// 	open := today.AddDate(0, 0, -4)
// 	closeDay := today.AddDate(0, 0, -2)
// 	listingID := uuid.New()
// 	posID := uuid.New()
// 	ps := &fakePositionStore{
// 		getByIDFn: func(ctx context.Context, id uuid.UUID) (*Position, error) {
// 			return &Position{
// 				ID:        posID,
// 				ListingID: &listingID,
// 				OpenDate:  open,
// 				CloseDate: &closeDay,
// 			}, nil
// 		},
// 	}
// 	ts := &fakeTransactionStore{
// 		getByPositionIDFn: func(ctx context.Context, positionID uuid.UUID, from *time.Time) ([]Transaction, error) {
// 			return []Transaction{}, nil
// 		},
// 	}
// 	ls := &fakeListingStore{
// 		fetchByIDFn: func(ctx context.Context, id uuid.UUID) (*marketdata.Listing, error) {
// 			return &marketdata.Listing{ID: listingID, Symbol: "AAA"}, nil
// 		},
// 	}
// 	mds := &fakeMarketDataService{
// 		getDailiesFn: func(ctx context.Context, symbol string, from, to *time.Time, limit, offset int) (*marketdata.DailyResponse, error) {
// 			return &marketdata.DailyResponse{Data: []marketdata.EOD{}}, nil
// 		},
// 	}
// 	svc := NewPositionBuilder(ps, ts, ls, mds)

// 	if err := svc.BuildPositionSnapshots(context.Background(), uuid.New(), posID); err != nil {
// 		t.Fatalf("expected no error, got %v", err)
// 	}
// 	if ps.createSnapshotCalls != 3 {
// 		t.Fatalf("expected snapshots from open through close day (3), got %d", ps.createSnapshotCalls)
// 	}
// 	if !ps.createdSnapshots[0].OccurredAt.Equal(open) {
// 		t.Fatalf("expected first snapshot at open day %s, got %s", open, ps.createdSnapshots[0].OccurredAt)
// 	}
// 	if !ps.createdSnapshots[2].OccurredAt.Equal(closeDay) {
// 		t.Fatalf("expected last snapshot at close day %s, got %s", closeDay, ps.createdSnapshots[2].OccurredAt)
// 	}
// }

// func TestBuildPositionSnapshots_StartsFromLastSnapshotPlusOneDay(t *testing.T) {
// 	today := date.StartOfDayUTC(time.Now())
// 	listingID := uuid.New()
// 	posID := uuid.New()
// 	lastDay := today.AddDate(0, 0, -2)
// 	ps := &fakePositionStore{
// 		getByIDFn: func(ctx context.Context, id uuid.UUID) (*Position, error) {
// 			return &Position{ID: posID, ListingID: &listingID, OpenDate: today.AddDate(0, 0, -10)}, nil
// 		},
// 		getLastSnapshotFn: func(ctx context.Context, positionID uuid.UUID) (*PositionSnapshot, error) {
// 			return &PositionSnapshot{
// 				ID:          uuid.New(),
// 				OccurredAt:  lastDay,
// 				Quantity:    1,
// 				CostBasis:   mustPrice(t, 10),
// 				MarketValue: mustPrice(t, 11),
// 			}, nil
// 		},
// 	}
// 	ts := &fakeTransactionStore{
// 		getByPositionIDFn: func(ctx context.Context, positionID uuid.UUID, from *time.Time) ([]Transaction, error) {
// 			return []Transaction{}, nil
// 		},
// 	}
// 	ls := &fakeListingStore{
// 		fetchByIDFn: func(ctx context.Context, id uuid.UUID) (*marketdata.Listing, error) {
// 			return &marketdata.Listing{ID: listingID, Symbol: "AAA"}, nil
// 		},
// 	}
// 	mds := &fakeMarketDataService{
// 		getDailiesFn: func(ctx context.Context, symbol string, from, to *time.Time, limit, offset int) (*marketdata.DailyResponse, error) {
// 			return &marketdata.DailyResponse{Data: []marketdata.EOD{}}, nil
// 		},
// 	}
// 	svc := NewPositionBuilder(ps, ts, ls, mds)

// 	if err := svc.BuildPositionSnapshots(context.Background(), uuid.New(), posID); err != nil {
// 		t.Fatalf("expected no error, got %v", err)
// 	}
// 	wantStart := today.AddDate(0, 0, -1)
// 	if ts.getByPositionIDArg.from == nil || !ts.getByPositionIDArg.from.Equal(wantStart) {
// 		t.Fatalf("expected GetByPositionID from=%s, got %v", wantStart, ts.getByPositionIDArg.from)
// 	}
// 	if ps.createSnapshotCalls != 1 {
// 		t.Fatalf("expected exactly one snapshot, got %d", ps.createSnapshotCalls)
// 	}
// 	if !ps.createdSnapshots[0].OccurredAt.Equal(wantStart) {
// 		t.Fatalf("expected snapshot at %s, got %s", wantStart, ps.createdSnapshots[0].OccurredAt)
// 	}
// }

// func TestBuildPositionSnapshots_CarriesForwardLastKnownPriceWhenNoDaily(t *testing.T) {
// 	today := date.StartOfDayUTC(time.Now())
// 	open := today.AddDate(0, 0, -2)
// 	listingID := uuid.New()
// 	posID := uuid.New()
// 	isin := "NL0000000111"
// 	ps := &fakePositionStore{
// 		getByIDFn: func(ctx context.Context, id uuid.UUID) (*Position, error) {
// 			return &Position{ID: posID, ListingID: &listingID, OpenDate: open}, nil
// 		},
// 	}
// 	ts := &fakeTransactionStore{
// 		getByPositionIDFn: func(ctx context.Context, positionID uuid.UUID, from *time.Time) ([]Transaction, error) {
// 			return []Transaction{
// 				mustTx(t, TxBuy, open.Add(3*time.Hour), &isin, nil, 1, 10, 0),
// 			}, nil
// 		},
// 	}
// 	ls := &fakeListingStore{
// 		fetchByIDFn: func(ctx context.Context, id uuid.UUID) (*marketdata.Listing, error) {
// 			return &marketdata.Listing{ID: listingID, Symbol: "AAA"}, nil
// 		},
// 	}
// 	mds := &fakeMarketDataService{
// 		getDailiesFn: func(ctx context.Context, symbol string, from, to *time.Time, limit, offset int) (*marketdata.DailyResponse, error) {
// 			return &marketdata.DailyResponse{Data: []marketdata.EOD{}}, nil
// 		},
// 	}
// 	svc := NewPositionBuilder(ps, ts, ls, mds)

// 	if err := svc.BuildPositionSnapshots(context.Background(), uuid.New(), posID); err != nil {
// 		t.Fatalf("expected no error, got %v", err)
// 	}
// 	if ps.createSnapshotCalls != 2 {
// 		t.Fatalf("expected 2 snapshots, got %d", ps.createSnapshotCalls)
// 	}
// 	if ps.createdSnapshots[0].UnitPrice != mustPrice(t, 10) {
// 		t.Fatalf("expected day1 unit price 10.000000, got %s", ps.createdSnapshots[0].UnitPrice)
// 	}
// 	if ps.createdSnapshots[1].UnitPrice != mustPrice(t, 10) {
// 		t.Fatalf("expected day2 carried unit price 10.000000, got %s", ps.createdSnapshots[1].UnitPrice)
// 	}
// }

// func TestBuildPositionSnapshots_CarriesForwardLastSnapshotPriceWhenNoDailyOrTx_Red(t *testing.T) {
// 	today := date.StartOfDayUTC(time.Now())
// 	listingID := uuid.New()
// 	posID := uuid.New()
// 	lastDay := today.AddDate(0, 0, -2)
// 	ps := &fakePositionStore{
// 		getByIDFn: func(ctx context.Context, id uuid.UUID) (*Position, error) {
// 			return &Position{ID: posID, ListingID: &listingID, OpenDate: today.AddDate(0, 0, -10)}, nil
// 		},
// 		getLastSnapshotFn: func(ctx context.Context, positionID uuid.UUID) (*PositionSnapshot, error) {
// 			return &PositionSnapshot{
// 				ID:          uuid.New(),
// 				OccurredAt:  lastDay,
// 				Quantity:    2,
// 				UnitPrice:   mustPrice(t, 42),
// 				CostBasis:   mustPrice(t, 80),
// 				MarketValue: mustPrice(t, 84),
// 			}, nil
// 		},
// 	}
// 	ts := &fakeTransactionStore{
// 		getByPositionIDFn: func(ctx context.Context, positionID uuid.UUID, from *time.Time) ([]Transaction, error) {
// 			return []Transaction{}, nil
// 		},
// 	}
// 	ls := &fakeListingStore{
// 		fetchByIDFn: func(ctx context.Context, id uuid.UUID) (*marketdata.Listing, error) {
// 			return &marketdata.Listing{ID: listingID, Symbol: "AAA"}, nil
// 		},
// 	}
// 	mds := &fakeMarketDataService{
// 		getDailiesFn: func(ctx context.Context, symbol string, from, to *time.Time, limit, offset int) (*marketdata.DailyResponse, error) {
// 			return &marketdata.DailyResponse{Data: []marketdata.EOD{}}, nil
// 		},
// 	}
// 	svc := NewPositionBuilder(ps, ts, ls, mds)

// 	if err := svc.BuildPositionSnapshots(context.Background(), uuid.New(), posID); err != nil {
// 		t.Fatalf("expected no error, got %v", err)
// 	}
// 	if ps.createSnapshotCalls != 1 {
// 		t.Fatalf("expected 1 snapshot after last snapshot date, got %d", ps.createSnapshotCalls)
// 	}
// 	if ps.createdSnapshots[0].UnitPrice != mustPrice(t, 42) {
// 		t.Fatalf("expected carried unit price 42.000000 from last snapshot, got %s", ps.createdSnapshots[0].UnitPrice)
// 	}
// }

// func TestBuildPositionSnapshots_UsesLatestBuySellPriceWhenNoDaily(t *testing.T) {
// 	today := date.StartOfDayUTC(time.Now())
// 	open := today.AddDate(0, 0, -1)
// 	listingID := uuid.New()
// 	posID := uuid.New()
// 	isin := "NL0000000112"
// 	ps := &fakePositionStore{
// 		getByIDFn: func(ctx context.Context, id uuid.UUID) (*Position, error) {
// 			return &Position{ID: posID, ListingID: &listingID, OpenDate: open}, nil
// 		},
// 	}
// 	ts := &fakeTransactionStore{
// 		getByPositionIDFn: func(ctx context.Context, positionID uuid.UUID, from *time.Time) ([]Transaction, error) {
// 			return []Transaction{
// 				mustTx(t, TxBuy, open.Add(2*time.Hour), &isin, nil, 1, 10, 0),
// 				mustTx(t, TxSell, open.Add(5*time.Hour), &isin, nil, 1, 12, 0),
// 			}, nil
// 		},
// 	}
// 	ls := &fakeListingStore{
// 		fetchByIDFn: func(ctx context.Context, id uuid.UUID) (*marketdata.Listing, error) {
// 			return &marketdata.Listing{ID: listingID, Symbol: "AAA"}, nil
// 		},
// 	}
// 	mds := &fakeMarketDataService{
// 		getDailiesFn: func(ctx context.Context, symbol string, from, to *time.Time, limit, offset int) (*marketdata.DailyResponse, error) {
// 			return &marketdata.DailyResponse{Data: []marketdata.EOD{}}, nil
// 		},
// 	}
// 	svc := NewPositionBuilder(ps, ts, ls, mds)

// 	if err := svc.BuildPositionSnapshots(context.Background(), uuid.New(), posID); err != nil {
// 		t.Fatalf("expected no error, got %v", err)
// 	}
// 	if ps.createSnapshotCalls != 1 {
// 		t.Fatalf("expected 1 snapshot, got %d", ps.createSnapshotCalls)
// 	}
// 	if ps.createdSnapshots[0].UnitPrice != mustPrice(t, 12) {
// 		t.Fatalf("expected latest buy/sell unit price 12.000000, got %s", ps.createdSnapshots[0].UnitPrice)
// 	}
// }

// func TestBuildPositionSnapshots_ZeroQuantityBuyOrSellReturnsError(t *testing.T) {
// 	today := date.StartOfDayUTC(time.Now())
// 	open := today.AddDate(0, 0, -1)
// 	listingID := uuid.New()
// 	posID := uuid.New()
// 	isin := "NL0000000113"
// 	ps := &fakePositionStore{
// 		getByIDFn: func(ctx context.Context, id uuid.UUID) (*Position, error) {
// 			return &Position{ID: posID, ListingID: &listingID, OpenDate: open}, nil
// 		},
// 	}
// 	ts := &fakeTransactionStore{
// 		getByPositionIDFn: func(ctx context.Context, positionID uuid.UUID, from *time.Time) ([]Transaction, error) {
// 			return []Transaction{
// 				mustTx(t, TxBuy, open.Add(2*time.Hour), &isin, nil, 0, 10, 0),
// 			}, nil
// 		},
// 	}
// 	ls := &fakeListingStore{
// 		fetchByIDFn: func(ctx context.Context, id uuid.UUID) (*marketdata.Listing, error) {
// 			return &marketdata.Listing{ID: listingID, Symbol: "AAA"}, nil
// 		},
// 	}
// 	mds := &fakeMarketDataService{
// 		getDailiesFn: func(ctx context.Context, symbol string, from, to *time.Time, limit, offset int) (*marketdata.DailyResponse, error) {
// 			return &marketdata.DailyResponse{Data: []marketdata.EOD{}}, nil
// 		},
// 	}
// 	svc := NewPositionBuilder(ps, ts, ls, mds)

// 	err := svc.BuildPositionSnapshots(context.Background(), uuid.New(), posID)
// 	if !errors.Is(err, money.ErrInvalidPrice) {
// 		t.Fatalf("expected ErrInvalidPrice for zero-quantity buy/sell fallback calculation, got %v", err)
// 	}
// }

// func TestGetPositionFromMap_ReturnsExistingPosition(t *testing.T) {
// 	accID := uuid.New()
// 	isin := "NL0000000991"
// 	existing, err := NewPosition(accID, &isin, nil, nil, time.Now().UTC())
// 	if err != nil {
// 		t.Fatalf("expected no error creating existing position, got %v", err)
// 	}
// 	positions := map[string]*Position{
// 		isin: existing,
// 	}
// 	tx := mustTx(t, TxBuy, time.Now().UTC(), &isin, nil, 1, 10, 0)

// 	got, err := getPositionFromMap(accID, &positions, tx)
// 	if err != nil {
// 		t.Fatalf("expected no error, got %v", err)
// 	}
// 	if got != existing {
// 		t.Fatalf("expected existing position pointer to be returned")
// 	}
// }

// func TestGetPositionFromMap_CreatesNewPositionWhenMissing(t *testing.T) {
// 	accID := uuid.New()
// 	isin := "NL0000000992"
// 	positions := map[string]*Position{}
// 	tx := mustTx(t, TxBuy, time.Now().UTC(), &isin, nil, 1, 10, 0)

// 	got, err := getPositionFromMap(accID, &positions, tx)
// 	if err != nil {
// 		t.Fatalf("expected no error, got %v", err)
// 	}
// 	if got == nil {
// 		t.Fatalf("expected non-nil new position")
// 	}
// 	if got.AccountID != accID {
// 		t.Fatalf("expected account id %s, got %s", accID, got.AccountID)
// 	}
// 	if _, ok := positions[isin]; !ok {
// 		t.Fatalf("expected position map to contain key %s", isin)
// 	}
// }

// func TestGetPositionFromMap_ErrorsWhenTransactionMissingID(t *testing.T) {
// 	accID := uuid.New()
// 	positions := map[string]*Position{}
// 	tx := mustTx(t, TxBuy, time.Now().UTC(), nil, nil, 1, 10, 0)

//		_, err := getPositionFromMap(accID, &positions, tx)
//		if !errors.Is(err, ErrTransactionISINAndSymbolMissing) {
//			t.Fatalf("expected ErrTransactionISINAndSymbolMissing, got %v", err)
//		}
//	}
package portfolio
