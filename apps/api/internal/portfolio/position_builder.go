package portfolio

import (
	"context"
	"fmt"
	"strings"
	"time"

	"github.com/google/uuid"
	"github.com/lennardclaproth/my-finances-tracker/internal/date"
	"github.com/lennardclaproth/my-finances-tracker/internal/marketdata"
	"github.com/lennardclaproth/my-finances-tracker/internal/money"
)

type PositionBuilder struct {
	ps  PositionStore
	ts  TransactionStore
	ls  ListingStore
	mds MarketDataService
}

type PositionStore interface {
	Clean(ctx context.Context, accID uuid.UUID) error
	CreateMany(ctx context.Context, positions []*Position) error
	GetByID(ctx context.Context, id uuid.UUID) (*Position, error)
	GetLastSnapshot(ctx context.Context, positionID uuid.UUID) (*PositionSnapshot, error)
	CreateSnapshot(ctx context.Context, snap *PositionSnapshot) error
}

type TransactionStore interface {
	GetASC(ctx context.Context, accID uuid.UUID) ([]Transaction, error)
	UpdatePositions(ctx context.Context, transactions []Transaction) error
	GetByPositionID(ctx context.Context, positionID uuid.UUID, from *time.Time) ([]Transaction, error)
}

type ListingStore interface {
	FetchBySymbolOrISIN(ctx context.Context, val string) (*marketdata.Listing, error)
	FetchByID(ctx context.Context, id uuid.UUID) (*marketdata.Listing, error)
}

type MarketDataService interface {
	GetDailies(ctx context.Context, symbol string, from, to *time.Time, limit, offset int) (*marketdata.DailyResponse, error)
}

func NewPositionBuilder(ps PositionStore, ts TransactionStore, ls ListingStore, mds MarketDataService) *PositionBuilder {
	return &PositionBuilder{ps: ps, ts: ts, ls: ls, mds: mds}
}

// CalculatePortfolioPerformance calculates the performance metrics of the portfolio
// based on the transactions. It does this by first determining the positions in the portfolio
// checks if they are opened or closed and then calculates the performance of those.
func CalculateROI(posID, accID uuid.UUID) {
	// 1. Get all transactions for the account
	// 2. Determine positions based on transactions (open/close)
	// 3. Map positions to listings to get current market prices
	// 4. For each position, calculate performance metrics (PnL, return %, etc.)
	// 5. Aggregate position performance into account-level performance metrics
}

