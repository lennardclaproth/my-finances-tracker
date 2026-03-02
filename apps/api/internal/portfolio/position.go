package portfolio

import (
	"fmt"
	"math"
	"time"

	"github.com/google/uuid"
	"github.com/lennardclaproth/my-finances-tracker/internal/money"
)

type PositionAcc struct {
	Quantity  float64
	CostBasis money.Price

	Income money.Price
	Fees   money.Price
	Taxes  money.Price

	RealizedPnL money.Price
}

func (a *PositionAcc) ApplyTx(tx Transaction) error {
	switch tx.Type {
	case TxBuy:
		err := a.applyBuy(tx)
		if err != nil {
			return err
		}
	case TxSell:
		err := a.applySell(tx)
		if err != nil {
			return err
		}
	case TxDividend:
		a.Income += money.Price(tx.AmountCents)
	case TxFee:
		a.Fees += money.Price(tx.AmountCents)
	case TxTax:
		a.Taxes += money.Price(tx.AmountCents)
	default:
		// ignore unknown types (or panic/error based on strictness)
	}
	return nil
}

func (a *PositionAcc) applyBuy(tx Transaction) error {
	if tx.Quantity <= 0 {
		return money.ErrInvalidPrice
	}
	a.Quantity += tx.Quantity
	// total notional added to cost basis (unit price * qty)
	notional, err := money.NewPrice(tx.UnitPrice.Float64() * tx.Quantity)
	if err != nil {
		return err
	}
	a.CostBasis += notional

	return nil
}

func (a *PositionAcc) applySell(tx Transaction) error {
	// guard against quantity being 0
	if a.Quantity <= 0 || tx.Quantity <= 0 {
		return money.ErrInvalidPrice
	}
	matchedQty := tx.Quantity
	if matchedQty > a.Quantity {
		matchedQty = a.Quantity // clamp oversell
	}
	// calculate average unit cost
	avgUnitCost, err := calcAvgUnitCost(a.CostBasis.Float64(), a.Quantity)
	if err != nil {
		return fmt.Errorf("applySell failed to calculate avg unit cost: %w", err)
	}
	costRemoved, err := money.NewPrice(avgUnitCost.Float64() * matchedQty)
	if err != nil {
		return err
	}
	// calculate RealizedPnL
	proceeds, err := money.NewPrice(tx.UnitPrice.Float64() * matchedQty)
	if err != nil {
		return err
	}
	a.RealizedPnL += proceeds - costRemoved

	a.Quantity -= matchedQty
	a.CostBasis -= costRemoved

	if a.Quantity == 0 || a.CostBasis < 0 {
		a.CostBasis = 0
	}
	return nil
}

func calcAvgUnitCost(costBasis, quantity float64) (money.Price, error) {
	if quantity == 0 || math.IsNaN(quantity) || math.IsInf(quantity, 0) {
		return 0, nil
	}
	return money.NewPrice(costBasis / quantity)
}

// Position represents an open or closed position in the portfolio. It maps back to a listing via the listing ID
type Position struct {
	ID        uuid.UUID  `db:"id"`
	AccountID uuid.UUID  `db:"account_id"`
	ISIN      *string    `db:"isin"`
	Symbol    *string    `db:"symbol"`
	ListingID *uuid.UUID `db:"listing_id"` // can be null if we couldn't map it to a listing (e.g. unknown symbol, cash transaction, etc.)

	OpenDate  time.Time  `db:"open_date"`
	CloseDate *time.Time `db:"close_date"` // nil if still open

	Quantity    float64     `db:"quantity"`
	Fees        money.Price `db:"fees"`
	CostBasis   money.Price `db:"cost_basis"` // total invested in this position
	Income      money.Price `db:"income"`     // dividends, coupon, interest etc.
	Taxes       money.Price `db:"taxes"`
	RealizedPnL money.Price `db:"realized_pnl"`

	CreatedAt time.Time `db:"created_at"`
	UpdatedAt time.Time `db:"updated_at"`
}

