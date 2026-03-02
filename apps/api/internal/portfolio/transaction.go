package portfolio

import (
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"io"
	"iter"
	"math"
	"strings"
	"time"

	"github.com/google/uuid"
	"github.com/lennardclaproth/my-finances-tracker/internal/money"
)

type TransactionType string

const (
	TxBuy      TransactionType = "BUY"
	TxSell     TransactionType = "SELL"
	TxDividend TransactionType = "DIVIDEND"
	TxTax      TransactionType = "TAX"
	TxFee      TransactionType = "FEE"
	TxCash     TransactionType = "CASH"
)

type Transaction struct {
	ID          uuid.UUID       `db:"id"`
	AccountID   *uuid.UUID      `db:"account_id"`
	ImportID    uuid.UUID       `db:"import_id"`
	Source      string          `db:"source"` // "degiro", "ibkr", ...
	OccurredAt  time.Time       `db:"occurred_at"`
	PositionID  *uuid.UUID      `db:"position_id"` // links to the position this transaction belongs to (can be null for cash transactions or if position mapping failed)
	ISIN        *string         `db:"isin"`
	Symbol      *string         `db:"symbol"`
	Description string          `db:"description"` // optional raw description from the broker, can be used for debugging or more complex parsing if needed
	Type        TransactionType `db:"type"`
	Quantity    float64         `db:"quantity"`
	UnitPrice   money.Price     `db:"unit_price"`   // per-unit for BUY/SELL
	AmountCents money.Price     `db:"amount_cents"` // absolute amount for cash-impact rows
	Checksum    string          `db:"checksum"`
	RowNumber   int             `db:"row_number"`
	CreatedAt   time.Time       `db:"created_at"`
	UpdatedAt   time.Time       `db:"updated_at"`
}

type CsvParser interface {
	ParseAll(rc io.ReadCloser) (iter.Seq2[int, TransactionData], error)
}

type TransactionData struct {
	Source      string
	OccurredAt  time.Time
	ValueDate   *time.Time
	ISIN        *string
	Symbol      *string
	Description string
	Type        TransactionType
	Quantity    float64
	Price       float64
	Amount      float64
}

var (
	ErrDuplicateTransaction            = fmt.Errorf("duplicate transaction")
	ErrTransactionISINAndSymbolMissing = fmt.Errorf("both ISIN and Symbol are missing, cannot determine position ID")
)

func NewTransaction(data TransactionData, rowNumber int, importID uuid.UUID, accountID, positionID *uuid.UUID) (*Transaction, error) {
	price, err := money.NewPrice(math.Abs(data.Price))
	if err != nil {
		return nil, fmt.Errorf("portfolio.NewTransaction price: %w", err)
	}
	amount, err := money.NewPrice(math.Abs(data.Amount))
	if err != nil {
		return nil, fmt.Errorf("portfolio.NewTransaction amount: %w", err)
	}
	quantity := data.Quantity
	if data.Type == TxCash {
		switch {
		case data.Amount < 0:
			quantity = -1
		case data.Amount > 0:
			quantity = 1
		default:
			quantity = 0
		}
	}
	if data.ISIN == nil && data.Symbol == nil {
		return nil, ErrTransactionISINAndSymbolMissing
	}
	if data.ISIN != nil {
		isin := strings.TrimSpace(*data.ISIN)
		data.ISIN = &isin
	}
	if data.Symbol != nil {
		symbol := strings.TrimSpace(*data.Symbol)
		data.Symbol = &symbol
	}
	tx := &Transaction{
		ID:          uuid.New(),
		AccountID:   accountID,
		ImportID:    importID,
		Source:      strings.TrimSpace(data.Source),
		OccurredAt:  data.OccurredAt,
		PositionID:  positionID,
		ISIN:        data.ISIN,
		Symbol:      data.Symbol,
		Description: data.Description,
		Type:        data.Type,
		Quantity:    quantity,
		UnitPrice:   price,
		AmountCents: amount,
		RowNumber:   rowNumber,
		CreatedAt:   time.Now().UTC(),
		UpdatedAt:   time.Now().UTC(),
	}
	tx.Checksum = tx.generateChecksum()
	return tx, nil
}

// GetID returns the identifier for the transaction, which is either the ISIN or Symbol. This is used for mapping transactions to positions. If both ISIN and Symbol are missing, it returns an error.
// ISIN takes precedence over Symbol for the ID.
func (t *Transaction) GetID() (string, error) {
	if t.ISIN != nil {
		return *t.ISIN, nil
	} else if t.Symbol != nil {
		return *t.Symbol, nil
	}
	return "", ErrTransactionISINAndSymbolMissing
}

func (t *Transaction) generateChecksum() string {
	accountID := ""
	if t.AccountID != nil {
		accountID = t.AccountID.String()
	}
	positionID := ""
	if t.PositionID != nil {
		positionID = t.PositionID.String()
	}
	if t.ISIN != nil {
		isin := strings.TrimSpace(*t.ISIN)
		t.ISIN = &isin
	}
	if t.Symbol != nil {
		symbol := strings.TrimSpace(*t.Symbol)
		t.Symbol = &symbol
	}
	isin := ""
	if t.ISIN != nil {
		isin = *t.ISIN
	}
	symbol := ""
	if t.Symbol != nil {
		symbol = *t.Symbol
	}
	const sep = "\x1F"
	payload := strings.Join([]string{
		strings.TrimSpace(t.Source),
		t.OccurredAt.Format("20060102"),
		isin,
		symbol,
		string(t.Type),
		fmt.Sprintf("%.8f", t.Quantity),
		fmt.Sprintf("%d", t.UnitPrice),
		fmt.Sprintf("%d", t.AmountCents),
		fmt.Sprintf("%d", t.RowNumber),
		accountID,
		positionID,
	}, sep)
	sum := sha256.Sum256([]byte(payload))
	return hex.EncodeToString(sum[:])
}
