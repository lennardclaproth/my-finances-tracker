package main

import (
	"context"
	"fmt"
	"net/http"
	"os"
	"os/signal"
	"syscall"

	_ "github.com/lennardclaproth/my-finances-tracker/docs"
	"github.com/lennardclaproth/my-finances-tracker/internal/account"
	"github.com/lennardclaproth/my-finances-tracker/internal/assets"
	"github.com/lennardclaproth/my-finances-tracker/internal/bootstrap"
	"github.com/lennardclaproth/my-finances-tracker/internal/cashflow"
	"github.com/lennardclaproth/my-finances-tracker/internal/config"
	"github.com/lennardclaproth/my-finances-tracker/internal/eventbus"
	memorybus "github.com/lennardclaproth/my-finances-tracker/internal/eventbus/memory"
	"github.com/lennardclaproth/my-finances-tracker/internal/files"
	"github.com/lennardclaproth/my-finances-tracker/internal/importer"
	importercashflow "github.com/lennardclaproth/my-finances-tracker/internal/importer/cashflow"
	cashflowparsers "github.com/lennardclaproth/my-finances-tracker/internal/importer/cashflow/parsers"
	importereod "github.com/lennardclaproth/my-finances-tracker/internal/importer/eod"
	eodparsers "github.com/lennardclaproth/my-finances-tracker/internal/importer/eod/parsers"
	importerportfolio "github.com/lennardclaproth/my-finances-tracker/internal/importer/portfolio"
	portfolioparsers "github.com/lennardclaproth/my-finances-tracker/internal/importer/portfolio/parsers"
	"github.com/lennardclaproth/my-finances-tracker/internal/logging"
	"github.com/lennardclaproth/my-finances-tracker/internal/marketdata"
	"github.com/lennardclaproth/my-finances-tracker/internal/marketdata/marketstack"
	"github.com/lennardclaproth/my-finances-tracker/internal/notify"
	"github.com/lennardclaproth/my-finances-tracker/internal/portfolio"
	"github.com/lennardclaproth/my-finances-tracker/internal/storage"
	"github.com/lennardclaproth/my-finances-tracker/internal/vendor"
	"github.com/lennardclaproth/my-finances-tracker/migrations"
	apphttp "github.com/lennardclaproth/my-finances-tracker/transport/http"
	basehandlers "github.com/lennardclaproth/my-finances-tracker/transport/http/handlers"
	accounthttp "github.com/lennardclaproth/my-finances-tracker/transport/http/handlers/account"
	assethttp "github.com/lennardclaproth/my-finances-tracker/transport/http/handlers/assets"
	cashflowhttp "github.com/lennardclaproth/my-finances-tracker/transport/http/handlers/cashflow"
	importerhttp "github.com/lennardclaproth/my-finances-tracker/transport/http/handlers/importer"
	marketdatahttp "github.com/lennardclaproth/my-finances-tracker/transport/http/handlers/marketdata"
	portfoliohttp "github.com/lennardclaproth/my-finances-tracker/transport/http/handlers/portfolio"
	vendorshttp "github.com/lennardclaproth/my-finances-tracker/transport/http/handlers/vendors"
	assetsevents "github.com/lennardclaproth/my-finances-tracker/transport/messaging/handlers/assets"
	importerevents "github.com/lennardclaproth/my-finances-tracker/transport/messaging/handlers/importer"
	portfolioevents "github.com/lennardclaproth/my-finances-tracker/transport/messaging/handlers/portfolio"
	httpSwagger "github.com/swaggo/http-swagger"
)

const eventQueueSize = 128

type application struct {
	log logging.Logger
	hub *notify.Hub

	accountCommands *account.Commands
	accountQueries  *account.Queries
	vendorCommands  *vendor.Commands
	vendorQueries   *vendor.Queries

	cashflowCommands *cashflow.Commands
	cashflowQueries  *cashflow.Queries

	portfolioCommands *portfolio.Commands
	portfolioQueries  *portfolio.Queries
	portfolioBuilder  *portfolio.Builder
	portfolioStore    *storage.SQLXPortfolioStore

	assetsCommands *assets.Commands
	assetsQueries  *assets.Queries
	assetsBuilder  *assets.Builder
	assetsSyncer   *assets.Syncer

	importerCommands   *importer.Commands
	marketDataCommands *marketdata.Commands
	marketDataQueries  *marketdata.Queries
}

