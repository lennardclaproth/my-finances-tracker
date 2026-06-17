package assets

import (
	"context"
	"testing"

	"github.com/google/uuid"
	"github.com/lennardclaproth/my-finances-tracker/internal/account"
	"github.com/lennardclaproth/my-finances-tracker/internal/assets"
	"github.com/lennardclaproth/my-finances-tracker/internal/eventbus"
)

// recordingAssetsStore records the account id of the created projection and
// no-ops the rest of the assets command store.
type recordingAssetsStore struct{ createdAccountID uuid.UUID }

func (s *recordingAssetsStore) CreateAccount(_ context.Context, acc *assets.Account) error {
	s.createdAccountID = acc.AccountID
	return nil
}
func (s *recordingAssetsStore) CreateAsset(_ context.Context, _ *assets.Asset) error { return nil }
func (s *recordingAssetsStore) SetWorth(_ context.Context, _ *assets.Asset) error    { return nil }
func (s *recordingAssetsStore) CreateClass(_ context.Context, _ *assets.Class) error { return nil }
func (s *recordingAssetsStore) UpdateClass(_ context.Context, _ *assets.Class) error { return nil }
func (s *recordingAssetsStore) CreateMutation(_ context.Context, _ *assets.Mutation) error {
	return nil
}
func (s *recordingAssetsStore) DeleteClass(_ context.Context, _ uuid.UUID) error { return nil }

func TestAccountCreatedCreatesAssetsProjection(t *testing.T) {
	store := &recordingAssetsStore{}
	commands := assets.NewCommands(store, nil, account.Queries{}, nil, nil, nil)
	handler := NewAccountCreatedHandler(commands, nil)

	accID := uuid.New()
	if err := handler.Handle(context.Background(), account.Created{AccID: accID}, eventbus.Metadata{}); err != nil {
		t.Fatalf("handle: %v", err)
	}
	if store.createdAccountID != accID {
		t.Fatalf("expected assets projection for %s, got %s", accID, store.createdAccountID)
	}
}
