package main

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"mime/multipart"
	"net/http"
	"net/textproto"
	"testing"

	"github.com/google/uuid"
	"github.com/lennardclaproth/my-finances-tracker/api"
	"github.com/lennardclaproth/my-finances-tracker/internal/account"
	"github.com/lennardclaproth/my-finances-tracker/internal/marketdata"
	"github.com/lennardclaproth/my-finances-tracker/internal/storage"
	"github.com/lennardclaproth/my-finances-tracker/internal/vendor"
)

func TestIntegration_SearchListingsEndpoint_ReturnsFilteredPagination(t *testing.T) {
	app := newIntegrationApp(t)
	ctx := context.Background()

	listingStore := storage.NewSQLXListingStore(app.db)
	l1, err := marketdata.NewListing("AAA.AS", "Alpha Growth", marketdata.SourceAlphaVantage)
	if err != nil {
		t.Fatalf("failed creating listing 1: %v", err)
	}
	l2, err := marketdata.NewListing("BBB.AS", "Beta Income", marketdata.SourceAlphaVantage)
	if err != nil {
		t.Fatalf("failed creating listing 2: %v", err)
	}
	if err := listingStore.Create(ctx, l1); err != nil {
		t.Fatalf("failed storing listing 1: %v", err)
	}
	if err := listingStore.Create(ctx, l2); err != nil {
		t.Fatalf("failed storing listing 2: %v", err)
	}

	req, err := http.NewRequest(http.MethodGet, app.server.URL+"/marketdata/listings/search?q=alpha&limit=25&offset=0", nil)
	if err != nil {
		t.Fatalf("failed creating request: %v", err)
	}
	res, err := http.DefaultClient.Do(req)
	if err != nil {
		t.Fatalf("GET /marketdata/listings/search failed: %v", err)
	}
	defer res.Body.Close()

	if res.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(res.Body)
		t.Fatalf("expected 200, got %d body=%s", res.StatusCode, string(body))
	}

	var payload api.ListingsSearchResponse
	if err := json.NewDecoder(res.Body).Decode(&payload); err != nil {
		t.Fatalf("failed decoding response: %v", err)
	}
	if payload.Pagination.Total != 1 {
		t.Fatalf("expected total 1, got %d", payload.Pagination.Total)
	}
	if len(payload.Data) != 1 {
		t.Fatalf("expected one result, got %d", len(payload.Data))
	}
	if payload.Data[0].Symbol != "AAA.AS" {
		t.Fatalf("expected AAA.AS, got %s", payload.Data[0].Symbol)
	}
}

func TestIntegration_ImportCsvEndpoint_RejectsImportDisabledVendor(t *testing.T) {
	app := newIntegrationApp(t)
	ctx := context.Background()

	v, err := vendor.NewVendor(vendor.VendorING, vendor.VendorTypeBank)
	if err != nil {
		t.Fatalf("failed creating vendor seed: %v", err)
	}
	if err := storage.NewSQLXVendorStore(app.db).Create(ctx, v); err != nil {
		t.Fatalf("failed storing vendor seed: %v", err)
	}

	vendorsTable := qualifiedTable(app.backend, storage.SchemaVendors, storage.TableVendors)
	updateQuery := app.db.Rebind(fmt.Sprintf("UPDATE %s SET import_disabled = ? WHERE id = ?", vendorsTable))
	if _, err := app.db.ExecContext(ctx, updateQuery, true, v.ID); err != nil {
		t.Fatalf("failed to set import_disabled: %v", err)
	}

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

	if res.StatusCode != http.StatusBadRequest {
		body, _ := io.ReadAll(res.Body)
		t.Fatalf("expected 400, got %d body=%s", res.StatusCode, string(body))
	}
	var payload map[string]string
	if err := json.NewDecoder(res.Body).Decode(&payload); err != nil {
		t.Fatalf("failed decoding response body: %v", err)
	}
	if payload["vendor_id"] != "vendor import disabled" {
		t.Fatalf("expected vendor import disabled message, got %+v", payload)
	}
}

