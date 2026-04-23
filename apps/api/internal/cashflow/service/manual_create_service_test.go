package service

import (
	"context"
	"errors"
	"testing"

	"github.com/google/uuid"
	"github.com/lennardclaproth/my-finances-tracker/internal/account"
	"github.com/lennardclaproth/my-finances-tracker/internal/cashflow"
	"github.com/lennardclaproth/my-finances-tracker/internal/importer"
	"github.com/lennardclaproth/my-finances-tracker/internal/vendor"
)

type fakeManualCashflowAccountFetcher struct {
	fetchFn func(ctx context.Context, id uuid.UUID) (*account.Account, error)
}

func (f *fakeManualCashflowAccountFetcher) FetchByID(ctx context.Context, id uuid.UUID) (*account.Account, error) {
	if f.fetchFn != nil {
		return f.fetchFn(ctx, id)
	}
	return nil, account.ErrAccountNotFound
}

type fakeManualCashflowProjectionCreator struct {
	createFn func(ctx context.Context, acc *cashflow.Account) error
}

func (f *fakeManualCashflowProjectionCreator) Create(ctx context.Context, acc *cashflow.Account) error {
	if f.createFn != nil {
		return f.createFn(ctx, acc)
	}
	return nil
}

type fakeManualImportAccountCreator struct {
	createFn func(ctx context.Context, acc *importer.Account) error
}

func (f *fakeManualImportAccountCreator) Create(ctx context.Context, acc *importer.Account) error {
	if f.createFn != nil {
		return f.createFn(ctx, acc)
	}
	return nil
}

type fakeManualCashflowVendorStore struct {
	fetchByNameFn func(ctx context.Context, name vendor.VendorID) (*vendor.Vendor, error)
	listActiveFn  func(ctx context.Context) ([]*vendor.Vendor, error)
}

func (f *fakeManualCashflowVendorStore) FetchByName(ctx context.Context, name vendor.VendorID) (*vendor.Vendor, error) {
	if f.fetchByNameFn != nil {
		return f.fetchByNameFn(ctx, name)
	}
	return nil, vendor.ErrVendorNotFound
}

func (f *fakeManualCashflowVendorStore) ListActive(ctx context.Context) ([]*vendor.Vendor, error) {
	if f.listActiveFn != nil {
		return f.listActiveFn(ctx)
	}
	return nil, nil
}

type fakeManualCashflowImportCreator struct {
	createFn func(ctx context.Context, imp *importer.Import) error
}

func (f *fakeManualCashflowImportCreator) Create(ctx context.Context, imp *importer.Import) error {
	if f.createFn != nil {
		return f.createFn(ctx, imp)
	}
	return nil
}

type fakeManualCashflowTransactionCreator struct {
	createFn func(ctx context.Context, tx *cashflow.Transaction) error
}

func (f *fakeManualCashflowTransactionCreator) Create(ctx context.Context, tx *cashflow.Transaction) error {
	if f.createFn != nil {
		return f.createFn(ctx, tx)
	}
	return nil
}

