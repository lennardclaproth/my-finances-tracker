package main

import (
	"context"
	"encoding/json"
	"io"
	"log/slog"
	"math"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/lennardclaproth/my-finances-tracker/api"
	"github.com/lennardclaproth/my-finances-tracker/internal/account"
	"github.com/lennardclaproth/my-finances-tracker/internal/config"
	"github.com/lennardclaproth/my-finances-tracker/internal/logging"
	"github.com/lennardclaproth/my-finances-tracker/internal/marketdata"
	"github.com/lennardclaproth/my-finances-tracker/internal/money"
	"github.com/lennardclaproth/my-finances-tracker/internal/portfolio"
)

func TestIntegration_GetPortfolioSnapshotsEndpoint_ReturnsChronologicalSeries(t *testing.T) {
	logger := logging.NewSlogLogger(slog.LevelDebug)
	db, _ := setupIntegrationDB(t, logger)
	cfg := &config.Config{
		DiskStorage: config.DiskStorage{
			BasePath: t.TempDir(),
		},
	}
	deps := newAppDependencies(logger, db, cfg)
	router := setupRouterWithDeps(logger, deps, nil, nil, nil)
	mux := http.NewServeMux()
	router.Register(mux)
	server := httptest.NewServer(mux)

	t.Cleanup(func() {
		server.Close()
		_ = db.Close()
	})

	ctx := context.Background()
	accID, _ := seedPortfolioReadFixture(t, ctx, deps)

	req, err := http.NewRequest(http.MethodGet, server.URL+"/portfolio/snapshots?account_id="+accID.String(), nil)
	if err != nil {
		t.Fatalf("failed creating request: %v", err)
	}
	res, err := http.DefaultClient.Do(req)
	if err != nil {
		t.Fatalf("GET /portfolio/snapshots failed: %v", err)
	}
	defer res.Body.Close()
	if res.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(res.Body)
		t.Fatalf("expected status 200, got %d body=%s", res.StatusCode, string(body))
	}

	var payload []api.PortfolioSnapshotPointResponse
	if err := json.NewDecoder(res.Body).Decode(&payload); err != nil {
		t.Fatalf("failed decoding response: %v", err)
	}
	if len(payload) != 2 {
		t.Fatalf("expected 2 snapshot points, got %d", len(payload))
	}
	if !payload[0].OccurredAt.Before(payload[1].OccurredAt) {
		t.Fatalf("expected chronological order, got %s then %s", payload[0].OccurredAt, payload[1].OccurredAt)
	}
	if payload[0].ValueIndex != 100 {
		t.Fatalf("expected first value_index 100, got %f", payload[0].ValueIndex)
	}
	if payload[1].ValueIndex != 104 {
		t.Fatalf("expected second value_index 104, got %f", payload[1].ValueIndex)
	}
	if math.Abs(payload[0].ReturnVsCostBasisPct-7.5) > 1e-9 {
		t.Fatalf("expected first return_vs_cost_basis_pct 7.5, got %f", payload[0].ReturnVsCostBasisPct)
	}
	if math.Abs(payload[1].ReturnVsCostBasisPct-14) > 1e-9 {
		t.Fatalf("expected second return_vs_cost_basis_pct 14, got %f", payload[1].ReturnVsCostBasisPct)
	}
}

func TestIntegration_GetPortfolioSnapshotsEndpoint_AppliesDateRangeFilter(t *testing.T) {
	logger := logging.NewSlogLogger(slog.LevelDebug)
	db, _ := setupIntegrationDB(t, logger)
	cfg := &config.Config{
		DiskStorage: config.DiskStorage{
			BasePath: t.TempDir(),
		},
	}
	deps := newAppDependencies(logger, db, cfg)
	router := setupRouterWithDeps(logger, deps, nil, nil, nil)
	mux := http.NewServeMux()
	router.Register(mux)
	server := httptest.NewServer(mux)

	t.Cleanup(func() {
		server.Close()
		_ = db.Close()
	})

	ctx := context.Background()
	accID, _ := seedPortfolioReadFixture(t, ctx, deps)

	req, err := http.NewRequest(http.MethodGet, server.URL+"/portfolio/snapshots?account_id="+accID.String()+"&from=2026-02-02&to=2026-02-02", nil)
	if err != nil {
		t.Fatalf("failed creating request: %v", err)
	}
	res, err := http.DefaultClient.Do(req)
	if err != nil {
		t.Fatalf("GET /portfolio/snapshots with range failed: %v", err)
	}
	defer res.Body.Close()
	if res.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(res.Body)
		t.Fatalf("expected status 200, got %d body=%s", res.StatusCode, string(body))
	}

	var payload []api.PortfolioSnapshotPointResponse
	if err := json.NewDecoder(res.Body).Decode(&payload); err != nil {
		t.Fatalf("failed decoding response: %v", err)
	}
	if len(payload) != 1 {
		t.Fatalf("expected exactly one snapshot point in range, got %d", len(payload))
	}
	if payload[0].OccurredAt.UTC().Format("2006-01-02") != "2026-02-02" {
		t.Fatalf("expected snapshot date 2026-02-02, got %s", payload[0].OccurredAt.UTC().Format("2006-01-02"))
	}
}

