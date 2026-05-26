package portfolio

import (
	"context"
	"errors"
	"fmt"
	"time"

	"github.com/google/uuid"
	"github.com/lennardclaproth/my-finances-tracker/internal/date"
	"github.com/lennardclaproth/my-finances-tracker/internal/marketdata"
	"github.com/lennardclaproth/my-finances-tracker/internal/sorting"

	"github.com/lennardclaproth/my-finances-tracker/internal/money"
)

type Builder struct {
	mdq *marketdata.Queries
	pfs PortfolioStore
	pss PositionStore
	ts  TransactionStore
	lk  Locker
}

type PositionStore interface {
	GetLastSnapshot(ctx context.Context, positionID uuid.UUID) (*PositionSnapshot, error)
	CreateSnapshot(ctx context.Context, snap *PositionSnapshot) error
	CreateMany(ctx context.Context, positions []*Position) error
	UpdatePositions(ctx context.Context, transactions []Transaction) error
}

type PortfolioStore interface {
	CreateSnapshot(ctx context.Context, snap *PortfolioSnapshot) error
	Clean(ctx context.Context, accID uuid.UUID) error
}

type TransactionStore interface {
	TransactionsForPosition(ctx context.Context, positionID uuid.UUID, from *time.Time) ([]Transaction, error)
	TransactionsForAccount(ctx context.Context, accID uuid.UUID, sort string) ([]Transaction, error)
}

type Locker interface {
	TryAcquireBuildLock(ctx context.Context, accID uuid.UUID) (bool, error)
	ReleaseBuildLock(ctx context.Context, accID uuid.UUID) error
}

func (b *Builder) buildPositionSnapshots(
	ctx context.Context,
	accID uuid.UUID,
	pos *Position,
) ([]*PositionSnapshot, error) {
	// Determine start date from when on we want the snapshots to be built.
	var startDate time.Time
	if pos.ListingID == nil {
		return nil, nil // if we don't have a listing for this position, we can't build snapshots, so we just return
	}
	// Get the listing for the position to get the historical market prices
	listing, err := b.mdq.Listing(ctx, *pos.ListingID)
	if err != nil {
		return nil, fmt.Errorf("build position snapshots: fetch listing: %w", err)
	}
	if listing == nil {
		return nil, nil // if we don't have a listing for this position, we can't build snapshots, so we just return
	}
	if listing.Symbol == "" {
		return nil, nil // if we don't have a symbol for this listing, we can't build snapshots, so we just return
	}
	// Setup the time window of a position. The start date is the position open date, and the end date is either
	// the position close date or today (if still open).
	startDate = pos.OpenDate
	endExclusive := date.StartOfDayUTC(time.Now())
	if pos.CloseDate != nil && date.StartOfDayUTC(*pos.CloseDate).AddDate(0, 0, 1).Before(endExclusive) {
		endExclusive = date.StartOfDayUTC(*pos.CloseDate).AddDate(0, 0, 1)
	}
	lps, err := b.pss.GetLastSnapshot(ctx, pos.ID)
	if err != nil {
		return nil, fmt.Errorf("build position snapshots: get last snapshot: %w", err)
	}
	if lps != nil && date.StartOfDayUTC(lps.OccurredAt).Before(endExclusive) {
		startDate = date.StartOfDayUTC(lps.OccurredAt).AddDate(0, 0, 1)
	}
	// Get all transactions for the position, sorted by time
	ts, err := b.ts.TransactionsForPosition(ctx, pos.ID, &startDate)
	if err != nil {
		return nil, fmt.Errorf("build position snapshots: get transactions: %w", err)
	}
	// Get historical market data for the listing from the position open date to today
	dailies, err := b.mdq.GetEODByListing(ctx, *pos.ListingID, &startDate, nil, 0, 0, sorting.ASC)
	if err != nil {
		return nil, fmt.Errorf("build position snapshots: get dailies: %w", err)
	}
	// Initialize iterators for transactions and daily data, and an accumulator for the position state
	txIdx := 0
	dIdx := 0
	var prevSnapshot *PositionSnapshot
	var snapshots []*PositionSnapshot
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
				return nil, fmt.Errorf("build position snapshots: apply tx: %w", err)
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
			// We prioritize the close price for the day for market value calculation, if available. If
			// not, we fallback to the open price. If there is no open and close price available, we keep
			// the previous market value (if any) to avoid having a gap in the market value for this day.
			if dailies.Data[dIdx].Close > 0 {
				unitPrice = dailies.Data[dIdx].Close
				unitPriceSet = true
			}
			if !unitPriceSet && dailies.Data[dIdx].Open > 0 {
				unitPrice = dailies.Data[dIdx].Open
				unitPriceSet = true
			}
		}
		// If we don't have daily data for this day, we carry forward the previous market value.
		// If a buy or sell transaction occurred on this day we can calculate the market value based on the
		// transaction price, this is an approximation but it's better than carrying forward the previous market
		// value without any adjustments.
		if !unitPriceSet && prevSnapshot != nil {
			unitPrice = prevSnapshot.UnitPrice
		}
		snap, err := NewPositionSnapshot(
			pos.ID,
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
			return nil, fmt.Errorf("build position snapshots: new snapshot: %w", err)
		}
		if err := b.pss.CreateSnapshot(ctx, snap); err != nil {
			return nil, fmt.Errorf("build position snapshots: create snapshot: %w", err)
		}
		prevSnapshot = snap
		snapshots = append(snapshots, snap)
	}

	return snapshots, nil
}

