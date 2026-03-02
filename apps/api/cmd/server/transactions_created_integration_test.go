package main

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"log/slog"
	"net/http"
	"net/http/httptest"
	"path/filepath"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/lennardclaproth/my-finances-tracker/api"
	"github.com/lennardclaproth/my-finances-tracker/internal/account"
	"github.com/lennardclaproth/my-finances-tracker/internal/bus"
	"github.com/lennardclaproth/my-finances-tracker/internal/config"
	"github.com/lennardclaproth/my-finances-tracker/internal/importer"
	"github.com/lennardclaproth/my-finances-tracker/internal/logging"
	"github.com/lennardclaproth/my-finances-tracker/internal/marketdata"
	"github.com/lennardclaproth/my-finances-tracker/internal/portfolio"
	"github.com/lennardclaproth/my-finances-tracker/internal/storage"
	"github.com/lennardclaproth/my-finances-tracker/internal/vendor"
)

func TestIntegration_TransactionsCreatedEvent_BuildsPortfolio(t *testing.T) {
	logger := logging.NewSlogLogger(slog.LevelDebug)
	db, backend := setupIntegrationDB(t, logger)

	cfg := &config.Config{
		DiskStorage: config.DiskStorage{
			BasePath: t.TempDir(),
		},
	}
	deps := newAppDependencies(logger, db, cfg)
	b, err := setupBus(logger, deps)
	if err != nil {
		t.Fatalf("failed setting up bus: %v", err)
	}

	t.Cleanup(func() {
		_ = b.Close()
		_ = db.Close()
	})

	ctx := context.Background()
	acc, err := account.NewAccount("transactions-created-integration", nil)
	if err != nil {
		t.Fatalf("failed creating account seed: %v", err)
	}
	if err := deps.accountStore.Create(ctx, acc); err != nil {
		t.Fatalf("failed storing account seed: %v", err)
	}
	if err := deps.portfolioAccountStore.Create(ctx, portfolio.NewAccount(acc.ID)); err != nil {
		t.Fatalf("failed storing portfolio account projection: %v", err)
	}

	v, err := vendor.NewVendor(vendor.VendorDEGIRO, vendor.VendorTypeBrokerage)
	if err != nil {
		t.Fatalf("failed creating vendor seed: %v", err)
	}
	if err := deps.vendorStore.Create(ctx, v); err != nil {
		t.Fatalf("failed storing vendor seed: %v", err)
	}

	symbol := "ITEST.AS"
	isin := "NLITEST000001"
	listing, err := marketdata.NewListing(
		symbol,
		"Integration Test Listing",
		marketdata.SourceAlphaVantage,
		marketdata.ListingWithISIN(isin),
	)
	if err != nil {
		t.Fatalf("failed creating listing seed: %v", err)
	}
	if err := deps.listingStore.Create(ctx, listing); err != nil {
		t.Fatalf("failed storing listing seed: %v", err)
	}

	txDate := time.Now().UTC().AddDate(0, 0, -1)
	daily, err := marketdata.NewDaily(symbol, txDate, 100.0, 100.0, 100.0, 100.0, 1000)
	if err != nil {
		t.Fatalf("failed creating daily seed: %v", err)
	}
	daily.ListingID = listing.ID
	if err := deps.dailyStore.Create(ctx, &daily); err != nil {
		t.Fatalf("failed storing daily seed: %v", err)
	}

	imp := importer.NewImport(*v, filepath.Join(t.TempDir(), "degiro.csv"))
	if err := deps.importStore.Create(ctx, imp); err != nil {
		t.Fatalf("failed storing import seed: %v", err)
	}

	ptx, err := portfolio.NewTransaction(portfolio.TransactionData{
		Source:      string(v.Name),
		OccurredAt:  txDate,
		ISIN:        &isin,
		Symbol:      &symbol,
		Description: "BUY integration row",
		Type:        portfolio.TxBuy,
		Quantity:    1,
		Price:       100,
		Amount:      100,
	}, 1, imp.ID, &acc.ID, nil)
	if err != nil {
		t.Fatalf("failed creating portfolio transaction seed: %v", err)
	}
	if err := deps.portfolioTransactionStore.Create(ctx, ptx); err != nil {
		t.Fatalf("failed storing portfolio transaction seed: %v", err)
	}

	msg, err := bus.NewJSONEnvelope(api.TransactionsCreated{AccID: acc.ID})
	if err != nil {
		t.Fatalf("failed creating transactions created envelope: %v", err)
	}
	if err := b.Publish(ctx, msg); err != nil {
		t.Fatalf("failed publishing transactions created event: %v", err)
	}

	positionsTable := qualifiedTable(backend, storage.SchemaPortfolio, storage.TablePositions)
	positionSnapshotsTable := qualifiedTable(backend, storage.SchemaPortfolio, storage.TablePosSnapshots)
	portfolioSnapshotsTable := qualifiedTable(backend, storage.SchemaPortfolio, storage.TablePortSnapshots)

	var positionsCount, positionSnapshotsCount, portfolioSnapshotsCount int
	deadline := time.Now().Add(5 * time.Second)
	for {
		positionsCount = countByAccountID(t, ctx, db, positionsTable, acc.ID)
		positionSnapshotsCount = countByAccountID(t, ctx, db, positionSnapshotsTable, acc.ID)
		portfolioSnapshotsCount = countByAccountID(t, ctx, db, portfolioSnapshotsTable, acc.ID)

		if positionsCount > 0 && positionSnapshotsCount > 0 && portfolioSnapshotsCount > 0 {
			break
		}
		if time.Now().After(deadline) {
			t.Fatalf(
				"timed out waiting for portfolio build after TransactionsCreated event (positions=%d position_snapshots=%d portfolio_snapshots=%d)",
				positionsCount,
				positionSnapshotsCount,
				portfolioSnapshotsCount,
			)
		}
		time.Sleep(25 * time.Millisecond)
	}
}

