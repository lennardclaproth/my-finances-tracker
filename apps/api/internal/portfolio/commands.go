package portfolio

import (
	"context"
	"errors"
	"fmt"
	"strings"
	"time"

	"github.com/google/uuid"
	"github.com/lennardclaproth/my-finances-tracker/internal/marketdata"
	"github.com/lennardclaproth/my-finances-tracker/internal/vendor"
)

type Commands struct {
	cs  CommandStore
	mdq marketdata.Queries
	vq  vendor.Queries
}

// CommandStore persists portfolio accounts and transactions.
type CommandStore interface {
	CreateAccount(ctx context.Context, acc *Account) error
	CreateTransaction(ctx context.Context, tx *Transaction) error
	CreateTransactions(ctx context.Context, txs []*Transaction) (int, error)
}

// NewCommands constructs the portfolio write-side use cases.
func NewCommands(cs CommandStore, mdq marketdata.Queries, vq vendor.Queries) *Commands {
	return &Commands{cs: cs, mdq: mdq, vq: vq}
}

type ManualTransactionInput struct {
	AccountID   uuid.UUID
	VendorID    uuid.UUID
	OccurredAt  string
	Type        string
	ListingID   *uuid.UUID
	Amount      string
	Quantity    *string
	Description *string
}

type ManualTransactionCreateResult struct {
	Transaction  *Transaction
	ListingID    *uuid.UUID
	SignedAmount float64
}

func (c *Commands) CreateTransaction(ctx context.Context, input ManualTransactionInput) (*ManualTransactionCreateResult, error) {
	v, err := c.vq.GetById(ctx, input.VendorID)
	if err != nil {
		if errors.Is(err, vendor.ErrVendorNotFound) {
			return nil, ErrManualVendorNotFound
		}
		return nil, fmt.Errorf("manual transaction: fetch vendor: %w", err)
	}
	if !v.Active {
		return nil, ErrManualVendorNotActive
	}
	if v.Type != vendor.VendorTypeBrokerage {
		return nil, ErrManualVendorTypeNotSupported
	}

	occurredAt, err := time.Parse("2006-01-02", strings.TrimSpace(input.OccurredAt))
	if err != nil {
		return nil, ErrManualInvalidOccurredAt
	}
	occurredAt = occurredAt.UTC()

	txType, err := parseManualType(input.Type)
	if err != nil {
		return nil, err
	}

	amount, err := parseDecimalString(input.Amount, ErrManualInvalidAmount)
	if err != nil {
		return nil, err
	}
	if txType == TxCash {
		if amount == 0 {
			return nil, ErrManualCashAmountMustBeNonZero
		}
	} else if amount <= 0 {
		return nil, ErrManualNonCashAmountMustBePos
	}

	quantity, err := parseManualQuantity(txType, input.Quantity)
	if err != nil {
		return nil, err
	}

	listingID, isin, symbol, err := c.resolveManualListing(ctx, txType, input.ListingID)
	if err != nil {
		return nil, err
	}

	unitPrice := 0.0
	if txType == TxBuy || txType == TxSell {
		unitPrice = amount / quantity
	}

	description := ""
	if input.Description != nil {
		description = strings.TrimSpace(*input.Description)
	}

	tx, err := NewManualTransaction(TransactionData{
		Source:      string(v.Name),
		OccurredAt:  occurredAt,
		ISIN:        isin,
		Symbol:      symbol,
		Description: description,
		Type:        txType,
		Quantity:    quantity,
		Price:       unitPrice,
		Amount:      amount,
	}, input.AccountID)
	if err != nil {
		return nil, fmt.Errorf("manual transaction: create transaction: %w", err)
	}

	if err := c.cs.CreateTransaction(ctx, tx); err != nil {
		return nil, err
	}

	signedAmount := tx.AmountCents.Float64()
	if tx.Type == TxCash && tx.Quantity < 0 {
		signedAmount = -signedAmount
	}

	return &ManualTransactionCreateResult{
		Transaction:  tx,
		ListingID:    listingID,
		SignedAmount: signedAmount,
	}, nil
}

func (c *Commands) resolveManualListing(ctx context.Context, txType TransactionType, listingID *uuid.UUID) (*uuid.UUID, *string, *string, error) {
	if txType == TxCash {
		if listingID != nil {
			return nil, nil, nil, ErrManualListingForbidden
		}
		return nil, nil, nil, nil
	}
	if listingID == nil {
		return nil, nil, nil, ErrManualListingRequired
	}

	listing, err := c.mdq.Listing(ctx, *listingID)
	if err != nil {
		return nil, nil, nil, fmt.Errorf("manual transaction: fetch listing: %w", err)
	}
	if listing == nil {
		return nil, nil, nil, ErrManualListingNotFound
	}

	var isin *string
	if listing.ISIN != nil {
		v := strings.TrimSpace(*listing.ISIN)
		if v != "" {
			isin = &v
		}
	}

	var symbol *string
	if v := strings.TrimSpace(listing.Symbol); v != "" {
		symbol = &v
	}

	if isin == nil && symbol == nil {
		return nil, nil, nil, ErrManualListingIdentityMissing
	}
	return listingID, isin, symbol, nil
}

func (c *Commands) CreateAccount(ctx context.Context, accountID uuid.UUID) (*Account, error) {
	acc := NewAccount(accountID)
	if err := c.cs.CreateAccount(ctx, acc); err != nil {
		return nil, fmt.Errorf("create account: failed to store account: %w", err)
	}
	return acc, nil
}

// CreateManyResult reports the outcome of a batch portfolio create.
type CreateManyResult struct {
	// Transactions are the transactions built from the input.
	Transactions []*Transaction
	// Imported is the number of rows newly inserted.
	Imported int
	// Duplicates is the number of rows skipped because they already existed.
	Duplicates int
}

// CreateMany builds a batch of imported portfolio transactions and persists them with a
// single bulk insert, skipping and counting rows that already exist. Each TransactionData
// must carry its source row number for deduplication.
func (c *Commands) CreateMany(ctx context.Context, importID uuid.UUID, accountID *uuid.UUID, rows []TransactionData) (CreateManyResult, error) {
	if len(rows) == 0 {
		return CreateManyResult{}, nil
	}

	txs := make([]*Transaction, 0, len(rows))
	for _, row := range rows {
		tx, err := NewTransaction(row, row.RowNumber, importID, accountID, nil)
		if err != nil {
			return CreateManyResult{}, fmt.Errorf("create many: build transaction: %w", err)
		}
		txs = append(txs, tx)
	}

	inserted, err := c.cs.CreateTransactions(ctx, txs)
	if err != nil {
		return CreateManyResult{}, fmt.Errorf("create many: %w", err)
	}

	return CreateManyResult{
		Transactions: txs,
		Imported:     inserted,
		Duplicates:   len(txs) - inserted,
	}, nil
}
