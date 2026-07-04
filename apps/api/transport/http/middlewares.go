package http

import (
	"context"
	"net/http"
	"strings"
	"time"

	"github.com/lennardclaproth/my-finances-tracker/internal/logging"
	"github.com/lennardclaproth/my-finances-tracker/internal/observability"
)

// WithRequestLogging returns middleware that logs request metadata after handler execution.
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
			// Context identifiers are already embedded in fields above.
			logger.Info(context.Background(), "request completed", observability.FilterFields(fields...)...)
		})
	}
}

// WithCORS returns middleware that applies CORS headers for the configured
// browser origins and short-circuits preflight (OPTIONS) requests with 204. It
// must wrap the router so preflight requests are answered before Go's
// method-based route matching would reject an unregistered OPTIONS pattern.
func WithCORS(allowedOrigins []string) func(http.Handler) http.Handler {
	allowAll := false
	allowed := make(map[string]struct{}, len(allowedOrigins))
	for _, o := range allowedOrigins {
		o = strings.TrimSpace(o)
		if o == "*" {
			allowAll = true
		}
		if o != "" {
			allowed[o] = struct{}{}
		}
	}
	defaultHeaders := "Content-Type, Accept, " + observability.HeaderRequestID + ", " + observability.HeaderCorrelationID
	exposedHeaders := observability.HeaderRequestID + ", " + observability.HeaderCorrelationID

	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			origin := r.Header.Get("Origin")
			if origin != "" {
				if _, ok := allowed[origin]; allowAll || ok {
					w.Header().Set("Access-Control-Allow-Origin", origin)
					w.Header().Add("Vary", "Origin")
					w.Header().Set("Access-Control-Allow-Methods", "GET, POST, PUT, PATCH, DELETE, OPTIONS")
					if reqHeaders := r.Header.Get("Access-Control-Request-Headers"); reqHeaders != "" {
						w.Header().Set("Access-Control-Allow-Headers", reqHeaders)
					} else {
						w.Header().Set("Access-Control-Allow-Headers", defaultHeaders)
					}
					w.Header().Set("Access-Control-Expose-Headers", exposedHeaders)
					w.Header().Set("Access-Control-Max-Age", "600")
				}
			}
			if r.Method == http.MethodOptions {
				w.WriteHeader(http.StatusNoContent)
				return
			}
			next.ServeHTTP(w, r)
		})
	}
}

// WithRequestIdentifiers ensures request/correlation IDs exist on context and response headers.
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
