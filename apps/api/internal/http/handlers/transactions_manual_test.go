package handlers

import (
	"bytes"
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/lennardclaproth/my-finances-tracker/api"
	"github.com/lennardclaproth/my-finances-tracker/internal/cashflow"
	cashflowservice "github.com/lennardclaproth/my-finances-tracker/internal/cashflow/service"
	"github.com/lennardclaproth/my-finances-tracker/internal/money"
)

type fakeManualCashflowCreateService struct {
	createManyFn func(ctx context.Context, input cashflowservice.ManualCashflowCreateInput) (*cashflowservice.ManualCashflowCreateResult, error)
}

func (f *fakeManualCashflowCreateService) CreateMany(ctx context.Context, input cashflowservice.ManualCashflowCreateInput) (*cashflowservice.ManualCashflowCreateResult, error) {
	if f.createManyFn != nil {
		return f.createManyFn(ctx, input)
	}
	return nil, nil
}

func TestCreateManualCashflowTransactions_Success(t *testing.T) {
	accID := uuid.New()
	amount, _ := money.NewPrice(10.50)
	occurredAt := time.Date(2026, 3, 10, 0, 0, 0, 0, time.UTC)
	txID := uuid.New()

	svc := &fakeManualCashflowCreateService{
		createManyFn: func(ctx context.Context, input cashflowservice.ManualCashflowCreateInput) (*cashflowservice.ManualCashflowCreateResult, error) {
			if input.AccountID != accID {
				t.Fatalf("expected account_id %s, got %s", accID, input.AccountID)
			}
			if len(input.Transactions) != 1 {
				t.Fatalf("expected 1 transaction, got %d", len(input.Transactions))
			}
			return &cashflowservice.ManualCashflowCreateResult{
				Transactions: []*cashflow.Transaction{
					{
						ID:          txID,
						Description: "Coffee",
						Note:        "Morning coffee",
						Source:      "manual:Cash",
						AmountCents: amount,
						Direction:   cashflow.CashOut,
						Date:        occurredAt,
						Tag:         "food",
					},
				},
			}, nil
		},
	}

	body, _ := json.Marshal(api.CreateManualCashflowTransactionsRequest{
		AccountID: accID,
		Transactions: []api.CreateManualCashflowTransactionEntryRequest{
			{
				Date:        "2026-03-10",
				Amount:      "10.50",
				Type:        "out",
				Description: "Coffee",
				Note:        "Morning coffee",
				Tag:         "food",
				Vendor:      "Cash",
			},
		},
	})
	req := httptest.NewRequest(http.MethodPost, "/cashflow/transactions/manual", bytes.NewReader(body))
	rec := httptest.NewRecorder()

	CreateManualCashflowTransactions(&testLogger{}, svc).ServeHTTP(rec, req)

	if rec.Code != http.StatusCreated {
		t.Fatalf("expected 201, got %d body=%s", rec.Code, rec.Body.String())
	}

	var response api.ManualCashflowTransactionsResponse
	if err := json.Unmarshal(rec.Body.Bytes(), &response); err != nil {
		t.Fatalf("failed to decode response: %v", err)
	}
	if response.CreatedCount != 1 {
		t.Fatalf("expected created_count=1, got %d", response.CreatedCount)
	}
	if len(response.Data) != 1 {
		t.Fatalf("expected one transaction in response, got %d", len(response.Data))
	}
	if response.Data[0].ID != txID {
		t.Fatalf("expected response transaction id %s, got %s", txID, response.Data[0].ID)
	}
}

func TestCreateManualCashflowTransactions_BadRequest(t *testing.T) {
	svc := &fakeManualCashflowCreateService{
		createManyFn: func(ctx context.Context, input cashflowservice.ManualCashflowCreateInput) (*cashflowservice.ManualCashflowCreateResult, error) {
			return nil, cashflowservice.ErrManualCashflowInvalidAmount
		},
	}

	body, _ := json.Marshal(api.CreateManualCashflowTransactionsRequest{
		AccountID: uuid.New(),
		Transactions: []api.CreateManualCashflowTransactionEntryRequest{
			{
				Date:        "2026-03-10",
				Amount:      "0",
				Type:        "out",
				Description: "Coffee",
				Note:        "Morning coffee",
				Tag:         "food",
			},
		},
	})
	req := httptest.NewRequest(http.MethodPost, "/cashflow/transactions/manual", bytes.NewReader(body))
	rec := httptest.NewRecorder()

	CreateManualCashflowTransactions(&testLogger{}, svc).ServeHTTP(rec, req)

	if rec.Code != http.StatusBadRequest {
		t.Fatalf("expected 400, got %d body=%s", rec.Code, rec.Body.String())
	}
}

func TestCreateManualCashflowTransactions_AccountNotFound(t *testing.T) {
	svc := &fakeManualCashflowCreateService{
		createManyFn: func(ctx context.Context, input cashflowservice.ManualCashflowCreateInput) (*cashflowservice.ManualCashflowCreateResult, error) {
			return nil, cashflowservice.ErrManualCashflowAccountNotFound
		},
	}

	body, _ := json.Marshal(api.CreateManualCashflowTransactionsRequest{
		AccountID: uuid.New(),
		Transactions: []api.CreateManualCashflowTransactionEntryRequest{
			{
				Date:        "2026-03-10",
				Amount:      "10",
				Type:        "out",
				Description: "Coffee",
				Note:        "Morning coffee",
				Tag:         "food",
			},
		},
	})
	req := httptest.NewRequest(http.MethodPost, "/cashflow/transactions/manual", bytes.NewReader(body))
	rec := httptest.NewRecorder()

	CreateManualCashflowTransactions(&testLogger{}, svc).ServeHTTP(rec, req)

	if rec.Code != http.StatusNotFound {
		t.Fatalf("expected 404, got %d body=%s", rec.Code, rec.Body.String())
	}
}

func TestCreateManualCashflowTransactions_Duplicate(t *testing.T) {
	svc := &fakeManualCashflowCreateService{
		createManyFn: func(ctx context.Context, input cashflowservice.ManualCashflowCreateInput) (*cashflowservice.ManualCashflowCreateResult, error) {
			return nil, cashflow.ErrDuplicateTransaction
		},
	}

	body, _ := json.Marshal(api.CreateManualCashflowTransactionsRequest{
		AccountID: uuid.New(),
		Transactions: []api.CreateManualCashflowTransactionEntryRequest{
			{
				Date:        "2026-03-10",
				Amount:      "10",
				Type:        "out",
				Description: "Coffee",
				Note:        "Morning coffee",
				Tag:         "food",
			},
		},
	})
	req := httptest.NewRequest(http.MethodPost, "/cashflow/transactions/manual", bytes.NewReader(body))
	rec := httptest.NewRecorder()

	CreateManualCashflowTransactions(&testLogger{}, svc).ServeHTTP(rec, req)

	if rec.Code != http.StatusConflict {
		t.Fatalf("expected 409, got %d body=%s", rec.Code, rec.Body.String())
	}
}
