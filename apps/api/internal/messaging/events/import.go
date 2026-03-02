package events

import "github.com/google/uuid"

type TransactionsImported struct {
	AccID uuid.UUID
}

func (TransactionsImported) MessageTopic() string {
	return "TransactionImported.v1"
}
