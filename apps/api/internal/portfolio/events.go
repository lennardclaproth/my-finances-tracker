package portfolio

import "github.com/google/uuid"

const (
	// TopicRebuilt is published after an account's portfolio positions and
	// snapshots are rebuilt. Consumers such as the assets feature react to it
	// to re-sync derived state, and the realtime layer uses it to notify clients.
	TopicRebuilt = "portfolio.rebuilt"
)

// Rebuilt notifies consumers that an account's portfolio was rebuilt. It is
// published through the eventbus.Bus after a successful rebuild so other
// features can react without depending on the portfolio build path directly.
type Rebuilt struct {
	AccID uuid.UUID
}
