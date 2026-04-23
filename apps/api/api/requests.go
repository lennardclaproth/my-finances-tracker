package api

import (
	"context"
	"fmt"
	"mime/multipart"
	"net/textproto"
	"strings"

	"github.com/google/uuid"
)

// ImportCsv represents a multipart CSV import request payload.
type ImportCsv struct {
	File      multipart.File       `multipart:"file"`
	Filename  string               `multipart:"filename"`
	Size      int64                `multipart:"size"`
	Header    textproto.MIMEHeader `multipart:"header"`
	VendorID  uuid.UUID            `form:"vendor_id"`
	AccountID uuid.UUID            `form:"account_id"`
}

// Valid validates required import identifiers.
func (r ImportCsv) Valid(ctx context.Context) map[string]string {
	problems := make(map[string]string)
	if r.VendorID == uuid.Nil {
		problems["vendor_id"] = "vendor_id is required"
	}
	if r.AccountID == uuid.Nil {
		problems["account_id"] = "account_id is required"
	}
	return problems
}

// UploadDailiesRequest represents a multipart manual daily upload request.
type UploadDailiesRequest struct {
	File      multipart.File       `multipart:"file"`
	Filename  string               `multipart:"filename"`
	Size      int64                `multipart:"size"`
	Header    textproto.MIMEHeader `multipart:"header"`
	ListingID uuid.UUID            `form:"listing_id"`
}

// Valid validates required daily upload identifiers.
func (r UploadDailiesRequest) Valid(ctx context.Context) map[string]string {
	problems := make(map[string]string)
	if r.ListingID == uuid.Nil {
		problems["listing_id"] = "listing_id is required"
	}
	return problems
}

// GetUntaggedTransactionsRequest contains pagination filters for untagged transactions.
type GetUntaggedTransactionsRequest struct {
	Page     int `json:"page" query:"page"`
	PageSize int `json:"page_size" query:"page_size"`
}

// GetCashflowTransactionsRequest contains filters, sorting, and pagination for cashflow queries.
type GetCashflowTransactionsRequest struct {
	Limit       int    `json:"limit,omitempty" query:"limit"`
	Offset      int    `json:"offset,omitempty" query:"offset"`
	SortBy      string `json:"sort_by,omitempty" query:"sort_by"`
	SortOrder   string `json:"sort_order,omitempty" query:"sort_order"`
	Q           string `json:"q,omitempty" query:"q"`
	Description string `json:"description,omitempty" query:"description"`
	Note        string `json:"note,omitempty" query:"note"`
	Source      string `json:"source,omitempty" query:"source"`
	Direction   string `json:"direction,omitempty" query:"direction"`
	Tags        string `json:"tags,omitempty" query:"tags"`
	Untagged    bool   `json:"untagged,omitempty" query:"untagged"`
	HideIgnored bool   `json:"hide_ignored,omitempty" query:"hide_ignored"`
	From        string `json:"from,omitempty" query:"from"`
	To          string `json:"to,omitempty" query:"to"`
}

// Valid validates cashflow query pagination and sort filters.
func (r GetCashflowTransactionsRequest) Valid(ctx context.Context) map[string]string {
	problems := make(map[string]string)
	if r.Limit < 0 {
		problems["limit"] = "limit must be greater than or equal to 0"
	}
	if r.Offset < 0 {
		problems["offset"] = "offset must be greater than or equal to 0"
	}
	if r.SortBy != "" {
		switch strings.ToLower(r.SortBy) {
		case "date", "description", "note", "tag", "source", "amount":
		default:
			problems["sort_by"] = "sort_by must be one of: date, description, note, tag, source, amount"
		}
	}
	if r.SortOrder != "" {
		order := strings.ToLower(r.SortOrder)
		if order != "asc" && order != "desc" {
			problems["sort_order"] = "sort_order must be either asc or desc"
		}
	}
	if r.Direction != "" {
		direction := strings.ToLower(r.Direction)
		if direction != "in" && direction != "out" {
			problems["direction"] = "direction must be either in or out"
		}
	}
	return problems
}

