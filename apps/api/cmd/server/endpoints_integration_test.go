package main

import (
	"bytes"
	"context"
	"encoding/json"
	"flag"
	"fmt"
	"io"
	"log/slog"
	"mime/multipart"
	"net/http"
	"net/http/httptest"
	"net/textproto"
	"os"
	"path/filepath"
	"runtime"
	"strings"
	"sync"
	"sync/atomic"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/lennardclaproth/my-finances-tracker/internal/account"
	"github.com/lennardclaproth/my-finances-tracker/internal/cashflow"
	"github.com/lennardclaproth/my-finances-tracker/internal/config"
	"github.com/lennardclaproth/my-finances-tracker/internal/importer"
	"github.com/lennardclaproth/my-finances-tracker/internal/jobs"
	"github.com/lennardclaproth/my-finances-tracker/internal/logging"
	"github.com/lennardclaproth/my-finances-tracker/internal/marketdata"
	"github.com/lennardclaproth/my-finances-tracker/internal/storage"
	"github.com/lennardclaproth/my-finances-tracker/internal/vendor"
	"github.com/lennardclaproth/my-finances-tracker/migrations"
)

var (
	itDBFlag         = flag.String("itdb", "", "integration test database backend: sqlite|postgres")
	itPostgresDSN    = flag.String("itpgdsn", "", "postgres DSN for integration tests")
	defaultSQLiteDSN = "file:mft-integration?mode=memory&cache=shared"
)

type integrationApp struct {
	db      *storage.DB
	backend storage.ConnectionType
	server  *httptest.Server
}

func TestIntegration_HealthEndpoint_HappyPath(t *testing.T) {
	app := newIntegrationApp(t)

	res, err := http.Get(app.server.URL + "/health")
	if err != nil {
		t.Fatalf("GET /health failed: %v", err)
	}
	defer res.Body.Close()

	if res.StatusCode != http.StatusOK {
		t.Fatalf("expected status 200, got %d", res.StatusCode)
	}

	body, err := io.ReadAll(res.Body)
	if err != nil {
		t.Fatalf("failed reading /health response: %v", err)
	}
	if !strings.Contains(string(body), `"status":"ok"`) {
		t.Fatalf("expected health response to include status ok, got: %s", string(body))
	}
}

func TestIntegration_ImportCsvEndpoint_HappyPath(t *testing.T) {
	app := newIntegrationApp(t)
	ctx := context.Background()

	v := seedVendorING(t, app.db)
	acc := seedImportAccount(t, app.db)

	var reqBody bytes.Buffer
	writer := multipart.NewWriter(&reqBody)
	partHeader := textproto.MIMEHeader{}
	partHeader.Set("Content-Disposition", `form-data; name="file"; filename="transactions.csv"`)
	partHeader.Set("Content-Type", "text/csv")
	part, err := writer.CreatePart(partHeader)
	if err != nil {
		t.Fatalf("failed creating multipart part: %v", err)
	}
	_, _ = part.Write([]byte("Date;Name / Description;Notifications;Amount (EUR);Debit/credit\n20250101;Coffee;Note;1,23;Debit\n"))
	if err := writer.WriteField("vendor_id", v.ID.String()); err != nil {
		t.Fatalf("failed writing vendor_id field: %v", err)
	}
	if err := writer.WriteField("account_id", acc.AccountID.String()); err != nil {
		t.Fatalf("failed writing account_id field: %v", err)
	}
	if err := writer.Close(); err != nil {
		t.Fatalf("failed closing multipart writer: %v", err)
	}

	req, err := http.NewRequest(http.MethodPost, app.server.URL+"/import/csv", &reqBody)
	if err != nil {
		t.Fatalf("failed creating request: %v", err)
	}
	req.Header.Set("Content-Type", writer.FormDataContentType())

	res, err := http.DefaultClient.Do(req)
	if err != nil {
		t.Fatalf("POST /import/csv failed: %v", err)
	}
	defer res.Body.Close()

	if res.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(res.Body)
		t.Fatalf("expected status 200, got %d body=%s", res.StatusCode, string(body))
	}

	var importID uuid.UUID
	if err := json.NewDecoder(res.Body).Decode(&importID); err != nil {
		t.Fatalf("failed decoding import id: %v", err)
	}
	if importID == uuid.Nil {
		t.Fatalf("expected non-nil import id")
	}

	var imp importer.Import
	importsTable := qualifiedTable(app.backend, storage.SchemaImports, storage.TableImports)
	query := app.db.Rebind(fmt.Sprintf("SELECT * FROM %s WHERE id = ?", importsTable))
	if err := app.db.GetContext(ctx, &imp, query, importID); err != nil {
		t.Fatalf("failed fetching created import: %v", err)
	}

	if imp.VendorID != v.ID {
		t.Fatalf("expected vendor id %s, got %s", v.ID, imp.VendorID)
	}
	if imp.AccountID == nil || *imp.AccountID != acc.AccountID {
		t.Fatalf("expected import account id %s, got %+v", acc.AccountID, imp.AccountID)
	}
	if _, err := os.Stat(imp.Path); err != nil {
		t.Fatalf("expected uploaded csv file at %s: %v", imp.Path, err)
	}
}

func TestIntegration_GetVendorsEndpoint_ReturnsActiveOnly(t *testing.T) {
	app := newIntegrationApp(t)
	ctx := context.Background()

	activeVendor := seedVendorING(t, app.db)
	inactiveVendor, err := vendor.NewVendor(vendor.VendorDEGIRO, vendor.VendorTypeBrokerage)
	if err != nil {
		t.Fatalf("failed creating inactive vendor seed: %v", err)
	}
	if err := storage.NewSQLXVendorStore(app.db).Create(ctx, inactiveVendor); err != nil {
		t.Fatalf("failed storing inactive vendor seed: %v", err)
	}

	vendorsTable := qualifiedTable(app.backend, storage.SchemaVendors, storage.TableVendors)
	updateInactive := app.db.Rebind(fmt.Sprintf("UPDATE %s SET active = ? WHERE id = ?", vendorsTable))
	if _, err := app.db.ExecContext(ctx, updateInactive, false, inactiveVendor.ID); err != nil {
		t.Fatalf("failed setting vendor inactive: %v", err)
	}

	res, err := http.Get(app.server.URL + "/vendors")
	if err != nil {
		t.Fatalf("GET /vendors failed: %v", err)
	}
	defer res.Body.Close()

	if res.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(res.Body)
		t.Fatalf("expected status 200, got %d body=%s", res.StatusCode, string(body))
	}

	var payload []struct {
		ID     uuid.UUID `json:"id"`
		Name   string    `json:"name"`
		Active bool      `json:"active"`
	}
	if err := json.NewDecoder(res.Body).Decode(&payload); err != nil {
		t.Fatalf("failed decoding /vendors response: %v", err)
	}

	if len(payload) == 0 {
		t.Fatalf("expected at least one active vendor")
	}

	var foundActive bool
	for _, row := range payload {
		if row.ID == inactiveVendor.ID {
			t.Fatalf("inactive vendor %s should not be returned", inactiveVendor.ID)
		}
		if row.ID == activeVendor.ID {
			foundActive = true
			if !row.Active {
				t.Fatalf("active vendor %s should be marked active", activeVendor.ID)
			}
		}
	}

	if !foundActive {
		t.Fatalf("expected active vendor %s in response", activeVendor.ID)
	}
}

func TestIntegration_CreateListingEndpoint_HappyPath(t *testing.T) {
	app := newIntegrationApp(t)
	ctx := context.Background()

	mockMarketStack, callCount := newMarketStackFixtureServer(t)
	seedMarketStackProvider(t, app.db, mockMarketStack.URL, "test-api-key")

	payload := map[string]any{
		"name":   "VanEck AEX UCITS ETF",
		"symbol": "TDT.AS",
		"source": string(marketdata.SourceAlphaVantage),
	}
	raw, err := json.Marshal(payload)
	if err != nil {
		t.Fatalf("failed marshalling listing payload: %v", err)
	}

	req, err := http.NewRequest(http.MethodPost, app.server.URL+"/marketdata/listing", bytes.NewReader(raw))
	if err != nil {
		t.Fatalf("failed creating request: %v", err)
	}
	req.Header.Set("Content-Type", "application/json")

	res, err := http.DefaultClient.Do(req)
	if err != nil {
		t.Fatalf("POST /marketdata/listing failed: %v", err)
	}
	defer res.Body.Close()

	if res.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(res.Body)
		t.Fatalf("expected status 200, got %d body=%s", res.StatusCode, string(body))
	}

	listingsTable := qualifiedTable(app.backend, storage.SchemaMarketData, storage.TableListings)
	var listing marketdata.Listing
	queryListing := app.db.Rebind(fmt.Sprintf("SELECT * FROM %s WHERE symbol = ?", listingsTable))
	if err := app.db.GetContext(ctx, &listing, queryListing, "TDT.AS"); err != nil {
		t.Fatalf("failed fetching listing: %v", err)
	}
	if listing.Symbol != "TDT.AS" {
		t.Fatalf("expected listing symbol TDT.AS, got %s", listing.Symbol)
	}

	// Validate async behavior: background sync should call MarketStack and persist at least one daily row.
	dailiesTable := qualifiedTable(app.backend, storage.SchemaMarketData, storage.TableHistories)
	queryDailyCount := app.db.Rebind(fmt.Sprintf("SELECT COUNT(1) FROM %s WHERE symbol = ?", dailiesTable))
	querySyncing := app.db.Rebind(fmt.Sprintf("SELECT syncing FROM %s WHERE symbol = ?", listingsTable))

	deadline := time.Now().Add(10 * time.Second)
	for {
		var dailyCount int
		if err := app.db.GetContext(ctx, &dailyCount, queryDailyCount, "TDT.AS"); err != nil {
			t.Fatalf("failed counting daily rows: %v", err)
		}
		var syncing bool
		if err := app.db.GetContext(ctx, &syncing, querySyncing, "TDT.AS"); err != nil {
			t.Fatalf("failed reading listing syncing flag: %v", err)
		}

		if dailyCount > 0 && !syncing {
			break
		}
		if time.Now().After(deadline) {
			t.Fatalf("timed out waiting for async sync (dailyCount=%d syncing=%t)", dailyCount, syncing)
		}
		time.Sleep(100 * time.Millisecond)
	}

	if atomic.LoadInt64(callCount) == 0 {
		t.Fatalf("expected marketstack mock to be called at least once")
	}
}

