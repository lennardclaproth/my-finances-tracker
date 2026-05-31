package main

import (
	"context"
	"fmt"
	stdhttp "net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/google/uuid"
	"github.com/lennardclaproth/my-finances-tracker/api"
	_ "github.com/lennardclaproth/my-finances-tracker/docs"
	"github.com/lennardclaproth/my-finances-tracker/internal/account"
	"github.com/lennardclaproth/my-finances-tracker/internal/agent"
	"github.com/lennardclaproth/my-finances-tracker/internal/assets"
	"github.com/lennardclaproth/my-finances-tracker/internal/bootstrap"
	"github.com/lennardclaproth/my-finances-tracker/internal/bus"
	memorybus "github.com/lennardclaproth/my-finances-tracker/internal/bus/memory"
	cashflowdomain "github.com/lennardclaproth/my-finances-tracker/internal/cashflow"
	cashflowservice "github.com/lennardclaproth/my-finances-tracker/internal/cashflow/service"
	"github.com/lennardclaproth/my-finances-tracker/internal/config"
	"github.com/lennardclaproth/my-finances-tracker/internal/http"
	handlers "github.com/lennardclaproth/my-finances-tracker/internal/http/handlers"
	"github.com/lennardclaproth/my-finances-tracker/internal/importer"
	"github.com/lennardclaproth/my-finances-tracker/internal/jobs"
	"github.com/lennardclaproth/my-finances-tracker/internal/logging"
	"github.com/lennardclaproth/my-finances-tracker/internal/marketdata"
	assetshandlers "github.com/lennardclaproth/my-finances-tracker/internal/messaging/handlers/assets"
	cashflowhandlers "github.com/lennardclaproth/my-finances-tracker/internal/messaging/handlers/cashflow"
	importhandlers "github.com/lennardclaproth/my-finances-tracker/internal/messaging/handlers/importer"
	portfoliohandlers "github.com/lennardclaproth/my-finances-tracker/internal/messaging/handlers/portfolio"
	"github.com/lennardclaproth/my-finances-tracker/internal/notify"
	"github.com/lennardclaproth/my-finances-tracker/internal/portfolio"
	"github.com/lennardclaproth/my-finances-tracker/internal/storage"
	"github.com/lennardclaproth/my-finances-tracker/migrations"
	httpSwagger "github.com/swaggo/http-swagger"
	"golang.org/x/sync/errgroup"
)

type appDependencies struct {
	vendorStore               *storage.SQLXVendorStore
	importStore               *storage.SQLXImportStore
	dailyUploadStore          *storage.SQLXDailyUploadStore
	accountStore              *storage.SQLXAccountStore
	importAccountStore        *storage.SQLXImportAccountStore
	cashflowAccountStore      *storage.SQLXCashflowAccountStore
	portfolioAccountStore     *storage.SQLXPortfolioAccountStore
	providerStore             *storage.SQLXProviderStore
	listingStore              *storage.SQLXListingStore
	dailyStore                *storage.SQLXDailyStore
	cashflowTransactionStore  *storage.SQLXBankTransactionStore
	portfolioTransactionStore *storage.SQLXPortfolioTransactionStore
	assetStore                *storage.SQLXAssetStore
	positionStore             *storage.SQLXPositionStore
	portfolioSnapshotStore    *storage.SQLXPortfolioSnapshotStore
	disk                      *storage.Disk
	dailyUploadDisk           *storage.Disk
	dailyUploadEnqueuer       jobs.DailyUploadEnqueuer
	marketDataService         *marketdata.Service
}

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
	defer func() {
		if err := db.Close(); err != nil {
			logger.Error(context.Background(), "failed closing database", err)
		}
	}()

	deps := newAppDependencies(logger, db, cfg)
	b, err := setupBus(logger, deps)
	if err != nil {
		return err
	}
	defer func() {
		if err := b.Close(); err != nil {
			logger.Error(context.Background(), "failed closing bus", err)
		}
	}()
	realtimeHub := notify.NewHub(logger)
	defer func() {
		if err := realtimeHub.Close(); err != nil {
			logger.Error(context.Background(), "failed closing realtime hub", err)
		}
	}()
	if err := setupRealtimeNotifications(logger, b, realtimeHub); err != nil {
		return err
	}
	// Bootstrap initial data
	bootstrapData(ctx, deps, b, logger, cfg)

	// Create background jobs first so the import endpoint can enqueue directly.
	jobMgr, importEnqueuer, bulkTagEnqueuer, dailyUploadEnqueuer := setupJobs(logger, deps, cfg, b)
	deps.dailyUploadEnqueuer = dailyUploadEnqueuer

	// Wiring: construct handlers and routes at the composition root
	router := setupRouterWithDeps(logger, deps, b, importEnqueuer, bulkTagEnqueuer, realtimeHub.Handler())

	// Create server and job manager
	srv := http.NewServer(fmt.Sprintf(":%d", cfg.Server.Port), router, logger)

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
	if err := migrator.EnsureDBExists(context.Background(), cfg.Database.ConnStr); err != nil {
		log.Error(context.Background(), "failed to ensure database exists", err)
		panic(err)
	}
	if err := migrator.RunMigrations(context.Background(), db, dbType); err != nil {
		log.Error(context.Background(), "failed to run migrations", err)
		panic(err)
	}
	return db
}

