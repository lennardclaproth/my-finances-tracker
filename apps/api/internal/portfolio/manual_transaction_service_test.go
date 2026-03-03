package portfolio

import (
	"context"
	"errors"
	"math"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/lennardclaproth/my-finances-tracker/internal/account"
	"github.com/lennardclaproth/my-finances-tracker/internal/marketdata"
	"github.com/lennardclaproth/my-finances-tracker/internal/vendor"
)

type fakeManualAccountFetcher struct {
	fetchByIDFn func(ctx context.Context, id uuid.UUID) (*account.Account, error)
}

func (f *fakeManualAccountFetcher) FetchByID(ctx context.Context, id uuid.UUID) (*account.Account, error) {
	if f.fetchByIDFn != nil {
		return f.fetchByIDFn(ctx, id)
	}
	return nil, account.ErrAccountNotFound
}

type fakeManualVendorFetcher struct {
	fetchByIDFn func(ctx context.Context, id uuid.UUID) (*vendor.Vendor, error)
}

func (f *fakeManualVendorFetcher) FetchById(ctx context.Context, id uuid.UUID) (*vendor.Vendor, error) {
	if f.fetchByIDFn != nil {
		return f.fetchByIDFn(ctx, id)
	}
	return nil, vendor.ErrVendorNotFound
}

type fakeManualListingFetcher struct {
	fetchByIDFn func(ctx context.Context, id uuid.UUID) (*marketdata.Listing, error)
}

func (f *fakeManualListingFetcher) FetchByID(ctx context.Context, id uuid.UUID) (*marketdata.Listing, error) {
	if f.fetchByIDFn != nil {
		return f.fetchByIDFn(ctx, id)
	}
	return nil, nil
}

type fakeManualTxStore struct {
	createFn func(ctx context.Context, tx *Transaction) error
	lastTx   *Transaction
}

func (f *fakeManualTxStore) Create(ctx context.Context, tx *Transaction) error {
	f.lastTx = tx
	if f.createFn != nil {
		return f.createFn(ctx, tx)
	}
	return nil
}

func TestManualTransactionService_CreateBuy_Success(t *testing.T) {
	accountID := uuid.New()
	vendorID := uuid.New()
	listingID := uuid.New()
	isin := "NL000TEST0001"

	txStore := &fakeManualTxStore{}
	svc := NewManualTransactionService(
		&fakeManualAccountFetcher{
			fetchByIDFn: func(ctx context.Context, id uuid.UUID) (*account.Account, error) {
				return &account.Account{ID: accountID}, nil
			},
		},
		&fakeManualVendorFetcher{
			fetchByIDFn: func(ctx context.Context, id uuid.UUID) (*vendor.Vendor, error) {
				return &vendor.Vendor{ID: vendorID, Name: vendor.VendorDEGIRO, Type: vendor.VendorTypeBrokerage, Active: true}, nil
			},
		},
		&fakeManualListingFetcher{
			fetchByIDFn: func(ctx context.Context, id uuid.UUID) (*marketdata.Listing, error) {
				return &marketdata.Listing{ID: listingID, Symbol: "TST.AS", ISIN: &isin}, nil
			},
		},
		txStore,
	)

	quantity := "2"
	result, err := svc.Create(context.Background(), ManualTransactionInput{
		AccountID:  accountID,
		VendorID:   vendorID,
		OccurredAt: "2026-03-01",
		Type:       string(TxBuy),
		ListingID:  &listingID,
		Amount:     "100",
		Quantity:   &quantity,
	})
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}
	if result == nil || result.Transaction == nil {
		t.Fatalf("expected transaction result")
	}
	tx := result.Transaction
	if tx.Origin != TransactionOriginManual {
		t.Fatalf("expected origin MANUAL, got %s", tx.Origin)
	}
	if tx.ImportID != nil {
		t.Fatalf("expected import_id to be nil for manual transaction")
	}
	if tx.Source != string(vendor.VendorDEGIRO) {
		t.Fatalf("expected source %s, got %s", vendor.VendorDEGIRO, tx.Source)
	}
	if tx.ISIN == nil || *tx.ISIN != isin {
		t.Fatalf("expected isin to be propagated from listing")
	}
	if math.Abs(tx.UnitPrice.Float64()-50) > 1e-9 {
		t.Fatalf("expected unit price 50, got %f", tx.UnitPrice.Float64())
	}
	if tx.AmountCents.Float64() != 100 {
		t.Fatalf("expected amount 100, got %f", tx.AmountCents.Float64())
	}
	if tx.OccurredAt.Format("2006-01-02") != "2026-03-01" {
		t.Fatalf("expected occurred_at 2026-03-01, got %s", tx.OccurredAt.Format("2006-01-02"))
	}
}

