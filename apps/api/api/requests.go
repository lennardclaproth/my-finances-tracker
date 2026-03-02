package api

import (
	"context"
	"mime/multipart"
	"net/textproto"
	"strings"

	"github.com/google/uuid"
)

type ImportCsv struct {
	File      multipart.File       `multipart:"file"`
	Filename  string               `multipart:"filename"`
	Size      int64                `multipart:"size"`
	Header    textproto.MIMEHeader `multipart:"header"`
	VendorID  uuid.UUID            `form:"vendor_id"`
	AccountID uuid.UUID            `form:"account_id"`
}

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

type GetUntaggedTransactionsRequest struct {
	Page     int `json:"page" query:"page"`
	PageSize int `json:"page_size" query:"page_size"`
}

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

type GetCashflowAnalyticsRequest struct {
	From           string `json:"from,omitempty" query:"from"`
	To             string `json:"to,omitempty" query:"to"`
	IncludeIgnored bool   `json:"include_ignored,omitempty" query:"include_ignored"`
}

type GetDailiesRequest struct {
	Symbol string `json:"symbol" query:"symbol"`
	From   string `json:"from,omitempty" query:"from"`
	To     string `json:"to,omitempty" query:"to"`
	Limit  int    `json:"limit,omitempty" query:"limit"`
	Offset int    `json:"offset,omitempty" query:"offset"`
}

func (r GetDailiesRequest) Valid(ctx context.Context) map[string]string {
	problems := make(map[string]string)
	if r.Symbol == "" {
		problems["symbol"] = "symbol is required"
	}
	if r.Limit < 0 {
		problems["limit"] = "limit must be greater than or equal to 0"
	}
	if r.Offset < 0 {
		problems["offset"] = "offset must be greater than or equal to 0"
	}
	return problems
}

type TagTransactionRequest struct {
	Id  uuid.UUID `json:"id"`
	Tag string    `json:"tag"`
}

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

type TagTransactionsBySelectionRequest struct {
	Tag string      `json:"tag"`
	IDs []uuid.UUID `json:"ids"`
}

func (r TagTransactionsBySelectionRequest) Valid(ctx context.Context) map[string]string {
	problems := make(map[string]string)
	if len(r.IDs) == 0 {
		problems["ids"] = "ids is required"
	}
	return problems
}

type TagTransactionsByFilterRequest struct {
	Tag     string             `json:"tag"`
	Filters CashflowTagFilters `json:"filters"`
}

type IgnoreTransactionsBySelectionRequest struct {
	Ignored *bool       `json:"ignored"`
	IDs     []uuid.UUID `json:"ids"`
}

func (r IgnoreTransactionsBySelectionRequest) Valid(ctx context.Context) map[string]string {
	problems := make(map[string]string)
	if len(r.IDs) == 0 {
		problems["ids"] = "ids is required"
	}
	return problems
}

type IgnoreTransactionsByFilterRequest struct {
	Ignored *bool              `json:"ignored"`
	Filters CashflowTagFilters `json:"filters"`
}

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

func (r UpdateListingFieldsRequest) Valid(ctx context.Context) map[string]string {
	problems := make(map[string]string)
	if r.Id == uuid.Nil {
		problems["id"] = "id is required"
	}
	return problems
}

type CreateAccountRequest struct {
	Name       string     `json:"name"`
	ExternalID *uuid.UUID `json:"external_id,omitempty"`
}

func (r CreateAccountRequest) Valid(ctx context.Context) map[string]string {
	problems := make(map[string]string)
	if strings.TrimSpace(r.Name) == "" {
		problems["name"] = "name is required"
	}
	return problems
}

type RebuildPortfolioRequest struct {
	AccountID uuid.UUID `json:"account_id"`
}

func (r RebuildPortfolioRequest) Valid(ctx context.Context) map[string]string {
	problems := make(map[string]string)
	if r.AccountID == uuid.Nil {
		problems["account_id"] = "account_id is required"
	}
	return problems
}

type GetPortfolioSnapshotsRequest struct {
	AccountID uuid.UUID `json:"account_id" query:"account_id"`
	From      string    `json:"from,omitempty" query:"from"`
	To        string    `json:"to,omitempty" query:"to"`
}

func (r GetPortfolioSnapshotsRequest) Valid(ctx context.Context) map[string]string {
	problems := make(map[string]string)
	if r.AccountID == uuid.Nil {
		problems["account_id"] = "account_id is required"
	}
	return problems
}

type GetPortfolioPositionsRequest struct {
	AccountID    uuid.UUID `json:"account_id" query:"account_id"`
	IncludeClosed bool     `json:"include_closed,omitempty" query:"include_closed"`
}

func (r GetPortfolioPositionsRequest) Valid(ctx context.Context) map[string]string {
	problems := make(map[string]string)
	if r.AccountID == uuid.Nil {
		problems["account_id"] = "account_id is required"
	}
	return problems
}
