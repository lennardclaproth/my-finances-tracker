package main

import (
	"context"
	"fmt"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/google/uuid"
	_ "github.com/lennardclaproth/my-finances-tracker/docs"
	"github.com/lennardclaproth/my-finances-tracker/internal/agent"
	"github.com/lennardclaproth/my-finances-tracker/internal/bootstrap"
	"github.com/lennardclaproth/my-finances-tracker/internal/config"
	"github.com/lennardclaproth/my-finances-tracker/internal/http"
	handlers "github.com/lennardclaproth/my-finances-tracker/internal/http/handlers"
	"github.com/lennardclaproth/my-finances-tracker/internal/jobs"
	"github.com/lennardclaproth/my-finances-tracker/internal/logging"
	"github.com/lennardclaproth/my-finances-tracker/internal/storage"
	"github.com/lennardclaproth/my-finances-tracker/migrations"
	httpSwagger "github.com/swaggo/http-swagger"
	"golang.org/x/sync/errgroup"
)

func run(ctx context.Context, args []string) error {
	ctx, stop := signal.NotifyContext(ctx, os.Interrupt, syscall.SIGTERM)
	defer stop()
	// Load configuration
	cfg, err := config.ReadConfig()
	if err != nil {
		return fmt.Errorf("failed to load config: %w", err)
	}

	// Setup
	logger := setupLogger(cfg)
	db := setupDatabase(logger, cfg)
	defer db.Close()

	// Bootstrap initial data
	bootstrapData(ctx, db, logger)

	// Wiring: construct handlers and routes at the composition root
	router := setupRouter(logger, db)

	// Create server and job manager
	srv := http.NewServer(fmt.Sprintf(":%d", cfg.Server.Port), router, logger)
	jobMgr := setupJobs(logger, db, cfg)

	// Run server and jobs concurrently with proper cleanup
	g, ctx := errgroup.WithContext(ctx)

	// Start HTTP server
	g.Go(func() error {
		return srv.Run(ctx)
	})

	// Start background jobs
	g.Go(func() error {
		return jobMgr.Start(ctx)
	})

	// Wait for both to finish
	if err := g.Wait(); err != nil {
		return fmt.Errorf("server or jobs exited with error: %w", err)
	}

	return nil
}

func main() {
	ctx := context.Background()
	if err := run(ctx, os.Args); err != nil {
		fmt.Fprintf(os.Stderr, "%s\n", err)
		os.Exit(1)
	}
}

// setupLogger creates and returns a structured logger based on config.
func setupLogger(cfg *config.Config) logging.Logger {
	return logging.NewSlogLogger(cfg.Logging.GetLogLevel())
}

// setupDatabase creates a database connection and returns it.
// It will panic on connection failure.
func setupDatabase(log logging.Logger, cfg *config.Config) *storage.DB {
	var dbType storage.ConnectionType
	if cfg.Database.Type == "sqlite3" {
		dbType = storage.Sqlite
	} else {
		dbType = storage.Postgres
	}

	db := storage.NewDB(cfg.Database.ConnStr, dbType)
	log.Info(context.Background(), "database connected", "type", dbType)
	migrator := migrations.NewMigrator(db, dbType, log)
	migrator.EnsureDBExists(context.Background(), cfg.Database.ConnStr)
	if err := migrator.RunMigrations(context.Background(), db, dbType); err != nil {
		log.Error(context.Background(), "failed to run migrations", err)
		panic(err)
	}
	return db
}

// setupRouter constructs all handlers and registers them with the router.
// This is the composition root where all dependencies are wired together.
func setupRouter(log logging.Logger, db *storage.DB) *http.Router {
	router := http.NewRouter()

	var transactionRepository = storage.NewSQLXTransactionStore(db)
	var importRepository = storage.NewSQLXImportStore(db)
	var vendorRepository = storage.NewSQLXVendorStore(db)

	var diskWriter = storage.NewDisk("./data/uploads")

	// Register routes with their handlers
	router.HandleWithMiddleware(
		"POST /import/csv",
		handlers.ImportCsv(
			log,
			importRepository,
			diskWriter,
			vendorRepository,
		),
		http.WithRequestLogging(log),
	)
	router.HandleWithMiddleware(
		"POST /transaction/tag",
		handlers.TagTransaction(log, transactionRepository),
		http.WithRequestLogging(log),
	)

	router.Handle("GET /swagger/", httpSwagger.WrapHandler)
	router.Handle("GET /health", handlers.HealthHandler())

	return router
}

func setupJobs(log logging.Logger, db *storage.DB, cfg *config.Config) *jobs.Manager {
	// Setup and start background jobs here
	importJob := jobs.NewImportJob(
		storage.NewSQLXVendorStore(db),
		storage.NewSQLXImportStore(db),
		storage.NewSQLXTransactionStore(db),
		storage.NewDisk(cfg.DiskStorage.BasePath+"/import"),
		log,
		5*time.Second,
	)

	var agentID uuid.UUID
	agentID, err := uuid.Parse(cfg.Agent.DefaultTagAgentID)
	if err != nil {
		agentID = uuid.Nil
	}
	taggerJob := jobs.NewTaggerJob(
		agent.NewRunner(
			cfg.Agent.AgentBaseURL,
			agentID,
		),
		storage.NewSQLXTransactionStore(db),
		100*time.Millisecond,
		log,
	)
	return jobs.NewManager(log, importJob, taggerJob)
}

func bootstrapData(ctx context.Context, db *storage.DB, log logging.Logger) {
	// Bootstrap vendors
	bootstrap.Vendors(ctx, storage.NewSQLXVendorStore(db), log)
	log.Info(ctx, "bootstrapped vendors")
}