// GetCashflowAnalyticsRequest contains date-range filters for analytics endpoints.
type GetCashflowAnalyticsRequest struct {
	From           string `json:"from,omitempty" query:"from"`
	To             string `json:"to,omitempty" query:"to"`
	IncludeIgnored bool   `json:"include_ignored,omitempty" query:"include_ignored"`
}

// GetDailiesRequest contains filters for listing/symbol daily market data retrieval.
type GetDailiesRequest struct {
	ListingID string `json:"listing_id,omitempty" query:"listing_id"`
	Symbol    string `json:"symbol" query:"symbol"`
	From      string `json:"from,omitempty" query:"from"`
	To        string `json:"to,omitempty" query:"to"`
	SortOrder string `json:"sort_order,omitempty" query:"sort_order"`
	Limit     int    `json:"limit,omitempty" query:"limit"`
	Offset    int    `json:"offset,omitempty" query:"offset"`
}

// Valid validates daily retrieval query constraints.
func (r GetDailiesRequest) Valid(ctx context.Context) map[string]string {
	problems := make(map[string]string)
	if strings.TrimSpace(r.Symbol) == "" && strings.TrimSpace(r.ListingID) == "" {
		problems["symbol"] = "symbol is required when listing_id is not provided"
	}
	if r.Limit < 0 {
		problems["limit"] = "limit must be greater than or equal to 0"
	}
	if r.Offset < 0 {
		problems["offset"] = "offset must be greater than or equal to 0"
	}
	if r.SortOrder != "" {
		order := strings.ToLower(strings.TrimSpace(r.SortOrder))
		if order != "asc" && order != "desc" {
			problems["sort_order"] = "sort_order must be either asc or desc"
		}
	}
	return problems
}

// TagTransactionRequest tags one transaction with a single tag.
type TagTransactionRequest struct {
	Id  uuid.UUID `json:"id"`
	Tag string    `json:"tag"`
}

// CashflowTagFilters defines reusable cashflow transaction filters for bulk actions.
type CashflowTagFilters struct {
	Q           string `json:"q,omitempty"`
	Description string `json:"description,omitempty"`
	Note        string `json:"note,omitempty"`
	Source      string `json:"source,omitempty"`
	Direction   string `json:"direction,omitempty"`
	Tags        string `json:"tags,omitempty"`
	Untagged    *bool  `json:"untagged,omitempty"`
	HideIgnored *bool  `json:"hide_ignored,omitempty"`
	From        string `json:"from,omitempty"`
	To          string `json:"to,omitempty"`
}

// TagTransactionsBySelectionRequest applies a tag to an explicit selection of transaction IDs.
type TagTransactionsBySelectionRequest struct {
	Tag string      `json:"tag"`
	IDs []uuid.UUID `json:"ids"`
}

// Valid validates selection-based tag requests.
func (r TagTransactionsBySelectionRequest) Valid(ctx context.Context) map[string]string {
	problems := make(map[string]string)
	if len(r.IDs) == 0 {
		problems["ids"] = "ids is required"
	}
	return problems
}

// TagTransactionsByFilterRequest applies a tag to transactions matching a filter query.
type TagTransactionsByFilterRequest struct {
	Tag       string             `json:"tag"`
	AccountID *uuid.UUID         `json:"account_id,omitempty"`
	Filters   CashflowTagFilters `json:"filters"`
}

// IgnoreTransactionsBySelectionRequest toggles ignored-state for selected transaction IDs.
type IgnoreTransactionsBySelectionRequest struct {
	Ignored *bool       `json:"ignored"`
	IDs     []uuid.UUID `json:"ids"`
}

// Valid validates selection-based ignore requests.
func (r IgnoreTransactionsBySelectionRequest) Valid(ctx context.Context) map[string]string {
	problems := make(map[string]string)
	if len(r.IDs) == 0 {
		problems["ids"] = "ids is required"
	}
	return problems
}

// IgnoreTransactionsByFilterRequest toggles ignored-state for transactions matching filters.
type IgnoreTransactionsByFilterRequest struct {
	Ignored *bool              `json:"ignored"`
	Filters CashflowTagFilters `json:"filters"`
}