func newAppDependencies(log logging.Logger, db *storage.DB, cfg *config.Config) *appDependencies {
	vendorStore := storage.NewSQLXVendorStore(db)
	importStore := storage.NewSQLXImportStore(db)
	dailyUploadStore := storage.NewSQLXDailyUploadStore(db)
	accountStore := storage.NewSQLXAccountStore(db)
	importAccountStore := storage.NewSQLXImportAccountStore(db)
	cashflowAccountStore := storage.NewSQLXCashflowAccountStore(db)
	portfolioAccountStore := storage.NewSQLXPortfolioAccountStore(db)
	providerStore := storage.NewSQLXProviderStore(db)
	listingStore := storage.NewSQLXListingStore(db)
	dailyStore := storage.NewSQLXDailyStore(db)
	disk := storage.NewDisk(cfg.DiskStorage.BasePath + "/import")
	dailyUploadDisk := storage.NewDisk(cfg.DiskStorage.BasePath + "/daily_uploads")
	marketStackClient := marketdata.NewMarketStackClient(providerStore, marketdata.ProviderMarketStack)
	mds := marketdata.NewService(listingStore, dailyStore, marketStackClient, log, providerStore)
	return &appDependencies{
		vendorStore:               vendorStore,
		importStore:               importStore,
		dailyUploadStore:          dailyUploadStore,
		accountStore:              accountStore,
		importAccountStore:        importAccountStore,
		cashflowAccountStore:      cashflowAccountStore,
		portfolioAccountStore:     portfolioAccountStore,
		providerStore:             providerStore,
		listingStore:              listingStore,
		dailyStore:                dailyStore,
		cashflowTransactionStore:  storage.NewSQLXBankTransactionStore(db),
		portfolioTransactionStore: storage.NewSQLXPortfolioTransactionStore(db),
		assetStore:                storage.NewSQLXAssetStore(db),
		positionStore:             storage.NewSQLXPositionStore(db),
		portfolioSnapshotStore:    storage.NewSQLXPortfolioSnapshotStore(db),
		disk:                      disk,
		dailyUploadDisk:           dailyUploadDisk,
		marketDataService:         mds,
	}
}

// setupRouter constructs all handlers and registers them with the router.
// This is the composition root where all dependencies are wired together.
func setupRouter(
	log logging.Logger,
	db *storage.DB,
	cfg *config.Config,
	importEnqueuer importer.ImportEnqueuer,
	bulkTagEnqueuer jobs.BulkTagEnqueuer,
) *http.Router {
	deps := newAppDependencies(log, db, cfg)
	return setupRouterWithDeps(log, deps, nil, importEnqueuer, bulkTagEnqueuer)
}

