package portfolio

import (
	"testing"
	"time"

	"github.com/google/uuid"
)

func TestNewTransaction_CashEncodesDirectionInQuantity(t *testing.T) {
	importID := uuid.New()
	tx, err := NewTransaction(
		TransactionData{
			Source:     "TEST",
			OccurredAt: time.Now().UTC(),
			ISIN:       ptrString("NLTEST0001"),
			Type:       TxCash,
			Amount:     -50,
		},
		1,
		importID,
		nil,
		nil,
	)
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}
	if tx.AmountCents <= 0 {
		t.Fatalf("expected absolute positive amount for cash transaction, got %d", tx.AmountCents)
	}
	if tx.Quantity != -1 {
		t.Fatalf("expected cash quantity sign marker -1, got %f", tx.Quantity)
	}
}

func TestNewTransaction_NonCashUsesAbsoluteAmount(t *testing.T) {
	importID := uuid.New()
	tx, err := NewTransaction(
		TransactionData{
			Source:     "TEST",
			OccurredAt: time.Now().UTC(),
			ISIN:       ptrString("NLTEST0002"),
			Type:       TxDividend,
			Amount:     -2,
		},
		1,
		importID,
		nil,
		nil,
	)
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}
	if tx.AmountCents <= 0 {
		t.Fatalf("expected absolute positive amount for non-cash transaction, got %d", tx.AmountCents)
	}
	if tx.Quantity != 0 {
		t.Fatalf("expected non-cash quantity to remain unchanged, got %f", tx.Quantity)
	}
}

func ptrString(v string) *string {
	return &v
}
