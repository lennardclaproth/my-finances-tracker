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
	FetchByName(ctx context.Context, name vendor.VendorID) (*vendor.Vendor, error)
}

type FromCsvHandler struct {
	ic  ImportCreator
	ifw ImportFileWriter
	fr  FileRemover
	vf  VendorFetcher
}

func NewFromCsvHandler(ic ImportCreator, ifw ImportFileWriter, fr FileRemover, vf VendorFetcher) *FromCsvHandler {
	return &FromCsvHandler{
		ic:  ic,
		ifw: ifw,
		fr:  fr,
		vf:  vf,
	}
}

// Handle processes the CSV import for a given vendor ID.
func (h *FromCsvHandler) Handle(ctx context.Context, r io.Reader, vendorId string) (uuid.UUID, error) {
	// Get vendor via VendorFetcher
	v, err := h.vf.FetchByName(ctx, vendor.VendorID(vendorId))
	if err != nil {
		return uuid.Nil, err
	}
	// Write file via ImportFileWriter
	path, err := h.ifw.WriteCsv(r)
	if err != nil {
		return uuid.Nil, err
	}
	// Create import via ImportCreator
	imp := NewImport(*v, path)
	if err := h.ic.Create(ctx, imp); err != nil {
		_ = h.fr.Remove(path) // best effort cleanup
		return uuid.Nil, err  // return original error
	}
	return imp.ID, nil
}