func setupRouterWithDeps(
	log logging.Logger,
	deps *appDependencies,
	b bus.Bus,
	importEnqueuer importer.ImportEnqueuer,
	bulkTagEnqueuer jobs.BulkTagEnqueuer,
	wsHandlers ...stdhttp.Handler,
) *http.Router {
	router := http.NewRouter()
	manualPortfolioTxService := portfolio.NewManualTransactionService(
		deps.accountStore,
		deps.vendorStore,
		deps.listingStore,
		deps.portfolioTransactionStore,
	)
	manualCashflowTxService := cashflowservice.NewManualCreateService(
		deps.accountStore,
		deps.cashflowAccountStore,
		deps.importAccountStore,
		deps.vendorStore,
		deps.importStore,
		deps.cashflowTransactionStore,
	)
	cashflowQueries := cashflowdomain.NewQueries(deps.cashflowTransactionStore)
	portfolioQueries := portfolio.NewQueries(deps.portfolioSnapshotStore, deps.positionStore)
	assetService := assets.NewService(
		deps.accountStore,
		deps.portfolioSnapshotStore,
		deps.assetStore,
		b,
	)

	// Register routes with their handlers
	router.HandleWithMiddleware(
		"POST /import/csv",
		handlers.ImportCsv(
			log,
			deps.importStore,
			deps.disk,
			deps.vendorStore,
			deps.importAccountStore,
			importEnqueuer,
		),
		http.WithRequestLogging(log),
	)
	router.HandleWithMiddleware(
		"POST /accounts",
		handlers.CreateAccount(log, account.NewCreateService(deps.accountStore, b)),
		http.WithRequestLogging(log),
	)
	router.HandleWithMiddleware(
		"GET /accounts",
		handlers.GetAccounts(log, deps.accountStore),
		http.WithRequestLogging(log),
	)
	router.HandleWithMiddleware(
		"GET /vendors",
		handlers.GetVendors(log, deps.vendorStore),
		http.WithRequestLogging(log),
	)
	router.HandleWithMiddleware(
		"POST /marketdata/listing", handlers.CreateListing(
			log,
			deps.marketDataService,
		), http.WithRequestLogging(log),
	)
	router.HandleWithMiddleware(
		"PATCH /marketdata/listing", handlers.UpdateListingFields(
			log,
			deps.marketDataService,
		), http.WithRequestLogging(log),
	)
	router.HandleWithMiddleware(
		"GET /marketdata/listings", handlers.GetListings(
			log,
			deps.marketDataService,
		), http.WithRequestLogging(log),
	)
	router.HandleWithMiddleware(
		"GET /marketdata/listings/search", handlers.SearchListings(
			log,
			deps.marketDataService,
		), http.WithRequestLogging(log),
	)
	router.HandleWithMiddleware(
		"POST /portfolio/rebuild",
		handlers.RebuildPortfolio(log, b, deps.accountStore),
		http.WithRequestLogging(log),
	)
	router.HandleWithMiddleware(
		"GET /portfolio/snapshots",
		handlers.GetPortfolioSnapshots(log, deps.accountStore, portfolioQueries),
		http.WithRequestLogging(log),
	)
	router.HandleWithMiddleware(
		"GET /portfolio/positions",
		handlers.GetPortfolioPositions(log, deps.accountStore, portfolioQueries),
		http.WithRequestLogging(log),
	)
	router.HandleWithMiddleware(
		"GET /portfolio/transactions",
		handlers.GetPortfolioTransactions(log, deps.accountStore, deps.portfolioTransactionStore),
		http.WithRequestLogging(log),
	)
	router.HandleWithMiddleware(
		"POST /portfolio/transactions/manual",
		handlers.CreateManualPortfolioTransaction(log, manualPortfolioTxService),
		http.WithRequestLogging(log),
	)

	router.HandleWithMiddleware(
		"GET /marketdata/dailies", handlers.GetDailies(
			log,
			deps.marketDataService,
		), http.WithRequestLogging(log),
	)
	router.HandleWithMiddleware(
		"POST /marketdata/dailies/upload", handlers.UploadDailiesFile(
			log,
			deps.dailyUploadStore,
			deps.listingStore,
			deps.providerStore,
			deps.dailyUploadDisk,
			deps.dailyUploadDisk,
			deps.dailyUploadEnqueuer,
		), http.WithRequestLogging(log),
	)
	router.HandleWithMiddleware(
		"GET /marketdata/dailies/uploads/{upload_id}", handlers.GetDailyUploadStatus(
			log,
			deps.dailyUploadStore,
		), http.WithRequestLogging(log),
	)
	router.HandleWithMiddleware(
		"GET /cashflow/transactions",
		handlers.GetCashflowTransactions(log, deps.cashflowTransactionStore),
		http.WithRequestLogging(log),
	)
	router.HandleWithMiddleware(
		"POST /cashflow/transactions/manual",
		handlers.CreateManualCashflowTransactions(log, manualCashflowTxService),
		http.WithRequestLogging(log),
	)
	router.HandleWithMiddleware(
		"GET /cashflow/analytics/monthly",
		handlers.GetCashflowMonthlyAnalytics(log, cashflowQueries),
		http.WithRequestLogging(log),
	)
	router.HandleWithMiddleware(
		"GET /cashflow/analytics/tags",
		handlers.GetCashflowTagDistribution(log, deps.cashflowTransactionStore),
		http.WithRequestLogging(log),
	)
	router.HandleWithMiddleware(
		"POST /cashflow/transactions/tag",
		handlers.TagTransaction(log, deps.cashflowTransactionStore),
		http.WithRequestLogging(log),
	)
	router.HandleWithMiddleware(
		"POST /cashflow/transactions/tag/selection",
		handlers.TagTransactionsBySelection(log, deps.cashflowTransactionStore),
		http.WithRequestLogging(log),
	)
	router.HandleWithMiddleware(
		"POST /cashflow/transactions/tag/filter",
		handlers.TagTransactionsByFilter(log, deps.cashflowTransactionStore, bulkTagEnqueuer),
		http.WithRequestLogging(log),
	)
	router.HandleWithMiddleware(
		"POST /cashflow/transactions/ignore/selection",
		handlers.IgnoreTransactionsBySelection(log, deps.cashflowTransactionStore),
		http.WithRequestLogging(log),
	)
	router.HandleWithMiddleware(
		"POST /cashflow/transactions/ignore/filter",
		handlers.IgnoreTransactionsByFilter(log, deps.cashflowTransactionStore),
		http.WithRequestLogging(log),
	)
	router.HandleWithMiddleware(
		"GET /assets/classes",
		handlers.GetAssetClasses(log, assetService),
		http.WithRequestLogging(log),
	)
	router.HandleWithMiddleware(
		"GET /assets/snapshots",
		handlers.GetAssetSnapshots(log, assetService),
		http.WithRequestLogging(log),
	)
	router.HandleWithMiddleware(
		"POST /assets/classes",
		handlers.CreateAssetClass(log, assetService),
		http.WithRequestLogging(log),
	)
	router.HandleWithMiddleware(
		"PATCH /assets/classes",
		handlers.UpdateAssetClass(log, assetService),
		http.WithRequestLogging(log),
	)
	router.HandleWithMiddleware(
		"DELETE /assets/classes/{class_id}",
		handlers.DeleteAssetClass(log, assetService),
		http.WithRequestLogging(log),
	)
	router.HandleWithMiddleware(
		"GET /assets/classes/{class_id}",
		handlers.GetAssetClassDetails(log, assetService),
		http.WithRequestLogging(log),
	)
	router.HandleWithMiddleware(
		"POST /assets/items",
		handlers.CreateAssetItem(log, assetService),
		http.WithRequestLogging(log),
	)
	router.HandleWithMiddleware(
		"POST /assets/items/worth/set",
		handlers.SetAssetItemWorth(log, assetService),
		http.WithRequestLogging(log),
	)
	router.HandleWithMiddleware(
		"POST /assets/items/worth/adjust",
		handlers.AdjustAssetItemWorth(log, assetService),
		http.WithRequestLogging(log),
	)
	if len(wsHandlers) > 0 && wsHandlers[0] != nil {
		router.HandleWithMiddleware(
			"GET /ws/accounts/{account_id}",
			wsHandlers[0],
			http.WithRequestLogging(log),
		)
	}

	router.HandleWithMiddleware(
		"GET /swagger/",
		httpSwagger.WrapHandler,
		http.WithRequestLogging(log),
	)
	router.HandleWithMiddleware(
		"GET /health",
		handlers.HealthHandler(log),
		http.WithRequestLogging(log),
	)

	return router
}