func (s *PositionBuilder) BuildPositionSnapshots(
	ctx context.Context,
	accID uuid.UUID,
	positionID uuid.UUID,
) error {
	_ = accID
	// Determine start date from when on we want the snapshots to be built.
	var startDate time.Time
	pos, err := s.ps.GetByID(ctx, positionID)
	if err != nil {
		return fmt.Errorf("build position snapshots: get position: %w", err)
	}
	if pos == nil {
		return ErrPositionNotFound
	}
	if pos.ListingID == nil {
		return nil // if we don't have a listing for this position, we can't build snapshots, so we just return
	}
	// Get the listing for the position to get the historical market prices
	listing, err := s.ls.FetchByID(ctx, *pos.ListingID)
	if err != nil {
		return fmt.Errorf("build position snapshots: fetch listing: %w", err)
	}
	if listing == nil {
		return nil // if we don't have a listing for this position, we can't build snapshots, so we just return
	}
	if listing.Symbol == "" {
		return nil // if we don't have a symbol for this listing, we can't build snapshots, so we just return
	}
	// Setup the time window of a position. The start date is the position open date, and the end date is either
	// the position close date or today (if still open).
	startDate = pos.OpenDate
	endExclusive := date.StartOfDayUTC(time.Now())
	if pos.CloseDate != nil && date.StartOfDayUTC(*pos.CloseDate).AddDate(0, 0, 1).Before(endExclusive) {
		endExclusive = date.StartOfDayUTC(*pos.CloseDate).AddDate(0, 0, 1)
	}
	lps, err := s.ps.GetLastSnapshot(ctx, positionID)
	if err != nil {
		return fmt.Errorf("build position snapshots: get last snapshot: %w", err)
	}
	if lps != nil && date.StartOfDayUTC(lps.OccurredAt).Before(endExclusive) {
		startDate = date.StartOfDayUTC(lps.OccurredAt).AddDate(0, 0, 1)
	}
	// Get all transactions for the position, sorted by time
	ts, err := s.ts.GetByPositionID(ctx, positionID, &startDate)
	if err != nil {
		return fmt.Errorf("build position snapshots: get transactions: %w", err)
	}
	// Get historical market data for the listing from the position open date to today
	dailies, err := s.mds.GetDailies(ctx, listing.Symbol, &startDate, nil, 0, 0)
	if err != nil {
		return fmt.Errorf("build position snapshots: get dailies: %w", err)
	}
	// Initialize iterators for transactions and daily data, and an accumulator for the position state
	txIdx := 0
	dIdx := 0
	var prevSnapshot *PositionSnapshot
	acc := PositionAcc{}
	// If we have a last snapshot, we should start with that as the initial state for the position,
	// and then apply any transactions that occurred after that snapshot to get to the current state.
	if lps != nil {
		acc.CostBasis = lps.CostBasis
		acc.Quantity = lps.Quantity
		acc.RealizedPnL = lps.RealizedPnL
		acc.Income = lps.Income
		acc.Fees = lps.Fees
		acc.Taxes = lps.Taxes
		prevSnapshot = lps
	}
	// Walk through the days from the start date to today, and build a snapshot for each day,
	// based on the transactions that occurred on that day and the market price for that day.
	for d := date.StartOfDayUTC(startDate); d.Before(endExclusive); d = d.AddDate(0, 0, 1) {
		// Set day end for the current day, this is used to determine which transactions
		// to include in the snapshot for this day (i.e. all transactions that occurred
		// before the end of the day)
		dayEnd := date.EndOfDayUTC(d)
		unitPriceSet := false
		var unitPrice money.Price
		// apply all tx up to this day
		for txIdx < len(ts) && !ts[txIdx].OccurredAt.UTC().After(dayEnd) {
			tx := ts[txIdx]
			if err := acc.ApplyTx(tx); err != nil {
				return fmt.Errorf("build position snapshots: apply tx: %w", err)
			}
			// If the transaction is a buy or sell, we can calculate the market value based on the
			// transaction price, this is an approximation but it's better than carrying forward
			// the previous market value without any adjustments.
			if tx.Type == TxBuy || tx.Type == TxSell {
				unitPrice = tx.UnitPrice
				unitPriceSet = true
			}
			txIdx++
		}
		// find marketclose for this day (advance pointer)
		// If daily data is available, calculate market value based on the close price for that day.
		for dIdx < len(dailies.Data) && date.StartOfDayUTC(dailies.Data[dIdx].Date).Before(d) {
			dIdx++
		}
		if dIdx < len(dailies.Data) && date.SameDayUTC(dailies.Data[dIdx].Date, d) {
			unitPrice = dailies.Data[dIdx].Close
			unitPriceSet = true
		}
		// If we don't have daily data for this day, we carry forward the previous market value.
		// If a buy or sell transaction occurred on this day we can calculate the market value based on the
		// transaction price, this is an approximation but it's better than carrying forward the previous market
		// value without any adjustments.
		if !unitPriceSet {
			unitPrice = prevSnapshot.UnitPrice
		}
		snap, err := NewPositionSnapshot(
			positionID,
			accID,
			listing.ID,
			listing.Symbol,
			listing.Name,
			acc.Quantity,
			unitPrice,
			acc.CostBasis,
			acc.RealizedPnL,
			acc.Income,
			acc.Fees,
			acc.Taxes,
			d,
			prevSnapshot,
		)
		if err != nil {
			return fmt.Errorf("build position snapshots: new snapshot: %w", err)
		}
		if err := s.ps.CreateSnapshot(ctx, snap); err != nil {
			return fmt.Errorf("build position snapshots: create snapshot: %w", err)
		}
		prevSnapshot = snap
	}

	return nil
}