func TestIntegration_UpdateListingEndpoint_HappyPath(t *testing.T) {
	app := newIntegrationApp(t)
	ctx := context.Background()

	store := storage.NewSQLXListingStore(app.db)
	listing, err := marketdata.NewListing(
		"TUPD.AS",
		"Update Listing",
		marketdata.SourceAlphaVantage,
		marketdata.ListingWithDescription("old description"),
		marketdata.ListingWithISIN("NLUPDATE0001"),
	)
	if err != nil {
		t.Fatalf("failed creating listing seed: %v", err)
	}
	if err := store.Create(ctx, listing); err != nil {
		t.Fatalf("failed storing listing seed: %v", err)
	}

	payload := map[string]any{
		"id":          listing.ID,
		"description": "updated description",
		"exchange":    "XAMS",
	}
	raw, err := json.Marshal(payload)
	if err != nil {
		t.Fatalf("failed marshalling update payload: %v", err)
	}

	req, err := http.NewRequest(http.MethodPatch, app.server.URL+"/marketdata/listing", bytes.NewReader(raw))
	if err != nil {
		t.Fatalf("failed creating request: %v", err)
	}
	req.Header.Set("Content-Type", "application/json")

	res, err := http.DefaultClient.Do(req)
	if err != nil {
		t.Fatalf("PATCH /listing failed: %v", err)
	}
	defer res.Body.Close()

	if res.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(res.Body)
		t.Fatalf("expected status 200, got %d body=%s", res.StatusCode, string(body))
	}

	listingsTable := qualifiedTable(app.backend, storage.SchemaMarketData, storage.TableListings)
	var persisted marketdata.Listing
	query := app.db.Rebind(fmt.Sprintf("SELECT * FROM %s WHERE id = ?", listingsTable))
	if err := app.db.GetContext(ctx, &persisted, query, listing.ID); err != nil {
		t.Fatalf("failed fetching updated listing: %v", err)
	}
	if persisted.Description == nil || *persisted.Description != "updated description" {
		t.Fatalf("expected updated description, got %+v", persisted.Description)
	}
	if persisted.Exchange == nil || *persisted.Exchange != "XAMS" {
		t.Fatalf("expected updated exchange, got %+v", persisted.Exchange)
	}
	if persisted.Name != "Update Listing" {
		t.Fatalf("expected name to remain unchanged, got %s", persisted.Name)
	}
	if persisted.ISIN == nil || *persisted.ISIN != "NLUPDATE0001" {
		t.Fatalf("expected isin to remain unchanged, got %+v", persisted.ISIN)
	}
}

func TestIntegration_GetListingsEndpoint_ReturnsSortedRows(t *testing.T) {
	app := newIntegrationApp(t)
	ctx := context.Background()

	store := storage.NewSQLXListingStore(app.db)
	listingA, err := marketdata.NewListing("ZZZ.AS", "Zulu", marketdata.SourceAlphaVantage)
	if err != nil {
		t.Fatalf("failed creating listing seed A: %v", err)
	}
	listingB, err := marketdata.NewListing("AAA.AS", "Alpha", marketdata.SourceAlphaVantage)
	if err != nil {
		t.Fatalf("failed creating listing seed B: %v", err)
	}
	if err := store.Create(ctx, listingA); err != nil {
		t.Fatalf("failed storing listing A: %v", err)
	}
	if err := store.Create(ctx, listingB); err != nil {
		t.Fatalf("failed storing listing B: %v", err)
	}

	req, err := http.NewRequest(http.MethodGet, app.server.URL+"/marketdata/listings", nil)
	if err != nil {
		t.Fatalf("failed creating request: %v", err)
	}
	res, err := http.DefaultClient.Do(req)
	if err != nil {
		t.Fatalf("GET /marketdata/listings failed: %v", err)
	}
	defer res.Body.Close()

	if res.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(res.Body)
		t.Fatalf("expected status 200, got %d body=%s", res.StatusCode, string(body))
	}

	var payload []struct {
		ID     uuid.UUID `json:"id"`
		Symbol string    `json:"symbol"`
		Name   string    `json:"name"`
	}
	if err := json.NewDecoder(res.Body).Decode(&payload); err != nil {
		t.Fatalf("failed decoding payload: %v", err)
	}
	if len(payload) < 2 {
		t.Fatalf("expected at least two listings in response, got %d", len(payload))
	}

	var idxA, idxZ int = -1, -1
	for i, row := range payload {
		if row.Symbol == "AAA.AS" {
			idxA = i
		}
		if row.Symbol == "ZZZ.AS" {
			idxZ = i
		}
	}
	if idxA == -1 || idxZ == -1 {
		t.Fatalf("expected seeded listings in response, got %+v", payload)
	}
	if idxA >= idxZ {
		t.Fatalf("expected symbol ordering ascending (AAA before ZZZ), got idxA=%d idxZ=%d", idxA, idxZ)
	}
}

func TestIntegration_GetDailiesEndpoint_HappyPath(t *testing.T) {
	app := newIntegrationApp(t)
	ctx := context.Background()

	mockMarketStack, _ := newMarketStackFixtureServer(t)
	seedMarketStackProvider(t, app.db, mockMarketStack.URL, "test-api-key")

	createPayload := map[string]any{
		"name":   "VanEck AEX UCITS ETF",
		"symbol": "TDT.AS",
		"source": string(marketdata.SourceAlphaVantage),
	}
	createRaw, err := json.Marshal(createPayload)
	if err != nil {
		t.Fatalf("failed marshalling listing payload: %v", err)
	}
	createReq, err := http.NewRequest(http.MethodPost, app.server.URL+"/marketdata/listing", bytes.NewReader(createRaw))
	if err != nil {
		t.Fatalf("failed creating listing request: %v", err)
	}
	createReq.Header.Set("Content-Type", "application/json")

	createRes, err := http.DefaultClient.Do(createReq)
	if err != nil {
		t.Fatalf("POST /listing failed: %v", err)
	}
	defer createRes.Body.Close()
	if createRes.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(createRes.Body)
		t.Fatalf("expected status 200 from POST /marketdata/listing, got %d body=%s", createRes.StatusCode, string(body))
	}

	// Wait until async sync has persisted dailies so GET /dailies returns data.
	listingsTable := qualifiedTable(app.backend, storage.SchemaMarketData, storage.TableListings)
	dailiesTable := qualifiedTable(app.backend, storage.SchemaMarketData, storage.TableHistories)
	queryDailyCount := app.db.Rebind(fmt.Sprintf("SELECT COUNT(1) FROM %s WHERE symbol = ?", dailiesTable))
	querySyncing := app.db.Rebind(fmt.Sprintf("SELECT syncing FROM %s WHERE symbol = ?", listingsTable))

	deadline := time.Now().Add(10 * time.Second)
	for {
		var dailyCount int
		if err := app.db.GetContext(ctx, &dailyCount, queryDailyCount, "TDT.AS"); err != nil {
			t.Fatalf("failed counting daily rows: %v", err)
		}
		var syncing bool
		if err := app.db.GetContext(ctx, &syncing, querySyncing, "TDT.AS"); err != nil {
			t.Fatalf("failed reading listing syncing flag: %v", err)
		}
		if dailyCount > 0 && !syncing {
			break
		}
		if time.Now().After(deadline) {
			t.Fatalf("timed out waiting for dailies to be synced (dailyCount=%d syncing=%t)", dailyCount, syncing)
		}
		time.Sleep(100 * time.Millisecond)
	}

	getReq, err := http.NewRequest(http.MethodGet, app.server.URL+"/marketdata/dailies?symbol=TDT.AS&limit=5&offset=0", nil)
	if err != nil {
		t.Fatalf("failed creating get dailies request: %v", err)
	}
	getRes, err := http.DefaultClient.Do(getReq)
	if err != nil {
		t.Fatalf("GET /marketdata/dailies failed: %v", err)
	}
	defer getRes.Body.Close()

	if getRes.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(getRes.Body)
		t.Fatalf("expected status 200, got %d body=%s", getRes.StatusCode, string(body))
	}

	var payload struct {
		Data []struct {
			Symbol string `json:"Symbol"`
		} `json:"Data"`
		Metadata struct {
			Message     string `json:"Message"`
			ResultCount int    `json:"ResultCount"`
			TotalCount  int    `json:"TotalCount"`
		} `json:"Metadata"`
	}
	if err := json.NewDecoder(getRes.Body).Decode(&payload); err != nil {
		t.Fatalf("failed decoding /dailies response: %v", err)
	}

	if len(payload.Data) == 0 {
		t.Fatalf("expected at least one daily item in response")
	}
	if payload.Data[0].Symbol != "TDT.AS" {
		t.Fatalf("expected first daily symbol TDT.AS, got %s", payload.Data[0].Symbol)
	}
	if payload.Metadata.ResultCount <= 0 {
		t.Fatalf("expected positive metadata result count, got %d", payload.Metadata.ResultCount)
	}
}

