-- +goose Up
-- +goose StatementBegin

CREATE EXTENSION IF NOT EXISTS pgcrypto;

CREATE SCHEMA IF NOT EXISTS cashflow;
CREATE SCHEMA IF NOT EXISTS marketdata;
CREATE SCHEMA IF NOT EXISTS import;
CREATE SCHEMA IF NOT EXISTS vendor;
CREATE SCHEMA IF NOT EXISTS portfolio;
CREATE SCHEMA IF NOT EXISTS account;

CREATE TABLE vendor.vendors (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    name VARCHAR(255) NOT NULL UNIQUE,
    type VARCHAR(50) NOT NULL,
    active BOOLEAN NOT NULL DEFAULT TRUE,
    created_at TIMESTAMPTZ NOT NULL DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMPTZ NOT NULL DEFAULT CURRENT_TIMESTAMP
);

CREATE TABLE account.accounts (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    external_id UUID,
    name TEXT NOT NULL UNIQUE,
    created_at TIMESTAMPTZ NOT NULL DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMPTZ NOT NULL DEFAULT CURRENT_TIMESTAMP
);

CREATE TABLE import.accounts (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    account_id UUID NOT NULL UNIQUE REFERENCES account.accounts(id) ON DELETE CASCADE,
    created_at TIMESTAMPTZ NOT NULL DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMPTZ NOT NULL DEFAULT CURRENT_TIMESTAMP
);

CREATE TABLE import.imports (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    vendor_id UUID NOT NULL REFERENCES vendor.vendors(id) ON DELETE CASCADE,
    account_id UUID REFERENCES import.accounts(account_id) ON DELETE RESTRICT,
    path TEXT NOT NULL,
    status TEXT NOT NULL DEFAULT 'pending' CHECK (status IN ('pending', 'in_progress', 'completed', 'failed')),
    status_msg TEXT NOT NULL DEFAULT '',
    duplicates INT NOT NULL DEFAULT 0,
    total_rows INT NOT NULL DEFAULT 0,
    imported INT NOT NULL DEFAULT 0,
    failed INT NOT NULL DEFAULT 0,
    created_at TIMESTAMPTZ NOT NULL DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMPTZ NOT NULL DEFAULT CURRENT_TIMESTAMP
);

CREATE TABLE cashflow.transactions (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    account_id UUID,
    description TEXT NOT NULL,
    note TEXT NOT NULL,
    source VARCHAR(255) NOT NULL,
    amount_cents BIGINT NOT NULL,
    direction TEXT NOT NULL CHECK (direction IN ('in', 'out')),
    date DATE NOT NULL,
    checksum VARCHAR(64) NOT NULL UNIQUE,
    created_at TIMESTAMPTZ NOT NULL DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMPTZ NOT NULL DEFAULT CURRENT_TIMESTAMP,
    tag TEXT,
    ignored BOOLEAN NOT NULL DEFAULT FALSE,
    row_number INT NOT NULL,
    import_id UUID NOT NULL REFERENCES import.imports(id) ON DELETE CASCADE,
    account_type TEXT CHECK (account_type IN ('checking', 'savings', 'credit', 'brokerage'))
);

CREATE TABLE cashflow.accounts (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    account_id UUID NOT NULL UNIQUE REFERENCES account.accounts(id) ON DELETE CASCADE,
    created_at TIMESTAMPTZ NOT NULL DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMPTZ NOT NULL DEFAULT CURRENT_TIMESTAMP
);

ALTER TABLE cashflow.transactions
    ADD CONSTRAINT fk_cashflow_transactions_account
    FOREIGN KEY (account_id) REFERENCES cashflow.accounts(account_id) ON DELETE SET NULL;

CREATE TABLE marketdata.listings (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    symbol VARCHAR(50) NOT NULL,
    name VARCHAR(255) NOT NULL,
    type VARCHAR(50),
    ticker VARCHAR(50),
    isin VARCHAR(50),
    description TEXT,
    exchange VARCHAR(50),
    region VARCHAR(50),
    active BOOLEAN NOT NULL DEFAULT TRUE,
    created_at TIMESTAMPTZ NOT NULL DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMPTZ NOT NULL DEFAULT CURRENT_TIMESTAMP,
    accumulated_start DATE,
    accumulated_end DATE,
    should_accumulate BOOLEAN NOT NULL DEFAULT FALSE,
    syncing BOOLEAN NOT NULL DEFAULT FALSE,
    source VARCHAR(50) NOT NULL,
    currency VARCHAR(10),
    CONSTRAINT uq_listing_symbol_source UNIQUE (symbol, source)
);

