package handlers

import (
	"net/http"

	"github.com/lennardclaproth/my-finances-tracker/internal/logging"
	apphttp "github.com/lennardclaproth/my-finances-tracker/transport/http"
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
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		apphttp.JSONEncode(w, http.StatusOK, map[string]string{"status": "ok"})
	})
}