func TestIntegration_TagTransactionEndpoint_HappyPath(t *testing.T) {
	app := newIntegrationApp(t)
	ctx := context.Background()

	v := seedVendorING(t, app.db)
	imp := importer.NewImport(*v, filepath.Join(t.TempDir(), "transactions.csv"))
	if err := storage.NewSQLXImportStore(app.db).Create(ctx, imp); err != nil {
		t.Fatalf("failed creating import seed: %v", err)
	}
	tx, err := cashflow.NewTransaction(
		"Coffee",
		"",
		"ING",
		cashflow.CashOut,
		4.20,
		time.Now().UTC(),
		1,
		imp.ID,
		nil,
	)
	if err != nil {
		t.Fatalf("failed creating transaction seed: %v", err)
	}
	if err := storage.NewSQLXBankTransactionStore(app.db).Create(ctx, tx); err != nil {
		t.Fatalf("failed saving transaction seed: %v", err)
	}

	payload := map[string]any{
		"id":  tx.ID,
		"tag": "food",
	}
	raw, err := json.Marshal(payload)
	if err != nil {
		t.Fatalf("failed marshalling tag payload: %v", err)
	}

	req, err := http.NewRequest(http.MethodPost, app.server.URL+"/cashflow/transactions/tag", bytes.NewReader(raw))
	if err != nil {
		t.Fatalf("failed creating request: %v", err)
	}
	req.Header.Set("Content-Type", "application/json")

	res, err := http.DefaultClient.Do(req)
	if err != nil {
		t.Fatalf("POST /cashflow/transactions/tag failed: %v", err)
	}
	defer res.Body.Close()

	if res.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(res.Body)
		t.Fatalf("expected status 200, got %d body=%s", res.StatusCode, string(body))
	}

	transactionsTable := qualifiedTable(app.backend, storage.SchemaCashflow, storage.TableTransactions)
	query := app.db.Rebind(fmt.Sprintf("SELECT tag FROM %s WHERE id = ?", transactionsTable))
	var gotTag string
	if err := app.db.GetContext(ctx, &gotTag, query, tx.ID); err != nil {
		t.Fatalf("failed fetching tagged transaction: %v", err)
	}
	if gotTag != "food" {
		t.Fatalf("expected tag food, got %s", gotTag)
	}
}

func TestIntegration_TagTransactionsBySelectionEndpoint_HappyPath(t *testing.T) {
	app := newIntegrationApp(t)
	ctx := context.Background()

	v := seedVendorING(t, app.db)
	imp := importer.NewImport(*v, filepath.Join(t.TempDir(), "transactions.csv"))
	if err := storage.NewSQLXImportStore(app.db).Create(ctx, imp); err != nil {
		t.Fatalf("failed creating import seed: %v", err)
	}

	store := storage.NewSQLXBankTransactionStore(app.db)
	tx1, err := cashflow.NewTransaction("Coffee", "", "ING", cashflow.CashOut, 4.20, time.Now().UTC(), 1, imp.ID, nil)
	if err != nil {
		t.Fatalf("failed creating tx1: %v", err)
	}
	tx2, err := cashflow.NewTransaction("Lunch", "", "ING", cashflow.CashOut, 8.70, time.Now().UTC(), 2, imp.ID, nil)
	if err != nil {
		t.Fatalf("failed creating tx2: %v", err)
	}
	tx3, err := cashflow.NewTransaction("Salary", "", "ING", cashflow.CashIn, 1000.00, time.Now().UTC(), 3, imp.ID, nil)
	if err != nil {
		t.Fatalf("failed creating tx3: %v", err)
	}
	for _, tx := range []*cashflow.Transaction{tx1, tx2, tx3} {
		if err := store.Create(ctx, tx); err != nil {
			t.Fatalf("failed storing seed transaction: %v", err)
		}
	}

	payload := map[string]any{
		"tag": "food",
		"ids": []uuid.UUID{tx1.ID, tx2.ID},
	}
	raw, err := json.Marshal(payload)
	if err != nil {
		t.Fatalf("failed marshalling payload: %v", err)
	}

	req, err := http.NewRequest(http.MethodPost, app.server.URL+"/cashflow/transactions/tag/selection", bytes.NewReader(raw))
	if err != nil {
		t.Fatalf("failed creating request: %v", err)
	}
	req.Header.Set("Content-Type", "application/json")

	res, err := http.DefaultClient.Do(req)
	if err != nil {
		t.Fatalf("POST /cashflow/transactions/tag/selection failed: %v", err)
	}
	defer res.Body.Close()
	if res.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(res.Body)
		t.Fatalf("expected status 200, got %d body=%s", res.StatusCode, string(body))
	}

	var body struct {
		UpdatedCount int    `json:"updated_count"`
		Status       string `json:"status"`
	}
	if err := json.NewDecoder(res.Body).Decode(&body); err != nil {
		t.Fatalf("failed decoding response: %v", err)
	}
	if body.UpdatedCount != 2 {
		t.Fatalf("expected updated_count=2, got %d", body.UpdatedCount)
	}

	transactionsTable := qualifiedTable(app.backend, storage.SchemaCashflow, storage.TableTransactions)
	query := app.db.Rebind(fmt.Sprintf("SELECT tag FROM %s WHERE id = ?", transactionsTable))

	for _, id := range []uuid.UUID{tx1.ID, tx2.ID} {
		var tag string
		if err := app.db.GetContext(ctx, &tag, query, id); err != nil {
			t.Fatalf("failed fetching tag: %v", err)
		}
		if tag != "food" {
			t.Fatalf("expected tag food for %s, got %s", id, tag)
		}
	}

	var untouched string
	if err := app.db.GetContext(ctx, &untouched, query, tx3.ID); err != nil {
		t.Fatalf("failed fetching untouched transaction tag: %v", err)
	}
	if untouched != "" {
		t.Fatalf("expected untouched transaction tag to remain empty, got %s", untouched)
	}
}

