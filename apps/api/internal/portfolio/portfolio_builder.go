// package portfolio

// import (
// 	"context"
// 	"errors"
// 	"fmt"
// 	"time"

// 	"github.com/google/uuid"
// 	"github.com/lennardclaproth/my-finances-tracker/internal/date"
// )

// type PortfolioBuilder struct {
// 	psb PositionBuilder
// 	psf PositionFetcher
// 	txs CashFlowTransactionStore
// 	as  AccountStore
// 	ps  PortfolioStore
// }

// type PositionFetcher interface {
// 	GetForAccount(ctx context.Context, accID uuid.UUID) ([]*Position, error)
// 	GetSnapshotsForAccount(ctx context.Context, accID uuid.UUID) ([]*PositionSnapshot, error)
// }

// type PortfolioStore interface {
// 	CreateSnapshot(ctx context.Context, snapshot *PortfolioSnapshot) error
// 	Clean(ctx context.Context, accID uuid.UUID) error
// }

// type CashFlowTransactionStore interface {
// 	GetASC(ctx context.Context, accID uuid.UUID) ([]Transaction, error)
// }

// type AccountStore interface {
// 	TryAcquireBuildLock(ctx context.Context, id uuid.UUID) (bool, error)
// 	ReleaseBuildLock(ctx context.Context, id uuid.UUID) error
// }

// func NewPortfolioBuilder(psb PositionBuilder, psf PositionFetcher, txs CashFlowTransactionStore, as AccountStore, ps PortfolioStore) PortfolioBuilder {
// 	return PortfolioBuilder{
// 		psb: psb,
// 		psf: psf,
// 		txs: txs,
// 		as:  as,
// 		ps:  ps,
// 	}
// }

// // Build builds the portfolio based on the transactions.
// func (s *PortfolioBuilder) Build(ctx context.Context, accID uuid.UUID) (err error) {
// 	// Acquire lock on account
// 	acquired, err := s.as.TryAcquireBuildLock(ctx, accID)
// 	if err != nil {
// 		return fmt.Errorf("portfolio build failed to acquire build lock: %w", err)
// 	}
// 	err = s.ps.Clean(ctx, accID)
// 	if err != nil {
// 		return fmt.Errorf("portfolio build failed to clean: %w", err)
// 	}
// 	if !acquired {
// 		return ErrBuildInProgress
// 	}
// 	defer func() {
// 		// Ensure we always clear the syncing flag, even if the request context is canceled.
// 		cleanupCtx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
// 		defer cancel()

// 		if releaseErr := s.as.ReleaseBuildLock(cleanupCtx, accID); releaseErr != nil {
// 			wrapped := fmt.Errorf("portfolio build failed to release build lock: %w", releaseErr)
// 			if err == nil {
// 				err = wrapped
// 			} else {
// 				err = errors.Join(err, wrapped)
// 			}
// 		}
// 	}()
// 	// Build the positions
// 	err = s.psb.BuildPositions(ctx, accID)
// 	if err != nil {
// 		return fmt.Errorf("portfolio build failed to build positions: %w", err)
// 	}
// 	// Build the position snapshots for all positions
// 	positions, err := s.psf.GetForAccount(ctx, accID)
// 	if err != nil {
// 		return fmt.Errorf("portfolio build failed to load positions: %w", err)
// 	}
// 	for _, ps := range positions {
// 		if err := s.psb.BuildPositionSnapshots(ctx, accID, ps.ID); err != nil {
// 			return fmt.Errorf("portfolio build failed to build position snapshots for position %s: %w", ps.ID, err)
// 		}
// 	}
// 	// Get the snapshots and build the portfolio from the snapshots
// 	positionSnapshots, err := s.psf.GetSnapshotsForAccount(ctx, accID) // this returns snapshots ordered ASC
// 	if err != nil {
// 		return fmt.Errorf("portfolio build failed to load position snapshots: %w", err)
// 	}
// 	if len(positionSnapshots) == 0 {
// 		return ErrPortfolioNoSnapshots
// 	}
// 	// Loop through days and build portfolio snapshots
// 	startDate := positionSnapshots[0].OccurredAt
// 	endExclusive := date.StartOfDayUTC(time.Now())
// 	idx := 0 // set iterator for iterating through snapshots because there ca be multiple snapshots in one day
// 	var prevSnapshot *PortfolioSnapshot
// 	var prevPss []*PositionSnapshot
// 	for d := date.StartOfDayUTC(startDate); d.Before(endExclusive); d = d.AddDate(0, 0, 1) {
// 		dayEnd := date.EndOfDayUTC(d)
// 		var pss []*PositionSnapshot
// 		// Consume snapshots of this day
// 		for idx < len(positionSnapshots) && !positionSnapshots[idx].OccurredAt.UTC().After(dayEnd) {
// 			pss = append(pss, positionSnapshots[idx])
// 			idx++
// 		}
// 		// set pss prevPss if no snapshots are in this day, to carry forward values
// 		if len(pss) == 0 {
// 			pss = prevPss
// 		}
// 		// Build portfolio snapshot
// 		snap := NewPortfolioSnapshot(
// 			accID,
// 			d,
// 			0,
// 			pss,
// 			prevSnapshot,
// 		)
// 		if err := s.ps.CreateSnapshot(ctx, snap); err != nil {
// 			return fmt.Errorf("portfolio build failed to persist portfolio snapshot: %w", err)
// 		}
// 		prevSnapshot = snap
// 		prevPss = pss
// 	}
// 	return nil
// }
package portfolio