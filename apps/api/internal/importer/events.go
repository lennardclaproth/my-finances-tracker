package importer

import "github.com/google/uuid"

const (
	// TopicAccepted is published after a CSV import record is persisted.
	TopicAccepted = "import.accepted"
	// TopicCompleted is published after a CSV import reaches completed state.
	TopicCompleted = "import.completed"
	// TopicFailed is published after a CSV import reaches failed state.
	TopicFailed = "import.failed"
)

// Accepted asks asynchronous workers to process the persisted import.
type Accepted struct {
	ImportID uuid.UUID
	Type     ImportType
}

// Completed notifies consumers that import processing completed.
type Completed struct {
	ImportID  uuid.UUID
	Type      ImportType
	AccountID *uuid.UUID
	ListingID *uuid.UUID
}

// Failed notifies consumers that import processing failed.
type Failed struct {
	ImportID  uuid.UUID
	Type      ImportType
	AccountID *uuid.UUID
	ListingID *uuid.UUID
	Reason    string
}
