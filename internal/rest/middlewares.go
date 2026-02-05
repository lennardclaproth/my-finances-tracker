package rest

import (
	"net/http"
	"time"

	"github.com/lennardclaproth/my-finances-tracker/internal/logging"
)

// withRequestLogging returns a http logging middleware function.
func WithRequestLogging(logger logging.Logger) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			start := time.Now()

			rw := NewResponseWriter(w)
			next.ServeHTTP(rw, r)

			duration := time.Since(start)
			logger.Info(r.Context(), "request completed",
				"method", r.Method,
				"path", r.URL.Path,
				"status", rw.StatusCode,
				"bytes", rw.Size,
				"duration_ms", duration.Milliseconds(),
			)
		})
	}
}
