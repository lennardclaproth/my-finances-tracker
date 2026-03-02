package importer

import (
	"context"
	"io"

	"github.com/google/uuid"
	"github.com/lennardclaproth/my-finances-tracker/internal/vendor"
)

// Single-use interfaces only used by FromCsvHandler

type ImportFileWriter interface {
	WriteCsv(r io.Reader) (string, error)
}

type FileRemover interface {
	Remove(path string) error
}

type VendorFetcher interface {
	FetchById(ctx context.Context, id uuid.UUID) (*vendor.Vendor, error)
}

type FromCsvHandler struct {
	ic  ImportCreator
	ifw ImportFileWriter
	fr  FileRemover
	vf  VendorFetcher
	af  AccountFetcher
}

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
		_ = h.fr.Remove(path) // best effort cleanup
		return uuid.Nil, err  // return original error
	}
	return imp.ID, nil
}
