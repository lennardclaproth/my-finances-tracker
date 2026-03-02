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

type Pagination struct {
	Limit  int `json:"limit"`
	Offset int `json:"offset"`
	Count  int `json:"count"`
	Total  int `json:"total"`
}

type CashflowTransactionsResponse struct {
	Pagination Pagination    `json:"pagination"`
	Data       []Transaction `json:"data"`
}

type CashflowMonthlyAnalyticsPoint struct {
	Month         string `json:"month" example:"2025-01-01"`
	IncomingCents int64  `json:"incomingCents" example:"250000"`
	OutgoingCents int64  `json:"outgoingCents" example:"120000"`
	NetCents      int64  `json:"netCents" example:"130000"`
}

type CashflowMonthlyAnalyticsResponse struct {
	Data []CashflowMonthlyAnalyticsPoint `json:"data"`
}

type CashflowTagDistributionEntry struct {
	Tag        string `json:"tag" example:"food"`
	TotalCents int64  `json:"totalCents" example:"4200"`
}

type CashflowTagDistributionResponse struct {
	Combined []CashflowTagDistributionEntry `json:"combined"`
	Incoming []CashflowTagDistributionEntry `json:"incoming"`
	Outgoing []CashflowTagDistributionEntry `json:"outgoing"`
}

type TagTransactionsResponse struct {
	UpdatedCount int    `json:"updated_count"`
	Status       string `json:"status"`
}

type VendorResponse struct {
	ID        uuid.UUID `json:"id"`
	Name      string    `json:"name"`
	Type      string    `json:"type"`
	Active    bool      `json:"active"`
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
}

type AccountResponse struct {
	ID         uuid.UUID  `json:"id"`
	ExternalID *uuid.UUID `json:"external_id,omitempty"`
	Name       string     `json:"name"`
	CreatedAt  time.Time  `json:"created_at"`
	UpdatedAt  time.Time  `json:"updated_at"`
}

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

type AsyncEventAcceptedResponse struct {
	MessageID uuid.UUID `json:"message_id"`
	Topic     string    `json:"topic"`
	AccountID uuid.UUID `json:"account_id"`
}

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

type PortfolioPositionsResponse struct {
	IncludeClosed bool                        `json:"include_closed"`
	Data          []PortfolioPositionResponse `json:"data"`
}
