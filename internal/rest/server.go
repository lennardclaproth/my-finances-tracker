package rest

import (
	"context"
	"fmt"
	"net/http"
	"time"

	"github.com/lennardclaproth/my-finances-tracker/internal/logging"
	"go.elastic.co/apm/module/apmhttp/v2"
)

type Server struct {
	addr   string
	router *Router
	log    logging.Logger
	mux    *http.ServeMux
}

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
	server := &http.Server{
		Addr:    s.addr,
		Handler: apmhttp.Wrap(s.mux),
	}

	s.log.Info(context.Background(),
		"My Fincances Tracker is listening for incoming requests...",
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

func (s *Server) registerRoutes() error {
	if s.router == nil {
		return fmt.Errorf("router is nil")
	}
	s.router.Register(s.mux)
	return nil
}
