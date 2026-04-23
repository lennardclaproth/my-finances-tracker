package api

import (
	"time"

	"github.com/google/uuid"
)

// ImportSummary represents the result of a CSV import operation, containing
// counts of processed, imported, duplicate, and failed rows along with detailed errors.
type ImportSummary struct {
	// TotalRows is the total number of data rows in the CSV file (excluding header)
	TotalRows int `json:"totalRows" example:"100"`
	// Imported is the number of rows successfully imported into the database
	Imported int `json:"imported" example:"98"`
	// Duplicates is the number of rows skipped due to duplicate checksums
	Duplicates int `json:"duplicates" example:"1"`
	// Failed is the number of rows that failed to import
	Failed int `json:"failed" example:"1"`
	// RowErrors contains detailed error information for each failed or problematic row
	RowErrors []RowError `json:"rowErrors"`
}

// Transaction represents a cashflow transaction in API responses.
type Transaction struct {
	ID          uuid.UUID `json:"id" example:"550e8400-e29b-41d4-a716-446655440000"`
	Description string    `json:"description" example:"Grocery shopping"`
	Note        string    `json:"note" example:"Bought fruits and vegetables"`
	Source      string    `json:"source" example:"MyBank"`
	AmountCents int64     `json:"amountCents" example:"4250"`
	Direction   string    `json:"direction" example:"out"`
	Date        time.Time `json:"date" example:"2025-01-15T00:00:00Z"`
	Tag         string    `json:"tag" example:"Food"`
	Ignored     bool      `json:"ignored" example:"false"`
}

// Pagination describes offset/limit paging metadata.
type Pagination struct {
	Limit  int `json:"limit"`
	Offset int `json:"offset"`
	Count  int `json:"count"`
	Total  int `json:"total"`
}

// CashflowTransactionsResponse returns paginated cashflow transactions.
type CashflowTransactionsResponse struct {
	Pagination Pagination    `json:"pagination"`
	Data       []Transaction `json:"data"`
}

// ManualCashflowTransactionsResponse returns the result of manual cashflow transaction creation.
type ManualCashflowTransactionsResponse struct {
	CreatedCount int           `json:"created_count"`
	Data         []Transaction `json:"data"`
}

// AssetClassResponse represents one asset class row.
type AssetClassResponse struct {
	ID           uuid.UUID  `json:"id"`
	Name         string     `json:"name"`
	Source       string     `json:"source"`
	Archived     bool       `json:"archived"`
	CurrentWorth string     `json:"current_worth"`
	LastChangeAt *time.Time `json:"last_change_at,omitempty"`
	GrowthPct    *float64   `json:"growth_pct,omitempty"`
	UpdatedAt    time.Time  `json:"updated_at"`
}

// AssetItemResponse represents one tracked item in a class.
type AssetItemResponse struct {
	ID           uuid.UUID `json:"id"`
	Name         string    `json:"name"`
	CurrentWorth string    `json:"current_worth"`
	Archived     bool      `json:"archived"`
	UpdatedAt    time.Time `json:"updated_at"`
}

// AssetGrowthPointResponse represents class total worth on a date.
type AssetGrowthPointResponse struct {
	Date       string `json:"date"`
	TotalWorth string `json:"total_worth"`
}

// AssetHistoryResponse represents one class history entry.
type AssetHistoryResponse struct {
	ID              uuid.UUID `json:"id"`
	ItemID          uuid.UUID `json:"item_id"`
	ChangeType      string    `json:"change_type"`
	Direction       *string   `json:"direction,omitempty"`
	Amount          string    `json:"amount"`
	PreviousWorth   string    `json:"previous_worth"`
	NewWorth        string    `json:"new_worth"`
	ClassTotalWorth string    `json:"class_total_worth"`
	EffectiveDate   string    `json:"effective_date"`
	Note            string    `json:"note"`
	CreatedAt       time.Time `json:"created_at"`
}

// AssetClassDetailsResponse returns slider data for one class.
type AssetClassDetailsResponse struct {
	Class   AssetClassResponse         `json:"class"`
	Items   []AssetItemResponse        `json:"items"`
	Growth  []AssetGrowthPointResponse `json:"growth"`
	History []AssetHistoryResponse     `json:"history"`
}

