package importer

import (
	"context"
	"fmt"
	"io"

	"github.com/google/uuid"
	"github.com/lennardclaproth/my-finances-tracker/internal/vendor"
)

// Single-use interfaces only used by FromCsvHandler

// ImportFileWriter persists uploaded CSV payloads and returns the stored path.
type ImportFileWriter interface {
	WriteCsv(r io.Reader) (string, error)
}

// FileRemover removes a previously persisted file path.
type FileRemover interface {
	Remove(path string) error
}

// VendorFetcher loads vendors by ID for import validation.
type VendorFetcher interface {
	FetchById(ctx context.Context, id uuid.UUID) (*vendor.Vendor, error)
}

// FromCsvHandler validates import metadata, persists the uploaded file, and creates an import job record.
type FromCsvHandler struct {
	ic  ImportCreator
	ifw ImportFileWriter
	fr  FileRemover
	vf  VendorFetcher
	af  AccountFetcher
}

// NewFromCsvHandler creates a CSV import handler with injected dependencies.
func NewFromCsvHandler(ic ImportCreator, ifw ImportFileWriter, fr FileRemover, vf VendorFetcher, af AccountFetcher) *FromCsvHandler {
	return &FromCsvHandler{
		ic:  ic,
		ifw: ifw,
		fr:  fr,
		vf:  vf,
		af:  af,
	}
}

// Handle processes the CSV import for a given vendor ID.
func (h *FromCsvHandler) Handle(ctx context.Context, r io.Reader, vendorID, accountID uuid.UUID) (uuid.UUID, error) {
	// Get vendor via VendorFetcher
	v, err := h.vf.FetchById(ctx, vendorID)
	if err != nil {
		return uuid.Nil, err
	}
	if v.ImportDisabled {
		return uuid.Nil, ErrVendorImportDisabled
	}
	if h.af == nil {
		return uuid.Nil, ErrImportAccountValidationNotReady
	}
	if _, err := h.af.FetchByID(ctx, accountID); err != nil {
		return uuid.Nil, err
	}
	// Write file via ImportFileWriter
	path, err := h.ifw.WriteCsv(r)
	if err != nil {
		return uuid.Nil, err
	}
	// Create import via ImportCreator
	imp := NewImport(*v, path, accountID)
	if err := h.ic.Create(ctx, imp); err != nil {
		if removeErr := h.fr.Remove(path); removeErr != nil {
			return uuid.Nil, fmt.Errorf("%w (cleanup failed removing %s: %v)", err, path, removeErr)
		}
		return uuid.Nil, err
	}
	return imp.ID, nil
}
