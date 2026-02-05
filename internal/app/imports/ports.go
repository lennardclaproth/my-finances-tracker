package imports

import (
	"io"

	"github.com/google/uuid"
	"github.com/lennardclaproth/my-finances-tracker/internal/domain"
)

type ImportCreator interface {
	Create(imp *domain.Import) error
}

type ImportFileWriter interface {
	WriteCsv(r io.Reader) (string, error)
}

type FileRemover interface {
	Remove(path string) error
}

type VendorFetcher interface {
	FetchById(id uuid.UUID) (*domain.Vendor, error)
	FetchByName(name domain.VendorID) (*domain.Vendor, error)
}