// CashflowMonthlyAnalyticsPoint represents one month of aggregated cashflow metrics.
type CashflowMonthlyAnalyticsPoint struct {
	Month         string `json:"month" example:"2025-01-01"`
	IncomingCents int64  `json:"incomingCents" example:"250000"`
	OutgoingCents int64  `json:"outgoingCents" example:"120000"`
	NetCents      int64  `json:"netCents" example:"130000"`
}

// CashflowMonthlyAnalyticsResponse returns monthly analytics time-series data.
type CashflowMonthlyAnalyticsResponse struct {
	Data []CashflowMonthlyAnalyticsPoint `json:"data"`
}

// CashflowTagDistributionEntry represents one tag aggregate bucket.
type CashflowTagDistributionEntry struct {
	Tag        string `json:"tag" example:"food"`
	TotalCents int64  `json:"totalCents" example:"4200"`
}

// CashflowTagDistributionResponse returns tag aggregates across combined/incoming/outgoing sets.
type CashflowTagDistributionResponse struct {
	Combined []CashflowTagDistributionEntry `json:"combined"`
	Incoming []CashflowTagDistributionEntry `json:"incoming"`
	Outgoing []CashflowTagDistributionEntry `json:"outgoing"`
}

// TagTransactionsResponse returns the result of a tag mutation operation.
type TagTransactionsResponse struct {
	UpdatedCount int    `json:"updated_count"`
	Status       string `json:"status"`
}

// VendorResponse represents one vendor record.
type VendorResponse struct {
	ID             uuid.UUID `json:"id"`
	Name           string    `json:"name"`
	Type           string    `json:"type"`
	Active         bool      `json:"active"`
	ImportDisabled bool      `json:"import_disabled"`
	CreatedAt      time.Time `json:"created_at"`
	UpdatedAt      time.Time `json:"updated_at"`
}

// AccountResponse represents one account record.
type AccountResponse struct {
	ID         uuid.UUID  `json:"id"`
	ExternalID *uuid.UUID `json:"external_id,omitempty"`
	Name       string     `json:"name"`
	CreatedAt  time.Time  `json:"created_at"`
	UpdatedAt  time.Time  `json:"updated_at"`
}

// ListingResponse represents one market listing record.
type ListingResponse struct {
	ID          uuid.UUID `json:"id"`
	Symbol      string    `json:"symbol"`
	Name        string    `json:"name"`
	Source      string    `json:"source"`
	Description *string   `json:"description,omitempty"`
	Exchange    *string   `json:"exchange,omitempty"`
	Region      *string   `json:"region,omitempty"`
	Currency    *string   `json:"currency,omitempty"`
	ISIN        *string   `json:"isin,omitempty"`
	Ticker      *string   `json:"ticker,omitempty"`
	Type        *string   `json:"type,omitempty"`
	CreatedAt   time.Time `json:"created_at"`
	UpdatedAt   time.Time `json:"updated_at"`
}

// ListingsSearchResponse returns paginated listing search results.
type ListingsSearchResponse struct {
	Pagination Pagination        `json:"pagination"`
	Data       []ListingResponse `json:"data"`
}

// DailyUploadAcceptedResponse confirms daily upload enqueue acceptance.
type DailyUploadAcceptedResponse struct {
	UploadID uuid.UUID `json:"upload_id"`
	Status   string    `json:"status"`
}

// DailyUploadRowErrorResponse represents a row-level daily upload parsing/validation error.
type DailyUploadRowErrorResponse struct {
	RowNumber int    `json:"row_number"`
	Reason    string `json:"reason"`
}

// DailyUploadStatusResponse represents the current processing state of a daily upload.
type DailyUploadStatusResponse struct {
	ID            uuid.UUID                     `json:"id"`
	ListingID     uuid.UUID                     `json:"listing_id"`
	Source        string                        `json:"source"`
	Status        string                        `json:"status"`
	StatusMessage string                        `json:"status_message"`
	TotalRows     int                           `json:"total_rows"`
	InsertedRows  int                           `json:"inserted_rows"`
	DuplicateRows int                           `json:"duplicate_rows"`
	ErrorRows     int                           `json:"error_rows"`
	RowErrors     []DailyUploadRowErrorResponse `json:"row_errors"`
	CreatedAt     time.Time                     `json:"created_at"`
	StartedAt     *time.Time                    `json:"started_at,omitempty"`
	FinishedAt    *time.Time                    `json:"finished_at,omitempty"`
	UpdatedAt     time.Time                     `json:"updated_at"`
}