func setupBus(log logging.Logger, deps *appDependencies) (bus.Bus, error) {
	b := memorybus.NewMemoryBus(
		memorybus.WithWorkers(8),
		memorybus.WithQueueSize(256),
		memorybus.WithBackpressure(memorybus.BackpressureDrop),
	)
	psb := portfolio.NewPositionBuilder(
		deps.positionStore,
		deps.portfolioTransactionStore,
		deps.listingStore,
		deps.marketDataService,
	)
	pb := portfolio.NewPortfolioBuilder(
		*psb,
		deps.positionStore,
		deps.portfolioTransactionStore,
		deps.portfolioAccountStore,
		deps.portfolioSnapshotStore,
	)

	reg := bus.NewRegistry(bus.JSONCodec{})
	closeBusWithContext := func(failure error) error {
		if closeErr := b.Close(); closeErr != nil {
			log.Error(context.Background(), "failed closing bus after subscription error", closeErr)
			return fmt.Errorf("%w (additionally failed to close bus: %v)", failure, closeErr)
		}
		return failure
	}
	handler := portfoliohandlers.NewTransactionsImportedHandler(pb, b, log)
	topic := api.TransactionsCreated{}.MessageTopic()
	if _, err := b.Subscribe(topic, bus.DecodeHandler(reg, handler.Handle)); err != nil {
		return nil, closeBusWithContext(fmt.Errorf("failed to subscribe portfolio handler on topic %s: %w", topic, err))
	}
	rebuildHandler := portfoliohandlers.NewPortfolioRebuildRequestedHandler(pb, b, log)
	rebuildTopic := api.PortfolioRebuildRequested{}.MessageTopic()
	if _, err := b.Subscribe(rebuildTopic, bus.DecodeHandler(reg, rebuildHandler.Handle)); err != nil {
		return nil, closeBusWithContext(fmt.Errorf("failed to subscribe portfolio rebuild handler on topic %s: %w", rebuildTopic, err))
	}

	accountTopic := api.AccountCreated{}.MessageTopic()

	portfolioAccountHandler := portfoliohandlers.NewAccountCreatedHandler(deps.portfolioAccountStore)
	if _, err := b.Subscribe(accountTopic, bus.DecodeHandler(reg, portfolioAccountHandler.Handle)); err != nil {
		return nil, closeBusWithContext(fmt.Errorf("failed to subscribe portfolio account handler: %w", err))
	}

	cashflowAccountHandler := cashflowhandlers.NewAccountCreatedHandler(deps.cashflowAccountStore)
	if _, err := b.Subscribe(accountTopic, bus.DecodeHandler(reg, cashflowAccountHandler.Handle)); err != nil {
		return nil, closeBusWithContext(fmt.Errorf("failed to subscribe cashflow account handler: %w", err))
	}
	assetService := assets.NewService(
		deps.accountStore,
		deps.portfolioSnapshotStore,
		deps.assetStore,
		b,
	)
	assetsAccountHandler := assetshandlers.NewAccountCreatedHandler(assetService)
	if _, err := b.Subscribe(accountTopic, bus.DecodeHandler(reg, assetsAccountHandler.Handle)); err != nil {
		return nil, closeBusWithContext(fmt.Errorf("failed to subscribe assets account handler: %w", err))
	}

	importAccountHandler := importhandlers.NewAccountCreatedHandler(deps.importAccountStore)
	if _, err := b.Subscribe(accountTopic, bus.DecodeHandler(reg, importAccountHandler.Handle)); err != nil {
		return nil, closeBusWithContext(fmt.Errorf("failed to subscribe import account handler: %w", err))
	}
	portfolioRebuiltTopic := api.PortfolioRebuilt{}.MessageTopic()
	assetsPortfolioRebuiltHandler := assetshandlers.NewPortfolioRebuiltHandler(assetService)
	if _, err := b.Subscribe(portfolioRebuiltTopic, bus.DecodeHandler(reg, assetsPortfolioRebuiltHandler.Handle)); err != nil {
		return nil, closeBusWithContext(fmt.Errorf("failed to subscribe assets portfolio rebuilt handler: %w", err))
	}
	assetsSnapshotsRebuildTopic := api.AssetsSnapshotsRebuildRequested{}.MessageTopic()
	assetsSnapshotsRebuildHandler := assetshandlers.NewSnapshotsRebuildRequestedHandler(assetService, b, log)
	if _, err := b.Subscribe(assetsSnapshotsRebuildTopic, bus.DecodeHandler(reg, assetsSnapshotsRebuildHandler.Handle)); err != nil {
		return nil, closeBusWithContext(fmt.Errorf("failed to subscribe assets snapshots rebuild handler: %w", err))
	}

	log.Info(context.Background(), "registered bus subscriptions", "topics", []string{topic, rebuildTopic, accountTopic, portfolioRebuiltTopic, assetsSnapshotsRebuildTopic})
	return b, nil
}

