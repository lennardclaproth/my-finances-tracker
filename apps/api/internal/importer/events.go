package importer

import "github.com/google/uuid"

const (
	// TopicImportAccepted is published after a CSV import record is persisted.
	TopicImportAccepted = "import.accepted"
	// TopicImportCompleted is published after a CSV import reaches completed state.
	TopicImportCompleted = "import.completed"
	// TopicImportFailed is published after a CSV import reaches failed state.
	TopicImportFailed = "import.failed"
)

// ImportAccepted asks asynchronous workers to process the persisted import.
type ImportAccepted struct {
	ImportID uuid.UUID
	Type     ImportType
}

// ImportCompleted notifies consumers that import processing completed.
type ImportCompleted struct {
	ImportID  uuid.UUID
	Type      ImportType
	AccountID *uuid.UUID
	ListingID *uuid.UUID
}

// ImportFailed notifies consumers that import processing failed.
type ImportFailed struct {
	ImportID  uuid.UUID
	Type      ImportType
	AccountID *uuid.UUID
	ListingID *uuid.UUID
	Reason    string
}
