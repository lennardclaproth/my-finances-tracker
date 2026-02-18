-- +goose Up
-- +goose StatementBegin

CREATE EXTENSION IF NOT EXISTS pgcrypto;

CREATE SCHEMA IF NOT EXISTS cashflow;
CREATE SCHEMA IF NOT EXISTS marketdata;
CREATE SCHEMA IF NOT EXISTS import;
CREATE SCHEMA IF NOT EXISTS vendor;
CREATE SCHEMA IF NOT EXISTS portfolio;

CREATE TABLE vendor.vendors (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    name VARCHAR(255) NOT NULL UNIQUE,
    type VARCHAR(50) NOT NULL,
    active BOOLEAN NOT NULL DEFAULT TRUE,
    created_at TIMESTAMPTZ NOT NULL DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMPTZ NOT NULL DEFAULT CURRENT_TIMESTAMP
);

CREATE TABLE import.imports (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    vendor_id UUID NOT NULL REFERENCES vendor.vendors(id) ON DELETE CASCADE,
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

CREATE TABLE portfolio.transactions (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    account_id UUID,
    import_id UUID NOT NULL REFERENCES import.imports(id) ON DELETE CASCADE,
    source VARCHAR(255) NOT NULL,
    occurred_at DATE NOT NULL,
    value_date DATE,
    listing_id UUID,
    isin VARCHAR(64),
    symbol VARCHAR(255),
    type TEXT NOT NULL CHECK (type IN ('BUY', 'SELL', 'DIVIDEND', 'TAX', 'FEE', 'CASH')),
    quantity DOUBLE PRECISION NOT NULL DEFAULT 0,
    price_cents BIGINT NOT NULL DEFAULT 0,
    amount_cents BIGINT NOT NULL DEFAULT 0,
    checksum VARCHAR(64) NOT NULL UNIQUE,
    raw_ref TEXT NOT NULL DEFAULT '',
    row_number INT NOT NULL,
    created_at TIMESTAMPTZ NOT NULL DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMPTZ NOT NULL DEFAULT CURRENT_TIMESTAMP
);

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

CREATE TABLE marketdata.dailies (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    symbol VARCHAR(50) NOT NULL,
    date DATE NOT NULL,
    open_cents BIGINT NOT NULL,
    high_cents BIGINT NOT NULL,
    low_cents BIGINT NOT NULL,
    close_cents BIGINT NOT NULL,
    volume BIGINT NOT NULL,
    created_at TIMESTAMPTZ NOT NULL DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMPTZ NOT NULL DEFAULT CURRENT_TIMESTAMP,

    CONSTRAINT uq_daily_symbol_date UNIQUE (symbol, date)
);

CREATE TABLE marketdata.providers (
    name VARCHAR(50) PRIMARY KEY,
    api_key TEXT NOT NULL,
    base_uri TEXT NOT NULL
);

CREATE INDEX idx_daily_symbol ON marketdata.dailies(symbol);
CREATE INDEX idx_imports_vendor_id ON import.imports(vendor_id);
CREATE INDEX idx_cashflow_transactions_date ON cashflow.transactions(date);
CREATE INDEX idx_cashflow_transactions_analytics ON cashflow.transactions(date, direction, tag) WHERE ignored = FALSE;
CREATE INDEX idx_portfolio_transactions_occurred_at ON portfolio.transactions(occurred_at);

-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin
DROP TABLE portfolio.transactions;
DROP TABLE cashflow.transactions;
DROP TABLE import.imports;
DROP TABLE vendor.vendors;
DROP TABLE marketdata.listings;
DROP TABLE marketdata.dailies;
DROP TABLE marketdata.providers;
-- +goose StatementEnd
