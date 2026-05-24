package assets

import (
	"errors"
	"net/http"

	"github.com/google/uuid"
	"github.com/lennardclaproth/my-finances-tracker/internal/assets"
	"github.com/lennardclaproth/my-finances-tracker/internal/logging"
	httpx "github.com/lennardclaproth/my-finances-tracker/internal/transport/http"
)

type DeleteClassRequest struct {
	AccountID uuid.UUID `json:"account_id"`
}

func (r DeleteClassRequest) isValid() (bool, map[string]string) {
	problems := make(map[string]string)
	if r.AccountID == uuid.Nil {
		problems["account_id"] = "account id cannot be an empty UUID"
	}
	return len(problems) == 0, problems
}

// DeleteClass deletes a manual class.
//
// @Summary Delete class
// @Description Deletes a manual class and its items/mutations.
// @Tags assets
// @Accept json
// @Produce json
// @Param class_id path string true "Class ID"
// @Param request body DeleteClassRequest true "Delete class payload"
// @Success 204 "No Content"
// @Failure 400 {object} map[string]string
// @Failure 404 {object} map[string]string
// @Failure 500 {object} map[string]string
// @Router /assets/classes/{class_id} [delete]
func DeleteClass(log logging.Logger, commands assets.Commands) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		classID, err := uuid.Parse(r.PathValue("class_id"))
		if err != nil || classID == uuid.Nil {
			_ = httpx.JSONEncode(w, http.StatusBadRequest, map[string]string{
				"asset_id": "asset_id must be a valid UUID",
			})
			return
		}
		req, err := httpx.JSONDecode[DeleteClassRequest](r)
		if err != nil {
			if httpx.WriteDecodeError(w, err) {
				return
			}
			httpx.JSONEncode(w, http.StatusBadRequest, map[string]string{"error": "invalid request payload"})
		}

		isValid, problems := req.isValid()
		if !isValid {
			httpx.JSONEncode(w, http.StatusBadRequest, problems)
		}

		err = commands.DeleteClass(r.Context(), req.AccountID, classID)

		switch {
		case errors.Is(assets.ErrClassAccountMismatch, err) || errors.Is(assets.ErrClassNotFound, err):
			httpx.JSONEncode(w, http.StatusNotFound, map[string]string{"error": "class not found"})
			return
		case err != nil:
			httpx.JSONEncode(w, http.StatusInternalServerError, map[string]string{"error": "failed to delete class"})
		}

		w.WriteHeader(http.StatusNoContent)
	})
}
