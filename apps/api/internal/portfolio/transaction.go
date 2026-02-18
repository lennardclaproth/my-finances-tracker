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
	ValueDate   *time.Time      `db:"value_date"`
	ListingID   *uuid.UUID      `db:"listing_id"`
	ISIN        *string         `db:"isin"`
	Symbol      *string         `db:"symbol"`
	Type        TransactionType `db:"type"`
	Quantity    float64         `db:"quantity"`
	PriceCents  money.Price     `db:"price_cents"`  // per-unit for BUY/SELL
	AmountCents money.Price     `db:"amount_cents"` // absolute amount for cash-impact rows
	Checksum    string          `db:"checksum"`
	RawRef      string          `db:"raw_ref"`
	RowNumber   int             `db:"row_number"`
	CreatedAt   time.Time       `db:"created_at"`
	UpdatedAt   time.Time       `db:"updated_at"`
}

type CsvParser interface {
	ParseAll(rc io.ReadCloser) (iter.Seq2[int, TransactionData], error)
}

type TransactionData struct {
	Source     string
	OccurredAt time.Time
	ValueDate  *time.Time
	ISIN       *string
	Symbol     *string
	Type       TransactionType
	Quantity   float64
	Price      float64
	Amount     float64
	RawRef     string
}

var (
	ErrDuplicateTransaction = fmt.Errorf("duplicate transaction")
)

func NewTransaction(data TransactionData, rowNumber int, importID uuid.UUID, accountID, listingID *uuid.UUID) (*Transaction, error) {
	price, err := money.NewPrice(math.Abs(data.Price))
	if err != nil {
		return nil, fmt.Errorf("portfolio.NewTransaction price: %w", err)
	}
	amount, err := money.NewPrice(math.Abs(data.Amount))
	if err != nil {
		return nil, fmt.Errorf("portfolio.NewTransaction amount: %w", err)
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
		ValueDate:   data.ValueDate,
		ListingID:   listingID,
		ISIN:        data.ISIN,
		Symbol:      data.Symbol,
		Type:        data.Type,
		Quantity:    data.Quantity,
		PriceCents:  price,
		AmountCents: amount,
		RawRef:      strings.TrimSpace(data.RawRef),
		RowNumber:   rowNumber,
		CreatedAt:   time.Now().UTC(),
		UpdatedAt:   time.Now().UTC(),
	}
	tx.Checksum = tx.generateChecksum()
	return tx, nil
}

func (t *Transaction) generateChecksum() string {
	valueDate := ""
	if t.ValueDate != nil {
		valueDate = t.ValueDate.Format("20060102")
	}
	accountID := ""
	if t.AccountID != nil {
		accountID = t.AccountID.String()
	}
	listingID := ""
	if t.ListingID != nil {
		listingID = t.ListingID.String()
	}
	if t.ISIN != nil {
		isin := strings.TrimSpace(*t.ISIN)
		t.ISIN = &isin
	}
	if t.Symbol != nil {
		symbol := strings.TrimSpace(*t.Symbol)
		t.Symbol = &symbol
	}
	const sep = "\x1F"
	payload := strings.Join([]string{
		strings.TrimSpace(t.Source),
		t.OccurredAt.Format("20060102"),
		valueDate,
		*t.ISIN,
		*t.Symbol,
		string(t.Type),
		fmt.Sprintf("%.8f", t.Quantity),
		fmt.Sprintf("%d", t.PriceCents),
		fmt.Sprintf("%d", t.AmountCents),
		strings.TrimSpace(t.RawRef),
		fmt.Sprintf("%d", t.RowNumber),
		t.ImportID.String(),
		accountID,
		listingID,
	}, sep)
	sum := sha256.Sum256([]byte(payload))
	return hex.EncodeToString(sum[:])
}