CREATE TABLE portfolio.transactions (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    account_id UUID REFERENCES account.accounts(id) ON DELETE SET NULL,
    import_id UUID NOT NULL REFERENCES import.imports(id) ON DELETE CASCADE,
    source VARCHAR(255) NOT NULL,
    occurred_at DATE NOT NULL,
    position_id UUID,
    isin VARCHAR(64),
    symbol VARCHAR(255),
    description TEXT NOT NULL DEFAULT '',
    type TEXT NOT NULL CHECK (type IN ('BUY', 'SELL', 'DIVIDEND', 'TAX', 'FEE', 'CASH')),
    quantity DOUBLE PRECISION NOT NULL DEFAULT 0,
    unit_price BIGINT NOT NULL DEFAULT 0,
    amount_cents BIGINT NOT NULL DEFAULT 0,
    checksum VARCHAR(64) NOT NULL UNIQUE,
    row_number INT NOT NULL,
    created_at TIMESTAMPTZ NOT NULL DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMPTZ NOT NULL DEFAULT CURRENT_TIMESTAMP
);

CREATE TABLE portfolio.accounts (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    account_id UUID NOT NULL UNIQUE REFERENCES account.accounts(id) ON DELETE CASCADE,
    building BOOLEAN NOT NULL DEFAULT FALSE,
    created_at TIMESTAMPTZ NOT NULL DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMPTZ NOT NULL DEFAULT CURRENT_TIMESTAMP
);

CREATE TABLE portfolio.positions (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    account_id UUID NOT NULL REFERENCES portfolio.accounts(account_id) ON DELETE CASCADE,
    isin VARCHAR(64),
    symbol VARCHAR(255),
    listing_id UUID REFERENCES marketdata.listings(id) ON DELETE SET NULL,
    open_date TIMESTAMPTZ NOT NULL,
    close_date TIMESTAMPTZ,
    quantity DOUBLE PRECISION NOT NULL DEFAULT 0,
    cost_basis BIGINT NOT NULL DEFAULT 0,
    fees BIGINT NOT NULL DEFAULT 0,
    income BIGINT NOT NULL DEFAULT 0,
    taxes BIGINT NOT NULL DEFAULT 0,
    realized_pnl BIGINT NOT NULL DEFAULT 0,
    created_at TIMESTAMPTZ NOT NULL DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMPTZ NOT NULL DEFAULT CURRENT_TIMESTAMP
);

CREATE TABLE portfolio.position_snapshots (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    account_id UUID NOT NULL REFERENCES portfolio.accounts(account_id) ON DELETE CASCADE,
    position_id UUID NOT NULL REFERENCES portfolio.positions(id) ON DELETE CASCADE,
    symbol VARCHAR(255) NOT NULL,
    name TEXT,
    listing_id UUID NOT NULL REFERENCES marketdata.listings(id) ON DELETE CASCADE,
    occurred_at DATE NOT NULL,
    quantity DOUBLE PRECISION NOT NULL DEFAULT 0,
    unit_price BIGINT NOT NULL DEFAULT 0,
    average_price BIGINT NOT NULL DEFAULT 0,
    market_value BIGINT NOT NULL DEFAULT 0,
    cost_basis BIGINT NOT NULL DEFAULT 0,
    income BIGINT NOT NULL DEFAULT 0,
    fees BIGINT NOT NULL DEFAULT 0,
    taxes BIGINT NOT NULL DEFAULT 0,
    total_pnl BIGINT NOT NULL DEFAULT 0,
    total_pnl_pct DOUBLE PRECISION NOT NULL DEFAULT 0,
    realized_pnl BIGINT NOT NULL DEFAULT 0,
    unrealized_pnl BIGINT NOT NULL DEFAULT 0,
    unrealized_pnl_pct DOUBLE PRECISION NOT NULL DEFAULT 0,
    daily_delta_pnl BIGINT NOT NULL DEFAULT 0,
    daily_delta_pnl_pct DOUBLE PRECISION NOT NULL DEFAULT 0,
    created_at TIMESTAMPTZ NOT NULL DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMPTZ NOT NULL DEFAULT CURRENT_TIMESTAMP,
    CONSTRAINT uq_position_snapshot UNIQUE (position_id, occurred_at)
);

