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
    account_id TEXT REFERENCES import_accounts(account_id) ON DELETE RESTRICT,
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
    account_id TEXT REFERENCES cashflow_accounts(account_id) ON DELETE SET NULL,
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

CREATE TABLE IF NOT EXISTS accounts (
    id TEXT PRIMARY KEY,
    external_id TEXT,
    name TEXT NOT NULL UNIQUE,
    created_at DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP,
    updated_at DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP
);

CREATE TABLE IF NOT EXISTS import_accounts (
    id TEXT PRIMARY KEY,
    account_id TEXT NOT NULL UNIQUE REFERENCES accounts(id) ON DELETE CASCADE,
    created_at DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP,
    updated_at DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP
);

CREATE TABLE IF NOT EXISTS cashflow_accounts (
    id TEXT PRIMARY KEY,
    account_id TEXT NOT NULL UNIQUE REFERENCES accounts(id) ON DELETE CASCADE,
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

CREATE TABLE IF NOT EXISTS portfolio_accounts (
    id TEXT PRIMARY KEY,
    account_id TEXT NOT NULL UNIQUE REFERENCES accounts(id) ON DELETE CASCADE,
    building INTEGER NOT NULL DEFAULT 0,
    created_at DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP,
    updated_at DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP
);

CREATE TABLE IF NOT EXISTS positions (
    id TEXT PRIMARY KEY,
    account_id TEXT NOT NULL REFERENCES portfolio_accounts(account_id) ON DELETE CASCADE,
    isin TEXT,
    symbol TEXT,
    listing_id TEXT REFERENCES listings(id) ON DELETE SET NULL,
    open_date DATETIME NOT NULL,
    close_date DATETIME,
    quantity REAL NOT NULL DEFAULT 0,
    cost_basis INTEGER NOT NULL DEFAULT 0,
    fees INTEGER NOT NULL DEFAULT 0,
    income INTEGER NOT NULL DEFAULT 0,
    taxes INTEGER NOT NULL DEFAULT 0,
    realized_pnl INTEGER NOT NULL DEFAULT 0,
    created_at DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP,
    updated_at DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP
);

CREATE TABLE IF NOT EXISTS position_snapshots (
    id TEXT PRIMARY KEY,
    account_id TEXT NOT NULL REFERENCES portfolio_accounts(account_id) ON DELETE CASCADE,
    position_id TEXT NOT NULL REFERENCES positions(id) ON DELETE CASCADE,
    symbol TEXT NOT NULL,
    name TEXT,
    listing_id TEXT NOT NULL REFERENCES listings(id) ON DELETE CASCADE,
    occurred_at DATE NOT NULL,
    quantity REAL NOT NULL DEFAULT 0,
    unit_price INTEGER NOT NULL DEFAULT 0,
    average_price INTEGER NOT NULL DEFAULT 0,
    market_value INTEGER NOT NULL DEFAULT 0,
    cost_basis INTEGER NOT NULL DEFAULT 0,
    income INTEGER NOT NULL DEFAULT 0,
    fees INTEGER NOT NULL DEFAULT 0,
    taxes INTEGER NOT NULL DEFAULT 0,
    total_pnl INTEGER NOT NULL DEFAULT 0,
    total_pnl_pct REAL NOT NULL DEFAULT 0,
    realized_pnl INTEGER NOT NULL DEFAULT 0,
    unrealized_pnl INTEGER NOT NULL DEFAULT 0,
    unrealized_pnl_pct REAL NOT NULL DEFAULT 0,
    daily_delta_pnl INTEGER NOT NULL DEFAULT 0,
    daily_delta_pnl_pct REAL NOT NULL DEFAULT 0,
    created_at DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP,
    updated_at DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP,
    CONSTRAINT uq_position_snapshot UNIQUE (position_id, occurred_at)
);

CREATE TABLE IF NOT EXISTS portfolio_snapshots (
    id TEXT PRIMARY KEY,
    account_id TEXT NOT NULL REFERENCES portfolio_accounts(account_id) ON DELETE CASCADE,
    occurred_at DATE NOT NULL,
    market_value INTEGER NOT NULL DEFAULT 0,
    cost_basis INTEGER NOT NULL DEFAULT 0,
    realized_pnl INTEGER NOT NULL DEFAULT 0,
    unrealized_pnl INTEGER NOT NULL DEFAULT 0,
    unrealized_pnl_pct REAL NOT NULL DEFAULT 0,
    income INTEGER NOT NULL DEFAULT 0,
    fees INTEGER NOT NULL DEFAULT 0,
    taxes INTEGER NOT NULL DEFAULT 0,
    cash_balance INTEGER NOT NULL DEFAULT 0,
    total_pnl INTEGER NOT NULL DEFAULT 0,
    total_pnl_pct REAL NOT NULL DEFAULT 0,
    daily_delta_pnl INTEGER NOT NULL DEFAULT 0,
    daily_delta_pnl_pct REAL NOT NULL DEFAULT 0,
    time_weighted_return_pct REAL NOT NULL DEFAULT 0,
    net_cashflow INTEGER NOT NULL DEFAULT 0,
    cumulative_net_cashflow INTEGER NOT NULL DEFAULT 0,
    created_at DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP,
    updated_at DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP,
    CONSTRAINT uq_portfolio_snapshot UNIQUE (account_id, occurred_at)
);

CREATE TABLE IF NOT EXISTS dailies (
    id TEXT PRIMARY KEY,
    listing_id TEXT NOT NULL REFERENCES listings(id) ON DELETE CASCADE,
    symbol TEXT NOT NULL,
    date DATE NOT NULL,
    open_cents INTEGER NOT NULL,
    high_cents INTEGER NOT NULL,
    low_cents INTEGER NOT NULL,
    close_cents INTEGER NOT NULL,
    volume INTEGER NOT NULL,
    created_at DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP,
    updated_at DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP,
    CONSTRAINT uq_daily_listing_date UNIQUE (listing_id, date)
);

CREATE TABLE IF NOT EXISTS providers (
    name TEXT NOT NULL,
    api_key TEXT NOT NULL,
    base_uri TEXT NOT NULL,
    remaining INTEGER NOT NULL DEFAULT 0,
    used INTEGER NOT NULL DEFAULT 0,
    total INTEGER NOT NULL DEFAULT 0,
    resets_at TEXT NOT NULL DEFAULT '',
    PRIMARY KEY (name, api_key)
);

CREATE INDEX IF NOT EXISTS idx_daily_listing_id ON dailies(listing_id);
CREATE INDEX IF NOT EXISTS idx_daily_symbol ON dailies(symbol);
CREATE INDEX IF NOT EXISTS idx_imports_vendor_id ON imports(vendor_id);
CREATE INDEX IF NOT EXISTS idx_imports_account_id ON imports(account_id);
CREATE INDEX IF NOT EXISTS idx_import_accounts_account_id ON import_accounts(account_id);
CREATE INDEX IF NOT EXISTS idx_transactions_date ON transactions(date);
CREATE INDEX IF NOT EXISTS idx_transactions_account_id ON transactions(account_id);
CREATE INDEX IF NOT EXISTS idx_cashflow_accounts_account_id ON cashflow_accounts(account_id);
CREATE INDEX IF NOT EXISTS idx_transactions_analytics ON transactions(date, direction, tag) WHERE ignored = 0;
CREATE INDEX IF NOT EXISTS idx_portfolio_transactions_occurred_at ON portfolio_transactions(occurred_at);
CREATE INDEX IF NOT EXISTS idx_portfolio_transactions_position_id ON portfolio_transactions(position_id);
CREATE INDEX IF NOT EXISTS idx_portfolio_accounts_account_id ON portfolio_accounts(account_id);
CREATE INDEX IF NOT EXISTS idx_portfolio_positions_account_id ON positions(account_id);
CREATE INDEX IF NOT EXISTS idx_portfolio_positions_listing_id ON positions(listing_id);
CREATE INDEX IF NOT EXISTS idx_position_snapshots_account_id ON position_snapshots(account_id);
CREATE INDEX IF NOT EXISTS idx_position_snapshots_position_id ON position_snapshots(position_id);
CREATE INDEX IF NOT EXISTS idx_portfolio_snapshots_account_id ON portfolio_snapshots(account_id);

-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin
DROP TABLE IF EXISTS portfolio_snapshots;
DROP TABLE IF EXISTS position_snapshots;
DROP TABLE IF EXISTS positions;
DROP TABLE IF EXISTS portfolio_accounts;
DROP TABLE IF EXISTS portfolio_transactions;
DROP TABLE IF EXISTS cashflow_accounts;
DROP TABLE IF EXISTS transactions;
DROP TABLE IF EXISTS imports;
DROP TABLE IF EXISTS import_accounts;
DROP TABLE IF EXISTS accounts;
DROP TABLE IF EXISTS vendors;
DROP TABLE IF EXISTS listings;
DROP TABLE IF EXISTS dailies;
DROP TABLE IF EXISTS providers;
-- +goose StatementEnd
