package assets

import (
	"net/http"

	"github.com/google/uuid"
	"github.com/lennardclaproth/my-finances-tracker/internal/assets"
	"github.com/lennardclaproth/my-finances-tracker/internal/logging"
	httpx "github.com/lennardclaproth/my-finances-tracker/internal/transport/http"
)

type CreateAssetClassRequest struct {
	AccountID uuid.UUID `json:"account_id"`
	Name      string    `json:"name"`
}

type CreateAssetClassResponse struct {
	ID uuid.UUID `json:"id"`
}

// CreateClass creates a new manual asset class
//
// @Summary Create asset class
// @Description Creates a manual asset class for the selected account.
// @Tags assets
// @Accept json
// @Produce json
// @Param request body CreateAssetClassRequest true "Create asset class payload"
// @Success 201 {object} CreateAssetClassResponse
// @Failure 400 {object} map[string]string
// @Failure 404 {object} map[string]string
// @Failure 409 {object} map[string]string
// @Failure 500 {object} map[string]string
// @Router /assets/classes [post]
func CreateClass(log logging.Logger, commands assets.Commands) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		req, err := httpx.JSONDecode[CreateAssetClassRequest](r)
		if err != nil {
			if httpx.WriteDecodeError(w, err) {
				return
			}
			httpx.JSONEncode(w, http.StatusBadRequest, map[string]string{"error": "invalid request payload"})
		}
		class, err := commands.CreateClass(r.Context(), req.AccountID, req.Name)
		if err != nil {
			httpx.JSONEncode(w, http.StatusInternalServerError, map[string]string{"error": "failed to create asset class"})
			log.Error(r.Context(), "create class: failed to create class.", err)
			return
		}
		httpx.JSONEncode(w, http.StatusCreated, CreateAssetClassResponse{ID: class.ID})
	})
}
