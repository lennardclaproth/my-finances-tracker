package assets

import "github.com/google/uuid"

const (
	// TopicSnapshotsRebuildRequested is published when an account's assets
	// snapshots should be rebuilt, for example after a class or asset mutation.
	TopicSnapshotsRebuildRequested = "assets.snapshots.rebuild_requested"
	// TopicSnapshotsRebuilt is published after an account's assets snapshots are rebuilt.
	TopicSnapshotsRebuilt = "assets.snapshots.rebuilt"
)

// SnapshotsRebuildRequested asks the assets feature to rebuild account-level
// snapshots from mutation history. It is published through the eventbus.Bus by
// the write side after a mutation and consumed by the assets snapshot handler.
type SnapshotsRebuildRequested struct {
	AccID uuid.UUID
}

// SnapshotsRebuilt notifies consumers that account-level assets snapshots were
// rebuilt. The realtime layer uses it to notify clients.
type SnapshotsRebuilt struct {
	AccID uuid.UUID
}
