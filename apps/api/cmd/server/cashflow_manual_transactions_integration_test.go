package main

import (
	"bytes"
	"context"
	"encoding/json"
	"io"
	"net/http"
	"testing"

	"github.com/google/uuid"
	"github.com/lennardclaproth/my-finances-tracker/api"
	"github.com/lennardclaproth/my-finances-tracker/internal/account"
	"github.com/lennardclaproth/my-finances-tracker/internal/storage"
)

func TestIntegration_CreateManualCashflowTransactionsEndpoint_HappyPath(t *testing.T) {
	app := newIntegrationApp(t)
	ctx := context.Background()

	seedVendorING(t, app.db)

	acc, err := account.NewAccount("manual-cashflow-account", nil)
	if err != nil {
		t.Fatalf("failed creating account seed: %v", err)
	}
	if err := storage.NewSQLXAccountStore(app.db).Create(ctx, acc); err != nil {
		t.Fatalf("failed storing account seed: %v", err)
	}

	payload := api.CreateManualCashflowTransactionsRequest{
		AccountID: acc.ID,
		Transactions: []api.CreateManualCashflowTransactionEntryRequest{
			{
				Date:        "2026-03-09",
				Amount:      "12.35",
				Type:        "out",
				Description: "Lunch",
				Note:        "Office lunch",
				Tag:         "food",
				Vendor:      "Cash",
			},
			{
				Date:        "2026-03-10",
				Amount:      "2100.00",
				Type:        "in",
				Description: "Salary",
				Note:        "Monthly salary",
				Tag:         "income",
				Vendor:      "Employer",
			},
		},
	}
	body, err := json.Marshal(payload)
	if err != nil {
		t.Fatalf("failed marshalling payload: %v", err)
	}
	req, err := http.NewRequest(http.MethodPost, app.server.URL+"/cashflow/transactions/manual", bytes.NewReader(body))
	if err != nil {
		t.Fatalf("failed creating request: %v", err)
	}
	req.Header.Set("Content-Type", "application/json")

	res, err := http.DefaultClient.Do(req)
	if err != nil {
		t.Fatalf("POST /cashflow/transactions/manual failed: %v", err)
	}
	defer func() {
		if err := res.Body.Close(); err != nil {
			t.Fatalf("failed closing response body: %v", err)
		}
	}()

	if res.StatusCode != http.StatusCreated {
		raw, _ := io.ReadAll(res.Body)
		t.Fatalf("expected status 201, got %d body=%s", res.StatusCode, string(raw))
	}

	var createResponse api.ManualCashflowTransactionsResponse
	if err := json.NewDecoder(res.Body).Decode(&createResponse); err != nil {
		t.Fatalf("failed decoding create response: %v", err)
	}
	if createResponse.CreatedCount != 2 || len(createResponse.Data) != 2 {
		t.Fatalf("expected created_count=2 with 2 rows, got created_count=%d len(data)=%d", createResponse.CreatedCount, len(createResponse.Data))
	}
	if createResponse.Data[0].Source != "manual:Cash" {
		t.Fatalf("expected first source manual:Cash, got %s", createResponse.Data[0].Source)
	}
	if createResponse.Data[1].Source != "manual:Employer" {
		t.Fatalf("expected second source manual:Employer, got %s", createResponse.Data[1].Source)
	}

	fetchReq, err := http.NewRequest(
		http.MethodGet,
		app.server.URL+"/cashflow/transactions?source=manual&limit=10&offset=0",
		nil,
	)
	if err != nil {
		t.Fatalf("failed creating fetch request: %v", err)
	}
	fetchRes, err := http.DefaultClient.Do(fetchReq)
	if err != nil {
		t.Fatalf("GET /cashflow/transactions failed: %v", err)
	}
	defer func() {
		if err := fetchRes.Body.Close(); err != nil {
			t.Fatalf("failed closing fetch response body: %v", err)
		}
	}()
	if fetchRes.StatusCode != http.StatusOK {
		raw, _ := io.ReadAll(fetchRes.Body)
		t.Fatalf("expected status 200, got %d body=%s", fetchRes.StatusCode, string(raw))
	}

	var fetched api.CashflowTransactionsResponse
	if err := json.NewDecoder(fetchRes.Body).Decode(&fetched); err != nil {
		t.Fatalf("failed decoding fetch response: %v", err)
	}
	if fetched.Pagination.Total != 2 {
		t.Fatalf("expected total=2 for source=manual filter, got %d", fetched.Pagination.Total)
	}
}

func TestIntegration_CreateManualCashflowTransactionsEndpoint_UnknownAccount(t *testing.T) {
	app := newIntegrationApp(t)
	seedVendorING(t, app.db)

	payload := api.CreateManualCashflowTransactionsRequest{
		AccountID: uuid.New(),
		Transactions: []api.CreateManualCashflowTransactionEntryRequest{
			{
				Date:        "2026-03-10",
				Amount:      "12.35",
				Type:        "out",
				Description: "Lunch",
				Note:        "Office lunch",
				Tag:         "food",
			},
		},
	}
	body, err := json.Marshal(payload)
	if err != nil {
		t.Fatalf("failed marshalling payload: %v", err)
	}
	req, err := http.NewRequest(http.MethodPost, app.server.URL+"/cashflow/transactions/manual", bytes.NewReader(body))
	if err != nil {
		t.Fatalf("failed creating request: %v", err)
	}
	req.Header.Set("Content-Type", "application/json")

	res, err := http.DefaultClient.Do(req)
	if err != nil {
		t.Fatalf("POST /cashflow/transactions/manual failed: %v", err)
	}
	defer func() {
		if err := res.Body.Close(); err != nil {
			t.Fatalf("failed closing response body: %v", err)
		}
	}()

	if res.StatusCode != http.StatusNotFound {
		raw, _ := io.ReadAll(res.Body)
		t.Fatalf("expected status 404, got %d body=%s", res.StatusCode, string(raw))
	}
}
