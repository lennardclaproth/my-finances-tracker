package handlers

import (
	"context"
	"errors"
	"net/http"

	"github.com/google/uuid"
	"github.com/lennardclaproth/my-finances-tracker/api"
	httpx "github.com/lennardclaproth/my-finances-tracker/internal/http"
	"github.com/lennardclaproth/my-finances-tracker/internal/importer"
	"github.com/lennardclaproth/my-finances-tracker/internal/logging"
	"github.com/lennardclaproth/my-finances-tracker/internal/storage"
)

// ImportCsv exposes an HTTP handler for importing csv files to be processed.
//
// @Summary Import transactions from CSV file
// @Description Upload a CSV file containing transaction data to import into a specific vendor
// @Tags imports
// @Accept multipart/form-data
// @Produce json
// @Param file formData file true "CSV file containing transaction data"
// @Param vendor_id formData string true "UUID of the vendor to import transactions for"
// @Param account_id formData string true "UUID of the account"
// @Success 200 {object} uuid.UUID "Import ID of the created import job"
// @Failure 400 {object} map[string]string "Invalid request (missing file, invalid vendor_id, etc.)"
// @Failure 413 {object} map[string]string "File too large (max 20MB)"
// @Failure 415 {object} map[string]string "Unsupported media type (only text/csv allowed)"
// @Failure 500 {object} map[string]string "Internal server error"
// @Router /import/csv [post]
func ImportCsv(
	log logging.Logger,
	ic importer.ImportCreator,
	dw *storage.Disk,
	vf importer.VendorFetcher,
	af importer.AccountFetcher,
	iq importer.ImportEnqueuer,
) http.Handler {
	// Setup the endpoint closure function.
	endpoint := func(ctx context.Context, req api.ImportCsv) (status int, res any, err error) {
		defer func() {
			if closeErr := req.File.Close(); closeErr != nil {
				log.Warn(ctx, "failed closing import file", "error", closeErr.Error())
			}
		}()
		handler := importer.NewFromCsvHandler(ic, dw, dw, vf, af)
		res, err = handler.Handle(ctx, req.File, req.VendorID, req.AccountID)
		if err != nil {
			if errors.Is(err, importer.ErrAccountIDRequired) {
				return http.StatusBadRequest, map[string]string{"account_id": importer.ErrAccountIDRequired.Error()}, nil
			}
			if errors.Is(err, importer.ErrImportAccountNotFound) {
				return http.StatusBadRequest, map[string]string{"account_id": importer.ErrImportAccountNotFound.Error()}, nil
			}
			if errors.Is(err, importer.ErrVendorImportDisabled) {
				return http.StatusBadRequest, map[string]string{"vendor_id": importer.ErrVendorImportDisabled.Error()}, nil
			}
			return http.StatusInternalServerError, struct{}{}, err
		}
		importID, ok := res.(uuid.UUID)
		if !ok {
			return http.StatusInternalServerError, struct{}{}, errors.New("unexpected import id type")
		}
		if iq != nil {
			if err := iq.Enqueue(ctx, importID); err != nil {
				// Import is already persisted; queue reconciliation will retry pending imports.
				log.Warn(ctx, "import persisted but enqueue failed", "import_id", importID, "error", err.Error())
			}
		}
		return http.StatusOK, importID, nil
	}
	// Setup the decoder function.
	decodeFn := httpx.DecoderFunc[api.ImportCsv](func(r *http.Request) (api.ImportCsv, error) {
		return httpx.DecodeMultipartFile[api.ImportCsv](r, httpx.MultipartFileDecoderOptions{
			FieldName:    "file",
			MaxBytes:     20 * 1024 * 1024, // 20 MB
			MaxMemory:    40 * 1024 * 1024, // 40 MB
			AllowedTypes: []string{"text/csv"},
		})
	})
	// Return the constructed endpoint handler.
	return httpx.Endpoint(decodeFn, log, endpoint)
}