func TestIntegration_CreateManualPortfolioTransaction_PersistsManualOriginWithoutRebuild(t *testing.T) {
	app := newIntegrationApp(t)
	ctx := context.Background()

	acc, err := account.NewAccount("manual-portfolio-it", nil)
	if err != nil {
		t.Fatalf("failed creating account: %v", err)
	}
	if err := storage.NewSQLXAccountStore(app.db).Create(ctx, acc); err != nil {
		t.Fatalf("failed storing account: %v", err)
	}

	v, err := vendor.NewVendor(vendor.VendorDEGIRO, vendor.VendorTypeBrokerage)
	if err != nil {
		t.Fatalf("failed creating vendor: %v", err)
	}
	if err := storage.NewSQLXVendorStore(app.db).Create(ctx, v); err != nil {
		t.Fatalf("failed storing vendor: %v", err)
	}

	isin := "NLMANUAL00001"
	listing, err := marketdata.NewListing(
		"MANUAL.AS",
		"Manual Listing",
		marketdata.SourceAlphaVantage,
		marketdata.ListingWithISIN(isin),
	)
	if err != nil {
		t.Fatalf("failed creating listing: %v", err)
	}
	if err := storage.NewSQLXListingStore(app.db).Create(ctx, listing); err != nil {
		t.Fatalf("failed storing listing: %v", err)
	}

	quantity := "2"
	raw, err := json.Marshal(api.CreateManualPortfolioTransactionRequest{
		AccountID:  acc.ID,
		VendorID:   v.ID,
		OccurredAt: "2026-03-01",
		Type:       "BUY",
		ListingID:  &listing.ID,
		Amount:     "100",
		Quantity:   &quantity,
	})
	if err != nil {
		t.Fatalf("failed marshalling request: %v", err)
	}

	req, err := http.NewRequest(http.MethodPost, app.server.URL+"/portfolio/transactions/manual", bytes.NewReader(raw))
	if err != nil {
		t.Fatalf("failed creating request: %v", err)
	}
	req.Header.Set("Content-Type", "application/json")

	res, err := http.DefaultClient.Do(req)
	if err != nil {
		t.Fatalf("POST /portfolio/transactions/manual failed: %v", err)
	}
	defer res.Body.Close()

	if res.StatusCode != http.StatusCreated {
		body, _ := io.ReadAll(res.Body)
		t.Fatalf("expected 201, got %d body=%s", res.StatusCode, string(body))
	}

	var response api.ManualPortfolioTransactionResponse
	if err := json.NewDecoder(res.Body).Decode(&response); err != nil {
		t.Fatalf("failed decoding response: %v", err)
	}
	if response.Origin != "MANUAL" {
		t.Fatalf("expected origin MANUAL, got %s", response.Origin)
	}

	portfolioTxTable := qualifiedTable(app.backend, storage.SchemaPortfolio, storage.TableTransactions)
	if app.backend == storage.Sqlite {
		portfolioTxTable = "portfolio_transactions"
	}
	var persisted struct {
		ID       uuid.UUID  `db:"id"`
		Origin   string     `db:"origin"`
		ImportID *uuid.UUID `db:"import_id"`
	}
	query := app.db.Rebind(fmt.Sprintf("SELECT id, origin, import_id FROM %s WHERE id = ?", portfolioTxTable))
	if err := app.db.GetContext(ctx, &persisted, query, response.ID); err != nil {
		t.Fatalf("failed fetching persisted manual transaction: %v", err)
	}
	if persisted.Origin != "MANUAL" {
		t.Fatalf("expected persisted origin MANUAL, got %s", persisted.Origin)
	}
	if persisted.ImportID != nil {
		t.Fatalf("expected import_id to be NULL for manual transaction")
	}

	positionsTable := qualifiedTable(app.backend, storage.SchemaPortfolio, storage.TablePositions)
	countQuery := app.db.Rebind(fmt.Sprintf("SELECT COUNT(1) FROM %s WHERE account_id = ?", positionsTable))
	var positionsCount int
	if err := app.db.GetContext(ctx, &positionsCount, countQuery, acc.ID); err != nil {
		t.Fatalf("failed counting positions: %v", err)
	}
	if positionsCount != 0 {
		t.Fatalf("expected no auto rebuild side effects, got positions count=%d", positionsCount)
	}
}

