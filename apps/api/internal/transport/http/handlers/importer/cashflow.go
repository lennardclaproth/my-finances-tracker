package importer

import (
	"context"
	"errors"
	"mime/multipart"
	"net/http"

	"github.com/google/uuid"
	"github.com/lennardclaproth/my-finances-tracker/internal/importer"
	"github.com/lennardclaproth/my-finances-tracker/internal/logging"
	httpx "github.com/lennardclaproth/my-finances-tracker/internal/transport/http"
	"github.com/lennardclaproth/my-finances-tracker/internal/vendor"
)

// maxImportUploadBytes caps the size of an uploaded import CSV file.
const maxImportUploadBytes = 10 << 20 // 10MB

// importCashflowRequest is the multipart payload accepted for a cashflow CSV import.
type importCashflowRequest struct {
	File      multipart.File `multipart:"file"`
	VendorID  uuid.UUID      `form:"vendor_id"`
	AccountID uuid.UUID      `form:"account_id"`
}

func (r importCashflowRequest) isValid() (bool, map[string]string) {
	problems := make(map[string]string)
	if r.VendorID == uuid.Nil {
		problems["vendor_id"] = "vendor_id is required"
	}
	if r.AccountID == uuid.Nil {
		problems["account_id"] = "account_id is required"
	}
	return len(problems) == 0, problems
}

// ImportAcceptedResponse is returned when an import upload is accepted for asynchronous processing.
type ImportAcceptedResponse struct {
	ImportID uuid.UUID `json:"import_id"`
	Status   string    `json:"status"`
}

// ImportCashflow accepts a cashflow CSV upload and queues it for asynchronous processing.
//
// @Summary     Import cashflow CSV
// @Description Uploads a vendor cashflow CSV file for an account and queues it for asynchronous processing.
// @Tags        imports
// @Accept      multipart/form-data
// @Produce     application/json
// @Param       file formData file true "Cashflow CSV file"
// @Param       vendor_id formData string true "Vendor UUID"
// @Param       account_id formData string true "Account UUID"
// @Success     202 {object} ImportAcceptedResponse
// @Failure     400 {object} map[string]string "Bad request"
// @Failure     404 {object} map[string]string "Not found"
// @Failure     422 {object} map[string]string "Unprocessable entity"
// @Failure     500 {object} map[string]string "Internal server error"
// @Router      /imports/cashflow [post]
func ImportCashflow(log logging.Logger, commands *importer.Commands) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		req, err := httpx.DecodeMultipartFile[importCashflowRequest](r, httpx.MultipartFileDecoderOptions{
			FieldName: "file",
			MaxBytes:  maxImportUploadBytes,
		})
		defer closeImportUpload(r.Context(), log, req.File, "import cashflow")
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

		importID, err := commands.ImportCashflowCSV(r.Context(), importer.CashflowCSVImportCommand{
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
			default:
				log.Error(r.Context(), "import cashflow: failed to accept import", err)
				_ = httpx.JSONEncode(w, http.StatusInternalServerError, map[string]string{"error": "failed to accept cashflow import"})
			}
			return
		}

		_ = httpx.JSONEncode(w, http.StatusAccepted, ImportAcceptedResponse{
			ImportID: importID,
			Status:   string(importer.ImportStatusPending),
		})
	})
}

// closeImportUpload closes an uploaded multipart file, logging any close failure.
func closeImportUpload(ctx context.Context, log logging.Logger, file multipart.File, op string) {
	if file == nil {
		return
	}
	if err := file.Close(); err != nil {
		log.Warn(ctx, op+": failed to close upload file", "error", err.Error())
	}
}
