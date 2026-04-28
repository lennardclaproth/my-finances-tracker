package handlers

import (
	"context"
	"errors"
	"fmt"
	"regexp"
	"strconv"
	"strings"
	"time"

	"github.com/google/uuid"
	"github.com/lennardclaproth/my-finances-tracker/internal/account"
	"github.com/lennardclaproth/my-finances-tracker/internal/marketdata"
	"github.com/lennardclaproth/my-finances-tracker/internal/vendor"
)

var (
	ErrManualAccountNotFound         = fmt.Errorf("account not found")
	ErrManualVendorNotFound          = fmt.Errorf("vendor not found")
	ErrManualListingNotFound         = fmt.Errorf("listing not found")
	ErrManualVendorNotActive         = fmt.Errorf("vendor must be active")
	ErrManualVendorTypeNotSupported  = fmt.Errorf("vendor type not allowed for portfolio manual transaction")
	ErrManualInvalidOccurredAt       = fmt.Errorf("occurred_at must be in YYYY-MM-DD format")
	ErrManualInvalidType             = fmt.Errorf("type must be one of: BUY, SELL, DIVIDEND, TAX, FEE, CASH")
	ErrManualInvalidAmount           = fmt.Errorf("amount must be a decimal string with up to 6 decimals")
	ErrManualInvalidQuantity         = fmt.Errorf("quantity must be a decimal string with up to 6 decimals")
	ErrManualQuantityRequired        = fmt.Errorf("quantity is required for BUY/SELL")
	ErrManualQuantityForbidden       = fmt.Errorf("quantity is only allowed for BUY/SELL")
	ErrManualListingRequired         = fmt.Errorf("listing_id is required for non-CASH transactions")
	ErrManualListingForbidden        = fmt.Errorf("listing_id is not allowed for CASH transactions")
	ErrManualNonCashAmountMustBePos  = fmt.Errorf("amount must be positive for non-CASH transactions")
	ErrManualCashAmountMustBeNonZero = fmt.Errorf("amount must be non-zero for CASH transactions")
	ErrManualQuantityMustBePositive  = fmt.Errorf("quantity must be positive")
	ErrManualListingIdentityMissing  = fmt.Errorf("listing must provide at least one of isin or symbol")
)

var decimalSixPattern = regexp.MustCompile(`^-?\d+(\.\d{1,6})?$`)

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

type manualAccountFetcher interface {
	FetchByID(ctx context.Context, id uuid.UUID) (*account.Account, error)
}

type manualVendorFetcher interface {
	FetchById(ctx context.Context, id uuid.UUID) (*vendor.Vendor, error)
}

type manualListingFetcher interface {
	FetchByID(ctx context.Context, id uuid.UUID) (*marketdata.Listing, error)
}

type manualTransactionCreator interface {
	Create(ctx context.Context, tx *Transaction) error
}

type ManualTransactionService struct {
	accounts manualAccountFetcher
	vendors  manualVendorFetcher
	listings manualListingFetcher
	txs      manualTransactionCreator
}

func NewManualTransactionService(
	accounts manualAccountFetcher,
	vendors manualVendorFetcher,
	listings manualListingFetcher,
	txs manualTransactionCreator,
) *ManualTransactionService {
	return &ManualTransactionService{
		accounts: accounts,
		vendors:  vendors,
		listings: listings,
		txs:      txs,
	}
}

func (s *ManualTransactionService) Create(ctx context.Context, input ManualTransactionInput) (*ManualTransactionCreateResult, error) {
	if _, err := s.accounts.FetchByID(ctx, input.AccountID); err != nil {
		if errors.Is(err, account.ErrAccountNotFound) {
			return nil, ErrManualAccountNotFound
		}
		return nil, fmt.Errorf("manual transaction: fetch account: %w", err)
	}

	v, err := s.vendors.FetchById(ctx, input.VendorID)
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

	listingID, isin, symbol, err := s.resolveManualListing(ctx, txType, input.ListingID)
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

	if err := s.txs.Create(ctx, tx); err != nil {
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

func parseManualType(raw string) (TransactionType, error) {
	normalized := TransactionType(strings.ToUpper(strings.TrimSpace(raw)))
	switch normalized {
	case TxBuy, TxSell, TxDividend, TxTax, TxFee, TxCash:
		return normalized, nil
	default:
		return "", ErrManualInvalidType
	}
}

func parseManualQuantity(txType TransactionType, raw *string) (float64, error) {
	if txType == TxBuy || txType == TxSell {
		if raw == nil || strings.TrimSpace(*raw) == "" {
			return 0, ErrManualQuantityRequired
		}
		quantity, err := parseDecimalString(*raw, ErrManualInvalidQuantity)
		if err != nil {
			return 0, err
		}
		if quantity <= 0 {
			return 0, ErrManualQuantityMustBePositive
		}
		return quantity, nil
	}

	if raw != nil && strings.TrimSpace(*raw) != "" {
		return 0, ErrManualQuantityForbidden
	}
	return 0, nil
}

func (s *ManualTransactionService) resolveManualListing(ctx context.Context, txType TransactionType, listingID *uuid.UUID) (*uuid.UUID, *string, *string, error) {
	if txType == TxCash {
		if listingID != nil {
			return nil, nil, nil, ErrManualListingForbidden
		}
		return nil, nil, nil, nil
	}
	if listingID == nil {
		return nil, nil, nil, ErrManualListingRequired
	}

	listing, err := s.listings.FetchByID(ctx, *listingID)
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

func parseDecimalString(raw string, invalidErr error) (float64, error) {
	trimmed := strings.TrimSpace(raw)
	if !decimalSixPattern.MatchString(trimmed) {
		return 0, invalidErr
	}
	val, err := strconv.ParseFloat(trimmed, 64)
	if err != nil {
		return 0, invalidErr
	}
	return val, nil
}