func (s *PositionBuilder) BuildPositions(ctx context.Context, accID uuid.UUID) error {
	// Clean slate: derived data gets rebuilt deterministically from transactions.
	if err := s.ps.Clean(ctx, accID); err != nil {
		return fmt.Errorf("build positions: clean: %w", err)
	}
	// Load the event stream (transactions) in chronological order.
	ts, err := s.ts.GetASC(ctx, accID)
	if err != nil {
		return fmt.Errorf("build positions: load transactions: %w", err)
	}
	// One active lifecycle ("cycle") per canonical instrument key at a time.
	// key = ISIN if available; otherwise symbol (with alias promotion later).
	activeByKey := make(map[string]*Position)
	// alias map to stitch symbol-only transactions to an ISIN once the ISIN appears later.
	// Example: first tx has Symbol="AAPL" only; later tx has ISIN="US0378331005".
	aliases := make(map[string]string) // symbol -> isin
	// We persist every created cycle (open or closed).
	allPositions := make([]*Position, 0)
	for i := range ts {
		tx := &ts[i]
		// Cash transactions don't map to security positions.
		if tx.Type == TxCash {
			continue
		}
		// Determine canonical key and a symbolKey (used for promotion).
		key, symbolKey, err := canonicalPositionKey(*tx, aliases)
		if err != nil {
			return fmt.Errorf("build positions: canonical key: %w", err)
		}
		// Fetch active position cycle; if ISIN appears later, promote symbol-cycle to ISIN key.
		pos := getOrPromoteCycle(activeByKey, key, symbolKey, *tx)
		// Apply transaction according to type. This helper is where the branching lives,
		// so the loop stays easy to read.
		pos, created, err := applyTxToCycle(accID, tx, key, symbolKey, aliases, activeByKey, pos)
		if err != nil {
			return fmt.Errorf("build positions: apply transaction: %w", err)
		}
		if created {
			allPositions = append(allPositions, pos)
		}
	}
	// Best-effort listing mapping: attach ListingID when possible, but never drop positions.
	posList := make([]*Position, 0, len(allPositions))
	for _, pos := range allPositions {
		id, err := pos.Identity()
		if err == nil {
			if listing, err := s.ls.FetchBySymbolOrISIN(ctx, id); err == nil && listing != nil {
				pos.ListingID = &listing.ID
			}
		}
		posList = append(posList, pos)
	}
	// Persist positions (cycles).
	if err := s.ps.CreateMany(ctx, posList); err != nil {
		return fmt.Errorf("build positions: persist positions: %w", err)
	}
	// Persist transaction->position mapping.
	if err := s.ts.UpdatePositions(ctx, ts); err != nil {
		return fmt.Errorf("build positions: persist transaction mapping: %w", err)
	}
	return nil
}

// getOrPromoteCycle returns the active cycle for "key".
// If tx now has an ISIN and we previously created a symbol-only cycle, we promote it:
// move activeByKey[symbolKey] -> activeByKey[key].
func getOrPromoteCycle(activeByKey map[string]*Position, key, symbolKey string, tx Transaction) *Position {
	pos := activeByKey[key]
	if pos != nil {
		return pos
	}
	// Promote symbol-only cycle to ISIN key once ISIN is known.
	if tx.ISIN != nil && symbolKey != "" {
		if promoted, ok := activeByKey[symbolKey]; ok {
			delete(activeByKey, symbolKey)
			activeByKey[key] = promoted
			return promoted
		}
	}
	return nil
}

