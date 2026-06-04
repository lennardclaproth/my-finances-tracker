package importer

import (
	"errors"
	"mime/multipart"
	"net/http"

	"github.com/google/uuid"
	"github.com/lennardclaproth/my-finances-tracker/internal/importer"
	"github.com/lennardclaproth/my-finances-tracker/internal/logging"
	"github.com/lennardclaproth/my-finances-tracker/internal/vendor"
	httpx "github.com/lennardclaproth/my-finances-tracker/transport/http"
)

// importPortfolioRequest is the multipart payload accepted for a portfolio CSV import.
type importPortfolioRequest struct {
	File      multipart.File `multipart:"file"`
	VendorID  uuid.UUID      `form:"vendor_id"`
	AccountID uuid.UUID      `form:"account_id"`
}

func (r importPortfolioRequest) isValid() (bool, map[string]string) {
	problems := make(map[string]string)
	if r.VendorID == uuid.Nil {
		problems["vendor_id"] = "vendor_id is required"
	}
	if r.AccountID == uuid.Nil {
		problems["account_id"] = "account_id is required"
	}
	return len(problems) == 0, problems
}

// ImportPortfolio accepts a brokerage portfolio CSV upload and queues it for asynchronous processing.
//
// @Summary     Import portfolio CSV
// @Description Uploads a brokerage portfolio CSV file for an account and queues it for asynchronous processing.
// @Tags        imports
// @Accept      multipart/form-data
// @Produce     application/json
// @Param       file formData file true "Portfolio CSV file"
// @Param       vendor_id formData string true "Brokerage vendor UUID"
// @Param       account_id formData string true "Account UUID"
// @Success     202 {object} ImportAcceptedResponse
// @Failure     400 {object} map[string]string "Bad request"
// @Failure     404 {object} map[string]string "Not found"
// @Failure     422 {object} map[string]string "Unprocessable entity"
// @Failure     500 {object} map[string]string "Internal server error"
// @Router      /imports/portfolio [post]
func ImportPortfolio(log logging.Logger, commands *importer.Commands) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		req, err := httpx.DecodeMultipartFile[importPortfolioRequest](r, httpx.MultipartFileDecoderOptions{
			FieldName: "file",
			MaxBytes:  maxImportUploadBytes,
		})
		defer closeImportUpload(r.Context(), log, req.File, "import portfolio")
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

		importID, err := commands.ImportPortfolioCSV(r.Context(), importer.PortfolioCSVImportCommand{
			File:      req.File,
			VendorID:  req.VendorID,
			AccountID: req.AccountID,
		})
		if err != nil {
			switch {
			case errors.Is(err, importer.ErrImportFileRequired):
				_ = httpx.JSONEncode(w, http.StatusBadRequest, map[string]string{"file": err.Error()})
			case errors.Is(err, importer.ErrVendorIDRequired):
				_ = httpx.JSONEncode(w, http.StatusBadRequest, map[string]string{"vendor_id": err.Error()})
			case errors.Is(err, importer.ErrAccountIDRequired):
				_ = httpx.JSONEncode(w, http.StatusBadRequest, map[string]string{"account_id": err.Error()})
			case errors.Is(err, importer.ErrAccountNotExists):
				_ = httpx.JSONEncode(w, http.StatusNotFound, map[string]string{"account_id": err.Error()})
			case errors.Is(err, vendor.ErrVendorNotFound):
				_ = httpx.JSONEncode(w, http.StatusNotFound, map[string]string{"vendor_id": err.Error()})
			case errors.Is(err, importer.ErrVendorImportDisabled):
				_ = httpx.JSONEncode(w, http.StatusUnprocessableEntity, map[string]string{"vendor_id": err.Error()})
			case errors.Is(err, importer.ErrVendorNotBrokerage):
				_ = httpx.JSONEncode(w, http.StatusUnprocessableEntity, map[string]string{"vendor_id": err.Error()})
			default:
				log.Error(r.Context(), "import portfolio: failed to accept import", err)
				_ = httpx.JSONEncode(w, http.StatusInternalServerError, map[string]string{"error": "failed to accept portfolio import"})
			}
			return
		}

		_ = httpx.JSONEncode(w, http.StatusAccepted, ImportAcceptedResponse{
			ImportID: importID,
			Status:   string(importer.ImportStatusPending),
		})
	})
}
