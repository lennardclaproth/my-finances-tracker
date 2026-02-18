package portfolio

import (
	"time"

	"github.com/google/uuid"
	"github.com/lennardclaproth/my-finances-tracker/internal/money"
)

type PositionSnapshot struct {
	ID                    uuid.UUID   `db:"id"`
	AccountSnapID         uuid.UUID   `db:"account_snapshot_id"`
	Symbol                string      `db:"symbol"`
	Name                  *string     `db:"name"`       // empty if unknown
	ListingID             uuid.UUID   `db:"listing_id"` // links back to the listing
	Quantity              float64     `db:"quantity"`   // shares held
	UnitPrice             money.Price `db:"unit_price"` // price per share
	MarketValue           money.Price `db:"market_value"`
	AllocationPct         float64     `db:"allocation_pct"`          // % of account value
	CostBasis             money.Price `db:"cost_basis"`              // total invested in this position (including fees)
	UnrealizedPnL         money.Price `db:"unrealized_pnl"`          // market value - cost basis
	UnrealizedReturnPct   float64     `db:"unrealized_return_pct"`   // (market value - cost basis) / cost basis * 100
	DailyPnL              money.Price `db:"daily_pnl"`               // change in market value since last snapshot
	DailyReturnPct        float64     `db:"daily_return_pct"`        // daily pnl / market value at previous snapshot * 100
	ReturnContributionPct float64     `db:"return_contribution_pct"` // (daily pnl / account market value) * 100, shows how much this position contributed to overall account return that day
	CreatedAt             time.Time   `db:"created_at"`
	UpdatedAt             time.Time   `db:"updated_at"`
}

func NewPositionSnapshot(
	accountSnapID uuid.UUID,
	symbol string,
	listingID uuid.UUID,
	quantity float64,
	unitPrice money.Price,
	costBasis money.Price,
	accountMarketValue money.Price, // needed for allocation %
	prevMarketValue *money.Price, // nil for first snapshot

) *PositionSnapshot {
	// Calculate market value
	marketValue := unitPrice.Float64() * quantity
	// Calculate allocation percentage
	var allocationPct float64
	if accountMarketValue.Float64() != 0 {
		allocationPct = (marketValue / accountMarketValue.Float64()) * 100
	}
	// Calculate unrealized PnL and return
	unrealizedPnL := marketValue - costBasis.Float64()
	// Guard against zero cost basis to avoid division by zero
	var unrealizedReturnPct float64
	if costBasis.Float64() != 0 {
		unrealizedReturnPct = (unrealizedPnL / costBasis.Float64()) * 100
	}
	// Calculate daily PnL and return
	var dailyPnL money.Price   // gets initialized to 0 if prevMarketValue is nil
	var dailyReturnPct float64 // gets initialized to 0 if prevMarketValue is nil
	if prevMarketValue != nil && prevMarketValue.Float64() != 0 {
		dailyPnL = money.Price(marketValue - prevMarketValue.Float64())
		dailyReturnPct = (dailyPnL.Float64() / prevMarketValue.Float64()) * 100
	}
	// Calculate return contribution to overall account return
	var contributionPct float64
	contributionPct = (allocationPct * dailyReturnPct) / 100
	return &PositionSnapshot{
		ID:                    uuid.New(),
		AccountSnapID:         accountSnapID,
		Symbol:                symbol,
		ListingID:             listingID,
		Quantity:              quantity,
		UnitPrice:             unitPrice,
		MarketValue:           money.Price(marketValue),
		AllocationPct:         allocationPct,
		CostBasis:             costBasis,
		UnrealizedPnL:         money.Price(unrealizedPnL),
		UnrealizedReturnPct:   unrealizedReturnPct,
		DailyPnL:              dailyPnL,
		DailyReturnPct:        dailyReturnPct,
		ReturnContributionPct: contributionPct,
	}
}