func TestIntegration_PortfolioRebuildEndpoint_PublishesAndBuilds(t *testing.T) {
	logger := logging.NewSlogLogger(slog.LevelDebug)
	db, backend := setupIntegrationDB(t, logger)

	cfg := &config.Config{
		DiskStorage: config.DiskStorage{
			BasePath: t.TempDir(),
		},
	}
	deps := newAppDependencies(logger, db, cfg)
	b, err := setupBus(logger, deps)
	if err != nil {
		t.Fatalf("failed setting up bus: %v", err)
	}
	router := setupRouterWithDeps(logger, deps, b, nil, nil)
	mux := http.NewServeMux()
	router.Register(mux)
	server := httptest.NewServer(mux)

	t.Cleanup(func() {
		server.Close()
		_ = b.Close()
		_ = db.Close()
	})

	ctx := context.Background()
	acc := seedPortfolioBuildFixture(t, ctx, deps)

	raw, err := json.Marshal(api.RebuildPortfolioRequest{AccountID: acc.ID})
	if err != nil {
		t.Fatalf("failed marshalling request: %v", err)
	}
	req, err := http.NewRequest(http.MethodPost, server.URL+"/portfolio/rebuild", bytes.NewReader(raw))
	if err != nil {
		t.Fatalf("failed creating request: %v", err)
	}
	req.Header.Set("Content-Type", "application/json")

	res, err := http.DefaultClient.Do(req)
	if err != nil {
		t.Fatalf("POST /portfolio/rebuild failed: %v", err)
	}
	defer res.Body.Close()
	if res.StatusCode != http.StatusAccepted {
		t.Fatalf("expected status 202, got %d body=%s", res.StatusCode, mustReadBody(t, res))
	}
	var accepted api.AsyncEventAcceptedResponse
	if err := json.NewDecoder(res.Body).Decode(&accepted); err != nil {
		t.Fatalf("failed decoding accepted response: %v", err)
	}
	if accepted.AccountID != acc.ID {
		t.Fatalf("expected account id %s, got %s", acc.ID, accepted.AccountID)
	}
	if accepted.Topic != (api.PortfolioRebuildRequested{}).MessageTopic() {
		t.Fatalf("expected topic %s, got %s", (api.PortfolioRebuildRequested{}).MessageTopic(), accepted.Topic)
	}

	positionsTable := qualifiedTable(backend, storage.SchemaPortfolio, storage.TablePositions)
	positionSnapshotsTable := qualifiedTable(backend, storage.SchemaPortfolio, storage.TablePosSnapshots)
	portfolioSnapshotsTable := qualifiedTable(backend, storage.SchemaPortfolio, storage.TablePortSnapshots)

	var positionsCount, positionSnapshotsCount, portfolioSnapshotsCount int
	deadline := time.Now().Add(5 * time.Second)
	for {
		positionsCount = countByAccountID(t, ctx, db, positionsTable, acc.ID)
		positionSnapshotsCount = countByAccountID(t, ctx, db, positionSnapshotsTable, acc.ID)
		portfolioSnapshotsCount = countByAccountID(t, ctx, db, portfolioSnapshotsTable, acc.ID)
		if positionsCount > 0 && positionSnapshotsCount > 0 && portfolioSnapshotsCount > 0 {
			break
		}
		if time.Now().After(deadline) {
			t.Fatalf(
				"timed out waiting for portfolio build after rebuild endpoint call (positions=%d position_snapshots=%d portfolio_snapshots=%d)",
				positionsCount,
				positionSnapshotsCount,
				portfolioSnapshotsCount,
			)
		}
		time.Sleep(25 * time.Millisecond)
	}
}