func main() {
	if err := run(); err != nil {
		_, _ = fmt.Fprintf(os.Stderr, "my-finances-tracker: %v\n", err)
		os.Exit(1)
	}
}

func run() error {
	cfg, err := config.ReadConfig()
	if err != nil {
		return fmt.Errorf("read config: %w", err)
	}

	log := logging.NewSlogLogger(cfg.Logging.GetLogLevel())
	ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
	defer stop()

	connType := storage.ConnectionType(cfg.Database.Type)
	db := storage.NewDB(cfg.Database.ConnStr, connType)
	defer closeDB(log, db)

	migrator := migrations.NewMigrator(db, connType, log)
	if err := migrator.EnsureDBExists(ctx, cfg.Database.ConnStr); err != nil {
		return fmt.Errorf("ensure database exists: %w", err)
	}
	if err := db.PingContext(ctx); err != nil {
		return fmt.Errorf("ping database: %w", err)
	}
	if err := migrator.RunMigrations(ctx, db, connType); err != nil {
		return fmt.Errorf("run migrations: %w", err)
	}

	bus := memorybus.NewMemoryBus(
		memorybus.WithQueueSize(eventQueueSize),
		memorybus.WithBackpressure(memorybus.BackpressureError),
	)
	defer closeBus(log, bus)

	app := buildApplication(cfg, log, bus, db)
	defer closeHub(log, app.hub)

	if err := registerEventHandlers(bus, app); err != nil {
		return err
	}

	bootstrap.Vendors(ctx, app.vendorCommands, log)
	bootstrap.Providers(ctx, app.marketDataCommands, cfg.Providers, log)
	bootstrap.Accounts(ctx, app.accountCommands, app.accountQueries, log)

	router := apphttp.NewRouter()
	registerRoutes(router, app)

	server := apphttp.NewServer(fmt.Sprintf(":%d", cfg.Server.Port), router, log)
	return server.Run(ctx)
}

func buildApplication(
	cfg *config.Config,
	log logging.Logger,
	bus eventbus.Bus,
	db *storage.DB,
) *application {
	accountStore := storage.NewSQLXAccountStore(db)
	vendorStore := storage.NewSQLXVendorStore(db)
	cashflowStore := storage.NewSQLXCashflowStore(db)
	portfolioStore := storage.NewSQLXPortfolioStore(db)
	assetsStore := storage.NewSQLXAssetsStore(db)
	importerStore := storage.NewSQLXImporterStore(db)
	marketDataStore := storage.NewSQLXMarketDataStore(db)
	fileStore := files.NewDisk(cfg.DiskStorage.BasePath)

	marketStackClient := marketstack.NewMarketStackClient(marketDataStore, marketdata.ProviderMarketStack)
	marketDataSyncer := marketdata.NewSyncer(marketDataStore, map[marketdata.Source]marketdata.EODFetcher{
		marketdata.SourceMarketStack: marketStackClient,
	})
	marketDataCommands := marketdata.NewCommands(marketDataStore, marketDataSyncer)
	marketDataQueries := marketdata.NewQueries(marketDataStore, marketDataSyncer)

	accountCommands := account.NewCommands(accountStore, bus)
	accountQueries := account.NewQueries(accountStore)
	vendorCommands := vendor.NewCommands(vendorStore)
	vendorQueries := vendor.NewQueries(vendorStore)
	cashflowCommands := cashflow.NewCommands(cashflowStore, cashflowStore, accountQueries)
	cashflowQueries := cashflow.NewQueries(cashflowStore)
	portfolioCommands := portfolio.NewCommands(portfolioStore, *marketDataQueries, *vendorQueries)
	portfolioQueries := portfolio.NewQueries(portfolioStore)
	portfolioBuilder := portfolio.NewBuilder(marketDataQueries, portfolioStore, portfolioStore, portfolioStore, portfolioStore, bus)
	assetsQueries := assets.NewQueries(assetsStore)
	assetsBuilder := assets.NewBuilder(assetsStore, assetsStore)
	assetsSyncer := assets.NewSyncer(portfolioQueries, assetsBuilder, assetsStore, assetsStore)
	assetsCommands := assets.NewCommands(assetsStore, assetsStore, *accountQueries, assetsStore, assetsStore, bus)
	fileQueries := files.NewQueries(fileStore)

	cashflowProcessor := importercashflow.NewProcessor(vendorQueries, fileQueries, cashflowparsers.CreateCsvParser, cashflowCommands)
	portfolioProcessor := importerportfolio.NewProcessor(vendorQueries, fileQueries, portfolioparsers.CreateCsvParser, portfolioCommands)
	eodProcessor := importereod.NewProcessor(marketDataQueries, fileQueries, eodparsers.CreateEODParser, marketDataCommands)
	importerCommands := importer.NewCommands(
		importerStore,
		fileStore,
		fileStore,
		*vendorQueries,
		*accountQueries,
		*marketDataQueries,
		bus,
		importer.WithProcessors(cashflowProcessor, portfolioProcessor, eodProcessor),
	)

	return &application{
		log: log,
		hub: notify.NewHub(log),

		accountCommands: accountCommands,
		accountQueries:  accountQueries,
		vendorCommands:  vendorCommands,
		vendorQueries:   vendorQueries,

		cashflowCommands: cashflowCommands,
		cashflowQueries:  cashflowQueries,

		portfolioCommands: portfolioCommands,
		portfolioQueries:  portfolioQueries,
		portfolioBuilder:  portfolioBuilder,
		portfolioStore:    portfolioStore,

		assetsCommands: assetsCommands,
		assetsQueries:  assetsQueries,
		assetsBuilder:  assetsBuilder,
		assetsSyncer:   assetsSyncer,

		importerCommands:   importerCommands,
		marketDataCommands: marketDataCommands,
		marketDataQueries:  marketDataQueries,
	}
}

