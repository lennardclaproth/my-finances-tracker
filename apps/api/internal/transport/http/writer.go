package http

import (
	"bufio"
	"fmt"
	"net"
	"net/http"
)

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

// Hijack passes through connection hijacking for websocket upgrades.
func (rw *responseWriter) Hijack() (net.Conn, *bufio.ReadWriter, error) {
	hj, ok := rw.ResponseWriter.(http.Hijacker)
	if !ok {
		return nil, nil, fmt.Errorf("underlying response writer does not support hijacking")
	}
	return hj.Hijack()
}

// Flush passes through HTTP streaming flush behavior when supported.
func (rw *responseWriter) Flush() {
	if fl, ok := rw.ResponseWriter.(http.Flusher); ok {
		fl.Flush()
	}
}

// Push passes through HTTP/2 server push behavior when supported.
func (rw *responseWriter) Push(target string, opts *http.PushOptions) error {
	pusher, ok := rw.ResponseWriter.(http.Pusher)
	if !ok {
		return http.ErrNotSupported
	}
	return pusher.Push(target, opts)
}
