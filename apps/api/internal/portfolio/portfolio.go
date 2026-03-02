package portfolio

import (
	"fmt"
	"math"
	"time"

	"github.com/google/uuid"
	"github.com/lennardclaproth/my-finances-tracker/internal/money"
)

type PortfolioSnapshot struct {
	ID         uuid.UUID `db:"id"`
	AccountID  uuid.UUID `db:"account_id"`
	OccurredAt time.Time `db:"occurred_at"`

	MarketValue money.Price `db:"market_value"` // MarketValue = Σ position.MarketValue + CashPosition; We omit CashPosition now because it is not calculated yet.
	CostBasis   money.Price `db:"cost_basis"`   // TotalCostbasis = Σ position.Costbasis

	RealizedPnL      money.Price `db:"realized_pnl"`   // RealizedPnL = Σ position.RealizedPnL
	UnrealizedPnL    money.Price `db:"unrealized_pnl"` // UnrealizedPnL = Σ position.UnrealizedPnL
	UnrealizedPnLPct float64     `db:"unrealized_pnl_pct"`
	Income           money.Price `db:"income"` // Income = Σ position.Income
	Fees             money.Price `db:"fees"`   // Fees = Σ position.Fees
	Taxes            money.Price `db:"taxes"`  // Taxes = Σ position.Taxes
	CashBalance      money.Price `db:"cash_balance"`

	TotalPnL    money.Price `db:"total_pnl"` // TotalPnL = Σ position.TotalPnL || TotalPnL = portfolioSnapshot.RealizedPnL + portfolioSnapshot.Unrealized + Income - fees - taxes
	TotalPnLPct float64     `db:"total_pnl_pct"`

	DailyDeltaPnL         money.Price `db:"daily_delta_pnl"` // DailyPnL = TotalPnL - prevSnapshot.TotalPnL
	DailyDeltaPnLPct      float64     `db:"daily_delta_pnl_pct"`
	NetCashflow           money.Price `db:"net_cashflow"`            // NetCashflow = (MarketValue(t) - MarketValue(t-1)) - (TotalPnL(t) - TotalPnL(t-1))
	CumulativeNetCashflow money.Price `db:"cumulative_net_cashflow"` // CumulativeNetCashflow = Σ NetCashFlow(d) where d <= t and t is end time.
	TimeWeightedReturnPct float64     `db:"time_weighted_return_pct"`

	CreatedAt time.Time `db:"created_at"`
	UpdatedAt time.Time `db:"updated_at"`
}

var (
	ErrBuildInProgress      = fmt.Errorf("sync is already in progress for this listing")
	ErrPortfolioNoSnapshots = fmt.Errorf("no position snapshots available to build portfolio")
)

func NewPortfolioSnapshot(
	accID uuid.UUID,
	occurredAt time.Time,
	cashBalance money.Price,
	posSnapshots []*PositionSnapshot,
	prevSnapshot *PortfolioSnapshot,
) *PortfolioSnapshot {
	now := time.Now()
	// loop over posSnapshots and set values
	var marketValue, costBasis money.Price
	var realizedPnL, unrealizedPnL, income, fees, taxes money.Price
	var totalPnL money.Price
	marketValue += cashBalance
	for _, pss := range posSnapshots {
		marketValue += pss.MarketValue
		costBasis += pss.CostBasis
		realizedPnL += pss.RealizedPnL
		unrealizedPnL += pss.UnrealizedPnL
		income += pss.Income
		fees += pss.Fees
		taxes += pss.Taxes
		totalPnL += pss.TotalPnL
	}
	var unrealizedPnLPct float64
	var totalPnLPct float64
	if cb := costBasis.Float64(); cb != 0 {
		unrealizedPnLPct = (unrealizedPnL.Float64() / cb) * 100
		totalPnLPct = (totalPnL.Float64() / cb) * 100
	}
	// set cashflow values
	var netCashflow money.Price
	var cumulativeNetCashflow money.Price
	if prevSnapshot != nil {
		netCashflow = (marketValue - prevSnapshot.MarketValue) - (totalPnL - prevSnapshot.TotalPnL)
		cumulativeNetCashflow = prevSnapshot.CumulativeNetCashflow + netCashflow
	} else {
		netCashflow = 0
		cumulativeNetCashflow = 0
	}
	// set daily values
	var dailyDeltaPnL money.Price
	var dailyDeltaPnLPct float64
	if prevSnapshot != nil && prevSnapshot.TotalPnL.Float64() != 0 {
		dailyDeltaPnL = totalPnL - prevSnapshot.TotalPnL
		if prevSnapshot.TotalPnL.Float64() != 0 {
			dailyDeltaPnLPct = (dailyDeltaPnL.Float64() / math.Abs(prevSnapshot.TotalPnL.Float64())) * 100
		}
	}
	// time-weighted daily return
	var timeWeightedReturnPct float64
	if prevSnapshot != nil && prevSnapshot.MarketValue.Float64() != 0 {
		r := (marketValue.Float64()-netCashflow.Float64())/prevSnapshot.MarketValue.Float64() - 1
		timeWeightedReturnPct = r * 100
	}
	return &PortfolioSnapshot{
		ID:         uuid.New(),
		AccountID:  accID,
		OccurredAt: occurredAt,

		MarketValue: marketValue,
		CostBasis:   costBasis,

		RealizedPnL:      realizedPnL,
		UnrealizedPnL:    unrealizedPnL,
		UnrealizedPnLPct: unrealizedPnLPct,
		Income:           income,
		Fees:             fees,
		Taxes:            taxes,
		CashBalance:      cashBalance,

		TotalPnL:    totalPnL,
		TotalPnLPct: totalPnLPct,

		DailyDeltaPnL:         dailyDeltaPnL,
		DailyDeltaPnLPct:      dailyDeltaPnLPct,
		TimeWeightedReturnPct: timeWeightedReturnPct,
		NetCashflow:           netCashflow,
		CumulativeNetCashflow: cumulativeNetCashflow,

		CreatedAt: now,
		UpdatedAt: now,
	}
}