// GetAssetClassesRequest contains query filters for asset classes.
type GetAssetClassesRequest struct {
	AccountID       uuid.UUID `json:"account_id" query:"account_id"`
	IncludeArchived bool      `json:"include_archived,omitempty" query:"include_archived"`
}

// Valid validates required asset class query fields.
func (r GetAssetClassesRequest) Valid(ctx context.Context) map[string]string {
	problems := make(map[string]string)
	if r.AccountID == uuid.Nil {
		problems["account_id"] = "account_id is required"
	}
	return problems
}

// GetAssetSnapshotsRequest contains query filters for account-level asset snapshots.
type GetAssetSnapshotsRequest struct {
	AccountID uuid.UUID `json:"account_id" query:"account_id"`
	From      string    `json:"from,omitempty" query:"from"`
	To        string    `json:"to,omitempty" query:"to"`
}

// Valid validates required asset snapshot query fields.
func (r GetAssetSnapshotsRequest) Valid(ctx context.Context) map[string]string {
	problems := make(map[string]string)
	if r.AccountID == uuid.Nil {
		problems["account_id"] = "account_id is required"
	}
	return problems
}

// CreateAssetClassRequest creates a manual asset class.
type CreateAssetClassRequest struct {
	AccountID uuid.UUID `json:"account_id"`
	Name      string    `json:"name"`
}

// Valid validates required asset class create fields.
func (r CreateAssetClassRequest) Valid(ctx context.Context) map[string]string {
	problems := make(map[string]string)
	if r.AccountID == uuid.Nil {
		problems["account_id"] = "account_id is required"
	}
	if strings.TrimSpace(r.Name) == "" {
		problems["name"] = "name is required"
	}
	return problems
}

// UpdateAssetClassRequest updates mutable fields of an asset class.
type UpdateAssetClassRequest struct {
	AccountID uuid.UUID `json:"account_id"`
	ID        uuid.UUID `json:"id"`
	Name      *string   `json:"name,omitempty"`
	Archived  *bool     `json:"archived,omitempty"`
}

// Valid validates required asset class update fields.
func (r UpdateAssetClassRequest) Valid(ctx context.Context) map[string]string {
	problems := make(map[string]string)
	if r.AccountID == uuid.Nil {
		problems["account_id"] = "account_id is required"
	}
	if r.ID == uuid.Nil {
		problems["id"] = "id is required"
	}
	return problems
}

// CreateAssetItemRequest creates one tracked asset item in a class.
type CreateAssetItemRequest struct {
	AccountID     uuid.UUID `json:"account_id"`
	ClassID       uuid.UUID `json:"class_id"`
	Name          string    `json:"name"`
	InitialWorth  string    `json:"initial_worth"`
	EffectiveDate string    `json:"effective_date"`
	Note          *string   `json:"note,omitempty"`
}

// Valid validates required asset item create fields.
func (r CreateAssetItemRequest) Valid(ctx context.Context) map[string]string {
	problems := make(map[string]string)
	if r.AccountID == uuid.Nil {
		problems["account_id"] = "account_id is required"
	}
	if r.ClassID == uuid.Nil {
		problems["class_id"] = "class_id is required"
	}
	if strings.TrimSpace(r.Name) == "" {
		problems["name"] = "name is required"
	}
	if strings.TrimSpace(r.InitialWorth) == "" {
		problems["initial_worth"] = "initial_worth is required"
	}
	if strings.TrimSpace(r.EffectiveDate) == "" {
		problems["effective_date"] = "effective_date is required"
	}
	return problems
}

// SetAssetItemWorthRequest sets an absolute worth value for an item.
type SetAssetItemWorthRequest struct {
	AccountID     uuid.UUID `json:"account_id"`
	ClassID       uuid.UUID `json:"class_id"`
	ItemID        uuid.UUID `json:"item_id"`
	Worth         string    `json:"worth"`
	EffectiveDate string    `json:"effective_date"`
	Note          *string   `json:"note,omitempty"`
}