func TestIntegration_TagTransactionsByFilterEndpoint_SyncAndAsync(t *testing.T) {
	t.Run("sync update under threshold", func(t *testing.T) {
		app := newIntegrationApp(t)
		ctx := context.Background()

		v := seedVendorING(t, app.db)
		imp := importer.NewImport(*v, filepath.Join(t.TempDir(), "transactions.csv"))
		if err := storage.NewSQLXImportStore(app.db).Create(ctx, imp); err != nil {
			t.Fatalf("failed creating import seed: %v", err)
		}

		store := storage.NewSQLXBankTransactionStore(app.db)
		tx1, err := cashflow.NewTransaction("Istanbul ferry", "Kadikoy", "ING", cashflow.CashOut, 12.50, time.Now().UTC(), 1, imp.ID, nil)
		if err != nil {
			t.Fatalf("failed creating tx1: %v", err)
		}
		tx2, err := cashflow.NewTransaction("Groceries", "Amsterdam", "ING", cashflow.CashOut, 20.00, time.Now().UTC(), 2, imp.ID, nil)
		if err != nil {
			t.Fatalf("failed creating tx2: %v", err)
		}
		for _, tx := range []*cashflow.Transaction{tx1, tx2} {
			if err := store.Create(ctx, tx); err != nil {
				t.Fatalf("failed storing seed transaction: %v", err)
			}
		}

		payload := map[string]any{
			"tag": "travel",
			"filters": map[string]any{
				"source": "ing",
				"q":      "istanbul",
			},
		}
		raw, err := json.Marshal(payload)
		if err != nil {
			t.Fatalf("failed marshalling payload: %v", err)
		}

		req, err := http.NewRequest(http.MethodPost, app.server.URL+"/cashflow/transactions/tag/filter", bytes.NewReader(raw))
		if err != nil {
			t.Fatalf("failed creating request: %v", err)
		}
		req.Header.Set("Content-Type", "application/json")

		res, err := http.DefaultClient.Do(req)
		if err != nil {
			t.Fatalf("POST /cashflow/transactions/tag/filter failed: %v", err)
		}
		defer res.Body.Close()
		if res.StatusCode != http.StatusOK {
			body, _ := io.ReadAll(res.Body)
			t.Fatalf("expected status 200, got %d body=%s", res.StatusCode, string(body))
		}

		var body struct {
			UpdatedCount int `json:"updated_count"`
		}
		if err := json.NewDecoder(res.Body).Decode(&body); err != nil {
			t.Fatalf("failed decoding response: %v", err)
		}
		if body.UpdatedCount != 1 {
			t.Fatalf("expected updated_count=1, got %d", body.UpdatedCount)
		}
	})

	t.Run("async scheduling over threshold", func(t *testing.T) {
		enq := &fakeBulkTagEnqueuer{}
		app := newIntegrationAppWithBulkTagEnqueuer(t, enq)
		ctx := context.Background()

		v := seedVendorING(t, app.db)
		imp := importer.NewImport(*v, filepath.Join(t.TempDir(), "transactions.csv"))
		if err := storage.NewSQLXImportStore(app.db).Create(ctx, imp); err != nil {
			t.Fatalf("failed creating import seed: %v", err)
		}

		store := storage.NewSQLXBankTransactionStore(app.db)
		for i := 0; i < 1001; i++ {
			tx, err := cashflow.NewTransaction(
				fmt.Sprintf("Bulk tx %d", i),
				"",
				"ING",
				cashflow.CashOut,
				1.00,
				time.Now().UTC(),
				i+1,
				imp.ID,
				nil,
			)
			if err != nil {
				t.Fatalf("failed creating bulk transaction %d: %v", i, err)
			}
			if err := store.Create(ctx, tx); err != nil {
				t.Fatalf("failed storing bulk transaction %d: %v", i, err)
			}
		}

		payload := map[string]any{
			"tag": "bulk",
			"filters": map[string]any{
				"source": "ing",
			},
		}
		raw, err := json.Marshal(payload)
		if err != nil {
			t.Fatalf("failed marshalling payload: %v", err)
		}

		req, err := http.NewRequest(http.MethodPost, app.server.URL+"/cashflow/transactions/tag/filter", bytes.NewReader(raw))
		if err != nil {
			t.Fatalf("failed creating request: %v", err)
		}
		req.Header.Set("Content-Type", "application/json")

		res, err := http.DefaultClient.Do(req)
		if err != nil {
			t.Fatalf("POST /cashflow/transactions/tag/filter failed: %v", err)
		}
		defer res.Body.Close()
		if res.StatusCode != http.StatusAccepted {
			body, _ := io.ReadAll(res.Body)
			t.Fatalf("expected status 202, got %d body=%s", res.StatusCode, string(body))
		}

		var body struct {
			UpdatedCount int    `json:"updated_count"`
			Status       string `json:"status"`
		}
		if err := json.NewDecoder(res.Body).Decode(&body); err != nil {
			t.Fatalf("failed decoding response: %v", err)
		}
		if body.UpdatedCount != 0 {
			t.Fatalf("expected updated_count=0 for scheduled job, got %d", body.UpdatedCount)
		}
		if !strings.Contains(strings.ToLower(body.Status), "scheduled") {
			t.Fatalf("expected scheduled status message, got %q", body.Status)
		}
		if enq.count() != 1 {
			t.Fatalf("expected one background enqueue call, got %d", enq.count())
		}
	})
}

func TestIntegration_IgnoreTransactionsBySelectionEndpoint_HappyPath(t *testing.T) {
	app := newIntegrationApp(t)
	ctx := context.Background()

	v := seedVendorING(t, app.db)
	imp := importer.NewImport(*v, filepath.Join(t.TempDir(), "transactions.csv"))
	if err := storage.NewSQLXImportStore(app.db).Create(ctx, imp); err != nil {
		t.Fatalf("failed creating import seed: %v", err)
	}

	store := storage.NewSQLXBankTransactionStore(app.db)
	tx1, err := cashflow.NewTransaction("Coffee", "", "ING", cashflow.CashOut, 4.20, time.Now().UTC(), 1, imp.ID, nil)
	if err != nil {
		t.Fatalf("failed creating tx1: %v", err)
	}
	tx2, err := cashflow.NewTransaction("Lunch", "", "ING", cashflow.CashOut, 8.70, time.Now().UTC(), 2, imp.ID, nil)
	if err != nil {
		t.Fatalf("failed creating tx2: %v", err)
	}
	tx3, err := cashflow.NewTransaction("Salary", "", "ING", cashflow.CashIn, 1000.00, time.Now().UTC(), 3, imp.ID, nil)
	if err != nil {
		t.Fatalf("failed creating tx3: %v", err)
	}
	for _, tx := range []*cashflow.Transaction{tx1, tx2, tx3} {
		if err := store.Create(ctx, tx); err != nil {
			t.Fatalf("failed storing seed transaction: %v", err)
		}
	}

	payload := map[string]any{
		"ignored": true,
		"ids":     []uuid.UUID{tx1.ID, tx2.ID},
	}
	raw, err := json.Marshal(payload)
	if err != nil {
		t.Fatalf("failed marshalling payload: %v", err)
	}

	req, err := http.NewRequest(http.MethodPost, app.server.URL+"/cashflow/transactions/ignore/selection", bytes.NewReader(raw))
	if err != nil {
		t.Fatalf("failed creating request: %v", err)
	}
	req.Header.Set("Content-Type", "application/json")

	res, err := http.DefaultClient.Do(req)
	if err != nil {
		t.Fatalf("POST /cashflow/transactions/ignore/selection failed: %v", err)
	}
	defer res.Body.Close()
	if res.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(res.Body)
		t.Fatalf("expected status 200, got %d body=%s", res.StatusCode, string(body))
	}

	var body struct {
		UpdatedCount int    `json:"updated_count"`
		Status       string `json:"status"`
	}
	if err := json.NewDecoder(res.Body).Decode(&body); err != nil {
		t.Fatalf("failed decoding response: %v", err)
	}
	if body.UpdatedCount != 2 {
		t.Fatalf("expected updated_count=2, got %d", body.UpdatedCount)
	}

	transactionsTable := qualifiedTable(app.backend, storage.SchemaCashflow, storage.TableTransactions)
	query := app.db.Rebind(fmt.Sprintf("SELECT ignored FROM %s WHERE id = ?", transactionsTable))

	for _, id := range []uuid.UUID{tx1.ID, tx2.ID} {
		var ignored bool
		if err := app.db.GetContext(ctx, &ignored, query, id); err != nil {
			t.Fatalf("failed fetching ignored flag: %v", err)
		}
		if !ignored {
			t.Fatalf("expected ignored=true for %s", id)
		}
	}

	var untouched bool
	if err := app.db.GetContext(ctx, &untouched, query, tx3.ID); err != nil {
		t.Fatalf("failed fetching untouched transaction ignored flag: %v", err)
	}
	if untouched {
		t.Fatalf("expected untouched transaction ignored flag to remain false")
	}
}

