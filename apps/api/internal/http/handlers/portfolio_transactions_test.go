package handlers

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/lennardclaproth/my-finances-tracker/api"
	"github.com/lennardclaproth/my-finances-tracker/internal/money"
	"github.com/lennardclaproth/my-finances-tracker/internal/portfolio"
)

type fakeManualPortfolioTransactionService struct {
	createFn func(ctx context.Context, input portfolio.ManualTransactionInput) (*portfolio.ManualTransactionCreateResult, error)
}

func (f *fakeManualPortfolioTransactionService) Create(ctx context.Context, input portfolio.ManualTransactionInput) (*portfolio.ManualTransactionCreateResult, error) {
	if f.createFn != nil {
		return f.createFn(ctx, input)
	}
	return nil, nil
}

func TestCreateManualPortfolioTransaction_Success(t *testing.T) {
	accID := uuid.New()
	listingID := uuid.New()
	isin := "NLTEST0001"
	symbol := "TST.AS"
	occurredAt := time.Date(2026, 3, 1, 0, 0, 0, 0, time.UTC)
	unitPrice, _ := money.NewPrice(50)
	amount, _ := money.NewPrice(100)

	svc := &fakeManualPortfolioTransactionService{
		createFn: func(ctx context.Context, input portfolio.ManualTransactionInput) (*portfolio.ManualTransactionCreateResult, error) {
			return &portfolio.ManualTransactionCreateResult{
				Transaction: &portfolio.Transaction{
					ID:          uuid.New(),
					AccountID:   &accID,
					Origin:      portfolio.TransactionOriginManual,
					Source:      "DEGIRO",
					OccurredAt:  occurredAt,
					Type:        portfolio.TxBuy,
					ISIN:        &isin,
					Symbol:      &symbol,
					Description: "manual buy",
					Quantity:    2,
					UnitPrice:   unitPrice,
					AmountCents: amount,
					CreatedAt:   occurredAt,
					UpdatedAt:   occurredAt,
				},
				ListingID:    &listingID,
				SignedAmount: 100,
			}, nil
		},
	}

	raw, _ := json.Marshal(api.CreateManualPortfolioTransactionRequest{
		AccountID:  accID,
		VendorID:   uuid.New(),
		OccurredAt: "2026-03-01",
		Type:       "BUY",
		ListingID:  &listingID,
		Amount:     "100",
		Quantity:   ptrString("2"),
	})
	req := httptest.NewRequest(http.MethodPost, "/portfolio/transactions/manual", bytes.NewReader(raw))
	rec := httptest.NewRecorder()

	CreateManualPortfolioTransaction(&testLogger{}, svc).ServeHTTP(rec, req)

	if rec.Code != http.StatusCreated {
		t.Fatalf("expected 201, got %d body=%s", rec.Code, rec.Body.String())
	}

	var response api.ManualPortfolioTransactionResponse
	if err := json.Unmarshal(rec.Body.Bytes(), &response); err != nil {
		t.Fatalf("failed to decode response: %v", err)
	}
	if response.Origin != string(portfolio.TransactionOriginManual) {
		t.Fatalf("expected origin MANUAL, got %s", response.Origin)
	}
	if response.Amount != "100" {
		t.Fatalf("expected amount 100, got %s", response.Amount)
	}
	if response.UnitPrice != "50" {
		t.Fatalf("expected unit_price 50, got %s", response.UnitPrice)
	}
}

func TestCreateManualPortfolioTransaction_Duplicate(t *testing.T) {
	svc := &fakeManualPortfolioTransactionService{
		createFn: func(ctx context.Context, input portfolio.ManualTransactionInput) (*portfolio.ManualTransactionCreateResult, error) {
			return nil, portfolio.ErrDuplicateTransaction
		},
	}

	raw, _ := json.Marshal(api.CreateManualPortfolioTransactionRequest{
		AccountID:  uuid.New(),
		VendorID:   uuid.New(),
		OccurredAt: "2026-03-01",
		Type:       "CASH",
		Amount:     "-1",
	})
	req := httptest.NewRequest(http.MethodPost, "/portfolio/transactions/manual", bytes.NewReader(raw))
	rec := httptest.NewRecorder()

	CreateManualPortfolioTransaction(&testLogger{}, svc).ServeHTTP(rec, req)

	if rec.Code != http.StatusConflict {
		t.Fatalf("expected 409, got %d body=%s", rec.Code, rec.Body.String())
	}
}

func TestCreateManualPortfolioTransaction_VendorTypeRejected(t *testing.T) {
	svc := &fakeManualPortfolioTransactionService{
		createFn: func(ctx context.Context, input portfolio.ManualTransactionInput) (*portfolio.ManualTransactionCreateResult, error) {
			return nil, portfolio.ErrManualVendorTypeNotSupported
		},
	}

	raw, _ := json.Marshal(api.CreateManualPortfolioTransactionRequest{
		AccountID:  uuid.New(),
		VendorID:   uuid.New(),
		OccurredAt: "2026-03-01",
		Type:       "CASH",
		Amount:     "-1",
	})
	req := httptest.NewRequest(http.MethodPost, "/portfolio/transactions/manual", bytes.NewReader(raw))
	rec := httptest.NewRecorder()

	CreateManualPortfolioTransaction(&testLogger{}, svc).ServeHTTP(rec, req)

	if rec.Code != http.StatusUnprocessableEntity {
		t.Fatalf("expected 422, got %d body=%s", rec.Code, rec.Body.String())
	}
}

func TestCreateManualPortfolioTransaction_BadRequestValidation(t *testing.T) {
	svc := &fakeManualPortfolioTransactionService{
		createFn: func(ctx context.Context, input portfolio.ManualTransactionInput) (*portfolio.ManualTransactionCreateResult, error) {
			return nil, errors.New("should not be called")
		},
	}

	raw, _ := json.Marshal(api.CreateManualPortfolioTransactionRequest{
		AccountID: uuid.Nil,
	})
	req := httptest.NewRequest(http.MethodPost, "/portfolio/transactions/manual", bytes.NewReader(raw))
	rec := httptest.NewRecorder()

	CreateManualPortfolioTransaction(&testLogger{}, svc).ServeHTTP(rec, req)

	if rec.Code != http.StatusBadRequest {
		t.Fatalf("expected 400, got %d body=%s", rec.Code, rec.Body.String())
	}
}
