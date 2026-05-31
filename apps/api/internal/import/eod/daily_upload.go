package eod

import (
	"encoding/json"
	"fmt"
	"time"

	"github.com/google/uuid"
)

// DailyUploadStatus tracks async daily-upload processing state.
type DailyUploadStatus string

const (
	DailyUploadStatusPending    DailyUploadStatus = "PENDING"
	DailyUploadStatusProcessing DailyUploadStatus = "PROCESSING"
	DailyUploadStatusSucceeded  DailyUploadStatus = "SUCCEEDED"
	DailyUploadStatusPartial    DailyUploadStatus = "PARTIAL"
	DailyUploadStatusFailed     DailyUploadStatus = "FAILED"
)

// DailyUploadRowError contains one failed row reason.
type DailyUploadRowError struct {
	RowNumber int    `json:"row_number"`
	Reason    string `json:"reason"`
}

// DailyUpload is the persisted aggregate for an async daily file upload.
type DailyUpload struct {
	ID               uuid.UUID         `db:"id"`
	ListingID        uuid.UUID         `db:"listing_id"`
	Source           Source            `db:"source"`
	Status           DailyUploadStatus `db:"status"`
	StoredFilename   string            `db:"stored_filename"`
	OriginalFilename string            `db:"original_filename"`
	StatusMessage    string            `db:"status_message"`
	TotalRows        int               `db:"total_rows"`
	InsertedRows     int               `db:"inserted_rows"`
	DuplicateRows    int               `db:"duplicate_rows"`
	ErrorRows        int               `db:"error_rows"`
	RowErrorsJSON    string            `db:"row_errors_json"`
	StartedAt        *time.Time        `db:"started_at"`
	FinishedAt       *time.Time        `db:"finished_at"`
	CreatedAt        time.Time         `db:"created_at"`
	UpdatedAt        time.Time         `db:"updated_at"`

	RowErrors []DailyUploadRowError `db:"-"`
}

var (
	// ErrDailyUploadNotFound indicates missing upload record.
	ErrDailyUploadNotFound = fmt.Errorf("daily upload not found")
	// ErrDailyUploadListingIDEmpty indicates missing listing identifier.
	ErrDailyUploadListingIDEmpty = fmt.Errorf("daily upload listing id cannot be empty")
	// ErrDailyUploadSourceEmpty indicates missing upload source.
	ErrDailyUploadSourceEmpty = fmt.Errorf("daily upload source cannot be empty")
	// ErrDailyUploadStoredFilenameEmpty indicates missing persisted filename.
	ErrDailyUploadStoredFilenameEmpty = fmt.Errorf("daily upload stored filename cannot be empty")
	// ErrDailyUploadOriginalNameEmpty indicates missing original filename.
	ErrDailyUploadOriginalNameEmpty = fmt.Errorf("daily upload original filename cannot be empty")
)

// NewDailyUpload constructs a new upload aggregate in pending state.
func NewDailyUpload(listingID uuid.UUID, source Source, storedFilename, originalFilename string) (*DailyUpload, error) {
	if listingID == uuid.Nil {
		return nil, ErrDailyUploadListingIDEmpty
	}
	if source == "" {
		return nil, ErrDailyUploadSourceEmpty
	}
	if storedFilename == "" {
		return nil, ErrDailyUploadStoredFilenameEmpty
	}
	if originalFilename == "" {
		return nil, ErrDailyUploadOriginalNameEmpty
	}
	now := time.Now().UTC()
	upload := &DailyUpload{
		ID:               uuid.New(),
		ListingID:        listingID,
		Source:           source,
		Status:           DailyUploadStatusPending,
		StoredFilename:   storedFilename,
		OriginalFilename: originalFilename,
		StatusMessage:    "",
		RowErrors:        []DailyUploadRowError{},
		CreatedAt:        now,
		UpdatedAt:        now,
	}
	if err := upload.SetRowErrors(nil, 0); err != nil {
		return nil, err
	}
	return upload, nil
}

// MarkProcessing transitions the upload into processing state.
func (u *DailyUpload) MarkProcessing() {
	now := time.Now().UTC()
	u.Status = DailyUploadStatusProcessing
	u.StartedAt = &now
	u.UpdatedAt = now
}

// MarkCompleted stores counters and terminal status for a processed upload.
func (u *DailyUpload) MarkCompleted(totalRows, insertedRows, duplicateRows, errorRows int, rowErrors []DailyUploadRowError, statusMsg string) error {
	now := time.Now().UTC()
	u.TotalRows = totalRows
	u.InsertedRows = insertedRows
	u.DuplicateRows = duplicateRows
	u.ErrorRows = errorRows
	u.StatusMessage = statusMsg
	u.FinishedAt = &now
	u.UpdatedAt = now
	if errorRows == 0 {
		u.Status = DailyUploadStatusSucceeded
	} else if insertedRows+duplicateRows > 0 {
		u.Status = DailyUploadStatusPartial
	} else {
		u.Status = DailyUploadStatusFailed
	}
	return u.SetRowErrors(rowErrors, 50)
}

// MarkFailed marks the upload as failed and persists sampled row errors.
func (u *DailyUpload) MarkFailed(statusMsg string) error {
	now := time.Now().UTC()
	u.Status = DailyUploadStatusFailed
	u.StatusMessage = statusMsg
	u.FinishedAt = &now
	u.UpdatedAt = now
	return u.SetRowErrors(u.RowErrors, 50)
}

// SetRowErrors stores row errors with optional sampling limit.
func (u *DailyUpload) SetRowErrors(rowErrors []DailyUploadRowError, sampleLimit int) error {
	out := rowErrors
	if out == nil {
		out = []DailyUploadRowError{}
	}
	if sampleLimit > 0 && len(out) > sampleLimit {
		out = out[:sampleLimit]
	}
	raw, err := json.Marshal(out)
	if err != nil {
		return err
	}
	u.RowErrors = out
	u.RowErrorsJSON = string(raw)
	return nil
}

// DecodeRowErrors parses persisted row error JSON into RowErrors.

func (u *DailyUpload) DecodeRowErrors() error {
	if u.RowErrorsJSON == "" {
		u.RowErrors = []DailyUploadRowError{}
		return nil
	}
	var rows []DailyUploadRowError
	if err := json.Unmarshal([]byte(u.RowErrorsJSON), &rows); err != nil {
		return err
	}
	u.RowErrors = rows
	return nil
}