func TestIntegration_IgnoreTransactionsByFilterEndpoint_HappyPath(t *testing.T) {
	app := newIntegrationApp(t)
	ctx := context.Background()

	v := seedVendorING(t, app.db)
	imp := importer.NewImport(*v, filepath.Join(t.TempDir(), "transactions.csv"))
	if err := storage.NewSQLXImportStore(app.db).Create(ctx, imp); err != nil {
		t.Fatalf("failed creating import seed: %v", err)
	}

	store := storage.NewSQLXBankTransactionStore(app.db)
	tx1, err := cashflow.NewTransaction("Istanbul ferry", "Kadikoy", "ING", cashflow.CashOut, 12.50, time.Now().UTC(), 1, imp.ID, nil)
	if err != nil {
		t.Fatalf("failed creating tx1: %v", err)
	}
	tx2, err := cashflow.NewTransaction("Groceries", "Amsterdam", "ING", cashflow.CashOut, 20.00, time.Now().UTC(), 2, imp.ID, nil)
	if err != nil {
		t.Fatalf("failed creating tx2: %v", err)
	}
	for _, tx := range []*cashflow.Transaction{tx1, tx2} {
		if err := store.Create(ctx, tx); err != nil {
			t.Fatalf("failed storing seed transaction: %v", err)
		}
	}

	payload := map[string]any{
		"ignored": true,
		"filters": map[string]any{
			"source": "ing",
			"q":      "istanbul",
		},
	}
	raw, err := json.Marshal(payload)
	if err != nil {
		t.Fatalf("failed marshalling payload: %v", err)
	}

	req, err := http.NewRequest(http.MethodPost, app.server.URL+"/cashflow/transactions/ignore/filter", bytes.NewReader(raw))
	if err != nil {
		t.Fatalf("failed creating request: %v", err)
	}
	req.Header.Set("Content-Type", "application/json")

	res, err := http.DefaultClient.Do(req)
	if err != nil {
		t.Fatalf("POST /cashflow/transactions/ignore/filter failed: %v", err)
	}
	defer res.Body.Close()
	if res.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(res.Body)
		t.Fatalf("expected status 200, got %d body=%s", res.StatusCode, string(body))
	}

	var body struct {
		UpdatedCount int `json:"updated_count"`
	}
	if err := json.NewDecoder(res.Body).Decode(&body); err != nil {
		t.Fatalf("failed decoding response: %v", err)
	}
	if body.UpdatedCount != 1 {
		t.Fatalf("expected updated_count=1, got %d", body.UpdatedCount)
	}

	transactionsTable := qualifiedTable(app.backend, storage.SchemaCashflow, storage.TableTransactions)
	query := app.db.Rebind(fmt.Sprintf("SELECT ignored FROM %s WHERE id = ?", transactionsTable))

	var ignored1 bool
	if err := app.db.GetContext(ctx, &ignored1, query, tx1.ID); err != nil {
		t.Fatalf("failed fetching tx1 ignored flag: %v", err)
	}
	if !ignored1 {
		t.Fatalf("expected tx1 ignored=true")
	}

	var ignored2 bool
	if err := app.db.GetContext(ctx, &ignored2, query, tx2.ID); err != nil {
		t.Fatalf("failed fetching tx2 ignored flag: %v", err)
	}
	if ignored2 {
		t.Fatalf("expected tx2 ignored=false")
	}
}

func TestIntegration_GetCashflowTransactionsEndpoint_SearchSortFilterPaginate(t *testing.T) {
	app := newIntegrationApp(t)
	ctx := context.Background()

	v := seedVendorING(t, app.db)
	imp := importer.NewImport(*v, filepath.Join(t.TempDir(), "transactions.csv"))
	if err := storage.NewSQLXImportStore(app.db).Create(ctx, imp); err != nil {
		t.Fatalf("failed creating import seed: %v", err)
	}

	store := storage.NewSQLXBankTransactionStore(app.db)

	tx1, err := cashflow.NewTransaction(
		"Groceries Istanbul",
		"Istanbul Kadikoy market",
		"ING",
		cashflow.CashOut,
		25.45,
		time.Date(2025, 2, 2, 0, 0, 0, 0, time.UTC),
		1,
		imp.ID,
		nil,
	)
	if err != nil {
		t.Fatalf("failed creating transaction 1 seed: %v", err)
	}
	tx2, err := cashflow.NewTransaction(
		"Flight Ticket",
		"Trip to Istanbul",
		"Revolut",
		cashflow.CashOut,
		320.00,
		time.Date(2025, 2, 1, 0, 0, 0, 0, time.UTC),
		2,
		imp.ID,
		nil,
	)
	if err != nil {
		t.Fatalf("failed creating transaction 2 seed: %v", err)
	}
	tx3, err := cashflow.NewTransaction(
		"Salary",
		"January payout",
		"ING",
		cashflow.CashIn,
		2500.00,
		time.Date(2025, 1, 31, 0, 0, 0, 0, time.UTC),
		3,
		imp.ID,
		nil,
	)
	if err != nil {
		t.Fatalf("failed creating transaction 3 seed: %v", err)
	}

	for _, tx := range []*cashflow.Transaction{tx1, tx2, tx3} {
		if err := store.Create(ctx, tx); err != nil {
			t.Fatalf("failed saving transaction seed: %v", err)
		}
	}
	if err := store.Tag(ctx, tx1.ID, "food"); err != nil {
		t.Fatalf("failed tagging transaction 1: %v", err)
	}
	if err := store.Tag(ctx, tx2.ID, "travel"); err != nil {
		t.Fatalf("failed tagging transaction 2: %v", err)
	}
	if err := store.Tag(ctx, tx3.ID, "income"); err != nil {
		t.Fatalf("failed tagging transaction 3: %v", err)
	}
	if _, err := store.UpdateIgnoredByIDs(ctx, []uuid.UUID{tx2.ID}, true); err != nil {
		t.Fatalf("failed setting transaction 2 ignored: %v", err)
	}

	req, err := http.NewRequest(
		http.MethodGet,
		app.server.URL+"/cashflow/transactions?tags=food,travel&source=ing&q=istanbul%20kadikoy&description=groc&sort_by=date&sort_order=desc&limit=10&offset=0&from=2025-01-01&to=2025-12-31",
		nil,
	)
	if err != nil {
		t.Fatalf("failed creating request: %v", err)
	}

	res, err := http.DefaultClient.Do(req)
	if err != nil {
		t.Fatalf("GET /cashflow/transactions failed: %v", err)
	}
	defer res.Body.Close()

	if res.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(res.Body)
		t.Fatalf("expected status 200, got %d body=%s", res.StatusCode, string(body))
	}

	var payload struct {
		Pagination struct {
			Limit  int `json:"limit"`
			Offset int `json:"offset"`
			Count  int `json:"count"`
			Total  int `json:"total"`
		} `json:"pagination"`
		Data []struct {
			ID        uuid.UUID `json:"id"`
			Tag       string    `json:"tag"`
			Source    string    `json:"source"`
			Direction string    `json:"direction"`
		} `json:"data"`
	}
	if err := json.NewDecoder(res.Body).Decode(&payload); err != nil {
		t.Fatalf("failed decoding /cashflow/transactions response: %v", err)
	}

	if payload.Pagination.Total != 1 || payload.Pagination.Count != 1 {
		t.Fatalf("expected total=1 count=1, got total=%d count=%d", payload.Pagination.Total, payload.Pagination.Count)
	}
	if len(payload.Data) != 1 {
		t.Fatalf("expected one transaction, got %d", len(payload.Data))
	}
	if payload.Data[0].ID != tx1.ID {
		t.Fatalf("expected transaction id %s, got %s", tx1.ID, payload.Data[0].ID)
	}
	if payload.Data[0].Tag != "food" {
		t.Fatalf("expected tag food, got %s", payload.Data[0].Tag)
	}
	if payload.Data[0].Direction != string(cashflow.CashOut) {
		t.Fatalf("expected direction %s, got %s", cashflow.CashOut, payload.Data[0].Direction)
	}

	pageReq, err := http.NewRequest(
		http.MethodGet,
		app.server.URL+"/cashflow/transactions?sort_by=date&sort_order=desc&limit=1&offset=1",
		nil,
	)
	if err != nil {
		t.Fatalf("failed creating pagination request: %v", err)
	}
	pageRes, err := http.DefaultClient.Do(pageReq)
	if err != nil {
		t.Fatalf("GET /cashflow/transactions pagination failed: %v", err)
	}
	defer pageRes.Body.Close()
	if pageRes.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(pageRes.Body)
		t.Fatalf("expected status 200, got %d body=%s", pageRes.StatusCode, string(body))
	}

	var pagePayload struct {
		Pagination struct {
			Limit  int `json:"limit"`
			Offset int `json:"offset"`
			Count  int `json:"count"`
			Total  int `json:"total"`
		} `json:"pagination"`
		Data []struct {
			ID uuid.UUID `json:"id"`
		} `json:"data"`
	}
	if err := json.NewDecoder(pageRes.Body).Decode(&pagePayload); err != nil {
		t.Fatalf("failed decoding paginated response: %v", err)
	}
	if pagePayload.Pagination.Total != 3 || pagePayload.Pagination.Count != 1 {
		t.Fatalf("expected total=3 count=1, got total=%d count=%d", pagePayload.Pagination.Total, pagePayload.Pagination.Count)
	}
	if len(pagePayload.Data) != 1 {
		t.Fatalf("expected one transaction in paginated data, got %d", len(pagePayload.Data))
	}
	if pagePayload.Data[0].ID != tx2.ID {
		t.Fatalf("expected second newest transaction id %s, got %s", tx2.ID, pagePayload.Data[0].ID)
	}

	hideIgnoredReq, err := http.NewRequest(
		http.MethodGet,
		app.server.URL+"/cashflow/transactions?sort_by=date&sort_order=desc&hide_ignored=true&limit=10&offset=0",
		nil,
	)
	if err != nil {
		t.Fatalf("failed creating hide_ignored request: %v", err)
	}
	hideIgnoredRes, err := http.DefaultClient.Do(hideIgnoredReq)
	if err != nil {
		t.Fatalf("GET /cashflow/transactions hide_ignored failed: %v", err)
	}
	defer hideIgnoredRes.Body.Close()
	if hideIgnoredRes.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(hideIgnoredRes.Body)
		t.Fatalf("expected status 200, got %d body=%s", hideIgnoredRes.StatusCode, string(body))
	}

	var hideIgnoredPayload struct {
		Pagination struct {
			Count int `json:"count"`
			Total int `json:"total"`
		} `json:"pagination"`
		Data []struct {
			ID      uuid.UUID `json:"id"`
			Ignored bool      `json:"ignored"`
		} `json:"data"`
	}
	if err := json.NewDecoder(hideIgnoredRes.Body).Decode(&hideIgnoredPayload); err != nil {
		t.Fatalf("failed decoding hide_ignored response: %v", err)
	}
	if hideIgnoredPayload.Pagination.Total != 2 || hideIgnoredPayload.Pagination.Count != 2 {
		t.Fatalf(
			"expected hide_ignored total=2 count=2, got total=%d count=%d",
			hideIgnoredPayload.Pagination.Total,
			hideIgnoredPayload.Pagination.Count,
		)
	}
	if len(hideIgnoredPayload.Data) != 2 {
		t.Fatalf("expected 2 transactions for hide_ignored, got %d", len(hideIgnoredPayload.Data))
	}
	if hideIgnoredPayload.Data[0].ID != tx1.ID || hideIgnoredPayload.Data[1].ID != tx3.ID {
		t.Fatalf(
			"expected hide_ignored order [%s, %s], got [%s, %s]",
			tx1.ID,
			tx3.ID,
			hideIgnoredPayload.Data[0].ID,
			hideIgnoredPayload.Data[1].ID,
		)
	}
	for _, entry := range hideIgnoredPayload.Data {
		if entry.Ignored {
			t.Fatalf("expected ignored=false for all hide_ignored results")
		}
	}

	amountSortReq, err := http.NewRequest(
		http.MethodGet,
		app.server.URL+"/cashflow/transactions?sort_by=amount&sort_order=asc&limit=10&offset=0",
		nil,
	)
	if err != nil {
		t.Fatalf("failed creating amount sort request: %v", err)
	}
	amountSortRes, err := http.DefaultClient.Do(amountSortReq)
	if err != nil {
		t.Fatalf("GET /cashflow/transactions amount sort failed: %v", err)
	}
	defer amountSortRes.Body.Close()
	if amountSortRes.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(amountSortRes.Body)
		t.Fatalf("expected status 200, got %d body=%s", amountSortRes.StatusCode, string(body))
	}

	var amountSortPayload struct {
		Data []struct {
			ID          uuid.UUID `json:"id"`
			AmountCents int64     `json:"amountCents"`
		} `json:"data"`
	}
	if err := json.NewDecoder(amountSortRes.Body).Decode(&amountSortPayload); err != nil {
		t.Fatalf("failed decoding amount sort response: %v", err)
	}
	if len(amountSortPayload.Data) != 3 {
		t.Fatalf("expected 3 transactions for amount sort, got %d", len(amountSortPayload.Data))
	}
	if amountSortPayload.Data[0].ID != tx1.ID || amountSortPayload.Data[1].ID != tx2.ID || amountSortPayload.Data[2].ID != tx3.ID {
		t.Fatalf(
			"expected amount order [%s, %s, %s], got [%s, %s, %s]",
			tx1.ID,
			tx2.ID,
			tx3.ID,
			amountSortPayload.Data[0].ID,
			amountSortPayload.Data[1].ID,
			amountSortPayload.Data[2].ID,
		)
	}

	directionReq, err := http.NewRequest(
		http.MethodGet,
		app.server.URL+"/cashflow/transactions?direction=in&sort_by=date&sort_order=desc&limit=10&offset=0",
		nil,
	)
	if err != nil {
		t.Fatalf("failed creating direction request: %v", err)
	}
	directionRes, err := http.DefaultClient.Do(directionReq)
	if err != nil {
		t.Fatalf("GET /cashflow/transactions direction filter failed: %v", err)
	}
	defer directionRes.Body.Close()
	if directionRes.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(directionRes.Body)
		t.Fatalf("expected status 200, got %d body=%s", directionRes.StatusCode, string(body))
	}

	var directionPayload struct {
		Pagination struct {
			Count int `json:"count"`
			Total int `json:"total"`
		} `json:"pagination"`
		Data []struct {
			ID        uuid.UUID `json:"id"`
			Direction string    `json:"direction"`
		} `json:"data"`
	}
	if err := json.NewDecoder(directionRes.Body).Decode(&directionPayload); err != nil {
		t.Fatalf("failed decoding direction response: %v", err)
	}
	if directionPayload.Pagination.Total != 1 || directionPayload.Pagination.Count != 1 {
		t.Fatalf(
			"expected direction=in total=1 count=1, got total=%d count=%d",
			directionPayload.Pagination.Total,
			directionPayload.Pagination.Count,
		)
	}
	if len(directionPayload.Data) != 1 || directionPayload.Data[0].ID != tx3.ID {
		t.Fatalf("expected direction=in to return only %s", tx3.ID)
	}
	if directionPayload.Data[0].Direction != string(cashflow.CashIn) {
		t.Fatalf("expected direction %s, got %s", cashflow.CashIn, directionPayload.Data[0].Direction)
	}

	invalidDirectionReq, err := http.NewRequest(
		http.MethodGet,
		app.server.URL+"/cashflow/transactions?direction=sideways&limit=10&offset=0",
		nil,
	)
	if err != nil {
		t.Fatalf("failed creating invalid direction request: %v", err)
	}
	invalidDirectionRes, err := http.DefaultClient.Do(invalidDirectionReq)
	if err != nil {
		t.Fatalf("GET /cashflow/transactions invalid direction failed: %v", err)
	}
	defer invalidDirectionRes.Body.Close()
	if invalidDirectionRes.StatusCode != http.StatusBadRequest {
		body, _ := io.ReadAll(invalidDirectionRes.Body)
		t.Fatalf("expected status 400, got %d body=%s", invalidDirectionRes.StatusCode, string(body))
	}
}