// Valid validates required set-worth fields.
func (r SetAssetItemWorthRequest) Valid(ctx context.Context) map[string]string {
	problems := make(map[string]string)
	if r.AccountID == uuid.Nil {
		problems["account_id"] = "account_id is required"
	}
	if r.ClassID == uuid.Nil {
		problems["class_id"] = "class_id is required"
	}
	if r.ItemID == uuid.Nil {
		problems["item_id"] = "item_id is required"
	}
	if strings.TrimSpace(r.Worth) == "" {
		problems["worth"] = "worth is required"
	}
	if strings.TrimSpace(r.EffectiveDate) == "" {
		problems["effective_date"] = "effective_date is required"
	}
	return problems
}

// AdjustAssetItemWorthRequest applies a directional delta to an item worth.
type AdjustAssetItemWorthRequest struct {
	AccountID     uuid.UUID `json:"account_id"`
	ClassID       uuid.UUID `json:"class_id"`
	ItemID        uuid.UUID `json:"item_id"`
	Direction     string    `json:"direction"`
	Amount        string    `json:"amount"`
	EffectiveDate string    `json:"effective_date"`
	Note          *string   `json:"note,omitempty"`
}

// Valid validates required adjust-worth fields.
func (r AdjustAssetItemWorthRequest) Valid(ctx context.Context) map[string]string {
	problems := make(map[string]string)
	if r.AccountID == uuid.Nil {
		problems["account_id"] = "account_id is required"
	}
	if r.ClassID == uuid.Nil {
		problems["class_id"] = "class_id is required"
	}
	if r.ItemID == uuid.Nil {
		problems["item_id"] = "item_id is required"
	}
	if strings.TrimSpace(r.Direction) == "" {
		problems["direction"] = "direction is required"
	}
	if strings.TrimSpace(r.Amount) == "" {
		problems["amount"] = "amount is required"
	}
	if strings.TrimSpace(r.EffectiveDate) == "" {
		problems["effective_date"] = "effective_date is required"
	}
	return problems
}

// CreateManualCashflowTransactionEntryRequest represents one manual cashflow transaction row.
type CreateManualCashflowTransactionEntryRequest struct {
	Date        string `json:"date"`
	Amount      string `json:"amount"`
	Type        string `json:"type"`
	Description string `json:"description"`
	Note        string `json:"note"`
	Tag         string `json:"tag"`
	Vendor      string `json:"vendor,omitempty"`
}

// CreateManualCashflowTransactionsRequest creates one or more manual cashflow transactions.
type CreateManualCashflowTransactionsRequest struct {
	AccountID    uuid.UUID                                     `json:"account_id"`
	Transactions []CreateManualCashflowTransactionEntryRequest `json:"transactions"`
}

// Valid validates required manual cashflow transaction create fields.
func (r CreateManualCashflowTransactionsRequest) Valid(ctx context.Context) map[string]string {
	problems := make(map[string]string)
	if r.AccountID == uuid.Nil {
		problems["account_id"] = "account_id is required"
	}
	if len(r.Transactions) == 0 {
		problems["transactions"] = "transactions is required"
		return problems
	}
	for i, row := range r.Transactions {
		prefix := fmt.Sprintf("transactions[%d]", i)
		if strings.TrimSpace(row.Date) == "" {
			problems[prefix+".date"] = "date is required"
		}
		if strings.TrimSpace(row.Amount) == "" {
			problems[prefix+".amount"] = "amount is required"
		}
		if strings.TrimSpace(row.Type) == "" {
			problems[prefix+".type"] = "type is required"
		}
		if strings.TrimSpace(row.Description) == "" {
			problems[prefix+".description"] = "description is required"
		}
		if strings.TrimSpace(row.Note) == "" {
			problems[prefix+".note"] = "note is required"
		}
		if strings.TrimSpace(row.Tag) == "" {
			problems[prefix+".tag"] = "tag is required"
		}
	}
	return problems
}

// CreateListingRequest creates a market listing.
type CreateListingRequest struct {
	Name   string `json:"name"`
	Symbol string `json:"symbol"`
	Source string `json:"source"`
	// Optional fields
	Description *string `json:"description,omitempty"`
	Exchange    *string `json:"exchange,omitempty"`
	Region      *string `json:"region,omitempty"`
	Currency    *string `json:"currency,omitempty"`
	ISIN        *string `json:"isin,omitempty"`
	Ticker      *string `json:"ticker,omitempty"`
	Type        *string `json:"type,omitempty"`
}