func TestIntegration_GetPortfolioPositionsEndpoint_RespectsIncludeClosed(t *testing.T) {
	logger := logging.NewSlogLogger(slog.LevelDebug)
	db, _ := setupIntegrationDB(t, logger)
	cfg := &config.Config{
		DiskStorage: config.DiskStorage{
			BasePath: t.TempDir(),
		},
	}
	deps := newAppDependencies(logger, db, cfg)
	router := setupRouterWithDeps(logger, deps, nil, nil, nil)
	mux := http.NewServeMux()
	router.Register(mux)
	server := httptest.NewServer(mux)

	t.Cleanup(func() {
		server.Close()
		_ = db.Close()
	})

	ctx := context.Background()
	accID, latestMarketValue := seedPortfolioReadFixture(t, ctx, deps)

	reqOpenOnly, err := http.NewRequest(http.MethodGet, server.URL+"/portfolio/positions?account_id="+accID.String(), nil)
	if err != nil {
		t.Fatalf("failed creating open-only request: %v", err)
	}
	resOpenOnly, err := http.DefaultClient.Do(reqOpenOnly)
	if err != nil {
		t.Fatalf("GET /portfolio/positions failed: %v", err)
	}
	defer resOpenOnly.Body.Close()
	if resOpenOnly.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resOpenOnly.Body)
		t.Fatalf("expected status 200, got %d body=%s", resOpenOnly.StatusCode, string(body))
	}

	var openOnlyPayload api.PortfolioPositionsResponse
	if err := json.NewDecoder(resOpenOnly.Body).Decode(&openOnlyPayload); err != nil {
		t.Fatalf("failed decoding open-only response: %v", err)
	}
	if openOnlyPayload.IncludeClosed {
		t.Fatalf("expected include_closed=false by default")
	}
	if len(openOnlyPayload.Data) != 1 {
		t.Fatalf("expected one open position, got %d", len(openOnlyPayload.Data))
	}
	if openOnlyPayload.Data[0].IsClosed {
		t.Fatalf("expected open position only")
	}
	if openOnlyPayload.Data[0].MarketValue == nil || *openOnlyPayload.Data[0].MarketValue != latestMarketValue {
		t.Fatalf("expected latest market value %d, got %+v", latestMarketValue, openOnlyPayload.Data[0].MarketValue)
	}

	reqAll, err := http.NewRequest(http.MethodGet, server.URL+"/portfolio/positions?account_id="+accID.String()+"&include_closed=true", nil)
	if err != nil {
		t.Fatalf("failed creating include_closed request: %v", err)
	}
	resAll, err := http.DefaultClient.Do(reqAll)
	if err != nil {
		t.Fatalf("GET /portfolio/positions include_closed failed: %v", err)
	}
	defer resAll.Body.Close()
	if resAll.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resAll.Body)
		t.Fatalf("expected status 200, got %d body=%s", resAll.StatusCode, string(body))
	}

	var allPayload api.PortfolioPositionsResponse
	if err := json.NewDecoder(resAll.Body).Decode(&allPayload); err != nil {
		t.Fatalf("failed decoding include_closed response: %v", err)
	}
	if !allPayload.IncludeClosed {
		t.Fatalf("expected include_closed=true in response")
	}
	if len(allPayload.Data) != 2 {
		t.Fatalf("expected two positions (open + closed), got %d", len(allPayload.Data))
	}
}