func setupJobs(log logging.Logger, deps *appDependencies, cfg *config.Config, b bus.Bus) (*jobs.Manager, importer.ImportEnqueuer, jobs.BulkTagEnqueuer, jobs.DailyUploadEnqueuer) {
	// Setup and start background jobs here
	importJob := jobs.NewImportJob(
		deps.vendorStore,
		deps.importStore,
		deps.cashflowTransactionStore,
		deps.portfolioTransactionStore,
		deps.disk,
		log,
		5*time.Second,
		256,
		b,
	)

	bulkTagJob := jobs.NewBulkTagJob(
		deps.cashflowTransactionStore,
		log,
		4,
		256,
		b,
	)
	dailyUploadJob := jobs.NewDailyUploadJob(
		deps.dailyUploadStore,
		deps.listingStore,
		deps.dailyStore,
		deps.dailyUploadDisk,
		log,
		5*time.Second,
		256,
	)

	managedJobs := []jobs.Job{importJob, dailyUploadJob}

	if cfg.Agent.Enabled {
		agentID, err := uuid.Parse(cfg.Agent.DefaultTagAgentID)
		if err != nil {
			log.Error(context.Background(), "agent auto-tagging disabled due to invalid default_tag_agent_id", err)
		} else {
			taggerJob := jobs.NewTaggerJob(
				agent.NewRunner(
					cfg.Agent.AgentBaseURL,
					agentID,
				),
				deps.cashflowTransactionStore,
				100*time.Millisecond,
				log,
			)
			managedJobs = append(managedJobs, taggerJob)
			log.Info(context.Background(), "agent auto-tagging background job enabled")
		}
	} else {
		log.Info(context.Background(), "agent auto-tagging background job disabled by configuration")
	}

	managedJobs = append(managedJobs, bulkTagJob)

	return jobs.NewManager(log, managedJobs...), importJob, bulkTagJob, dailyUploadJob
}