// applyTxToCycle applies a transaction to a position cycle and assigns tx.PositionID.
// It may create a new cycle if needed (returns created=true in that case).
func applyTxToCycle(
	accID uuid.UUID,
	tx *Transaction,
	key, symbolKey string,
	aliases map[string]string,
	activeByKey map[string]*Position,
	pos *Position,
) (*Position, bool, error) {
	ensure := func() (*Position, bool, error) {
		if pos != nil {
			return pos, false, nil
		}
		// Start a new cycle with the best identity we can infer at this point.
		pos, err := newPositionCycle(accID, *tx, symbolKey, aliases)
		if err != nil {
			return nil, false, err
		}
		activeByKey[key] = pos
		return pos, true, nil
	}
	switch tx.Type {
	case TxBuy:
		pos, created, err := ensure()
		if err != nil {
			return nil, false, err
		}
		mergePositionIdentity(pos, *tx)
		pos.ApplyTx(*tx)
		tx.PositionID = &pos.ID
		return pos, created, nil
	case TxSell:
		// No active cycle to reduce/close. ignore silently (current behavior)
		if pos == nil {
			return nil, false, nil
		}
		mergePositionIdentity(pos, *tx)
		pos.ApplyTx(*tx)
		tx.PositionID = &pos.ID
		// Close the cycle when quantity reaches zero.
		if pos.Quantity == 0 {
			delete(activeByKey, key)
		}
		return pos, false, nil
	case TxDividend, TxTax, TxFee:
		// These affect cost basis; if we don't have a cycle yet, start one.
		pos, created, err := ensure()
		if err != nil {
			return nil, false, err
		}
		mergePositionIdentity(pos, *tx)
		pos.ApplyTx(*tx)
		tx.PositionID = &pos.ID
		return pos, created, nil
	default:
		// Unknown tx type: ignore or error depending on your domain strictness.
		return pos, false, nil
	}
}

func canonicalPositionKey(tx Transaction, aliases map[string]string) (key, symbolKey string, err error) {
	isin := normalizedID(tx.ISIN)
	symbol := normalizedID(tx.Symbol)
	if isin == "" && symbol == "" {
		return "", "", ErrTransactionISINAndSymbolMissing
	}
	if symbol != "" {
		symbolKey = symbol
	}
	if isin != "" {
		key = isin
		if symbol != "" {
			aliases[symbol] = isin
		}
		return key, symbolKey, nil
	}
	if mapped, ok := aliases[symbol]; ok && mapped != "" {
		key = mapped
	} else {
		key = symbol
	}
	return key, symbolKey, nil
}

func newPositionCycle(accID uuid.UUID, tx Transaction, symbolKey string, aliases map[string]string) (*Position, error) {
	isin := tx.ISIN
	if isin == nil && symbolKey != "" {
		if mapped, ok := aliases[symbolKey]; ok && mapped != "" && mapped != symbolKey {
			v := mapped
			isin = &v
		}
	}
	return NewPosition(accID, isin, tx.Symbol, nil, tx.OccurredAt)
}

func mergePositionIdentity(pos *Position, tx Transaction) {
	if pos == nil {
		return
	}
	if pos.ISIN == nil && tx.ISIN != nil && normalizedID(tx.ISIN) != "" {
		v := normalizedID(tx.ISIN)
		pos.ISIN = &v
	}
	if pos.Symbol == nil && tx.Symbol != nil && normalizedID(tx.Symbol) != "" {
		v := normalizedID(tx.Symbol)
		pos.Symbol = &v
	}
}

func normalizedID(v *string) string {
	if v == nil {
		return ""
	}
	return strings.TrimSpace(*v)
}

// getPositionFromMap tries to find an existing position for the given transaction based on ISIN or symbol.
// If no position exists, it creates a new one and adds it to the map.
func getPositionFromMap(accID uuid.UUID, positions *map[string]*Position, tx Transaction) (*Position, error) {
	txID, err := tx.GetID()
	if err != nil {
		return nil, err
	}
	// Fallback to symbol-based mapping if ISIN is not available
	if pos, ok := (*positions)[txID]; ok {
		return pos, nil
	}
	// No existing position found for this transaction, create a new one
	pos, err := NewPosition(accID, tx.ISIN, tx.Symbol, nil, tx.OccurredAt)
	if err != nil {
		return nil, err
	}
	id, err := pos.Identity()
	if err != nil {
		return nil, err
	}
	(*positions)[id] = pos
	return pos, nil
}
