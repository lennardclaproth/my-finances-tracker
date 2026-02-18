package portfolio

import (
	"time"

	"github.com/google/uuid"
	"github.com/lennardclaproth/my-finances-tracker/internal/money"
)

type Account struct {
	ID   uuid.UUID
	Name string
}

func NewAccount(name string) *Account {
	return &Account{
		ID:   uuid.New(),
		Name: name,
	}
}

type AccountSnapshot struct {
	ID                 uuid.UUID   `db:"id"`
	AccountID          uuid.UUID   `db:"account_id"`
	InvestedGross      money.Price `db:"invested_gross"`
	InvestedNet        money.Price `db:"invested_net"`
	MarketValue        money.Price `db:"market_value"`
	ProfitLoss         money.Price `db:"profit_loss"`
	ReturnPct          float64     `db:"return_pct"`
	DailyProfitLoss    money.Price `db:"daily_profit_loss"`
	DailyReturnPct     float64     `db:"daily_return_pct"`
	ValueIndex         float64     `db:"value_index"`
	TimeWeightedReturn float64     `db:"time_weighted_return"`
	Time               time.Time   `db:"time"`
	NetCashFlow        money.Price `db:"net_cash_flow"` // deposits - withdrawals in this period (0 if none)
	CreatedAt          time.Time   `db:"created_at"`
	UpdatedAt          time.Time   `db:"updated_at"`
}

func NewAccountSnapshot(
	accID uuid.UUID,
	date time.Time,
	investedGross, investedNet, marketValue, originalValue money.Price,
	prevMarketValue *money.Price, // nil for first snapshot
	netCashFlow money.Price, // deposits - withdrawals in this period (0 if none)
) *AccountSnapshot {
	// Calculate profitLoss
	profitLoss := marketValue - investedNet
	var returnPct float64
	// Guard against zero bought net to avoid division by zero
	if bn := investedNet.Float64(); bn != 0 {
		returnPct = (profitLoss.Float64() / bn) * 100
	}
	// Calculate valueIndex (relative performance to initial snapshot)
	var valueIndex float64
	if ow := originalValue.Float64(); ow != 0 {
		valueIndex = (marketValue.Float64() / ow) * 100
	}
	// Calculate daily movement
	var dailyProfitLoss money.Price   // gets initialized to 0 if prevMarketValue is nil
	var dailyReturnPct float64        // gets initialized to 0 if prevMarketValue is nil
	var timeWeightedReturnPct float64 // gets initialized to 0 if prevMarketValue is nil
	// If prevWorth is not nil, calculate daily delta and weighted performance.
	if prevMarketValue != nil && prevMarketValue.Float64() != 0 {
		dailyProfitLoss = marketValue - *prevMarketValue
		dailyReturnPct = (dailyProfitLoss.Float64() / prevMarketValue.Float64()) * 100
		// Cash-flow adjusted daily return (time-weighted-ish)
		// WeightedReturn = (W_t - W_{t-1} - CF_t) / W_{t-1}
		timeWeightedReturnPct = ((marketValue.Float64() - prevMarketValue.Float64() - netCashFlow.Float64()) / prevMarketValue.Float64()) * 100
	}
	return &AccountSnapshot{
		ID:                 uuid.New(),
		AccountID:          accID,
		InvestedGross:      investedGross,
		InvestedNet:        investedNet,
		MarketValue:        marketValue,
		Time:               date,
		ProfitLoss:         profitLoss,
		ReturnPct:          returnPct,
		ValueIndex:         valueIndex,
		DailyProfitLoss:    dailyProfitLoss,
		DailyReturnPct:     dailyReturnPct,
		TimeWeightedReturn: timeWeightedReturnPct,
		NetCashFlow:        netCashFlow,
		CreatedAt:          time.Now(),
		UpdatedAt:          time.Now(),
	}
}
