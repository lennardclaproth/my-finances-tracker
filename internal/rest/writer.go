package rest

import "net/http"

// responseWriter is a custom http.ResponseWriter that captures status code and size of the response.
type responseWriter struct {
	http.ResponseWriter
	StatusCode int
	Size       int
}

// NewResponseWriter wraps an http.ResponseWriter to capture status code and size.
func NewResponseWriter(w http.ResponseWriter) *responseWriter {
	return &responseWriter{ResponseWriter: w, StatusCode: http.StatusOK}
}

func (rw *responseWriter) WriteHeader(code int) {
	rw.StatusCode = code
	rw.ResponseWriter.WriteHeader(code)
}

func (rw *responseWriter) Write(b []byte) (int, error) {
	n, err := rw.ResponseWriter.Write(b)
	rw.Size += n
	return n, err
}
