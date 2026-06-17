//go:build integration

package integration

import (
	"context"
	"errors"
	"testing"

	"github.com/google/uuid"

	"github.com/lennardclaproth/my-finances-tracker/internal/account"
	"github.com/lennardclaproth/my-finances-tracker/internal/storage"
)

// TestAccountStoreCreateAndGet round-trips an account through the real migrated
// schema: a created account can be read back by ID and reports as existing.
func TestAccountStoreCreateAndGet(t *testing.T) {
	eachDialect(t, func(t *testing.T, db *storage.DB) {
		ctx := context.Background()
		store := storage.NewSQLXAccountStore(db)

		acc, err := account.NewAccount("Checking", nil, nil)
		if err != nil {
			t.Fatalf("new account: %v", err)
		}
		if err := store.Create(ctx, acc); err != nil {
			t.Fatalf("create: %v", err)
		}

		got, err := store.GetByID(ctx, acc.ID)
		if err != nil {
			t.Fatalf("get by id: %v", err)
		}
		if got.ID != acc.ID || got.Name != "Checking" {
			t.Fatalf("round-trip mismatch: got id=%s name=%q, want id=%s name=Checking", got.ID, got.Name, acc.ID)
		}

		exists, err := store.Exists(ctx, acc.ID)
		if err != nil {
			t.Fatalf("exists: %v", err)
		}
		if !exists {
			t.Fatal("expected created account to exist")
		}
	})
}

// TestAccountStoreGetByIDNotFound verifies the store maps a missing row to the
// feature's typed account.ErrAccountNotFound (sql.ErrNoRows translation).
func TestAccountStoreGetByIDNotFound(t *testing.T) {
	eachDialect(t, func(t *testing.T, db *storage.DB) {
		store := storage.NewSQLXAccountStore(db)
		if _, err := store.GetByID(context.Background(), uuid.New()); !errors.Is(err, account.ErrAccountNotFound) {
			t.Fatalf("expected ErrAccountNotFound, got %v", err)
		}
	})
}

// TestAccountStoreCreateDuplicateIsRejected verifies a unique-constraint
// violation is mapped to account.ErrAccountAlreadyExists on each dialect.
func TestAccountStoreCreateDuplicateIsRejected(t *testing.T) {
	eachDialect(t, func(t *testing.T, db *storage.DB) {
		ctx := context.Background()
		store := storage.NewSQLXAccountStore(db)

		acc, err := account.NewAccount("Checking", nil, nil)
		if err != nil {
			t.Fatalf("new account: %v", err)
		}
		if err := store.Create(ctx, acc); err != nil {
			t.Fatalf("first create: %v", err)
		}
		if err := store.Create(ctx, acc); !errors.Is(err, account.ErrAccountAlreadyExists) {
			t.Fatalf("expected ErrAccountAlreadyExists on duplicate, got %v", err)
		}
	})
}

// TestWithTxRollsBackOnError verifies DB.WithTx rolls back when fn returns an
// error: a row written through the transaction-aware executor must not be
// visible after the transaction unwinds.
func TestWithTxRollsBackOnError(t *testing.T) {
	eachDialect(t, func(t *testing.T, db *storage.DB) {
		ctx := context.Background()
		store := storage.NewSQLXAccountStore(db)

		acc, err := account.NewAccount("Rollback", nil, nil)
		if err != nil {
			t.Fatalf("new account: %v", err)
		}

		sentinel := errors.New("forced rollback")
		err = db.WithTx(ctx, func(ctx context.Context) error {
			if createErr := store.Create(ctx, acc); createErr != nil {
				return createErr
			}
			return sentinel
		})
		if !errors.Is(err, sentinel) {
			t.Fatalf("expected sentinel error from WithTx, got %v", err)
		}

		exists, err := store.Exists(ctx, acc.ID)
		if err != nil {
			t.Fatalf("exists: %v", err)
		}
		if exists {
			t.Fatal("expected the account to be rolled back, but it persists")
		}
	})
}

// TestWithTxCommitsOnSuccess verifies DB.WithTx commits when fn returns nil, so
// work done inside the transaction is durable afterward.
func TestWithTxCommitsOnSuccess(t *testing.T) {
	eachDialect(t, func(t *testing.T, db *storage.DB) {
		ctx := context.Background()
		store := storage.NewSQLXAccountStore(db)

		acc, err := account.NewAccount("Committed", nil, nil)
		if err != nil {
			t.Fatalf("new account: %v", err)
		}

		if err := db.WithTx(ctx, func(ctx context.Context) error {
			return store.Create(ctx, acc)
		}); err != nil {
			t.Fatalf("WithTx: %v", err)
		}

		exists, err := store.Exists(ctx, acc.ID)
		if err != nil {
			t.Fatalf("exists: %v", err)
		}
		if !exists {
			t.Fatal("expected the committed account to persist")
		}
	})
}