func TestIntegration_GetCashflowMonthlyAnalyticsEndpoint_HappyPath(t *testing.T) {
	app := newIntegrationApp(t)
	ctx := context.Background()

	v := seedVendorING(t, app.db)
	imp := importer.NewImport(*v, filepath.Join(t.TempDir(), "transactions.csv"))
	if err := storage.NewSQLXImportStore(app.db).Create(ctx, imp); err != nil {
		t.Fatalf("failed creating import seed: %v", err)
	}

	store := storage.NewSQLXBankTransactionStore(app.db)
	janIn, err := cashflow.NewTransaction("Salary Jan", "", "ING", cashflow.CashIn, 1000.00, time.Date(2025, 1, 5, 0, 0, 0, 0, time.UTC), 1, imp.ID, nil)
	if err != nil {
		t.Fatalf("failed creating janIn: %v", err)
	}
	janOut, err := cashflow.NewTransaction("Rent Jan", "", "ING", cashflow.CashOut, 250.00, time.Date(2025, 1, 10, 0, 0, 0, 0, time.UTC), 2, imp.ID, nil)
	if err != nil {
		t.Fatalf("failed creating janOut: %v", err)
	}
	janIgnored, err := cashflow.NewTransaction("Ignored Jan", "", "ING", cashflow.CashOut, 50.00, time.Date(2025, 1, 15, 0, 0, 0, 0, time.UTC), 3, imp.ID, nil)
	if err != nil {
		t.Fatalf("failed creating janIgnored: %v", err)
	}
	febIn, err := cashflow.NewTransaction("Salary Feb", "", "ING", cashflow.CashIn, 700.00, time.Date(2025, 2, 1, 0, 0, 0, 0, time.UTC), 4, imp.ID, nil)
	if err != nil {
		t.Fatalf("failed creating febIn: %v", err)
	}
	febOut, err := cashflow.NewTransaction("Groceries Feb", "", "ING", cashflow.CashOut, 200.00, time.Date(2025, 2, 2, 0, 0, 0, 0, time.UTC), 5, imp.ID, nil)
	if err != nil {
		t.Fatalf("failed creating febOut: %v", err)
	}

	for _, tx := range []*cashflow.Transaction{janIn, janOut, janIgnored, febIn, febOut} {
		if err := store.Create(ctx, tx); err != nil {
			t.Fatalf("failed storing seed transaction: %v", err)
		}
	}
	if _, err := store.UpdateIgnoredByIDs(ctx, []uuid.UUID{janIgnored.ID}, true); err != nil {
		t.Fatalf("failed setting ignored transaction: %v", err)
	}

	req, err := http.NewRequest(
		http.MethodGet,
		app.server.URL+"/cashflow/analytics/monthly?from=2025-01-01&to=2025-02-28",
		nil,
	)
	if err != nil {
		t.Fatalf("failed creating request: %v", err)
	}
	res, err := http.DefaultClient.Do(req)
	if err != nil {
		t.Fatalf("GET /cashflow/analytics/monthly failed: %v", err)
	}
	defer res.Body.Close()
	if res.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(res.Body)
		t.Fatalf("expected status 200, got %d body=%s", res.StatusCode, string(body))
	}

	var payload struct {
		Data []struct {
			Month         string `json:"month"`
			IncomingCents int64  `json:"incomingCents"`
			OutgoingCents int64  `json:"outgoingCents"`
			NetCents      int64  `json:"netCents"`
		} `json:"data"`
	}
	if err := json.NewDecoder(res.Body).Decode(&payload); err != nil {
		t.Fatalf("failed decoding analytics response: %v", err)
	}
	if len(payload.Data) != 2 {
		t.Fatalf("expected 2 monthly points, got %d", len(payload.Data))
	}

	if payload.Data[0].Month != "2025-01-01" {
		t.Fatalf("expected first month 2025-01-01, got %s", payload.Data[0].Month)
	}
	if payload.Data[0].IncomingCents != 1000000000 || payload.Data[0].OutgoingCents != 250000000 || payload.Data[0].NetCents != 750000000 {
		t.Fatalf("unexpected january totals: %+v", payload.Data[0])
	}
	if payload.Data[1].Month != "2025-02-01" {
		t.Fatalf("expected second month 2025-02-01, got %s", payload.Data[1].Month)
	}
	if payload.Data[1].IncomingCents != 700000000 || payload.Data[1].OutgoingCents != 200000000 || payload.Data[1].NetCents != 500000000 {
		t.Fatalf("unexpected february totals: %+v", payload.Data[1])
	}

	includeIgnoredReq, err := http.NewRequest(
		http.MethodGet,
		app.server.URL+"/cashflow/analytics/monthly?from=2025-01-01&to=2025-02-28&include_ignored=true",
		nil,
	)
	if err != nil {
		t.Fatalf("failed creating include_ignored request: %v", err)
	}
	includeIgnoredRes, err := http.DefaultClient.Do(includeIgnoredReq)
	if err != nil {
		t.Fatalf("GET /cashflow/analytics/monthly include_ignored failed: %v", err)
	}
	defer includeIgnoredRes.Body.Close()
	if includeIgnoredRes.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(includeIgnoredRes.Body)
		t.Fatalf("expected status 200, got %d body=%s", includeIgnoredRes.StatusCode, string(body))
	}
	var includeIgnoredPayload struct {
		Data []struct {
			Month         string `json:"month"`
			IncomingCents int64  `json:"incomingCents"`
			OutgoingCents int64  `json:"outgoingCents"`
			NetCents      int64  `json:"netCents"`
		} `json:"data"`
	}
	if err := json.NewDecoder(includeIgnoredRes.Body).Decode(&includeIgnoredPayload); err != nil {
		t.Fatalf("failed decoding include_ignored response: %v", err)
	}
	if len(includeIgnoredPayload.Data) == 0 {
		t.Fatalf("expected at least one monthly point in include_ignored response")
	}
	jan := includeIgnoredPayload.Data[0]
	if jan.Month != "2025-01-01" {
		t.Fatalf("expected january row first, got %s", jan.Month)
	}
	if jan.OutgoingCents != 300000000 || jan.NetCents != 700000000 {
		t.Fatalf("expected january outgoing=300000000 net=700000000 with include_ignored, got %+v", jan)
	}
}