// ManualPortfolioTransactionResponse represents a created manual portfolio transaction.
type ManualPortfolioTransactionResponse struct {
	ID          uuid.UUID  `json:"id"`
	AccountID   uuid.UUID  `json:"account_id"`
	Origin      string     `json:"origin"`
	Source      string     `json:"source"`
	OccurredAt  time.Time  `json:"occurred_at"`
	Type        string     `json:"type"`
	ListingID   *uuid.UUID `json:"listing_id,omitempty"`
	ISIN        *string    `json:"isin,omitempty"`
	Symbol      *string    `json:"symbol,omitempty"`
	Description string     `json:"description"`
	Amount      string     `json:"amount"`
	Quantity    string     `json:"quantity"`
	UnitPrice   string     `json:"unit_price"`
	CreatedAt   time.Time  `json:"created_at"`
	UpdatedAt   time.Time  `json:"updated_at"`
}

// PortfolioTransactionResponse represents one portfolio transaction in read APIs.
type PortfolioTransactionResponse struct {
	ID          uuid.UUID  `json:"id"`
	AccountID   uuid.UUID  `json:"account_id"`
	Origin      string     `json:"origin"`
	Source      string     `json:"source"`
	OccurredAt  time.Time  `json:"occurred_at"`
	Type        string     `json:"type"`
	ListingID   *uuid.UUID `json:"listing_id,omitempty"`
	ISIN        *string    `json:"isin,omitempty"`
	Symbol      *string    `json:"symbol,omitempty"`
	Description string     `json:"description"`
	Amount      string     `json:"amount"`
	Quantity    string     `json:"quantity"`
	UnitPrice   string     `json:"unit_price"`
	CreatedAt   time.Time  `json:"created_at"`
	UpdatedAt   time.Time  `json:"updated_at"`
}

// PortfolioTransactionsResponse returns paginated portfolio transactions.
type PortfolioTransactionsResponse struct {
	Pagination Pagination                     `json:"pagination"`
	Data       []PortfolioTransactionResponse `json:"data"`
}

// AsyncEventAcceptedResponse confirms asynchronous event publication for client polling flows.
type AsyncEventAcceptedResponse struct {
	MessageID uuid.UUID `json:"message_id"`
	Topic     string    `json:"topic"`
	AccountID uuid.UUID `json:"account_id"`
}

// PortfolioSnapshotPointResponse represents one snapshot point in the portfolio timeline.
type PortfolioSnapshotPointResponse struct {
	OccurredAt            time.Time `json:"occurred_at"`
	MarketValue           int64     `json:"market_value"`
	TotalPnL              int64     `json:"total_pnl"`
	TotalPnLPct           float64   `json:"total_pnl_pct"`
	TotalCostBasis        int64     `json:"total_cost_basis"`
	ReturnVsCostBasisPct  float64   `json:"return_vs_cost_basis_pct"`
	DailyReturnPct        float64   `json:"daily_return_pct"`
	TimeWeightedReturnPct float64   `json:"time_weighted_return_pct"`
	ValueIndex            float64   `json:"value_index"`
}

// PortfolioPositionResponse represents one current or closed position for an account.
type PortfolioPositionResponse struct {
	ID               uuid.UUID  `json:"id"`
	Symbol           *string    `json:"symbol,omitempty"`
	Name             *string    `json:"name,omitempty"`
	Quantity         float64    `json:"quantity"`
	CostBasis        int64      `json:"cost_basis"`
	RealizedPnL      int64      `json:"realized_pnl"`
	MarketValue      *int64     `json:"market_value,omitempty"`
	UnrealizedPnLPct *float64   `json:"unrealized_pnl_pct,omitempty"`
	LastSnapshotAt   *time.Time `json:"last_snapshot_at,omitempty"`
	OpenDate         time.Time  `json:"open_date"`
	CloseDate        *time.Time `json:"close_date,omitempty"`
	IsClosed         bool       `json:"is_closed"`
}

// PortfolioPositionsResponse returns account position rows with include-closed metadata.
type PortfolioPositionsResponse struct {
	IncludeClosed bool                        `json:"include_closed"`
	Data          []PortfolioPositionResponse `json:"data"`
}
