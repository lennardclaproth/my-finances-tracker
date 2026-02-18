package handlers

import (
	"net/http"
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
func HealthHandler() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		_, _ = w.Write([]byte(`{"status":"ok"}`))
	}
}
