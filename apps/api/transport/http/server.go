package http

import (
	"context"
	"fmt"
	"net/http"
	"strings"
	"time"

	"github.com/lennardclaproth/my-finances-tracker/internal/logging"
	"go.elastic.co/apm/module/apmhttp/v2"
	"go.elastic.co/apm/v2"
)

type Server struct {
	addr   string
	router *Router
	log    logging.Logger
	mux    *http.ServeMux
}

const (
	readTimeout       = 15 * time.Second
	readHeaderTimeout = 5 * time.Second
	writeTimeout      = 30 * time.Second
	idleTimeout       = 60 * time.Second
)

// NewServer creates and returns a new Server for the given address and database.
func NewServer(addr string, router *Router, log logging.Logger) *Server {
	s := &Server{
		addr:   addr,
		router: router,
		log:    log,
		mux:    http.NewServeMux(),
	}

	s.registerRoutes()
	return s
}

// Run starts the http server on the address provided
func (s *Server) Run(ctx context.Context) error {
	handler := apmhttp.Wrap(s.mux, apmhttp.WithServerRequestIgnorer(func(r *http.Request) bool {
		if shouldIgnoreAPMServerRequest(r) {
			return true
		}
		return apm.DefaultTracer().IgnoredTransactionURL(r.URL)
	}))
	handler = WithRequestIdentifiers()(handler)

	server := s.newHTTPServer(handler)

	s.log.Info(context.Background(),
		"My Finances Tracker is listening for incoming requests...",
		"addr", s.addr,
		"swagger_url", fmt.Sprintf("http://localhost%s/swagger/index.html", s.addr),
	)
	// Run server in a goroutine
	errChan := make(chan error, 1)
	go func() {
		errChan <- server.ListenAndServe()
	}()

	// Wait for interrupt or server error
	select {
	case <-ctx.Done():
		s.log.Info(ctx, "Shutting down gracefully...")
		shutdownCtx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
		defer cancel()
		if err := server.Shutdown(shutdownCtx); err != nil {
			s.log.Error(ctx, "graceful shutdown failed", err)
			return err
		}
		s.log.Info(ctx, "Server stopped cleanly.")
		return nil
	case err := <-errChan:
		if err != nil && err != http.ErrServerClosed {
			s.log.Error(ctx, "server failed", err)
			return err
		}
		return nil
	}
}

// shouldIgnoreAPMServerRequest returns true for requests that should not create APM server transactions.
func shouldIgnoreAPMServerRequest(r *http.Request) bool {
	if r == nil {
		return false
	}
	if strings.EqualFold(strings.TrimSpace(r.Method), http.MethodGet) {
		if strings.TrimSpace(r.Pattern) == "GET /ws/accounts/{account_id}" {
			return true
		}
		if strings.HasPrefix(strings.TrimSpace(r.URL.Path), "/ws/accounts/") {
			return true
		}
	}
	return false
}

func (s *Server) newHTTPServer(handler http.Handler) *http.Server {
	return &http.Server{
		Addr:              s.addr,
		Handler:           handler,
		ReadTimeout:       readTimeout,
		ReadHeaderTimeout: readHeaderTimeout,
		WriteTimeout:      writeTimeout,
		IdleTimeout:       idleTimeout,
	}
}

func (s *Server) registerRoutes() {
	if s.router == nil {
		panic("router is nil")
	}
	s.router.Register(s.mux)
}