var (
	ErrPositionISINAndSymbolMissing  = fmt.Errorf("isin and symbol are both missing, cannot determine position ID")
	ErrPositionNotFound              = fmt.Errorf("position not found")
	ErrPositionSnapshotAlreadyExists = fmt.Errorf("position snapshot already exists for the occurred day")
)

func NewPosition(accountID uuid.UUID, isin, symbol *string, listingID *uuid.UUID, openDate time.Time) (*Position, error) {
	if isin == nil && symbol == nil {
		return nil, ErrPositionISINAndSymbolMissing
	}
	return &Position{
		ID:        uuid.New(),
		AccountID: accountID,
		ISIN:      isin,
		Symbol:    symbol,
		ListingID: listingID,
		OpenDate:  openDate,
		Quantity:  0,
		CostBasis: 0,
		CreatedAt: time.Now(),
		UpdatedAt: time.Now(),
	}, nil
}

func (p *Position) Identity() (string, error) {
	if p.ISIN != nil {
		return *p.ISIN, nil
	} else if p.Symbol != nil {
		return *p.Symbol, nil
	}
	return "", ErrPositionISINAndSymbolMissing
}

func (p *Position) ApplyTx(tx Transaction) error {
	acc := PositionAcc{
		Quantity:    p.Quantity,
		CostBasis:   p.CostBasis,
		Income:      p.Income,
		Fees:        p.Fees,
		Taxes:       p.Taxes,
		RealizedPnL: p.RealizedPnL,
	}

	if err := acc.ApplyTx(tx); err != nil {
		return err
	}

	p.Quantity = acc.Quantity
	p.CostBasis = acc.CostBasis
	p.Income = acc.Income
	p.Fees = acc.Fees
	p.Taxes = acc.Taxes
	p.RealizedPnL = acc.RealizedPnL

	// lifecycle
	if p.Quantity == 0 {
		p.CloseDate = &tx.OccurredAt
	} else {
		p.CloseDate = nil
	}

	// timestamps
	if p.Quantity > 0 && p.OpenDate.IsZero() {
		p.OpenDate = tx.OccurredAt
	}
	p.UpdatedAt = time.Now()
	return nil
}

type PositionSnapshot struct {
	ID         uuid.UUID `db:"id"`
	AccountID  uuid.UUID `db:"account_id"`  // Reference to the account
	PositionID uuid.UUID `db:"position_id"` // Reference to the position

	Symbol     string    `db:"symbol"` // The symbol of the stock
	Name       *string   `db:"name"`
	ListingID  uuid.UUID `db:"listing_id"`  // Reference to the listing
	OccurredAt time.Time `db:"occurred_at"` // The day of the snapshot

	Quantity    float64     `db:"quantity"`      // The quantity of the units in the position
	UnitPrice   money.Price `db:"unit_price"`    // The price of a single unit
	AvgPrice    money.Price `db:"average_price"` // The average price, this is excluding fees: AveragePrice = Costbasis / Quantity
	MarketValue money.Price `db:"market_value"`  // The complete market value of the position: MarketValue = UnitPrice * Quantity

	CostBasis        money.Price `db:"cost_basis"`         // The complete cost of the position excluding dividends and fees
	Income           money.Price `db:"income"`             // Dividends/ interest received
	Fees             money.Price `db:"fees"`               // Transaction fees
	Taxes            money.Price `db:"taxes"`              // Tax fees
	TotalPnL         money.Price `db:"total_pnl"`          // TotalPnl = Realized + Unrealized + Income - Fees - Taxes
	TotalPnLPct      float64     `db:"total_pnl_pct"`      // TotalPnlPct = TotalPnL / CostBasis * 100
	RealizedPnL      money.Price `db:"realized_pnl"`       // The realized profit: RealizedPnL = Proceeds - Costbasis; Technically speaking transaction costs should be included here. We omit them for now.
	UnrealizedPnL    money.Price `db:"unrealized_pnl"`     // Unrealized hypothetical profit: UnrealizedPnL = MarketValue - Costbasis
	UnrealizedPnLPct float64     `db:"unrealized_pnl_pct"` // The unrealized hypothetical profit in pct

	DailyDeltaPnL    money.Price `db:"daily_delta_pnl"`     // Daily profit
	DailyDeltaPnLPct float64     `db:"daily_delta_pnl_pct"` // Daily profit in pct to see which positions perform well

	CreatedAt time.Time `db:"created_at"`
	UpdatedAt time.Time `db:"updated_at"`
}