func TestIntegration_GetCashflowTagDistributionEndpoint_HappyPath(t *testing.T) {
	app := newIntegrationApp(t)
	ctx := context.Background()

	v := seedVendorING(t, app.db)
	imp := importer.NewImport(*v, filepath.Join(t.TempDir(), "transactions.csv"))
	if err := storage.NewSQLXImportStore(app.db).Create(ctx, imp); err != nil {
		t.Fatalf("failed creating import seed: %v", err)
	}

	store := storage.NewSQLXBankTransactionStore(app.db)
	txFood, err := cashflow.NewTransaction("Food", "", "ING", cashflow.CashOut, 20.00, time.Date(2025, 1, 5, 0, 0, 0, 0, time.UTC), 1, imp.ID, nil)
	if err != nil {
		t.Fatalf("failed creating txFood: %v", err)
	}
	txTravel, err := cashflow.NewTransaction("Travel", "", "ING", cashflow.CashOut, 50.00, time.Date(2025, 1, 6, 0, 0, 0, 0, time.UTC), 2, imp.ID, nil)
	if err != nil {
		t.Fatalf("failed creating txTravel: %v", err)
	}
	txOutUntagged, err := cashflow.NewTransaction("Out Untagged", "", "ING", cashflow.CashOut, 10.00, time.Date(2025, 1, 7, 0, 0, 0, 0, time.UTC), 3, imp.ID, nil)
	if err != nil {
		t.Fatalf("failed creating txOutUntagged: %v", err)
	}
	txSalary, err := cashflow.NewTransaction("Salary", "", "ING", cashflow.CashIn, 100.00, time.Date(2025, 1, 8, 0, 0, 0, 0, time.UTC), 4, imp.ID, nil)
	if err != nil {
		t.Fatalf("failed creating txSalary: %v", err)
	}
	txInUntagged, err := cashflow.NewTransaction("In Untagged", "", "ING", cashflow.CashIn, 5.00, time.Date(2025, 1, 9, 0, 0, 0, 0, time.UTC), 5, imp.ID, nil)
	if err != nil {
		t.Fatalf("failed creating txInUntagged: %v", err)
	}
	txIgnored, err := cashflow.NewTransaction("Ignored", "", "ING", cashflow.CashOut, 30.00, time.Date(2025, 1, 10, 0, 0, 0, 0, time.UTC), 6, imp.ID, nil)
	if err != nil {
		t.Fatalf("failed creating txIgnored: %v", err)
	}
	for _, tx := range []*cashflow.Transaction{txFood, txTravel, txOutUntagged, txSalary, txInUntagged, txIgnored} {
		if err := store.Create(ctx, tx); err != nil {
			t.Fatalf("failed storing seed transaction: %v", err)
		}
	}
	if err := store.Tag(ctx, txFood.ID, "food"); err != nil {
		t.Fatalf("failed tagging txFood: %v", err)
	}
	if err := store.Tag(ctx, txTravel.ID, "travel"); err != nil {
		t.Fatalf("failed tagging txTravel: %v", err)
	}
	if err := store.Tag(ctx, txSalary.ID, "salary"); err != nil {
		t.Fatalf("failed tagging txSalary: %v", err)
	}
	if err := store.Tag(ctx, txIgnored.ID, "ignoredtag"); err != nil {
		t.Fatalf("failed tagging txIgnored: %v", err)
	}
	if _, err := store.UpdateIgnoredByIDs(ctx, []uuid.UUID{txIgnored.ID}, true); err != nil {
		t.Fatalf("failed setting ignored transaction: %v", err)
	}

	req, err := http.NewRequest(
		http.MethodGet,
		app.server.URL+"/cashflow/analytics/tags?from=2025-01-01&to=2025-01-31",
		nil,
	)
	if err != nil {
		t.Fatalf("failed creating request: %v", err)
	}
	res, err := http.DefaultClient.Do(req)
	if err != nil {
		t.Fatalf("GET /cashflow/analytics/tags failed: %v", err)
	}
	defer res.Body.Close()
	if res.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(res.Body)
		t.Fatalf("expected status 200, got %d body=%s", res.StatusCode, string(body))
	}

	var payload struct {
		Combined []struct {
			Tag        string `json:"tag"`
			TotalCents int64  `json:"totalCents"`
		} `json:"combined"`
		Incoming []struct {
			Tag        string `json:"tag"`
			TotalCents int64  `json:"totalCents"`
		} `json:"incoming"`
		Outgoing []struct {
			Tag        string `json:"tag"`
			TotalCents int64  `json:"totalCents"`
		} `json:"outgoing"`
	}
	if err := json.NewDecoder(res.Body).Decode(&payload); err != nil {
		t.Fatalf("failed decoding tag distribution response: %v", err)
	}

	toMap := func(entries []struct {
		Tag        string `json:"tag"`
		TotalCents int64  `json:"totalCents"`
	}) map[string]int64 {
		out := make(map[string]int64, len(entries))
		for _, entry := range entries {
			out[entry.Tag] = entry.TotalCents
		}
		return out
	}

	combined := toMap(payload.Combined)
	incoming := toMap(payload.Incoming)
	outgoing := toMap(payload.Outgoing)

	if combined["food"] != 20000000 || combined["travel"] != 50000000 || combined["salary"] != 100000000 || combined["untagged"] != 15000000 {
		t.Fatalf("unexpected combined distribution: %+v", combined)
	}
	if incoming["salary"] != 100000000 || incoming["untagged"] != 5000000 {
		t.Fatalf("unexpected incoming distribution: %+v", incoming)
	}
	if outgoing["food"] != 20000000 || outgoing["travel"] != 50000000 || outgoing["untagged"] != 10000000 {
		t.Fatalf("unexpected outgoing distribution: %+v", outgoing)
	}
	if _, exists := combined["ignoredtag"]; exists {
		t.Fatalf("ignored transactions must be excluded by default")
	}
}