func setupRealtimeNotifications(log logging.Logger, b bus.Bus, hub *notify.Hub) error {
	if b == nil || hub == nil {
		return nil
	}

	reg := bus.NewRegistry(bus.JSONCodec{})
	importTopic := api.ImportCompleted{}.MessageTopic()
	if _, err := b.Subscribe(importTopic, bus.DecodeHandler(reg, notify.NewImportCompletedHandler(hub).Handle)); err != nil {
		return fmt.Errorf("failed to subscribe realtime import completed handler: %w", err)
	}

	portfolioTopic := api.PortfolioRebuilt{}.MessageTopic()
	if _, err := b.Subscribe(portfolioTopic, bus.DecodeHandler(reg, notify.NewPortfolioRebuiltHandler(hub).Handle)); err != nil {
		return fmt.Errorf("failed to subscribe realtime portfolio rebuilt handler: %w", err)
	}

	bulkTagTopic := api.BulkTagCompleted{}.MessageTopic()
	if _, err := b.Subscribe(bulkTagTopic, bus.DecodeHandler(reg, notify.NewBulkTagCompletedHandler(hub).Handle)); err != nil {
		return fmt.Errorf("failed to subscribe realtime bulk tag completed handler: %w", err)
	}
	assetsSnapshotsTopic := api.AssetsSnapshotsRebuilt{}.MessageTopic()
	if _, err := b.Subscribe(assetsSnapshotsTopic, bus.DecodeHandler(reg, notify.NewAssetsSnapshotsRebuiltHandler(hub).Handle)); err != nil {
		return fmt.Errorf("failed to subscribe realtime assets snapshots rebuilt handler: %w", err)
	}

	log.Info(context.Background(), "registered realtime subscriptions", "topics", []string{importTopic, portfolioTopic, bulkTagTopic, assetsSnapshotsTopic})
	return nil
}

func bootstrapData(ctx context.Context, deps *appDependencies, b bus.Bus, log logging.Logger, cfg *config.Config) {
	// Bootstrap vendors
	bootstrap.Vendors(ctx, deps.vendorStore, log)
	log.Info(ctx, "bootstrapped vendors")

	// Bootstrap default account
	bootstrap.Accounts(ctx, account.NewCreateService(deps.accountStore, b), deps.accountStore, log)
	log.Info(ctx, "bootstrapped accounts")

	// Bootstrap provider API keys from environment
	bootstrap.Providers(ctx, deps.providerStore, cfg.Providers, log)
	log.Info(ctx, "bootstrapped providers")
}
