package account

import "github.com/google/uuid"

const (
	TopicAccountCreated = "account.created"
)

type AccountCreated struct {
	AccID uuid.UUID
}
