package storage

import (
	"context"
	"fmt"

	"github.com/jmoiron/sqlx"

	_ "github.com/lib/pq"
	"go.elastic.co/apm/module/apmsql/v2"
	_ "go.elastic.co/apm/module/apmsql/v2/pq"
)

type ConnectionType string

const (
	Sqlite   ConnectionType = "sqlite3"
	Postgres ConnectionType = "postgres"
)

const (
	TableVendors      = "vendors"
	TableTransactions = "transactions"
	TableImports      = "imports"
)

type DB struct {
	*sqlx.DB
}

func NewDB(connStr string, connType ConnectionType) *DB {
	db, err := apmsql.Open(string(connType), connStr)
	if err != nil {
		panic(fmt.Errorf("db: failed to open connection to database: %w", err))
	}

	sqlxDB := sqlx.NewDb(db, string(connType))

	return &DB{DB: sqlxDB}
}

func (db *DB) GetExecutor(ctx context.Context) sqlx.ExtContext {
	tx, ok := ctx.Value("tx").(*sqlx.Tx)
	if ok {
		return tx
	}
	return db
}