// Valid validates required listing creation fields.
func (r CreateListingRequest) Valid(ctx context.Context) map[string]string {
	problems := make(map[string]string)
	if r.Name == "" {
		problems["name"] = "name is required"
	}
	if r.Symbol == "" {
		problems["symbol"] = "symbol is required"
	}
	if r.Source == "" {
		problems["source"] = "source is required"
	}
	return problems
}

// UpdateListingFieldsRequest updates mutable listing metadata fields.
type UpdateListingFieldsRequest struct {
	Id uuid.UUID `json:"id"`
	// Optional fields
	Description *string `json:"description,omitempty"`
	Exchange    *string `json:"exchange,omitempty"`
	Region      *string `json:"region,omitempty"`
	Currency    *string `json:"currency,omitempty"`
	ISIN        *string `json:"isin,omitempty"`
	Ticker      *string `json:"ticker,omitempty"`
	Type        *string `json:"type,omitempty"`
}

// Valid validates required listing update fields.
func (r UpdateListingFieldsRequest) Valid(ctx context.Context) map[string]string {
	problems := make(map[string]string)
	if r.Id == uuid.Nil {
		problems["id"] = "id is required"
	}
	return problems
}

// SearchListingsRequest contains search and pagination inputs for listings.
type SearchListingsRequest struct {
	Q      string `json:"q" query:"q"`
	Limit  int    `json:"limit,omitempty" query:"limit"`
	Offset int    `json:"offset,omitempty" query:"offset"`
}

// Valid validates listing search query constraints.
func (r SearchListingsRequest) Valid(ctx context.Context) map[string]string {
	problems := make(map[string]string)
	if strings.TrimSpace(r.Q) == "" {
		problems["q"] = "q is required"
	}
	if r.Limit < 0 {
		problems["limit"] = "limit must be greater than or equal to 0"
	}
	if r.Limit > 100 {
		problems["limit"] = "limit must be less than or equal to 100"
	}
	if r.Offset < 0 {
		problems["offset"] = "offset must be greater than or equal to 0"
	}
	return problems
}

// CreateManualPortfolioTransactionRequest creates a manual portfolio transaction.
type CreateManualPortfolioTransactionRequest struct {
	AccountID   uuid.UUID  `json:"account_id"`
	VendorID    uuid.UUID  `json:"vendor_id"`
	OccurredAt  string     `json:"occurred_at"`
	Type        string     `json:"type"`
	ListingID   *uuid.UUID `json:"listing_id,omitempty"`
	Amount      string     `json:"amount"`
	Quantity    *string    `json:"quantity,omitempty"`
	Description *string    `json:"description,omitempty"`
}

// Valid validates required manual portfolio transaction fields.
func (r CreateManualPortfolioTransactionRequest) Valid(ctx context.Context) map[string]string {
	problems := make(map[string]string)
	if r.AccountID == uuid.Nil {
		problems["account_id"] = "account_id is required"
	}
	if r.VendorID == uuid.Nil {
		problems["vendor_id"] = "vendor_id is required"
	}
	if strings.TrimSpace(r.OccurredAt) == "" {
		problems["occurred_at"] = "occurred_at is required"
	}
	if strings.TrimSpace(r.Type) == "" {
		problems["type"] = "type is required"
	}
	if strings.TrimSpace(r.Amount) == "" {
		problems["amount"] = "amount is required"
	}
	return problems
}

// CreateAccountRequest creates an account record.
type CreateAccountRequest struct {
	Name       string     `json:"name"`
	ExternalID *uuid.UUID `json:"external_id,omitempty"`
}

// Valid validates required account creation fields.
func (r CreateAccountRequest) Valid(ctx context.Context) map[string]string {
	problems := make(map[string]string)
	if strings.TrimSpace(r.Name) == "" {
		problems["name"] = "name is required"
	}
	return problems
}

// RebuildPortfolioRequest requests an asynchronous portfolio rebuild for an account.
type RebuildPortfolioRequest struct {
	AccountID uuid.UUID `json:"account_id"`
}

