package handlers

import (
	"net/http"

	"github.com/lennardclaproth/my-finances-tracker/internal/logging"
)

// HealthHandler returns a simple health check handler.
//
// @Summary     Health check
// @Description Returns 200 when service is healthy
// @Accept      json
// @Produce     application/json
// @Success     200 {object} map[string]string "status"
// @Router      /health [get]
// @Tags        Health
func HealthHandler(log logging.Logger) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		if _, err := w.Write([]byte(`{"status":"ok"}`)); err != nil {
			log.Warn(r.Context(), "failed writing health response", "error", err.Error())
		}
	}
}