func seedPortfolioBuildFixture(t *testing.T, ctx context.Context, deps *appDependencies) *account.Account {
	t.Helper()

	acc, err := account.NewAccount("portfolio-rebuild-it", nil)
	if err != nil {
		t.Fatalf("failed creating account seed: %v", err)
	}
	if err := deps.accountStore.Create(ctx, acc); err != nil {
		t.Fatalf("failed storing account seed: %v", err)
	}
	if err := deps.portfolioAccountStore.Create(ctx, portfolio.NewAccount(acc.ID)); err != nil {
		t.Fatalf("failed storing portfolio account projection: %v", err)
	}

	v, err := vendor.NewVendor(vendor.VendorDEGIRO, vendor.VendorTypeBrokerage)
	if err != nil {
		t.Fatalf("failed creating vendor seed: %v", err)
	}
	if err := deps.vendorStore.Create(ctx, v); err != nil {
		t.Fatalf("failed storing vendor seed: %v", err)
	}

	symbol := "ITEST.AS"
	isin := "NLITEST000001"
	listing, err := marketdata.NewListing(
		symbol,
		"Integration Test Listing",
		marketdata.SourceAlphaVantage,
		marketdata.ListingWithISIN(isin),
	)
	if err != nil {
		t.Fatalf("failed creating listing seed: %v", err)
	}
	if err := deps.listingStore.Create(ctx, listing); err != nil {
		t.Fatalf("failed storing listing seed: %v", err)
	}

	txDate := time.Now().UTC().AddDate(0, 0, -1)
	daily, err := marketdata.NewDaily(symbol, txDate, 100.0, 100.0, 100.0, 100.0, 1000)
	if err != nil {
		t.Fatalf("failed creating daily seed: %v", err)
	}
	daily.ListingID = listing.ID
	if err := deps.dailyStore.Create(ctx, &daily); err != nil {
		t.Fatalf("failed storing daily seed: %v", err)
	}

	imp := importer.NewImport(*v, filepath.Join(t.TempDir(), "degiro.csv"))
	if err := deps.importStore.Create(ctx, imp); err != nil {
		t.Fatalf("failed storing import seed: %v", err)
	}

	ptx, err := portfolio.NewTransaction(portfolio.TransactionData{
		Source:      string(v.Name),
		OccurredAt:  txDate,
		ISIN:        &isin,
		Symbol:      &symbol,
		Description: "BUY integration row",
		Type:        portfolio.TxBuy,
		Quantity:    1,
		Price:       100,
		Amount:      100,
	}, 1, imp.ID, &acc.ID, nil)
	if err != nil {
		t.Fatalf("failed creating portfolio transaction seed: %v", err)
	}
	if err := deps.portfolioTransactionStore.Create(ctx, ptx); err != nil {
		t.Fatalf("failed storing portfolio transaction seed: %v", err)
	}

	return acc
}

func mustReadBody(t *testing.T, res *http.Response) string {
	t.Helper()
	var body any
	dec := json.NewDecoder(res.Body)
	if err := dec.Decode(&body); err != nil {
		return "<unable to decode body>"
	}
	raw, err := json.Marshal(body)
	if err != nil {
		return "<unable to marshal body>"
	}
	return string(raw)
}

func countByAccountID(t *testing.T, ctx context.Context, db *storage.DB, table string, accountID uuid.UUID) int {
	t.Helper()

	query := db.Rebind(fmt.Sprintf("SELECT COUNT(1) FROM %s WHERE account_id = ?", table))
	var count int
	if err := db.GetContext(ctx, &count, query, accountID); err != nil {
		t.Fatalf("failed counting rows in %s for account %s: %v", table, accountID, err)
	}
	return count
}
