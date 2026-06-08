package importer

import (
	"errors"
	"mime/multipart"
	"net/http"

	"github.com/google/uuid"
	"github.com/lennardclaproth/my-finances-tracker/internal/importer"
	"github.com/lennardclaproth/my-finances-tracker/internal/logging"
	httpx "github.com/lennardclaproth/my-finances-tracker/transport/http"
)

// importEODRequest is the multipart payload accepted for an end-of-day market-data CSV import.
type importEODRequest struct {
	File      multipart.File `multipart:"file"`
	ListingID uuid.UUID      `form:"listing_id"`
}

func (r importEODRequest) isValid() (bool, map[string]string) {
	problems := make(map[string]string)
	if r.ListingID == uuid.Nil {
		problems["listing_id"] = "listing_id is required"
	}
	return len(problems) == 0, problems
}

// ImportEOD accepts an end-of-day market-data CSV upload and queues it for asynchronous processing.
//
// @Summary     Import end-of-day market-data CSV
// @Description Uploads an end-of-day market-data CSV file for a manually ingested listing and queues it for asynchronous processing.
// @Tags        imports
// @Accept      multipart/form-data
// @Produce     application/json
// @Param       file formData file true "End-of-day market-data CSV file"
// @Param       listing_id formData string true "Listing UUID"
// @Success     202 {object} ImportAcceptedResponse
// @Failure     400 {object} map[string]string "Bad request"
// @Failure     404 {object} map[string]string "Not found"
// @Failure     422 {object} map[string]string "Unprocessable entity"
// @Failure     500 {object} map[string]string "Internal server error"
// @Router      /imports/eod [post]
func ImportEOD(log logging.Logger, commands *importer.Commands) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		req, err := httpx.DecodeMultipartFile[importEODRequest](r, httpx.MultipartFileDecoderOptions{
			FieldName: "file",
			MaxBytes:  maxImportUploadBytes,
		})
		defer closeImportUpload(r.Context(), log, req.File, "import eod")
		if err != nil {
			if httpx.WriteDecodeError(w, err) {
				return
			}
			_ = httpx.JSONEncode(w, http.StatusBadRequest, map[string]string{"error": "invalid upload"})
			return
		}
		if isValid, problems := req.isValid(); !isValid {
			_ = httpx.JSONEncode(w, http.StatusBadRequest, problems)
			return
		}

		importID, err := commands.ImportEODCSV(r.Context(), importer.EODCSVImportCommand{
			File:      req.File,
			ListingID: req.ListingID,
		})
		if err != nil {
			switch {
			case errors.Is(err, importer.ErrImportFileRequired):
				_ = httpx.JSONEncode(w, http.StatusBadRequest, map[string]string{"file": err.Error()})
			case errors.Is(err, importer.ErrImportListingNotFound):
				_ = httpx.JSONEncode(w, http.StatusNotFound, map[string]string{"listing_id": err.Error()})
			case errors.Is(err, importer.ErrImportListingInactive),
				errors.Is(err, importer.ErrImportProviderUnavailable),
				errors.Is(err, importer.ErrImportProviderNotManual):
				_ = httpx.JSONEncode(w, http.StatusUnprocessableEntity, map[string]string{"listing_id": err.Error()})
			default:
				log.Error(r.Context(), "import eod: failed to accept import", err)
				_ = httpx.JSONEncode(w, http.StatusInternalServerError, map[string]string{"error": "failed to accept eod import"})
			}
			return
		}

		_ = httpx.JSONEncode(w, http.StatusAccepted, ImportAcceptedResponse{
			ImportID: importID,
			Status:   string(importer.ImportStatusPending),
		})
	})
}
