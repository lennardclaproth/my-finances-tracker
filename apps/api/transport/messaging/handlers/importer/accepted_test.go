package importer

import (
	"context"
	"testing"

	"github.com/google/uuid"
	"github.com/lennardclaproth/my-finances-tracker/internal/account"
	"github.com/lennardclaproth/my-finances-tracker/internal/eventbus"
	"github.com/lennardclaproth/my-finances-tracker/internal/importer"
	"github.com/lennardclaproth/my-finances-tracker/internal/marketdata"
	"github.com/lennardclaproth/my-finances-tracker/internal/vendor"
)

// recordingImportStore satisfies the importer command store and records the id
// passed to FetchByID so the test can assert the handler forwarded it. Returning
// a nil import makes Process short-circuit, keeping the test focused on the
// handler's delegation.
type recordingImportStore struct{ fetchedID uuid.UUID }

func (s *recordingImportStore) Create(_ context.Context, _ *importer.Import) error { return nil }
func (s *recordingImportStore) FetchByID(_ context.Context, id uuid.UUID) (*importer.Import, error) {
	s.fetchedID = id
	return nil, nil
}
func (s *recordingImportStore) UpdateState(_ context.Context, _ *importer.Import) error {
	return nil
}

func TestAcceptedHandlerProcessesAcceptedImport(t *testing.T) {
	store := &recordingImportStore{}
	commands := importer.NewCommands(store, nil, nil, vendor.Queries{}, account.Queries{}, marketdata.Queries{}, nil)
	handler := NewAcceptedHandler(commands, nil)

	importID := uuid.New()
	err := handler.Handle(context.Background(), importer.Accepted{ImportID: importID}, eventbus.Metadata{})
	if err != nil {
		t.Fatalf("handle: %v", err)
	}
	if store.fetchedID != importID {
		t.Fatalf("expected handler to process import %s, fetched %s", importID, store.fetchedID)
	}
}
