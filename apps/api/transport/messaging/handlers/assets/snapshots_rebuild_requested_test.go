package assets

import (
	"context"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/lennardclaproth/my-finances-tracker/internal/assets"
	"github.com/lennardclaproth/my-finances-tracker/internal/eventbus"
	memorybus "github.com/lennardclaproth/my-finances-tracker/internal/eventbus/memory"
	"github.com/lennardclaproth/my-finances-tracker/internal/sorting"
)

// fakeBuilderStore is a no-op assets builder store: no mutations, so RebuildAll
// stores a single zero snapshot and succeeds.
type fakeBuilderStore struct{}

func (fakeBuilderStore) Mutations(_ context.Context, _ uuid.UUID, _ *sorting.Direction, _, _ *uint64) ([]*assets.Mutation, error) {
	return nil, nil
}
func (fakeBuilderStore) DeleteSnapshots(_ context.Context, _ uuid.UUID) error         { return nil }
func (fakeBuilderStore) StoreSnapshots(_ context.Context, _ []*assets.Snapshot) error { return nil }

// immediateUOW runs the transaction body inline.
type immediateUOW struct{}

func (immediateUOW) Do(ctx context.Context, fn func(txCtx context.Context) error) error {
	return fn(ctx)
}

func TestSnapshotsRebuildRequestedRebuildsAndAnnounces(t *testing.T) {
	bus := memorybus.NewMemoryBus(memorybus.WithWorkers(2), memorybus.WithQueueSize(8))
	defer bus.Close()

	got := make(chan eventbus.Envelope, 1)
	sub, err := bus.Subscribe(assets.TopicSnapshotsRebuilt, func(_ context.Context, env eventbus.Envelope) error {
		got <- env
		return nil
	})
	if err != nil {
		t.Fatalf("subscribe: %v", err)
	}
	defer sub.Close()

	builder := assets.NewBuilder(fakeBuilderStore{}, immediateUOW{})
	handler := NewSnapshotsRebuildRequestedHandler(builder, bus, nil)

	accID := uuid.New()
	if err := handler.Handle(context.Background(), assets.SnapshotsRebuildRequested{AccID: accID}, eventbus.Metadata{}); err != nil {
		t.Fatalf("handle: %v", err)
	}

	select {
	case env := <-got:
		payload, ok := env.Payload.(assets.SnapshotsRebuilt)
		if !ok {
			t.Fatalf("unexpected payload type %T", env.Payload)
		}
		if payload.AccID != accID {
			t.Fatalf("expected AccID %s, got %s", accID, payload.AccID)
		}
	case <-time.After(2 * time.Second):
		t.Fatal("timed out waiting for SnapshotsRebuilt")
	}
}
