-- +goose Up
-- +goose StatementBegin

PRAGMA foreign_keys = ON;

CREATE TABLE IF NOT EXISTS vendors (
    id TEXT PRIMARY KEY,
    name TEXT NOT NULL UNIQUE,
    type TEXT NOT NULL,
    active INTEGER NOT NULL DEFAULT 1,
    created_at DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP,
    updated_at DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP
);

CREATE TABLE IF NOT EXISTS imports (
    id TEXT PRIMARY KEY,
    vendor_id TEXT NOT NULL REFERENCES vendors(id) ON DELETE CASCADE,
    path TEXT NOT NULL,
    status TEXT NOT NULL DEFAULT 'pending' CHECK (status IN ('pending', 'in_progress', 'completed', 'failed')),
    status_msg TEXT NOT NULL DEFAULT '',
    duplicates INTEGER NOT NULL DEFAULT 0,
    total_rows INTEGER NOT NULL DEFAULT 0,
    imported INTEGER NOT NULL DEFAULT 0,
    failed INTEGER NOT NULL DEFAULT 0,
    created_at DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP,
    updated_at DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP
);

CREATE TABLE IF NOT EXISTS transactions (
    id TEXT PRIMARY KEY,
    description TEXT NOT NULL,
    note TEXT NOT NULL,
    source TEXT NOT NULL,
    amount_cents INTEGER NOT NULL,
    direction TEXT NOT NULL CHECK (direction IN ('in', 'out')),
    date DATE NOT NULL,
    checksum TEXT NOT NULL UNIQUE,
    created_at DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP,
    updated_at DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP,
    tag TEXT,
    ignored INTEGER NOT NULL DEFAULT 0,
    row_number INTEGER NOT NULL,
    import_id TEXT NOT NULL REFERENCES imports(id) ON DELETE CASCADE,
    account_type TEXT CHECK (account_type IN ('checking', 'savings', 'credit', 'brokerage'))
);

CREATE TABLE IF NOT EXISTS portfolio_transactions (
    id TEXT PRIMARY KEY,
    account_id TEXT,
    import_id TEXT NOT NULL REFERENCES imports(id) ON DELETE CASCADE,
    source TEXT NOT NULL,
    occurred_at DATE NOT NULL,
    value_date DATE,
    listing_id TEXT,
    isin TEXT NOT NULL DEFAULT '',
    symbol TEXT NOT NULL DEFAULT '',
    type TEXT NOT NULL CHECK (type IN ('BUY', 'SELL', 'DIVIDEND', 'TAX', 'FEE', 'CASH')),
    quantity REAL NOT NULL DEFAULT 0,
    price_cents INTEGER NOT NULL DEFAULT 0,
    amount_cents INTEGER NOT NULL DEFAULT 0,
    checksum TEXT NOT NULL UNIQUE,
    raw_ref TEXT NOT NULL DEFAULT '',
    row_number INTEGER NOT NULL,
    created_at DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP,
    updated_at DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP
);

CREATE TABLE IF NOT EXISTS listings (
    id TEXT PRIMARY KEY,
    symbol TEXT NOT NULL,
    name TEXT NOT NULL,
    type TEXT,
    ticker TEXT,
    isin TEXT,
    description TEXT,
    exchange TEXT,
    region TEXT,
    active INTEGER NOT NULL DEFAULT 1,
    created_at DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP,
    updated_at DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP,
    accumulated_start DATE,
    accumulated_end DATE,
    should_accumulate INTEGER NOT NULL DEFAULT 0,
    syncing INTEGER NOT NULL DEFAULT 0,
    source TEXT NOT NULL,
    currency TEXT,
    CONSTRAINT uq_listing_symbol_source UNIQUE (symbol, source)
);

CREATE TABLE IF NOT EXISTS dailies (
    id TEXT PRIMARY KEY,
    symbol TEXT NOT NULL,
    date DATE NOT NULL,
    open_cents INTEGER NOT NULL,
    high_cents INTEGER NOT NULL,
    low_cents INTEGER NOT NULL,
    close_cents INTEGER NOT NULL,
    volume INTEGER NOT NULL,
    created_at DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP,
    updated_at DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP,
    CONSTRAINT uq_daily_symbol_date UNIQUE (symbol, date)
);

CREATE TABLE IF NOT EXISTS providers (
    name TEXT PRIMARY KEY,
    api_key TEXT NOT NULL,
    base_uri TEXT NOT NULL
);

CREATE INDEX IF NOT EXISTS idx_daily_symbol ON dailies(symbol);
CREATE INDEX IF NOT EXISTS idx_imports_vendor_id ON imports(vendor_id);
CREATE INDEX IF NOT EXISTS idx_transactions_date ON transactions(date);
CREATE INDEX IF NOT EXISTS idx_transactions_analytics ON transactions(date, direction, tag) WHERE ignored = 0;
CREATE INDEX IF NOT EXISTS idx_portfolio_transactions_occurred_at ON portfolio_transactions(occurred_at);

-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin
DROP TABLE IF EXISTS portfolio_transactions;
DROP TABLE IF EXISTS transactions;
DROP TABLE IF EXISTS imports;
DROP TABLE IF EXISTS vendors;
DROP TABLE IF EXISTS listings;
DROP TABLE IF EXISTS dailies;
DROP TABLE IF EXISTS providers;
-- +goose StatementEnd
