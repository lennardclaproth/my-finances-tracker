package storage

import (
	"context"
	"database/sql"
	"fmt"

	"github.com/jmoiron/sqlx"

	_ "github.com/lib/pq"
	"go.elastic.co/apm/module/apmsql/v2"
	_ "go.elastic.co/apm/module/apmsql/v2/pq"
	_ "modernc.org/sqlite"
)

type ConnectionType string

const (
	Sqlite   ConnectionType = "sqlite3"
	Postgres ConnectionType = "postgres"
)

const (
	SchemaVendors    = "vendor"
	SchemaCashflow   = "cashflow"
	SchemaPortfolio  = "portfolio"
	SchemaAccount    = "account"
	SchemaImports    = "import"
	SchemaMarketData = "marketdata"
)

const (
	TableVendors       = "vendors"
	TableTransactions  = "transactions"
	TableImports       = "imports"
	TableListings      = "listings"
	TableHistories     = "dailies"
	TableProviders     = "providers"
	TableAccounts      = "accounts"
	TablePositions     = "positions"
	TablePosSnapshots  = "position_snapshots"
	TablePortSnapshots = "portfolio_snapshots"
)

type DB struct {
	*sqlx.DB
}

func NewDB(connStr string, connType ConnectionType) *DB {
	var (
		db  *sql.DB
		err error
	)
	if connType == Sqlite {
		db, err = sql.Open("sqlite", connStr)
	} else {
		db, err = apmsql.Open(string(connType), connStr)
	}
	if err != nil {
		panic(fmt.Errorf("db: failed to open connection to database: %w", err))
	}

	sqlxDB := sqlx.NewDb(db, string(connType))

	return &DB{DB: sqlxDB}
}

func qualifyTable(db *DB, schema, table string) string {
	if db == nil || db.DriverName() == string(Sqlite) {
		return table
	}
	return fmt.Sprintf("%s.%s", schema, table)
}

func (db *DB) GetExecutor(ctx context.Context) sqlx.ExtContext {
	tx, ok := ctx.Value("tx").(*sqlx.Tx)
	if ok {
		return tx
	}
	return db
}