func registerEventHandlers(bus eventbus.Bus, app *application) error {
	if err := subscribe(bus, account.TopicCreated, portfolioevents.NewAccountCreatedHandler(app.portfolioCommands, app.log).Handle); err != nil {
		return err
	}
	if err := subscribe(bus, account.TopicCreated, assetsevents.NewAccountCreatedHandler(app.assetsCommands, app.log).Handle); err != nil {
		return err
	}
	if err := subscribe(bus, importer.TopicAccepted, importerevents.NewAcceptedHandler(app.importerCommands, app.log).Handle); err != nil {
		return err
	}
	if err := subscribe(bus, importer.TopicCompleted, portfolioevents.NewImportCompletedHandler(app.portfolioBuilder, app.log).Handle); err != nil {
		return err
	}
	if err := subscribe(bus, importer.TopicCompleted, notify.NewImportCompletedHandler(app.hub).Handle); err != nil {
		return err
	}
	if err := subscribe(bus, portfolio.TopicRebuilt, assetsevents.NewPortfolioRebuiltHandler(app.assetsSyncer, app.assetsBuilder, bus, app.log).Handle); err != nil {
		return err
	}
	if err := subscribe(bus, portfolio.TopicRebuilt, notify.NewPortfolioRebuiltHandler(app.hub).Handle); err != nil {
		return err
	}
	if err := subscribe(bus, assets.TopicSnapshotsRebuildRequested, assetsevents.NewSnapshotsRebuildRequestedHandler(app.assetsBuilder, bus, app.log).Handle); err != nil {
		return err
	}
	if err := subscribe(bus, assets.TopicSnapshotsRebuilt, notify.NewAssetsSnapshotsRebuiltHandler(app.hub).Handle); err != nil {
		return err
	}
	return nil
}

func subscribe[T any](bus eventbus.Bus, topic string, handler eventbus.Handler[T]) error {
	if _, err := eventbus.Subscribe(bus, topic, handler); err != nil {
		return fmt.Errorf("subscribe %s: %w", topic, err)
	}
	return nil
}