func TestIntegration_GetPortfolioTransactionsEndpoint_ReturnsDateFilteredRows(t *testing.T) {
	app := newIntegrationApp(t)
	ctx := context.Background()

	acc, err := account.NewAccount("manual-portfolio-list-it", nil)
	if err != nil {
		t.Fatalf("failed creating account: %v", err)
	}
	if err := storage.NewSQLXAccountStore(app.db).Create(ctx, acc); err != nil {
		t.Fatalf("failed storing account: %v", err)
	}

	v, err := vendor.NewVendor(vendor.VendorDEGIRO, vendor.VendorTypeBrokerage)
	if err != nil {
		t.Fatalf("failed creating vendor: %v", err)
	}
	if err := storage.NewSQLXVendorStore(app.db).Create(ctx, v); err != nil {
		t.Fatalf("failed storing vendor: %v", err)
	}

	isin := "NLMANUAL00002"
	listing, err := marketdata.NewListing(
		"MANUAL2.AS",
		"Manual Listing Two",
		marketdata.SourceAlphaVantage,
		marketdata.ListingWithISIN(isin),
	)
	if err != nil {
		t.Fatalf("failed creating listing: %v", err)
	}
	if err := storage.NewSQLXListingStore(app.db).Create(ctx, listing); err != nil {
		t.Fatalf("failed storing listing: %v", err)
	}

	createManualTx := func(occurredAt, amount, quantity string) {
		raw, err := json.Marshal(api.CreateManualPortfolioTransactionRequest{
			AccountID:  acc.ID,
			VendorID:   v.ID,
			OccurredAt: occurredAt,
			Type:       "BUY",
			ListingID:  &listing.ID,
			Amount:     amount,
			Quantity:   &quantity,
		})
		if err != nil {
			t.Fatalf("failed marshalling request: %v", err)
		}
		req, err := http.NewRequest(http.MethodPost, app.server.URL+"/portfolio/transactions/manual", bytes.NewReader(raw))
		if err != nil {
			t.Fatalf("failed creating request: %v", err)
		}
		req.Header.Set("Content-Type", "application/json")
		res, err := http.DefaultClient.Do(req)
		if err != nil {
			t.Fatalf("POST /portfolio/transactions/manual failed: %v", err)
		}
		defer res.Body.Close()
		if res.StatusCode != http.StatusCreated {
			body, _ := io.ReadAll(res.Body)
			t.Fatalf("expected 201, got %d body=%s", res.StatusCode, string(body))
		}
	}

	createManualTx("2026-01-15", "100", "2")
	createManualTx("2026-03-01", "50", "1")

	req, err := http.NewRequest(http.MethodGet, app.server.URL+"/portfolio/transactions?account_id="+acc.ID.String()+"&from=2026-02-01&to=2026-03-31", nil)
	if err != nil {
		t.Fatalf("failed creating request: %v", err)
	}
	res, err := http.DefaultClient.Do(req)
	if err != nil {
		t.Fatalf("GET /portfolio/transactions failed: %v", err)
	}
	defer res.Body.Close()

	if res.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(res.Body)
		t.Fatalf("expected 200, got %d body=%s", res.StatusCode, string(body))
	}

	var payload api.PortfolioTransactionsResponse
	if err := json.NewDecoder(res.Body).Decode(&payload); err != nil {
		t.Fatalf("failed decoding response: %v", err)
	}
	if len(payload.Data) != 1 {
		t.Fatalf("expected 1 transaction after date filter, got %d", len(payload.Data))
	}
	if payload.Data[0].OccurredAt.Format("2006-01-02") != "2026-03-01" {
		t.Fatalf("expected occurred_at 2026-03-01, got %s", payload.Data[0].OccurredAt.Format("2006-01-02"))
	}
}

func TestIntegration_GetPortfolioTransactionsEndpoint_UnknownAccountReturns404(t *testing.T) {
	app := newIntegrationApp(t)

	req, err := http.NewRequest(http.MethodGet, app.server.URL+"/portfolio/transactions?account_id="+uuid.New().String(), nil)
	if err != nil {
		t.Fatalf("failed creating request: %v", err)
	}
	res, err := http.DefaultClient.Do(req)
	if err != nil {
		t.Fatalf("GET /portfolio/transactions failed: %v", err)
	}
	defer res.Body.Close()

	if res.StatusCode != http.StatusNotFound {
		body, _ := io.ReadAll(res.Body)
		t.Fatalf("expected 404, got %d body=%s", res.StatusCode, string(body))
	}
}
