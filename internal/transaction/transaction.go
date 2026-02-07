package transaction

import (
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"math"
	"strings"
	"time"

	"github.com/google/uuid"
)

type CashFlowDirection string

const (
	CashIn  CashFlowDirection = "in"
	CashOut CashFlowDirection = "out"
)

type Transaction struct {
	ID          uuid.UUID         `db:"id"`
	Description string            `db:"description"`
	Note        string            `db:"note"`
	Source      string            `db:"source"`
	AmountCents int64             `db:"amount_cents"`
	Direction   CashFlowDirection `db:"direction"`
	Date        time.Time         `db:"date"`
	Checksum    string            `db:"checksum"`
	CreatedAt   time.Time         `db:"created_at"`
	UpdatedAt   time.Time         `db:"updated_at"`
	Tag         string            `db:"tag"`
	RowNumber   int               `db:"row_number"`
	Ignored     bool              `db:"ignored"`
	ImportID    uuid.UUID         `db:"import_id"`
}

type TransactionData struct {
	Description string
	Note        string
	Source      string
	Direction   CashFlowDirection
	Amount      float64
	Date        time.Time
}

var (
	ErrDuplicateTransaction = fmt.Errorf("duplicate transaction")
	ErrInvalidAmount        = fmt.Errorf("invalid amount")
	ErrUnsupportedDirection = fmt.Errorf("unsupported direction")
	ErrNoTransactionFound   = fmt.Errorf("no transaction found with the given ID")
)

// NewTransaction creates a new Transaction instance and generates its checksum.
func NewTransaction(desc, note, source string, direction CashFlowDirection, amount float64, date time.Time, rowNumber int, importID uuid.UUID) (*Transaction, error) {
	// Guard on domain level against invalid amount values
	if math.IsNaN(amount) || math.IsInf(amount, 0) || amount < 0 {
		return nil, ErrInvalidAmount
	}

	t := &Transaction{
		ID:          uuid.New(),
		Description: desc,
		Note:        note,
		Source:      source,
		Direction:   direction,
		AmountCents: int64(amount * 100),
		Date:        date,
		CreatedAt:   time.Now().UTC(),
		UpdatedAt:   time.Now().UTC(),
		RowNumber:   rowNumber,
		ImportID:    importID,
	}
	t.Checksum = t.generateChecksum()
	return t, nil
}

// generateChecksum creates a checksum for the transaction based on the fields
// description, note, source, amountCents, and date. It uses amountCents instead
// of amount to avoid floating-point precision issues.
func (t *Transaction) generateChecksum() string {
	// initialize fields to be used in checksum generation, these fields need to be
	// of type string
	desc := strings.TrimSpace(t.Description)
	note := strings.TrimSpace(t.Note)
	source := strings.TrimSpace(t.Source)
	direction := string(t.Direction)
	amountCents := fmt.Sprintf("%d", t.AmountCents)
	rowNumber := fmt.Sprintf("%d", t.RowNumber)
	importID := t.ImportID.String()
	date := t.Date.Format("20060102") // Standard date format
	// concatenate all fields to form the payload string to generate a checksum
	const sep = "\x1F" // Unit Separator character see -> https://www.ascii-code.com/character/%E2%90%9F
	payload := strings.Join([]string{desc, note, source, direction, amountCents, date, rowNumber, importID}, sep)
	// digest the payload in byte format and encode it to hexadecimal string
	sum := sha256.Sum256([]byte(payload))
	return hex.EncodeToString(sum[:])
}