func registerRoutes(router *apphttp.Router, app *application) {
	withRequestLogging := apphttp.WithRequestLogging(app.log)
	handle := func(pattern string, handler http.Handler) {
		router.HandleWithMiddleware(pattern, handler, withRequestLogging)
	}

	handle("GET /health", basehandlers.HealthHandler(app.log))
	handle("GET /swagger/", httpSwagger.WrapHandler)
	handle("GET /ws/accounts/{account_id}", app.hub.Handler())

	handle("POST /accounts", accounthttp.Create(app.log, *app.accountCommands))
	handle("GET /vendors", vendorshttp.List(app.log, app.vendorQueries))

	handle("POST /imports/cashflow", importerhttp.ImportCashflow(app.log, app.importerCommands))
	handle("POST /imports/portfolio", importerhttp.ImportPortfolio(app.log, app.importerCommands))
	handle("POST /imports/eod", importerhttp.ImportEOD(app.log, app.importerCommands))

	handle("POST /marketdata/listing", marketdatahttp.CreateListing(app.log, app.marketDataCommands))
	handle("PATCH /marketdata/listing", marketdatahttp.UpdateListingFields(app.log, app.marketDataCommands))
	handle("GET /marketdata/listings", marketdatahttp.GetListings(app.log, app.marketDataQueries))
	handle("GET /marketdata/listings/search", marketdatahttp.SearchListings(app.log, app.marketDataQueries))
	handle("GET /marketdata/eods", marketdatahttp.GetEOD(app.log, app.marketDataQueries))

	handle("GET /cashflow/transactions", cashflowhttp.GetTransactions(app.log, app.cashflowQueries))
	handle("POST /cashflow/transactions/manual", cashflowhttp.CreateTransactions(app.log, app.cashflowCommands))
	handle("GET /cashflow/analytics/monthly", cashflowhttp.GetMonthlyAnalytics(app.log, app.cashflowQueries))
	handle("GET /cashflow/analytics/tags", cashflowhttp.GetCashflowTagDistribution(app.log, app.cashflowQueries))
	handle("POST /cashflow/transactions/tag", cashflowhttp.TagTransaction(app.log, app.cashflowCommands))
	handle("POST /cashflow/transactions/tag/selection", cashflowhttp.TagTransactionsBySelection(app.log, app.cashflowCommands))
	handle("POST /cashflow/transactions/tag/filter", cashflowhttp.TagTransactionsByFilter(app.log, app.cashflowCommands))
	handle("POST /cashflow/transactions/ignore/selection", cashflowhttp.IgnoreTransactionsBySelection(app.log, app.cashflowCommands))
	handle("POST /cashflow/transactions/ignore/filter", cashflowhttp.IgnoreTransactionsByFilter(app.log, app.cashflowCommands))

	handle("GET /portfolio/positions", portfoliohttp.GetPortfolioPositions(app.log, app.portfolioQueries))
	handle("GET /portfolio/snapshots", portfoliohttp.GetPortfolioSnapshots(app.log, app.accountQueries, app.portfolioQueries))
	handle("GET /portfolio/transactions", portfoliohttp.GetPortfolioTransactions(app.log, app.accountQueries, app.portfolioStore))
	handle("POST /portfolio/transactions/manual", portfoliohttp.CreateManualPortfolioTransaction(app.log, app.portfolioCommands))
	handle("POST /portfolio/rebuild", portfoliohttp.RebuildPortfolio(app.log, app.portfolioBuilder))

	handle("GET /assets/classes", assethttp.GetClasses(app.log, *app.assetsQueries))
	handle("POST /assets/classes", assethttp.CreateClass(app.log, *app.assetsCommands))
	handle("PATCH /assets/classes", assethttp.UpdateClass(app.log, *app.assetsCommands))
	handle("GET /assets/classes/{class_id}", assethttp.GetClassDetails(app.log, *app.assetsQueries))
	handle("DELETE /assets/classes/{class_id}", assethttp.DeleteClass(app.log, *app.assetsCommands))
	handle("POST /assets", assethttp.CreateAsset(app.log, *app.assetsCommands))
	handle("PUT /assets/{asset_id}/worth", assethttp.SetAssetWorth(app.log, *app.assetsCommands))
	handle("PUT /assets/{asset_id}/adjust", assethttp.AdjustAssetWorth(app.log, *app.assetsCommands))
	handle("GET /assets/snapshots", assethttp.GetSnapshots(app.log, *app.assetsQueries))
}

func closeDB(log logging.Logger, db *storage.DB) {
	if err := db.Close(); err != nil {
		log.Warn(context.Background(), "failed closing database", "error", err.Error())
	}
}

func closeBus(log logging.Logger, bus eventbus.Bus) {
	if err := bus.Close(); err != nil {
		log.Warn(context.Background(), "failed closing event bus", "error", err.Error())
	}
}

func closeHub(log logging.Logger, hub *notify.Hub) {
	if err := hub.Close(); err != nil {
		log.Warn(context.Background(), "failed closing websocket hub", "error", err.Error())
	}
}