func newIntegrationApp(t *testing.T) *integrationApp {
	t.Helper()

	logger := logging.NewSlogLogger(slog.LevelDebug)
	db, backend := setupIntegrationDB(t, logger)

	cfg := &config.Config{
		DiskStorage: config.DiskStorage{
			BasePath: t.TempDir(),
		},
	}
	router := setupRouter(logger, db, cfg, nil, nil)
	mux := http.NewServeMux()
	router.Register(mux)
	server := httptest.NewServer(mux)

	t.Cleanup(func() {
		server.Close()
		_ = db.Close()
	})

	return &integrationApp{
		db:      db,
		backend: backend,
		server:  server,
	}
}

func newIntegrationAppWithBulkTagEnqueuer(t *testing.T, enqueuer jobs.BulkTagEnqueuer) *integrationApp {
	t.Helper()

	logger := logging.NewSlogLogger(slog.LevelDebug)
	db, backend := setupIntegrationDB(t, logger)

	cfg := &config.Config{
		DiskStorage: config.DiskStorage{
			BasePath: t.TempDir(),
		},
	}
	router := setupRouter(logger, db, cfg, nil, enqueuer)
	mux := http.NewServeMux()
	router.Register(mux)
	server := httptest.NewServer(mux)

	t.Cleanup(func() {
		server.Close()
		_ = db.Close()
	})

	return &integrationApp{
		db:      db,
		backend: backend,
		server:  server,
	}
}

type fakeBulkTagEnqueuer struct {
	mu sync.Mutex
	n  int
}

func (f *fakeBulkTagEnqueuer) EnqueueFilter(_ context.Context, _ storage.CashflowTransactionQuery, _ string) error {
	f.mu.Lock()
	defer f.mu.Unlock()
	f.n++
	return nil
}

func (f *fakeBulkTagEnqueuer) count() int {
	f.mu.Lock()
	defer f.mu.Unlock()
	return f.n
}

func setupIntegrationDB(t *testing.T, logger logging.Logger) (*storage.DB, storage.ConnectionType) {
	t.Helper()

	backend := selectedIntegrationBackend()
	ctx := context.Background()

	switch backend {
	case storage.Sqlite:
		// Use a shared in-memory sqlite DB and force a single connection so all queries hit the same DB.
		connStr := defaultSQLiteDSN
		db := storage.NewDB(connStr, storage.Sqlite)
		db.SetMaxOpenConns(1)
		db.SetMaxIdleConns(1)

		m := migrations.NewMigrator(db, storage.Sqlite, logger)
		if err := m.RunMigrations(ctx, db, storage.Sqlite); err != nil {
			if strings.Contains(err.Error(), "go-sqlite3 requires cgo") {
				t.Skip("sqlite backend requires CGO; run with -itdb=postgres and provide -itpgdsn (or MFT_TEST_POSTGRES_DSN)")
			}
			t.Fatalf("failed running sqlite migrations: %v", err)
		}
		return db, storage.Sqlite

	case storage.Postgres:
		dsn := selectedIntegrationPostgresDSN()
		if dsn == "" {
			t.Skip("postgres backend selected but no DSN provided (set -itpgdsn or MFT_TEST_POSTGRES_DSN)")
		}

		db := storage.NewDB(dsn, storage.Postgres)
		m := migrations.NewMigrator(db, storage.Postgres, logger)
		if err := m.EnsureDBExists(ctx, dsn); err != nil {
			t.Fatalf("failed ensuring postgres DB exists: %v", err)
		}
		if err := m.RunMigrations(ctx, db, storage.Postgres); err != nil {
			t.Fatalf("failed running postgres migrations: %v", err)
		}
		if _, err := db.ExecContext(ctx, `
			TRUNCATE TABLE
				portfolio.portfolio_snapshots,
				portfolio.position_snapshots,
				portfolio.positions,
				portfolio.accounts,
				portfolio.transactions,
				cashflow.accounts,
				marketdata.dailies,
				marketdata.listings,
				marketdata.providers,
				cashflow.transactions,
				import.imports,
				import.accounts,
				account.accounts,
				vendor.vendors
			RESTART IDENTITY CASCADE
		`); err != nil {
			t.Fatalf("failed truncating postgres tables: %v", err)
		}
		return db, storage.Postgres

	default:
		t.Fatalf("unsupported integration backend: %s", backend)
		return nil, ""
	}
}

func selectedIntegrationBackend() storage.ConnectionType {
	val := strings.TrimSpace(*itDBFlag)
	if val == "" {
		val = strings.TrimSpace(os.Getenv("MFT_TEST_DB"))
	}
	if val == "" {
		return storage.Sqlite
	}
	val = strings.ToLower(val)
	if val == "sqlite" || val == "sqlite3" {
		return storage.Sqlite
	}
	if val == "postgres" {
		return storage.Postgres
	}
	return storage.Sqlite
}

func selectedIntegrationPostgresDSN() string {
	if dsn := strings.TrimSpace(*itPostgresDSN); dsn != "" {
		return dsn
	}
	if dsn := strings.TrimSpace(os.Getenv("MFT_TEST_POSTGRES_DSN")); dsn != "" {
		return dsn
	}
	return strings.TrimSpace(os.Getenv("DATABASE_URL"))
}

func seedVendorING(t *testing.T, db *storage.DB) *vendor.Vendor {
	t.Helper()
	v, err := vendor.NewVendor(vendor.VendorING, vendor.VendorTypeBank)
	if err != nil {
		t.Fatalf("failed creating vendor seed: %v", err)
	}
	if err := storage.NewSQLXVendorStore(db).Create(context.Background(), v); err != nil {
		t.Fatalf("failed storing vendor seed: %v", err)
	}
	return v
}

func seedImportAccount(t *testing.T, db *storage.DB) *importer.Account {
	t.Helper()
	ctx := context.Background()

	acc, err := account.NewAccount("Integration Test Account", nil)
	if err != nil {
		t.Fatalf("failed creating account seed: %v", err)
	}
	if err := storage.NewSQLXAccountStore(db).Create(ctx, acc); err != nil {
		t.Fatalf("failed storing account seed: %v", err)
	}
	importAcc := importer.NewAccount(acc.ID)
	if err := storage.NewSQLXImportAccountStore(db).Create(ctx, importAcc); err != nil {
		t.Fatalf("failed storing import account seed: %v", err)
	}
	return importAcc
}

func seedMarketStackProvider(t *testing.T, db *storage.DB, baseURI, apiKey string) {
	t.Helper()
	provider, err := marketdata.NewAPIProviderWithAPIKey(marketdata.ProviderMarketStack, baseURI, apiKey)
	if err != nil {
		t.Fatalf("failed creating marketstack provider seed: %v", err)
	}
	if err := storage.NewSQLXProviderStore(db).Create(context.Background(), provider); err != nil {
		t.Fatalf("failed creating marketstack provider seed: %v", err)
	}
}

func seedBrandNewDayManualProvider(t *testing.T, db *storage.DB) {
	t.Helper()
	provider, err := marketdata.NewManualProvider(marketdata.ProviderBrandNewDay)
	if err != nil {
		t.Fatalf("failed creating brandnewday provider seed: %v", err)
	}
	if err := storage.NewSQLXProviderStore(db).Create(context.Background(), provider); err != nil {
		t.Fatalf("failed creating brandnewday provider seed: %v", err)
	}
}

func newMarketStackFixtureServer(t *testing.T) (*httptest.Server, *int64) {
	t.Helper()

	fixture := loadFixture(t, filepath.Join(".fixture-data", "market_stack_eod.json"))
	var calls int64
	handler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/eod" {
			http.NotFound(w, r)
			return
		}

		atomic.AddInt64(&calls, 1)
		w.Header().Set("Content-Type", "application/json")

		if r.URL.Query().Get("offset") == "0" {
			_, _ = w.Write(fixture)
			return
		}
		_, _ = w.Write([]byte(`{"pagination":{"limit":1000,"offset":1000,"count":0,"total":211},"data":[]}`))
	})
	server := httptest.NewServer(handler)
	t.Cleanup(server.Close)
	return server, &calls
}

func loadFixture(t *testing.T, relPath string) []byte {
	t.Helper()
	_, thisFile, _, ok := runtime.Caller(0)
	if !ok {
		t.Fatalf("failed resolving test file path")
	}
	repoRoot := filepath.Join(filepath.Dir(thisFile), "..", "..")
	path := filepath.Join(repoRoot, relPath)
	b, err := os.ReadFile(path)
	if err != nil {
		t.Fatalf("failed reading fixture %s: %v", path, err)
	}
	return b
}

func qualifiedTable(connType storage.ConnectionType, schema, table string) string {
	if connType == storage.Sqlite {
		return table
	}
	return fmt.Sprintf("%s.%s", schema, table)
}
