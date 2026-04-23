package api

import "github.com/google/uuid"

// TransactionsCreated is emitted after an import creates cashflow/portfolio transactions.
type TransactionsCreated struct {
	AccID uuid.UUID
}

// MessageTopic returns the bus topic name for TransactionsCreated events.
func (TransactionsCreated) MessageTopic() string {
	return "TransactionCreated.v1"
}

// AccountCreated is emitted after an account is created.
type AccountCreated struct {
	AccID uuid.UUID
}

// MessageTopic returns the bus topic name for AccountCreated events.
func (AccountCreated) MessageTopic() string {
	return "AccountCreated.v1"
}

// PortfolioRebuildRequested is emitted when a portfolio rebuild is requested.
type PortfolioRebuildRequested struct {
	AccID uuid.UUID
}

// MessageTopic returns the bus topic name for PortfolioRebuildRequested events.
func (PortfolioRebuildRequested) MessageTopic() string {
	return "PortfolioRebuildRequested.v1"
}

// PortfolioRebuilt is emitted after a portfolio rebuild completes.
type PortfolioRebuilt struct {
	AccID uuid.UUID
}

// MessageTopic returns the bus topic name for PortfolioRebuilt events.
func (PortfolioRebuilt) MessageTopic() string {
	return "PortfolioRebuilt.v1"
}

// ImportCompleted is emitted when an import job finishes processing.
type ImportCompleted struct {
	AccID    uuid.UUID
	ImportID uuid.UUID
}

// MessageTopic returns the bus topic name for ImportCompleted events.
func (ImportCompleted) MessageTopic() string {
	return "ImportCompleted.v1"
}

// BulkTagCompleted is emitted when an async bulk-tag job finishes.
type BulkTagCompleted struct {
	AccID        uuid.UUID
	UpdatedCount int
}

// MessageTopic returns the bus topic name for BulkTagCompleted events.
func (BulkTagCompleted) MessageTopic() string {
	return "BulkTagCompleted.v1"
}

// AssetsSnapshotsRebuildRequested is emitted when account-level assets snapshots should be rebuilt.
type AssetsSnapshotsRebuildRequested struct {
	AccID uuid.UUID
}

// MessageTopic returns the bus topic name for AssetsSnapshotsRebuildRequested events.
func (AssetsSnapshotsRebuildRequested) MessageTopic() string {
	return "AssetsSnapshotsRebuildRequested.v1"
}

// AssetsSnapshotsRebuilt is emitted after account-level assets snapshots are rebuilt.
type AssetsSnapshotsRebuilt struct {
	AccID uuid.UUID
}

// MessageTopic returns the bus topic name for AssetsSnapshotsRebuilt events.
func (AssetsSnapshotsRebuilt) MessageTopic() string {
	return "AssetsSnapshotsRebuilt.v1"
}