func TestManualTransactionService_CreateCash_SignedAmountDirection(t *testing.T) {
	accountID := uuid.New()
	vendorID := uuid.New()
	txStore := &fakeManualTxStore{}
	svc := NewManualTransactionService(
		&fakeManualAccountFetcher{
			fetchByIDFn: func(ctx context.Context, id uuid.UUID) (*account.Account, error) {
				return &account.Account{ID: accountID}, nil
			},
		},
		&fakeManualVendorFetcher{
			fetchByIDFn: func(ctx context.Context, id uuid.UUID) (*vendor.Vendor, error) {
				return &vendor.Vendor{ID: vendorID, Name: vendor.VendorDEGIRO, Type: vendor.VendorTypeBrokerage, Active: true}, nil
			},
		},
		&fakeManualListingFetcher{},
		txStore,
	)

	result, err := svc.Create(context.Background(), ManualTransactionInput{
		AccountID:  accountID,
		VendorID:   vendorID,
		OccurredAt: "2026-03-01",
		Type:       string(TxCash),
		Amount:     "-12.5",
	})
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}
	if result.SignedAmount != -12.5 {
		t.Fatalf("expected signed amount -12.5, got %f", result.SignedAmount)
	}
	if txStore.lastTx == nil {
		t.Fatalf("expected created transaction")
	}
	if txStore.lastTx.Quantity != -1 {
		t.Fatalf("expected cash quantity marker -1, got %f", txStore.lastTx.Quantity)
	}
	if txStore.lastTx.ISIN != nil || txStore.lastTx.Symbol != nil {
		t.Fatalf("expected no listing identity for cash transaction")
	}
}

func TestManualTransactionService_RejectsNonBrokerageVendor(t *testing.T) {
	accountID := uuid.New()
	vendorID := uuid.New()
	svc := NewManualTransactionService(
		&fakeManualAccountFetcher{
			fetchByIDFn: func(ctx context.Context, id uuid.UUID) (*account.Account, error) {
				return &account.Account{ID: accountID}, nil
			},
		},
		&fakeManualVendorFetcher{
			fetchByIDFn: func(ctx context.Context, id uuid.UUID) (*vendor.Vendor, error) {
				return &vendor.Vendor{
					ID:             vendorID,
					Name:           vendor.VendorING,
					Type:           vendor.VendorTypeBank,
					Active:         true,
					CreatedAt:      time.Now().UTC(),
					UpdatedAt:      time.Now().UTC(),
					ImportDisabled: false,
				}, nil
			},
		},
		&fakeManualListingFetcher{},
		&fakeManualTxStore{},
	)

	_, err := svc.Create(context.Background(), ManualTransactionInput{
		AccountID:  accountID,
		VendorID:   vendorID,
		OccurredAt: "2026-03-01",
		Type:       string(TxCash),
		Amount:     "10",
	})
	if !errors.Is(err, ErrManualVendorTypeNotSupported) {
		t.Fatalf("expected ErrManualVendorTypeNotSupported, got %v", err)
	}
}

func TestManualTransactionService_DuplicatePropagates(t *testing.T) {
	accountID := uuid.New()
	vendorID := uuid.New()
	svc := NewManualTransactionService(
		&fakeManualAccountFetcher{
			fetchByIDFn: func(ctx context.Context, id uuid.UUID) (*account.Account, error) {
				return &account.Account{ID: accountID}, nil
			},
		},
		&fakeManualVendorFetcher{
			fetchByIDFn: func(ctx context.Context, id uuid.UUID) (*vendor.Vendor, error) {
				return &vendor.Vendor{ID: vendorID, Name: vendor.VendorDEGIRO, Type: vendor.VendorTypeBrokerage, Active: true}, nil
			},
		},
		&fakeManualListingFetcher{},
		&fakeManualTxStore{
			createFn: func(ctx context.Context, tx *Transaction) error {
				return ErrDuplicateTransaction
			},
		},
	)

	_, err := svc.Create(context.Background(), ManualTransactionInput{
		AccountID:  accountID,
		VendorID:   vendorID,
		OccurredAt: "2026-03-01",
		Type:       string(TxCash),
		Amount:     "-1",
	})
	if !errors.Is(err, ErrDuplicateTransaction) {
		t.Fatalf("expected ErrDuplicateTransaction, got %v", err)
	}
}