CREATE TABLE portfolio.portfolio_snapshots (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    account_id UUID NOT NULL REFERENCES portfolio.accounts(account_id) ON DELETE CASCADE,
    occurred_at DATE NOT NULL,
    market_value BIGINT NOT NULL DEFAULT 0,
    cost_basis BIGINT NOT NULL DEFAULT 0,
    realized_pnl BIGINT NOT NULL DEFAULT 0,
    unrealized_pnl BIGINT NOT NULL DEFAULT 0,
    unrealized_pnl_pct DOUBLE PRECISION NOT NULL DEFAULT 0,
    income BIGINT NOT NULL DEFAULT 0,
    fees BIGINT NOT NULL DEFAULT 0,
    taxes BIGINT NOT NULL DEFAULT 0,
    cash_balance BIGINT NOT NULL DEFAULT 0,
    total_pnl BIGINT NOT NULL DEFAULT 0,
    total_pnl_pct DOUBLE PRECISION NOT NULL DEFAULT 0,
    daily_delta_pnl BIGINT NOT NULL DEFAULT 0,
    daily_delta_pnl_pct DOUBLE PRECISION NOT NULL DEFAULT 0,
    time_weighted_return_pct DOUBLE PRECISION NOT NULL DEFAULT 0,
    net_cashflow BIGINT NOT NULL DEFAULT 0,
    cumulative_net_cashflow BIGINT NOT NULL DEFAULT 0,
    created_at TIMESTAMPTZ NOT NULL DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMPTZ NOT NULL DEFAULT CURRENT_TIMESTAMP,
    CONSTRAINT uq_portfolio_snapshot UNIQUE (account_id, occurred_at)
);

CREATE TABLE marketdata.dailies (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    listing_id UUID NOT NULL REFERENCES marketdata.listings(id) ON DELETE CASCADE,
    symbol VARCHAR(50) NOT NULL,
    date DATE NOT NULL,
    open_cents BIGINT NOT NULL,
    high_cents BIGINT NOT NULL,
    low_cents BIGINT NOT NULL,
    close_cents BIGINT NOT NULL,
    volume BIGINT NOT NULL,
    created_at TIMESTAMPTZ NOT NULL DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMPTZ NOT NULL DEFAULT CURRENT_TIMESTAMP,

    CONSTRAINT uq_daily_listing_date UNIQUE (listing_id, date)
);

CREATE TABLE marketdata.providers (
    name VARCHAR(50) NOT NULL,
    api_key TEXT NOT NULL,
    base_uri TEXT NOT NULL,
    remaining INT NOT NULL DEFAULT 0,
    used INT NOT NULL DEFAULT 0,
    total INT NOT NULL DEFAULT 0,
    resets_at TEXT NOT NULL DEFAULT '',
    PRIMARY KEY (name, api_key)
);

CREATE INDEX idx_daily_listing_id ON marketdata.dailies(listing_id);
CREATE INDEX idx_daily_symbol ON marketdata.dailies(symbol);
CREATE INDEX idx_imports_vendor_id ON import.imports(vendor_id);
CREATE INDEX idx_imports_account_id ON import.imports(account_id);
CREATE INDEX idx_import_accounts_account_id ON import.accounts(account_id);
CREATE INDEX idx_cashflow_transactions_date ON cashflow.transactions(date);
CREATE INDEX idx_cashflow_transactions_account_id ON cashflow.transactions(account_id);
CREATE INDEX idx_cashflow_accounts_account_id ON cashflow.accounts(account_id);
CREATE INDEX idx_cashflow_transactions_analytics ON cashflow.transactions(date, direction, tag) WHERE ignored = FALSE;
CREATE INDEX idx_portfolio_transactions_occurred_at ON portfolio.transactions(occurred_at);
CREATE INDEX idx_portfolio_transactions_position_id ON portfolio.transactions(position_id);
CREATE INDEX idx_portfolio_accounts_account_id ON portfolio.accounts(account_id);
CREATE INDEX idx_portfolio_positions_account_id ON portfolio.positions(account_id);
CREATE INDEX idx_portfolio_positions_listing_id ON portfolio.positions(listing_id);
CREATE INDEX idx_position_snapshots_account_id ON portfolio.position_snapshots(account_id);
CREATE INDEX idx_position_snapshots_position_id ON portfolio.position_snapshots(position_id);
CREATE INDEX idx_portfolio_snapshots_account_id ON portfolio.portfolio_snapshots(account_id);

-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin
DROP TABLE IF EXISTS portfolio.portfolio_snapshots;
DROP TABLE IF EXISTS portfolio.position_snapshots;
DROP TABLE IF EXISTS portfolio.positions;
DROP TABLE IF EXISTS portfolio.accounts;
DROP TABLE IF EXISTS portfolio.transactions;
DROP TABLE IF EXISTS cashflow.transactions;
DROP TABLE IF EXISTS cashflow.accounts;
DROP TABLE IF EXISTS import.imports;
DROP TABLE IF EXISTS import.accounts;
DROP TABLE IF EXISTS account.accounts;
DROP TABLE IF EXISTS vendor.vendors;
DROP TABLE IF EXISTS marketdata.listings;
DROP TABLE IF EXISTS marketdata.dailies;
DROP TABLE IF EXISTS marketdata.providers;
-- +goose StatementEnd
