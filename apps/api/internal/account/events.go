package account

import "github.com/google/uuid"

const (
	// TopicCreated is the eventbus topic published after an account is created.
	TopicCreated = "account.created"
)

// Created notifies consumers that a new account was created. It is published
// through the eventbus.Bus and fans out to the per-feature account projections.
type Created struct {
	AccID uuid.UUID
}