func (b *Builder) buildPositions(ctx context.Context, accID uuid.UUID) ([]*Position, error) {
	// Load the event stream (transactions) in chronological order.
	ts, err := b.ts.TransactionsForAccount(ctx, accID, "ASC")
	if err != nil {
		return nil, fmt.Errorf("build positions: failed toload transactions: %w", err)
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
			return nil, fmt.Errorf("build positions: canonical key: %w", err)
		}
		// Fetch active position cycle; if ISIN appears later, promote symbol-cycle to ISIN key.
		pos := getOrPromoteCycle(activeByKey, key, symbolKey, *tx)
		// Apply transaction according to type. This helper is where the branching lives,
		// so the loop stays easy to read.
		pos, created, err := applyTxToCycle(accID, tx, key, symbolKey, aliases, activeByKey, pos)
		if err != nil {
			return nil, fmt.Errorf("build positions: apply transaction: %w", err)
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
			if listing, _, err := b.mdq.SearchListings(ctx, id, 1, 0); err == nil && len(listing) > 0 {
				pos.ListingID = &listing[0].ID
			}
		}
		posList = append(posList, pos)
	}
	// Persist positions (cycles).
	if err := b.pss.CreateMany(ctx, posList); err != nil {
		return nil, fmt.Errorf("build positions: persist positions: %w", err)
	}
	// Persist transaction->position mapping.
	// TODO: think of better naming.
	if err := b.pss.UpdatePositions(ctx, ts); err != nil {
		return nil, fmt.Errorf("build positions: persist transaction mapping: %w", err)
	}
	return posList, nil
}

// Build builds the portfolio based on the transactions.
func (b *Builder) Build(ctx context.Context, accID uuid.UUID) (err error) {
	// Acquire lock on account
	acquired, err := b.lk.TryAcquireBuildLock(ctx, accID)
	if err != nil {
		return fmt.Errorf("portfolio build failed to acquire build lock: %w", err)
	}
	if err != nil {
		return fmt.Errorf("portfolio build failed to clean: %w", err)
	}
	if !acquired {
		return ErrBuildInProgress
	}
	// Clean up existing snapshots and positions for this account, since we're
	// doing a full rebuild. Only do this after we have acquired the lock.
	err = b.pfs.Clean(ctx, accID)
	defer func() {
		// Ensure we always clear the syncing flag, even if the request context is canceled.
		cleanupCtx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
		defer cancel()

		if releaseErr := b.lk.ReleaseBuildLock(cleanupCtx, accID); releaseErr != nil {
			wrapped := fmt.Errorf("portfolio build failed to release build lock: %w", releaseErr)
			if err == nil {
				err = wrapped
			} else {
				err = errors.Join(err, wrapped)
			}
		}
	}()
	// Build the positions
	positions, err := b.buildPositions(ctx, accID)
	if err != nil {
		return fmt.Errorf("portfolio build failed to build positions: %w", err)
	}
	// Build the position snapshots for all positions and store them
	var positionSnapshots []*PositionSnapshot
	for _, ps := range positions {
		snapshots, err := b.buildPositionSnapshots(ctx, accID, ps)
		if err != nil {
			return fmt.Errorf("portfolio build failed to build position snapshots for position %s: %w", ps.ID, err)
		}
		positionSnapshots = append(positionSnapshots, snapshots...)
	}
	if len(positionSnapshots) == 0 {
		return ErrPortfolioNoSnapshots
	}
	// Loop through days and build portfolio snapshots
	startDate := positionSnapshots[0].OccurredAt
	endExclusive := date.StartOfDayUTC(time.Now())
	idx := 0 // set iterator for iterating through snapshots because there ca be multiple snapshots in one day
	var prevSnapshot *PortfolioSnapshot
	var prevPss []*PositionSnapshot
	for d := date.StartOfDayUTC(startDate); d.Before(endExclusive); d = d.AddDate(0, 0, 1) {
		dayEnd := date.EndOfDayUTC(d)
		var pss []*PositionSnapshot
		// Consume snapshots of this day
		for idx < len(positionSnapshots) && !positionSnapshots[idx].OccurredAt.UTC().After(dayEnd) {
			pss = append(pss, positionSnapshots[idx])
			idx++
		}
		// set pss prevPss if no snapshots are in this day, to carry forward values
		if len(pss) == 0 {
			pss = prevPss
		}
		// Build portfolio snapshot
		snap := NewPortfolioSnapshot(
			accID,
			d,
			0,
			pss,
			prevSnapshot,
		)
		if err := b.pfs.CreateSnapshot(ctx, snap); err != nil {
			return fmt.Errorf("portfolio build failed to persist portfolio snapshot: %w", err)
		}
		prevSnapshot = snap
		prevPss = pss
	}
	return nil
}
