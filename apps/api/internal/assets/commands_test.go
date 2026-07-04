package assets_test

import (
	"context"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/lennardclaproth/my-finances-tracker/internal/account"
	"github.com/lennardclaproth/my-finances-tracker/internal/assets"
	"github.com/lennardclaproth/my-finances-tracker/internal/eventbus"
	memorybus "github.com/lennardclaproth/my-finances-tracker/internal/eventbus/memory"
)

// fakeAccountStore satisfies the account query store so account.Queries reports
// the account as existing.
type fakeAccountStore struct{ exists bool }

func (f fakeAccountStore) Exists(_ context.Context, _ uuid.UUID) (bool, error) { return f.exists, nil }
func (f fakeAccountStore) GetByID(_ context.Context, _ uuid.UUID) (*account.Account, error) {
	return nil, nil
}
func (f fakeAccountStore) List(_ context.Context) ([]*account.Account, error) { return nil, nil }

// fakeAssetsStore is a no-op assets command store; persistence is not under test.
type fakeAssetsStore struct{}

func (fakeAssetsStore) CreateAsset(_ context.Context, _ *assets.Asset) error       { return nil }
func (fakeAssetsStore) CreateAccount(_ context.Context, _ *assets.Account) error   { return nil }
func (fakeAssetsStore) SetWorth(_ context.Context, _ *assets.Asset) error          { return nil }
func (fakeAssetsStore) CreateClass(_ context.Context, _ *assets.Class) error       { return nil }
func (fakeAssetsStore) UpdateClass(_ context.Context, _ *assets.Class) error       { return nil }
func (fakeAssetsStore) CreateMutation(_ context.Context, _ *assets.Mutation) error { return nil }
func (fakeAssetsStore) DeleteClass(_ context.Context, _ uuid.UUID) error           { return nil }

// TestCreateClassPublishesSnapshotsRebuildRequested verifies the constructor
// wires the bus and that a mutation publishes the rebuild-requested event.
func TestCreateClassPublishesSnapshotsRebuildRequested(t *testing.T) {
	bus := memorybus.NewMemoryBus(memorybus.WithWorkers(2), memorybus.WithQueueSize(8))
	defer bus.Close()

	got := make(chan eventbus.Envelope, 1)
	sub, err := bus.Subscribe(assets.TopicSnapshotsRebuildRequested, func(_ context.Context, env eventbus.Envelope) error {
		got <- env
		return nil
	})
	if err != nil {
		t.Fatalf("subscribe: %v", err)
	}
	defer sub.Close()

	// account.Queries is held by value on Commands; deref the constructed pointer.
	aq := account.NewQueries(fakeAccountStore{exists: true})
	commands := assets.NewCommands(fakeAssetsStore{}, nil, *aq, nil, nil, bus)

	accID := uuid.New()
	if _, err := commands.CreateClass(context.Background(), accID, "Savings"); err != nil {
		t.Fatalf("create class: %v", err)
	}

	select {
	case env := <-got:
		payload, ok := env.Payload.(assets.SnapshotsRebuildRequested)
		if !ok {
			t.Fatalf("unexpected payload type %T", env.Payload)
		}
		if payload.AccID != accID {
			t.Fatalf("expected AccID %s, got %s", accID, payload.AccID)
		}
	case <-time.After(2 * time.Second):
		t.Fatal("timed out waiting for SnapshotsRebuildRequested")
	}
}
