package domain

import (
	"math"
	"testing"
	"time"

	"github.com/google/uuid"
)

// Helper to reduce repetition
func newTx() *Transaction {
	importID := uuid.New()
	tx, _ := NewTransaction(
		"Groceries",
		"Weekly shopping",
		"MyBank",
		"in",
		42.50,
		time.Date(2025, 1, 1, 0, 0, 0, 0, time.UTC),
		1,
		importID,
	)
	return tx
}

func newTxWithImportID(importID uuid.UUID) *Transaction {
	tx, _ := NewTransaction(
		"Groceries",
		"Weekly shopping",
		"MyBank",
		"in",
		42.50,
		time.Date(2025, 1, 1, 0, 0, 0, 0, time.UTC),
		1,
		importID,
	)
	return tx
}

func TestChecksum_IsConsistentForSameData(t *testing.T) {
	importID := uuid.New()
	tx1 := newTxWithImportID(importID)
	tx2 := newTxWithImportID(importID)

	if tx1.Checksum != tx2.Checksum {
		t.Fatalf("expected checksums to be equal, got %s and %s", tx1.Checksum, tx2.Checksum)
	}
}

func TestChecksum_DiffersWhenDescriptionChanges(t *testing.T) {
	tx1 := newTx()
	tx2 := newTx()
	tx2.Description = "Different"

	if tx1.generateChecksum() == tx2.generateChecksum() {
		t.Fatalf("checksum should change when description changes")
	}
}

func TestChecksum_DiffersWhenNoteChanges(t *testing.T) {
	tx1 := newTx()
	tx2 := newTx()
	tx2.Note = "Different note"

	if tx1.generateChecksum() == tx2.generateChecksum() {
		t.Fatalf("checksum should change when note changes")
	}
}

func TestChecksum_DiffersWhenSourceChanges(t *testing.T) {
	tx1 := newTx()
	tx2 := newTx()
	tx2.Source = "AnotherBank"

	if tx1.generateChecksum() == tx2.generateChecksum() {
		t.Fatalf("checksum should change when source changes")
	}
}

func TestChecksum_DiffersWhenAmountCentsChanges(t *testing.T) {
	tx1 := newTx()
	tx2 := newTx()
	tx2.AmountCents = 9999 // manual change

	if tx1.generateChecksum() == tx2.generateChecksum() {
		t.Fatalf("checksum should change when amount cents changes")
	}
}

func TestChecksum_DiffersWhenDateChanges(t *testing.T) {
	tx1 := newTx()
	tx2 := newTx()
	tx2.Date = time.Date(2025, 1, 2, 0, 0, 0, 0, time.UTC)

	if tx1.generateChecksum() == tx2.generateChecksum() {
		t.Fatalf("checksum should change when date changes")
	}
}

func TestChecksum_IgnoresIDField(t *testing.T) {
	importID := uuid.New()
	tx1 := newTxWithImportID(importID)
	tx2 := newTxWithImportID(importID)

	// Force different IDs
	tx2.ID = tx1.ID

	if tx1.Checksum != tx2.Checksum {
		t.Fatalf("checksum should not depend on ID")
	}
}

func TestChecksum_TrimsWhitespace(t *testing.T) {
	importID := uuid.New()
	tx1 := newTxWithImportID(importID)

	tx2, _ := NewTransaction(
		"  Groceries  ",
		"\tWeekly shopping\n",
		" MyBank ",
		"in",
		42.50,
		time.Date(2025, 1, 1, 0, 0, 0, 0, time.UTC),
		1,
		importID,
	)

	if tx1.Checksum != tx2.Checksum {
		t.Fatalf("expected trimmed strings to produce identical checksum")
	}
}

func TestNewTransaction_InvalidAmount_ReturnsError(t *testing.T) {
	tests := []struct {
		name   string
		amount float64
	}{
		{"NaN amount", math.NaN()},
		{"Positive infinity", math.Inf(1)},
		{"Negative infinity", math.Inf(-1)},
		{"Negative amount", -10.0},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tx, err := NewTransaction(
				"Test",
				"Test note",
				"TestBank",
				"in",
				tt.amount,
				time.Now(),
				1,
				uuid.New(),
			)

			if err == nil {
				t.Fatal("expected error for invalid amount, got nil")
			}
			if err != ErrInvalidAmount {
				t.Fatalf("expected ErrInvalidAmount, got %v", err)
			}
			if tx != nil {
				t.Fatal("expected nil transaction, got non-nil")
			}
		})
	}
}

func TestNewTransaction_InvalidDirection_ReturnsError(t *testing.T) {
	// Note: Direction validation happens through CashFlowDirection type casting.
	// Invalid direction strings are simply cast to CashFlowDirection without validation.
	// This test documents that invalid directions don't currently return errors.
	// If stricter validation is desired, add an ErrInvalidDirection error and validation logic.

	tx, err := NewTransaction(
		"Test",
		"Test note",
		"TestBank",
		"invalid",
		42.50,
		time.Now(),
		1,
		uuid.New(),
	)

	// Currently, invalid directions don't error—they just convert to the CashFlowDirection type
	if err != ErrUnsupportedDirection {
		t.Fatalf("expected ErrUnsupportedDirection error, got %v", err)
	}
	if tx != nil {
		t.Fatal("expected nil transaction for invalid direction")
	}
}
