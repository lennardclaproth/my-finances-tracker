package migrations

import (
	"context"
	"embed"
	"fmt"
	"io/fs"
	"net/url"

	"github.com/jmoiron/sqlx"
	"github.com/lennardclaproth/my-finances-tracker/internal/infra/storage"
	"github.com/lennardclaproth/my-finances-tracker/internal/logging"
	"github.com/pressly/goose/v3"
)

//go:embed postgres
var migrations embed.FS

type Migrator struct {
	db       *storage.DB
	connType storage.ConnectionType
	log      logging.Logger
}

func GetFS(connType storage.ConnectionType) fs.FS {
	var dir string
	switch connType {
	case storage.Sqlite:
		dir = "sqlite"
	case storage.Postgres:
		dir = "postgres"
	default:
		dir = "sqlite"
	}

	fsys, err := fs.Sub(migrations, dir)
	if err != nil {
		panic(err)
	}
	return fsys
}

func NewMigrator(db *storage.DB, connType storage.ConnectionType, log logging.Logger) *Migrator {
	return &Migrator{
		db:       db,
		connType: connType,
		log:      log,
	}
}

func (m *Migrator) EnsureDBExists(ctx context.Context, connStr string) error {
	// Parse the DSN
	u, err := url.Parse(connStr)
	if err != nil {
		return fmt.Errorf("invalid connection string: %w", err)
	}

	targetDB := u.Path[1:] // remove leading "/"
	if targetDB == "" {
		return fmt.Errorf("no database name found in connection string")
	}

	// Connect to postgres default DB instead
	u.Path = "/postgres"
	bootstrapConn := u.String()

	db, err := sqlx.Connect("postgres", bootstrapConn)
	if err != nil {
		return fmt.Errorf("unable to connect to bootstrap DB: %w", err)
	}
	defer db.Close()

	// Check if DB exists
	var exists bool
	err = db.Get(&exists, "SELECT EXISTS(SELECT 1 FROM pg_database WHERE datname=$1)", targetDB)
	if err != nil {
		return fmt.Errorf("fetching database existence: %w", err)
	}

	if exists {
		m.log.Info(ctx, "Database %q already exists\n", targetDB)
		return nil
	}

	// Create database
	_, err = db.Exec("CREATE DATABASE " + targetDB)
	if err != nil {
		return fmt.Errorf("creating database: %w", err)
	}

	m.log.Info(ctx, "Database %q created successfully\n", targetDB)
	return nil
}

func (m *Migrator) RunMigrations(ctx context.Context, db *storage.DB, connType storage.ConnectionType) error {
	// Set the correct SQL dialect
	dialect := goose.DialectSQLite3
	if connType == storage.Postgres {
		dialect = goose.DialectPostgres
	}

	// Get the embedded filesystem for the connection type
	fsys := GetFS(storage.ConnectionType(connType))

	// Create a new provider with the embedded migrations
	provider, err := goose.NewProvider(
		dialect,
		db.DB.DB,
		fsys,
		goose.WithVerbose(true),
	)
	if err != nil {
		return fmt.Errorf("failed to create goose provider: %w", err)
	}

	// Run migrations up
	results, err := provider.Up(ctx)
	if err != nil {
		return fmt.Errorf("failed to run migrations: %w", err)
	}

	if len(results) > 0 {
		m.log.Info(ctx, "Applied %d migration(s)\n", len(results))
	}

	return nil
}
