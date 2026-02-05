package imports

import (
	"io"

	"github.com/google/uuid"
	"github.com/lennardclaproth/my-finances-tracker/internal/domain"
)

type ImportFromCsvHandler struct {
	ic  ImportCreator
	ifw ImportFileWriter
	fr  FileRemover
	vf  VendorFetcher
}

func NewImportFromCsvHandler(ic ImportCreator, ifw ImportFileWriter, fr FileRemover, vf VendorFetcher) *ImportFromCsvHandler {
	return &ImportFromCsvHandler{
		ic:  ic,
		ifw: ifw,
		fr:  fr,
		vf:  vf,
	}
}

// Handle processes the CSV import for a given vendor ID.
func (h *ImportFromCsvHandler) Handle(r io.Reader, vendorId string) (uuid.UUID, error) {
	// Get vendor via VendorFetcher
	vendor, err := h.vf.FetchByName(domain.VendorID(vendorId))
	if err != nil {
		return uuid.Nil, err
	}
	// Write file via ImportFileWriter
	path, err := h.ifw.WriteCsv(r)
	if err != nil {
		return uuid.Nil, err
	}
	// Create import via ImportCreator
	imp := domain.NewImport(*vendor, path)
	if err := h.ic.Create(imp); err != nil {
		_ = h.fr.Remove(path) // best effort cleanup
		return uuid.Nil, err  // return original error
	}
	return imp.ID, nil
}