// Valid validates rebuild request account identity.
func (r RebuildPortfolioRequest) Valid(ctx context.Context) map[string]string {
	problems := make(map[string]string)
	if r.AccountID == uuid.Nil {
		problems["account_id"] = "account_id is required"
	}
	return problems
}

// GetPortfolioSnapshotsRequest contains query filters for portfolio snapshot history.
type GetPortfolioSnapshotsRequest struct {
	AccountID uuid.UUID `json:"account_id" query:"account_id"`
	From      string    `json:"from,omitempty" query:"from"`
	To        string    `json:"to,omitempty" query:"to"`
}

// Valid validates snapshot request account identity.
func (r GetPortfolioSnapshotsRequest) Valid(ctx context.Context) map[string]string {
	problems := make(map[string]string)
	if r.AccountID == uuid.Nil {
		problems["account_id"] = "account_id is required"
	}
	return problems
}

// GetPortfolioPositionsRequest contains query filters for current portfolio positions.
type GetPortfolioPositionsRequest struct {
	AccountID     uuid.UUID `json:"account_id" query:"account_id"`
	IncludeClosed bool      `json:"include_closed,omitempty" query:"include_closed"`
}

// Valid validates positions request account identity.
func (r GetPortfolioPositionsRequest) Valid(ctx context.Context) map[string]string {
	problems := make(map[string]string)
	if r.AccountID == uuid.Nil {
		problems["account_id"] = "account_id is required"
	}
	return problems
}

// GetPortfolioTransactionsRequest contains filters, sorting, and pagination for portfolio transactions.
type GetPortfolioTransactionsRequest struct {
	AccountID uuid.UUID `json:"account_id" query:"account_id"`
	From      string    `json:"from,omitempty" query:"from"`
	To        string    `json:"to,omitempty" query:"to"`
	Limit     int       `json:"limit,omitempty" query:"limit"`
	Offset    int       `json:"offset,omitempty" query:"offset"`
	SortBy    string    `json:"sort_by,omitempty" query:"sort_by"`
	SortOrder string    `json:"sort_order,omitempty" query:"sort_order"`
	Q         string    `json:"q,omitempty" query:"q"`
	Type      string    `json:"type,omitempty" query:"type"`
	Origin    string    `json:"origin,omitempty" query:"origin"`
	Source    string    `json:"source,omitempty" query:"source"`
	Listing   string    `json:"listing,omitempty" query:"listing"`
}

// Valid validates portfolio transaction query constraints.
func (r GetPortfolioTransactionsRequest) Valid(ctx context.Context) map[string]string {
	problems := make(map[string]string)
	if r.AccountID == uuid.Nil {
		problems["account_id"] = "account_id is required"
	}
	if r.Limit < 0 {
		problems["limit"] = "limit must be greater than or equal to 0"
	}
	if r.Limit != 0 && r.Limit != 10 && r.Limit != 25 && r.Limit != 50 && r.Limit != 100 {
		problems["limit"] = "limit must be one of: 10, 25, 50, 100"
	}
	if r.Offset < 0 {
		problems["offset"] = "offset must be greater than or equal to 0"
	}
	if r.SortBy != "" && strings.ToLower(strings.TrimSpace(r.SortBy)) != "date" {
		problems["sort_by"] = "sort_by must be one of: date"
	}
	if r.SortOrder != "" {
		order := strings.ToLower(strings.TrimSpace(r.SortOrder))
		if order != "asc" && order != "desc" {
			problems["sort_order"] = "sort_order must be either asc or desc"
		}
	}
	if r.Type != "" {
		switch strings.ToUpper(strings.TrimSpace(r.Type)) {
		case "BUY", "SELL", "DIVIDEND", "TAX", "FEE", "CASH":
		default:
			problems["type"] = "type must be one of: BUY, SELL, DIVIDEND, TAX, FEE, CASH"
		}
	}
	if r.Origin != "" {
		switch strings.ToUpper(strings.TrimSpace(r.Origin)) {
		case "IMPORT", "MANUAL":
		default:
			problems["origin"] = "origin must be one of: IMPORT, MANUAL"
		}
	}
	return problems
}