func TestManualCreateService_CreateManySuccess(t *testing.T) {
	accID := uuid.New()
	vendorID := uuid.New()

	var imported *importer.Import
	persistedTransactions := make([]*cashflow.Transaction, 0)

	svc := NewManualCreateService(
		&fakeManualCashflowAccountFetcher{
			fetchFn: func(ctx context.Context, id uuid.UUID) (*account.Account, error) {
				return &account.Account{ID: id, Name: "cashflow-account"}, nil
			},
		},
		&fakeManualCashflowProjectionCreator{},
		&fakeManualImportAccountCreator{},
		&fakeManualCashflowVendorStore{
			fetchByNameFn: func(ctx context.Context, name vendor.VendorID) (*vendor.Vendor, error) {
				return &vendor.Vendor{ID: vendorID, Name: vendor.VendorING, Active: true, Type: vendor.VendorTypeBank}, nil
			},
		},
		&fakeManualCashflowImportCreator{
			createFn: func(ctx context.Context, imp *importer.Import) error {
				imported = imp
				return nil
			},
		},
		&fakeManualCashflowTransactionCreator{
			createFn: func(ctx context.Context, tx *cashflow.Transaction) error {
				persistedTransactions = append(persistedTransactions, tx)
				return nil
			},
		},
	)

	result, err := svc.CreateMany(context.Background(), ManualCashflowCreateInput{
		AccountID: accID,
		Transactions: []ManualCashflowCreateTransactionInput{
			{
				Date:        "2026-03-08",
				Amount:      "10.50",
				Type:        "out",
				Description: "Coffee",
				Note:        "Barista",
				Tag:         "food",
				Vendor:      "Cash",
			},
			{
				Date:        "2026-03-09",
				Amount:      "1000",
				Type:        "in",
				Description: "Salary",
				Note:        "March salary",
				Tag:         "income",
			},
		},
	})
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}
	if result == nil || len(result.Transactions) != 2 {
		t.Fatalf("expected 2 transactions in result, got %+v", result)
	}
	if imported == nil {
		t.Fatalf("expected import row to be created")
	}
	if imported.Status != importer.ImportStatusCompleted {
		t.Fatalf("expected import status completed, got %s", imported.Status)
	}
	if imported.Imported != 2 || imported.TotalRows != 2 {
		t.Fatalf("expected import counts imported=2 total_rows=2, got imported=%d total_rows=%d", imported.Imported, imported.TotalRows)
	}
	if len(persistedTransactions) != 2 {
		t.Fatalf("expected 2 persisted transactions, got %d", len(persistedTransactions))
	}
	if persistedTransactions[0].Source != "manual:Cash" {
		t.Fatalf("expected first source manual:Cash, got %s", persistedTransactions[0].Source)
	}
	if persistedTransactions[1].Source != "manual" {
		t.Fatalf("expected second source manual, got %s", persistedTransactions[1].Source)
	}
	if persistedTransactions[0].Tag != "food" || persistedTransactions[1].Tag != "income" {
		t.Fatalf("expected tags to be persisted, got first=%s second=%s", persistedTransactions[0].Tag, persistedTransactions[1].Tag)
	}
}

func TestManualCreateService_AccountNotFound(t *testing.T) {
	svc := NewManualCreateService(
		&fakeManualCashflowAccountFetcher{
			fetchFn: func(ctx context.Context, id uuid.UUID) (*account.Account, error) {
				return nil, account.ErrAccountNotFound
			},
		},
		&fakeManualCashflowProjectionCreator{},
		&fakeManualImportAccountCreator{},
		&fakeManualCashflowVendorStore{},
		&fakeManualCashflowImportCreator{},
		&fakeManualCashflowTransactionCreator{},
	)

	_, err := svc.CreateMany(context.Background(), ManualCashflowCreateInput{
		AccountID: uuid.New(),
		Transactions: []ManualCashflowCreateTransactionInput{
			{
				Date:        "2026-03-08",
				Amount:      "10",
				Type:        "out",
				Description: "Coffee",
				Note:        "Barista",
				Tag:         "food",
			},
		},
	})
	if !errors.Is(err, ErrManualCashflowAccountNotFound) {
		t.Fatalf("expected ErrManualCashflowAccountNotFound, got %v", err)
	}
}

func TestManualCreateService_InvalidAmount(t *testing.T) {
	svc := NewManualCreateService(
		&fakeManualCashflowAccountFetcher{
			fetchFn: func(ctx context.Context, id uuid.UUID) (*account.Account, error) {
				return &account.Account{ID: id, Name: "ok"}, nil
			},
		},
		&fakeManualCashflowProjectionCreator{},
		&fakeManualImportAccountCreator{},
		&fakeManualCashflowVendorStore{},
		&fakeManualCashflowImportCreator{},
		&fakeManualCashflowTransactionCreator{},
	)

	_, err := svc.CreateMany(context.Background(), ManualCashflowCreateInput{
		AccountID: uuid.New(),
		Transactions: []ManualCashflowCreateTransactionInput{
			{
				Date:        "2026-03-08",
				Amount:      "-1",
				Type:        "out",
				Description: "Coffee",
				Note:        "Barista",
				Tag:         "food",
			},
		},
	})
	if !errors.Is(err, ErrManualCashflowInvalidAmount) {
		t.Fatalf("expected ErrManualCashflowInvalidAmount, got %v", err)
	}
}
