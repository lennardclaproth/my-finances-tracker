package assets

import (
	"context"
	"errors"
	"fmt"
	"time"

	"github.com/google/uuid"
	"github.com/lennardclaproth/my-finances-tracker/internal/money"
	"github.com/lennardclaproth/my-finances-tracker/internal/portfolio"
	"github.com/lennardclaproth/my-finances-tracker/internal/sorting"
)

type Syncer struct {
	pq  *portfolio.Queries
	b   *Builder
	ss  SyncStore
	uow uow
}

type SyncStore interface {
	TryAcquireSyncLock(ctx context.Context, id uuid.UUID) (bool, error)
	ReleaseSyncLock(ctx context.Context, id uuid.UUID) error
	CleanPortfolio(ctx context.Context, accountID uuid.UUID) error
	ClassBySource(ctx context.Context, accountID uuid.UUID, source ClassSource) (*Class, error)
	CreateClass(ctx context.Context, class *Class) error
	AssetByClassAndName(ctx context.Context, accountID, classID uuid.UUID, name string) (*Asset, error)
	CreateAsset(ctx context.Context, asset *Asset) error
	UpdateAssetWorth(ctx context.Context, accountID, classID, assetID uuid.UUID, worth money.Price) error
	CreateMutation(ctx context.Context, mutation *Mutation) error
}

type uow interface {
	Do(ctx context.Context, fn func(txCtx context.Context) error) error
}

// NewSyncer constructs the assets portfolio Syncer.
func NewSyncer(pq *portfolio.Queries, b *Builder, ss SyncStore, u uow) *Syncer {
	return &Syncer{pq: pq, b: b, ss: ss, uow: u}
}

// SyncPortfolio
func (s *Syncer) SyncPortfolio(ctx context.Context, accID uuid.UUID) error {
	// Acquire sync lock for account
	acquired, err := s.ss.TryAcquireSyncLock(ctx, accID)
	if err != nil {
		return fmt.Errorf("sync portfolio: failed to execute acquire sync lock: %w", err)
	}
	if !acquired {
		return fmt.Errorf("sync portfolio: failed to acquire sync lock: %w", ErrSyncInProgress)
	}
	defer func() {
		cleanupCtx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
		defer cancel()

		if releaseErr := s.ss.ReleaseSyncLock(cleanupCtx, accID); releaseErr != nil {
			err = errors.Join(
				err,
				fmt.Errorf("sync eod: %w: %w", ErrReleaseSyncLock, releaseErr),
			)
		}
	}()
	// TODO: check if we can simplify this and only delete the mutations
	// Clean up existing portfolio class, assets and mutations
	err = s.ss.CleanPortfolio(ctx, accID)
	if err != nil {
		return fmt.Errorf("sync portfolio: failed to clean up portfolio data: %w", err)
	}
	// List portfolio snapshots for account
	direction := sorting.ASC
	snapshots, err := s.pq.SnapshotsForAccount(ctx, accID, nil, nil, nil, nil, &direction)
	if err != nil {
		return fmt.Errorf("sync portfolio: failed to list portfolio snapshots: %w", err)
	}
	// If there are no snapshots we assume there is nothing to sync and we can skip the process.
	// This prevents us from doing unnecessary work.
	if len(snapshots) == 0 {
		return nil
	}
	// TODO: move transaction logic to this function, EnsureClassCreated returns class and asset, creates them if they do not exist yet.
	// Sync the mutations and create the class if it does not exist yet.
	if err := s.syncMutationsFromSnapshots(ctx, accID, snapshots); err != nil {
		return err
	}

	//TODO: Notify socket clients about the update
	return nil
}

func (s *Syncer) syncMutationsFromSnapshots(ctx context.Context, accountID uuid.UUID, snapshots []*portfolio.PortfolioSnapshot) error {
	if len(snapshots) == 0 {
		return nil
	}
	now := time.Now().UTC()
	// Execute a scoped transaction to sync the portfolio class
	err := s.uow.Do(ctx, func(txCtx context.Context) error {
		// Fetch the portfolio class, if it does not exist
		// create it.
		class, err := s.ss.ClassBySource(txCtx, accountID, ClassSourcePortfolio)
		if err != nil {
			return err
		}
		if class == nil {
			class = &Class{
				ID:        uuid.New(),
				AccountID: accountID,
				Name:      PortfolioClassName,
				Source:    ClassSourcePortfolio,
				Archived:  false,
				CreatedAt: now,
				UpdatedAt: now,
			}
			if err := s.ss.CreateClass(txCtx, class); err != nil {
				return err
			}
		}
		// Get or create the portfolio asset, this is a static
		// asset that represents the total worth of the portfolio.
		asset, err := s.ss.AssetByClassAndName(txCtx, accountID, class.ID, PortfolioAssetName)
		if err != nil {
			return err
		}
		if asset == nil {
			asset = &Asset{
				ID:           uuid.New(),
				ClassID:      class.ID,
				AccountID:    accountID,
				Name:         PortfolioAssetName,
				CurrentWorth: 0,
				Archived:     false,
				CreatedAt:    now,
				UpdatedAt:    now,
			}
			if err := s.ss.CreateAsset(txCtx, asset); err != nil {
				return err
			}
		}
		// Update the portfolio asset worth to the latest snapshot market value
		latest := snapshots[len(snapshots)-1]
		if latest == nil {
			return nil
		}
		if err := s.ss.UpdateAssetWorth(txCtx, accountID, class.ID, asset.ID, latest.MarketValue); err != nil {
			return err
		}

		previousWorth := money.Price(0)
		// TODO: Batch mutation creation
		// For each snapshot, create a mutation with the market value change
		// since the previous snapshot
		for _, snapshot := range snapshots {
			if snapshot == nil {
				continue
			}
			note := "synced from portfolio snapshot"
			date := snapshot.OccurredAt.UTC()
			entry := &Mutation{
				ID:              uuid.New(),
				AccountID:       accountID,
				ClassID:         class.ID,
				AssetID:         asset.ID,
				ChangeType:      ChangeTypeSet,
				Direction:       nil,
				Amount:          snapshot.MarketValue,
				PreviousWorth:   previousWorth,
				NewWorth:        snapshot.MarketValue,
				ClassTotalWorth: snapshot.MarketValue,
				EffectiveDate:   time.Date(date.Year(), date.Month(), date.Day(), 0, 0, 0, 0, time.UTC),
				Note:            &note,
				CreatedAt:       now,
			}
			if err := s.ss.CreateMutation(txCtx, entry); err != nil {
				return err
			}
			previousWorth = snapshot.MarketValue
		}
		return nil
	})
	if err != nil {
		return fmt.Errorf("sync portfolio: failed to execute transaction: %w", err)
	}
	return nil
}