func seedPortfolioReadFixture(
	t *testing.T,
	ctx context.Context,
	deps *appDependencies,
) (uuid.UUID, int64) {
	t.Helper()

	acc, err := account.NewAccount("portfolio-read-"+uuid.NewString(), nil)
	if err != nil {
		t.Fatalf("failed creating account: %v", err)
	}
	if err := deps.accountStore.Create(ctx, acc); err != nil {
		t.Fatalf("failed persisting account: %v", err)
	}
	if err := deps.portfolioAccountStore.Create(ctx, portfolio.NewAccount(acc.ID)); err != nil {
		t.Fatalf("failed persisting portfolio account projection: %v", err)
	}

	listing, err := marketdata.NewListing(
		"READ.AS",
		"Read Listing",
		marketdata.SourceAlphaVantage,
		marketdata.ListingWithISIN("NLREAD000001"),
	)
	if err != nil {
		t.Fatalf("failed creating listing: %v", err)
	}
	if err := deps.listingStore.Create(ctx, listing); err != nil {
		t.Fatalf("failed persisting listing: %v", err)
	}

	symbol := listing.Symbol
	isin := "NLREAD000001"
	openPos, err := portfolio.NewPosition(
		acc.ID,
		&isin,
		&symbol,
		&listing.ID,
		time.Date(2026, 1, 1, 0, 0, 0, 0, time.UTC),
	)
	if err != nil {
		t.Fatalf("failed creating open position: %v", err)
	}
	openPos.Quantity = 2
	openPos.CostBasis = money.Price(20000)

	closedPos, err := portfolio.NewPosition(
		acc.ID,
		&isin,
		&symbol,
		&listing.ID,
		time.Date(2026, 1, 2, 0, 0, 0, 0, time.UTC),
	)
	if err != nil {
		t.Fatalf("failed creating closed position: %v", err)
	}
	closedPos.Quantity = 0
	closedAt := time.Date(2026, 1, 10, 0, 0, 0, 0, time.UTC)
	closedPos.CloseDate = &closedAt

	if err := deps.positionStore.CreateMany(ctx, []*portfolio.Position{openPos, closedPos}); err != nil {
		t.Fatalf("failed persisting positions: %v", err)
	}

	latestOpenSnapshotAt := time.Date(2026, 2, 2, 0, 0, 0, 0, time.UTC)
	latestOpenSnapshotMarketValue := int64(24000)
	openSnapshotOld := &portfolio.PositionSnapshot{
		ID:               uuid.New(),
		AccountID:        acc.ID,
		PositionID:       openPos.ID,
		Symbol:           symbol,
		Name:             ptrString("Open Position"),
		ListingID:        listing.ID,
		OccurredAt:       time.Date(2026, 2, 1, 0, 0, 0, 0, time.UTC),
		Quantity:         2,
		UnitPrice:        money.Price(11000),
		MarketValue:      money.Price(22000),
		CostBasis:        money.Price(20000),
		RealizedPnL:      money.Price(0),
		UnrealizedPnL:    money.Price(2000),
		UnrealizedPnLPct: 10,
	}
	openSnapshotLatest := &portfolio.PositionSnapshot{
		ID:               uuid.New(),
		AccountID:        acc.ID,
		PositionID:       openPos.ID,
		Symbol:           symbol,
		Name:             ptrString("Open Position"),
		ListingID:        listing.ID,
		OccurredAt:       latestOpenSnapshotAt,
		Quantity:         2,
		UnitPrice:        money.Price(12000),
		MarketValue:      money.Price(latestOpenSnapshotMarketValue),
		CostBasis:        money.Price(20000),
		RealizedPnL:      money.Price(0),
		UnrealizedPnL:    money.Price(4000),
		UnrealizedPnLPct: 20,
	}
	closedSnapshot := &portfolio.PositionSnapshot{
		ID:               uuid.New(),
		AccountID:        acc.ID,
		PositionID:       closedPos.ID,
		Symbol:           symbol,
		Name:             ptrString("Closed Position"),
		ListingID:        listing.ID,
		OccurredAt:       closedAt,
		Quantity:         0,
		UnitPrice:        money.Price(0),
		MarketValue:      money.Price(0),
		CostBasis:        money.Price(0),
		RealizedPnL:      money.Price(1000),
		UnrealizedPnL:    money.Price(0),
		UnrealizedPnLPct: 0,
	}
	if err := deps.positionStore.CreateSnapshot(ctx, openSnapshotOld); err != nil {
		t.Fatalf("failed persisting old open snapshot: %v", err)
	}
	if err := deps.positionStore.CreateSnapshot(ctx, openSnapshotLatest); err != nil {
		t.Fatalf("failed persisting latest open snapshot: %v", err)
	}
	if err := deps.positionStore.CreateSnapshot(ctx, closedSnapshot); err != nil {
		t.Fatalf("failed persisting closed snapshot: %v", err)
	}

	oldPortfolioSnapshot := &portfolio.PortfolioSnapshot{
		ID:                    uuid.New(),
		AccountID:             acc.ID,
		OccurredAt:            time.Date(2026, 2, 1, 0, 0, 0, 0, time.UTC),
		MarketValue:           money.Price(22000),
		CostBasis:             money.Price(20000),
		TotalPnL:              money.Price(1500),
		DailyDeltaPnLPct:      1.0,
		TimeWeightedReturnPct: 0,
	}
	newPortfolioSnapshot := &portfolio.PortfolioSnapshot{
		ID:                    uuid.New(),
		AccountID:             acc.ID,
		OccurredAt:            latestOpenSnapshotAt,
		MarketValue:           money.Price(24000),
		CostBasis:             money.Price(20000),
		TotalPnL:              money.Price(2800),
		DailyDeltaPnLPct:      2.0,
		TimeWeightedReturnPct: 4,
	}
	// Insert out of order to prove API ordering is by occurred_at.
	if err := deps.portfolioSnapshotStore.CreateSnapshot(ctx, newPortfolioSnapshot); err != nil {
		t.Fatalf("failed persisting newer portfolio snapshot: %v", err)
	}
	if err := deps.portfolioSnapshotStore.CreateSnapshot(ctx, oldPortfolioSnapshot); err != nil {
		t.Fatalf("failed persisting older portfolio snapshot: %v", err)
	}

	return acc.ID, latestOpenSnapshotMarketValue
}

func ptrString(v string) *string {
	return &v
}
