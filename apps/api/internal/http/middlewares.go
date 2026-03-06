package http

import (
	"net/http"
	"strings"
	"time"

	"github.com/lennardclaproth/my-finances-tracker/internal/logging"
	"github.com/lennardclaproth/my-finances-tracker/internal/observability"
)

// withRequestLogging returns a http logging middleware function.
func WithRequestLogging(logger logging.Logger) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			start := time.Now()

			rw := NewResponseWriter(w)
			next.ServeHTTP(rw, r)

			duration := time.Since(start)
			outcome := "success"
			if rw.StatusCode >= http.StatusBadRequest {
				outcome = "failure"
			}
			route := strings.TrimSpace(r.Pattern)
			if route == "" {
				route = r.URL.Path
			}
			fields := observability.AppendContextFields(r.Context(),
				"method", r.Method,
				"path", r.URL.Path,
				"status", rw.StatusCode,
				"bytes", rw.Size,
				"duration_ms", duration.Milliseconds(),
				"operation", observability.HTTPOperation(r.Method, route),
				"component", "http",
				"outcome", outcome,
			)
			logger.Info(r.Context(), "request completed", observability.FilterFields(fields...)...)
		})
	}
}

func WithRequestIdentifiers() func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			ctx, requestID, correlationID := observability.EnsureRequestAndCorrelationIDs(
				r.Context(),
				r.Header.Get(observability.HeaderRequestID),
				r.Header.Get(observability.HeaderCorrelationID),
			)
			w.Header().Set(observability.HeaderRequestID, requestID)
			w.Header().Set(observability.HeaderCorrelationID, correlationID)
			next.ServeHTTP(w, r.WithContext(ctx))
		})
	}
}