type PositionWithLatestSnapshot struct {
	ID               uuid.UUID   `db:"id"`
	Symbol           *string     `db:"symbol"`
	Name             *string     `db:"name"`
	Quantity         float64     `db:"quantity"`
	CostBasis        money.Price `db:"cost_basis"`
	RealizedPnL      money.Price `db:"realized_pnl"`
	MarketValue      *int64      `db:"market_value"`
	UnrealizedPnLPct *float64    `db:"unrealized_pnl_pct"`
	LastSnapshotAt   *time.Time  `db:"last_snapshot_at"`
	OpenDate         time.Time   `db:"open_date"`
	CloseDate        *time.Time  `db:"close_date"`
	IsClosed         bool        `db:"is_closed"`
}

func NewPositionSnapshot(
	positionID, accountID, listingID uuid.UUID,
	symbol, name string,
	quantity float64,
	unitPrice money.Price,
	costBasis money.Price,
	realizedPnL money.Price,
	income money.Price,
	fees money.Price,
	taxes money.Price,
	occurredAt time.Time,
	prevSnapshot *PositionSnapshot,
) (*PositionSnapshot, error) {
	now := time.Now()
	// Calculate market value
	marketValue, err := money.NewPrice(unitPrice.Float64() * quantity)
	if err != nil {
		return nil, err
	}
	// Calculate totals
	unrealizedPnL := marketValue - costBasis
	totalPnL := unrealizedPnL + realizedPnL + income - taxes - fees
	var unrealizedPnLPct float64
	var totalPnLPct float64
	if cb := costBasis.Float64(); cb != 0 {
		unrealizedPnLPct = (unrealizedPnL.Float64() / cb) * 100
		totalPnLPct = (totalPnL.Float64() / cb) * 100
	}
	// Calculate daily change
	var dailyDeltaPnL money.Price
	var dailyDeltaPnLPct float64
	if prevSnapshot != nil && prevSnapshot.TotalPnL.Float64() != 0 {
		dailyDeltaPnL = totalPnL - prevSnapshot.TotalPnL
		if prevSnapshot.TotalPnL.Float64() != 0 {
			dailyDeltaPnLPct = (dailyDeltaPnL.Float64() / math.Abs(prevSnapshot.TotalPnL.Float64())) * 100
		}
	}
	// Calc avg price
	// TODO: fix error here
	avg, _ := calcAvgUnitCost(costBasis.Float64(), quantity)
	return &PositionSnapshot{
		ID:         uuid.New(),
		AccountID:  accountID,
		PositionID: positionID,

		Symbol:    symbol,
		Name:      &name,
		ListingID: listingID,

		Quantity:    quantity,
		UnitPrice:   unitPrice,
		AvgPrice:    avg,
		MarketValue: marketValue,

		CostBasis:        costBasis,
		Income:           income,
		Fees:             fees,
		Taxes:            taxes,
		TotalPnL:         totalPnL,
		TotalPnLPct:      totalPnLPct,
		RealizedPnL:      realizedPnL,
		UnrealizedPnL:    unrealizedPnL,
		UnrealizedPnLPct: unrealizedPnLPct,

		DailyDeltaPnL:    dailyDeltaPnL,
		DailyDeltaPnLPct: dailyDeltaPnLPct,

		OccurredAt: occurredAt,
		CreatedAt:  now,
		UpdatedAt:  now,
	}, nil
}
