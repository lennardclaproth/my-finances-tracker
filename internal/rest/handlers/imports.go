package handlers

import (
	"context"
	"net/http"

	"github.com/google/uuid"
	"github.com/lennardclaproth/my-finances-tracker/api/contracts"
	"github.com/lennardclaproth/my-finances-tracker/internal/app/imports"
	"github.com/lennardclaproth/my-finances-tracker/internal/infra/storage"
	"github.com/lennardclaproth/my-finances-tracker/internal/logging"
	"github.com/lennardclaproth/my-finances-tracker/internal/rest"
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
// @Success 200 {object} uuid.UUID "Import ID of the created import job"
// @Failure 400 {object} map[string]string "Invalid request (missing file, invalid vendor_id, etc.)"
// @Failure 413 {object} map[string]string "File too large (max 20MB)"
// @Failure 415 {object} map[string]string "Unsupported media type (only text/csv and application/vnd.ms-excel allowed)"
// @Failure 500 {object} map[string]string "Internal server error"
// @Router /import/csv [post]
func ImportCsv(
	log logging.Logger,
	ic imports.ImportCreator,
	dw *storage.DiskWriter,
	vf imports.VendorFetcher,
) http.Handler {
	// Setup the endpoint closure function.
	endpoint := func(ctx context.Context, req contracts.ImportCsv) (status int, res uuid.UUID, err error) {
		defer req.File.Close()
		handler := imports.NewImportFromCsvHandler(ic, dw, dw, vf)
		res, err = handler.Handle(req.File, req.VendorID)
		if err != nil {
			return http.StatusInternalServerError, uuid.Nil, err
		}
		return http.StatusOK, res, nil
	}
	// Setup the decoder function.
	decodeFn := rest.DecoderFunc[contracts.ImportCsv](func(r *http.Request) (contracts.ImportCsv, error) {
		return rest.DecodeMultipartFile[contracts.ImportCsv](r, rest.MultipartFileDecoderOptions{
			FieldName:    "file",
			MaxBytes:     20 * 1024 * 1024, // 20 MB
			MaxMemory:    40 * 1024 * 1024, // 40 MB
			AllowedTypes: []string{"text/csv", "application/vnd.ms-excel"},
		})
	})
	// Return the constructed endpoint handler.
	return rest.Endpoint(decodeFn, log, endpoint)
}
