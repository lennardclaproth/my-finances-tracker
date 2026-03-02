package api

import "github.com/google/uuid"

type TransactionsCreated struct {
	AccID uuid.UUID
}

func (TransactionsCreated) MessageTopic() string {
	return "TransactionCreated.v1"
}

type AccountCreated struct {
	AccID uuid.UUID
}

func (AccountCreated) MessageTopic() string {
	return "AccountCreated.v1"
}

type PortfolioRebuildRequested struct {
	AccID uuid.UUID
}

func (PortfolioRebuildRequested) MessageTopic() string {
	return "PortfolioRebuildRequested.v1"
}
