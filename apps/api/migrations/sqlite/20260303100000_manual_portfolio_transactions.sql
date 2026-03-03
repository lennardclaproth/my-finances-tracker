-- +goose Up
-- +goose StatementBegin
ALTER TABLE vendors
    ADD COLUMN import_disabled INTEGER NOT NULL DEFAULT 0;

PRAGMA foreign_keys = OFF;

CREATE TABLE portfolio_transactions_new (
    id TEXT PRIMARY KEY,
    account_id TEXT,
    import_id TEXT REFERENCES imports(id) ON DELETE CASCADE,
    source TEXT NOT NULL,
    occurred_at DATE NOT NULL,
    isin TEXT NOT NULL DEFAULT '',
    symbol TEXT NOT NULL DEFAULT '',
    description TEXT NOT NULL DEFAULT '',
    position_id TEXT,
    type TEXT NOT NULL CHECK (type IN ('BUY', 'SELL', 'DIVIDEND', 'TAX', 'FEE', 'CASH')),
    quantity REAL NOT NULL DEFAULT 0,
    unit_price INTEGER NOT NULL DEFAULT 0,
    amount_cents INTEGER NOT NULL DEFAULT 0,
    checksum TEXT NOT NULL UNIQUE,
    row_number INTEGER NOT NULL,
    origin TEXT NOT NULL DEFAULT 'IMPORT' CHECK (origin IN ('IMPORT', 'MANUAL')),
    created_at DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP,
    updated_at DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP,
    CONSTRAINT ck_portfolio_transactions_import_origin
        CHECK (
            (origin = 'IMPORT' AND import_id IS NOT NULL) OR
            (origin = 'MANUAL' AND import_id IS NULL)
        )
);

INSERT INTO portfolio_transactions_new (
    id, account_id, import_id, source, occurred_at, isin, symbol, description, position_id,
    type, quantity, unit_price, amount_cents, checksum, row_number, origin, created_at, updated_at
)
SELECT
    id, account_id, import_id, source, occurred_at, isin, symbol, description, position_id,
    type, quantity, unit_price, amount_cents, checksum, row_number, 'IMPORT', created_at, updated_at
FROM portfolio_transactions;

DROP TABLE portfolio_transactions;
ALTER TABLE portfolio_transactions_new RENAME TO portfolio_transactions;

CREATE INDEX IF NOT EXISTS idx_portfolio_transactions_occurred_at ON portfolio_transactions(occurred_at);
CREATE INDEX IF NOT EXISTS idx_portfolio_transactions_position_id ON portfolio_transactions(position_id);

PRAGMA foreign_keys = ON;
-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin
PRAGMA foreign_keys = OFF;

CREATE TABLE portfolio_transactions_old (
    id TEXT PRIMARY KEY,
    account_id TEXT,
    import_id TEXT NOT NULL REFERENCES imports(id) ON DELETE CASCADE,
    source TEXT NOT NULL,
    occurred_at DATE NOT NULL,
    isin TEXT NOT NULL DEFAULT '',
    symbol TEXT NOT NULL DEFAULT '',
    description TEXT NOT NULL DEFAULT '',
    position_id TEXT,
    type TEXT NOT NULL CHECK (type IN ('BUY', 'SELL', 'DIVIDEND', 'TAX', 'FEE', 'CASH')),
    quantity REAL NOT NULL DEFAULT 0,
    unit_price INTEGER NOT NULL DEFAULT 0,
    amount_cents INTEGER NOT NULL DEFAULT 0,
    checksum TEXT NOT NULL UNIQUE,
    row_number INTEGER NOT NULL,
    created_at DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP,
    updated_at DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP
);

INSERT INTO portfolio_transactions_old (
    id, account_id, import_id, source, occurred_at, isin, symbol, description, position_id,
    type, quantity, unit_price, amount_cents, checksum, row_number, created_at, updated_at
)
SELECT
    id, account_id, import_id, source, occurred_at, isin, symbol, description, position_id,
    type, quantity, unit_price, amount_cents, checksum, row_number, created_at, updated_at
FROM portfolio_transactions
WHERE import_id IS NOT NULL;

DROP TABLE portfolio_transactions;
ALTER TABLE portfolio_transactions_old RENAME TO portfolio_transactions;

CREATE INDEX IF NOT EXISTS idx_portfolio_transactions_occurred_at ON portfolio_transactions(occurred_at);
CREATE INDEX IF NOT EXISTS idx_portfolio_transactions_position_id ON portfolio_transactions(position_id);

-- SQLite does not support dropping a column directly; table rewrite is intentionally omitted here.
PRAGMA foreign_keys = ON;
-- +goose StatementEnd
