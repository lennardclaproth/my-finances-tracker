package cashflow

import (
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"io"
	"iter"
	"strings"
	"time"

	"github.com/google/uuid"
	"github.com/lennardclaproth/my-finances-tracker/internal/money"
)

type CashFlowDirection string
type AccountType string

type CsvParser interface {
	ParseAll(rc io.ReadCloser) (iter.Seq2[int, TransactionData], error)
}

const (
	AccountTypeChecking  AccountType = "checking"
	AccountTypeSavings   AccountType = "savings"
	AccountTypeCredit    AccountType = "credit"
	AccountTypeBrokerage AccountType = "brokerage"
)

const (
	CashIn  CashFlowDirection = "in"
	CashOut CashFlowDirection = "out"
)

type Transaction struct {
	ID          uuid.UUID         `db:"id"`
	AccountID   *uuid.UUID        `db:"account_id"`
	Description string            `db:"description"`
	Note        string            `db:"note"`
	Source      string            `db:"source"`
	AmountCents money.Price       `db:"amount_cents"`
	Direction   CashFlowDirection `db:"direction"`
	Date        time.Time         `db:"date"`
	Checksum    string            `db:"checksum"`
	CreatedAt   time.Time         `db:"created_at"`
	UpdatedAt   time.Time         `db:"updated_at"`
	Tag         string            `db:"tag"`
	RowNumber   int               `db:"row_number"`
	Ignored     bool              `db:"ignored"`
	ImportID    uuid.UUID         `db:"import_id"`
	AccountType *AccountType      `db:"account_type"` // Allow nullable account type.
}

type TransactionData struct {
	Description string
	Note        string
	Source      string
	Direction   CashFlowDirection
	Amount      float64
	Date        time.Time
	AccountType *AccountType
}

var (
	ErrDuplicateTransaction = fmt.Errorf("duplicate transaction")
	ErrUnsupportedDirection = fmt.Errorf("unsupported direction")
	ErrNoTransactionFound   = fmt.Errorf("no transaction found with the given ID")
)

// NewTransaction creates a new Transaction instance and generates its checksum.
func NewTransaction(desc, note, source string, direction CashFlowDirection, amount float64, date time.Time, rowNumber int, importID uuid.UUID, accountType *AccountType, accountID ...*uuid.UUID) (*Transaction, error) {
	// Guard on domain level against invalid amount values
	amountCents, err := money.NewPrice(amount)
	if err != nil {
		return nil, fmt.Errorf("NewTransaction failed: %w", err)
	}
	var accID *uuid.UUID
	if len(accountID) > 0 {
		accID = accountID[0]
	}
	t := &Transaction{
		ID:          uuid.New(),
		AccountID:   accID,
		Description: desc,
		Note:        note,
		Source:      source,
		Direction:   direction,
		AmountCents: amountCents,
		Date:        date,
		CreatedAt:   time.Now().UTC(),
		UpdatedAt:   time.Now().UTC(),
		RowNumber:   rowNumber,
		ImportID:    importID,
		AccountType: accountType,
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
	date := t.Date.Format("20060102") // Standard date format
	accountID := ""
	if t.AccountID != nil {
		accountID = t.AccountID.String()
	}
	// concatenate all fields to form the payload string to generate a checksum
	const sep = "\x1F" // Unit Separator character see -> https://www.ascii-code.com/character/%E2%90%9F
	payload := strings.Join([]string{desc, note, source, direction, amountCents, date, rowNumber, accountID}, sep)
	// digest the payload in byte format and encode it to hexadecimal string
	sum := sha256.Sum256([]byte(payload))
	return hex.EncodeToString(sum[:])
}
