-- +goose Up
-- +goose StatementBegin

-- FK enforcement is per-connection in SQLite; the application DSN must also
-- enable it (e.g. _pragma=foreign_keys(1)). This sets it for the migration.
PRAGMA foreign_keys = ON;

-- SQLite has no schemas, so the Postgres schema.table names are flattened to the
-- prefixed names the storage layer resolves to on SQLite (see qualify* helpers).
-- VARCHAR/TIMESTAMPTZ map to TEXT/DATETIME affinity; size limits are enforced on
-- Postgres only. Money (money.Price, int64) and booleans (0/1) are INTEGER.

-- ---------------------------------------------------------------------------
-- Roots.
-- ---------------------------------------------------------------------------

CREATE TABLE vendors (
    id              TEXT PRIMARY KEY,
    name            TEXT     NOT NULL UNIQUE,
    type            TEXT     NOT NULL CHECK (type IN ('brokerage', 'bank')),
    active          INTEGER  NOT NULL DEFAULT 1,
    import_disabled INTEGER  NOT NULL DEFAULT 0,
    created_at      DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP,
    updated_at      DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP
);

CREATE TABLE accounts (
    id          TEXT PRIMARY KEY,
    external_id TEXT,
    name        TEXT     NOT NULL UNIQUE,
    created_at  DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP,
    updated_at  DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP
);

-- ---------------------------------------------------------------------------
-- Market data: providers + listings.
-- ---------------------------------------------------------------------------

CREATE TABLE providers (
    id             TEXT PRIMARY KEY,
    name           TEXT    NOT NULL CHECK (name IN ('marketstack', 'alphavantage', 'brandnewday')),
    ingestion_mode TEXT    NOT NULL CHECK (ingestion_mode IN ('API', 'MANUAL')),
    api_key        TEXT,
    base_uri       TEXT,
    remaining      INTEGER NOT NULL DEFAULT 0,
    used           INTEGER NOT NULL DEFAULT 0,
    total          INTEGER NOT NULL DEFAULT 0,
    resets_at      TEXT,
    CONSTRAINT ck_providers_ingestion_fields CHECK (
        (ingestion_mode = 'MANUAL' AND api_key IS NULL AND base_uri IS NULL AND resets_at IS NULL)
        OR
        (ingestion_mode = 'API' AND api_key IS NOT NULL AND base_uri IS NOT NULL AND resets_at IS NOT NULL)
    )
);

CREATE TABLE listings (
    id                TEXT PRIMARY KEY,
    symbol            TEXT    NOT NULL,
    name              TEXT    NOT NULL,
    source            TEXT    NOT NULL CHECK (source IN ('alpha_vantage', 'market_stack', 'brandnewday')),
    type              TEXT,
    ticker            TEXT,
    isin              TEXT,
    description       TEXT,
    exchange          TEXT,
    region            TEXT,
    currency          TEXT,
    provider          TEXT    REFERENCES providers(id) ON DELETE SET NULL,
    active            INTEGER NOT NULL DEFAULT 1,
    should_accumulate INTEGER NOT NULL DEFAULT 0,
    syncing           INTEGER NOT NULL DEFAULT 0,
    accumulated_start DATE,
    accumulated_end   DATE,
    created_at        DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP,
    updated_at        DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP,
    CONSTRAINT uq_listings_symbol_source UNIQUE (symbol, source)
);

-- ---------------------------------------------------------------------------
-- Imports. type/source/listing_id modelled on importer.Import; not yet
-- persisted by the store (see CHANGELOG) so kept nullable/defaulted.
-- ---------------------------------------------------------------------------

CREATE TABLE imports (
    id          TEXT PRIMARY KEY,
    vendor_id   TEXT     NOT NULL REFERENCES vendors(id) ON DELETE CASCADE,
    account_id  TEXT     REFERENCES accounts(id) ON DELETE SET NULL,
    listing_id  TEXT     REFERENCES listings(id) ON DELETE SET NULL,
    type        TEXT     CHECK (type IS NULL OR type IN ('cashflow', 'portfolio', 'eod')),
    source      TEXT     NOT NULL DEFAULT '',
    path        TEXT     NOT NULL,
    status      TEXT     NOT NULL DEFAULT 'pending'
                    CHECK (status IN ('pending', 'processing', 'in_progress', 'completed', 'failed')),
    status_msg  TEXT     NOT NULL DEFAULT '',
    duplicates  INTEGER  NOT NULL DEFAULT 0,
    total_rows  INTEGER  NOT NULL DEFAULT 0,
    imported    INTEGER  NOT NULL DEFAULT 0,
    failed      INTEGER  NOT NULL DEFAULT 0,
    created_at  DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP,
    updated_at  DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP
);

-- ---------------------------------------------------------------------------
-- Cashflow.
-- ---------------------------------------------------------------------------

CREATE TABLE cashflow_accounts (
    id         TEXT PRIMARY KEY,
    account_id TEXT     NOT NULL UNIQUE REFERENCES accounts(id) ON DELETE CASCADE,
    created_at DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP,
    updated_at DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP
);

CREATE TABLE transactions (
    id           TEXT PRIMARY KEY,
    account_id   TEXT     NOT NULL REFERENCES cashflow_accounts(account_id) ON DELETE CASCADE,
    import_id    TEXT     REFERENCES imports(id) ON DELETE CASCADE,
    description  TEXT     NOT NULL,
    note         TEXT     NOT NULL DEFAULT '',
    source       TEXT     NOT NULL,
    amount_cents INTEGER  NOT NULL,
    direction    TEXT     NOT NULL CHECK (direction IN ('in', 'out')),
    date         DATE     NOT NULL,
    tag          TEXT     NOT NULL DEFAULT '',
    account_type TEXT     CHECK (account_type IS NULL OR account_type IN ('checking', 'savings', 'credit', 'brokerage')),
    ignored      INTEGER  NOT NULL DEFAULT 0,
    row_number   INTEGER  NOT NULL,
    checksum     TEXT     NOT NULL UNIQUE,
    created_at   DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP,
    updated_at   DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP
);

-- ---------------------------------------------------------------------------
-- Portfolio.
-- ---------------------------------------------------------------------------

CREATE TABLE portfolio_accounts (
    id         TEXT PRIMARY KEY,
    account_id TEXT     NOT NULL UNIQUE REFERENCES accounts(id) ON DELETE CASCADE,
    building   INTEGER  NOT NULL DEFAULT 0,
    created_at DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP,
    updated_at DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP
);

CREATE TABLE positions (
    id           TEXT PRIMARY KEY,
    account_id   TEXT     NOT NULL REFERENCES portfolio_accounts(account_id) ON DELETE CASCADE,
    listing_id   TEXT     REFERENCES listings(id) ON DELETE SET NULL,
    isin         TEXT,
    symbol       TEXT,
    open_date    DATETIME NOT NULL,
    close_date   DATETIME,
    quantity     REAL     NOT NULL DEFAULT 0,
    cost_basis   INTEGER  NOT NULL DEFAULT 0,
    fees         INTEGER  NOT NULL DEFAULT 0,
    income       INTEGER  NOT NULL DEFAULT 0,
    taxes        INTEGER  NOT NULL DEFAULT 0,
    realized_pnl INTEGER  NOT NULL DEFAULT 0,
    created_at   DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP,
    updated_at   DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP
);

CREATE TABLE portfolio_transactions (
    id           TEXT PRIMARY KEY,
    account_id   TEXT     REFERENCES portfolio_accounts(account_id) ON DELETE SET NULL,
    import_id    TEXT     REFERENCES imports(id) ON DELETE CASCADE,
    position_id  TEXT     REFERENCES positions(id) ON DELETE SET NULL,
    origin       TEXT     NOT NULL DEFAULT 'IMPORT' CHECK (origin IN ('IMPORT', 'MANUAL')),
    source       TEXT     NOT NULL,
    occurred_at  DATE     NOT NULL,
    isin         TEXT,
    symbol       TEXT,
    description  TEXT     NOT NULL DEFAULT '',
    type         TEXT     NOT NULL CHECK (type IN ('BUY', 'SELL', 'DIVIDEND', 'TAX', 'FEE', 'CASH')),
    quantity     REAL     NOT NULL DEFAULT 0,
    unit_price   INTEGER  NOT NULL DEFAULT 0,
    amount_cents INTEGER  NOT NULL DEFAULT 0,
    checksum     TEXT     NOT NULL UNIQUE,
    row_number   INTEGER  NOT NULL,
    created_at   DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP,
    updated_at   DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP,
    CONSTRAINT ck_portfolio_transactions_import_origin CHECK (
        (origin = 'IMPORT' AND import_id IS NOT NULL) OR
        (origin = 'MANUAL' AND import_id IS NULL)
    )
);

CREATE TABLE position_snapshots (
    id                  TEXT PRIMARY KEY,
    account_id          TEXT     NOT NULL REFERENCES portfolio_accounts(account_id) ON DELETE CASCADE,
    position_id         TEXT     NOT NULL REFERENCES positions(id) ON DELETE CASCADE,
    listing_id          TEXT     NOT NULL REFERENCES listings(id) ON DELETE CASCADE,
    symbol              TEXT     NOT NULL,
    name                TEXT,
    occurred_at         DATE     NOT NULL,
    quantity            REAL     NOT NULL DEFAULT 0,
    unit_price          INTEGER  NOT NULL DEFAULT 0,
    average_price       INTEGER  NOT NULL DEFAULT 0,
    market_value        INTEGER  NOT NULL DEFAULT 0,
    cost_basis          INTEGER  NOT NULL DEFAULT 0,
    income              INTEGER  NOT NULL DEFAULT 0,
    fees                INTEGER  NOT NULL DEFAULT 0,
    taxes               INTEGER  NOT NULL DEFAULT 0,
    total_pnl           INTEGER  NOT NULL DEFAULT 0,
    total_pnl_pct       REAL     NOT NULL DEFAULT 0,
    realized_pnl        INTEGER  NOT NULL DEFAULT 0,
    unrealized_pnl      INTEGER  NOT NULL DEFAULT 0,
    unrealized_pnl_pct  REAL     NOT NULL DEFAULT 0,
    daily_delta_pnl     INTEGER  NOT NULL DEFAULT 0,
    daily_delta_pnl_pct REAL     NOT NULL DEFAULT 0,
    created_at          DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP,
    updated_at          DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP,
    CONSTRAINT uq_position_snapshots_position_day UNIQUE (position_id, occurred_at)
);

CREATE TABLE portfolio_snapshots (
    id                       TEXT PRIMARY KEY,
    account_id               TEXT     NOT NULL REFERENCES portfolio_accounts(account_id) ON DELETE CASCADE,
    occurred_at              DATE     NOT NULL,
    market_value             INTEGER  NOT NULL DEFAULT 0,
    cost_basis               INTEGER  NOT NULL DEFAULT 0,
    realized_pnl             INTEGER  NOT NULL DEFAULT 0,
    unrealized_pnl           INTEGER  NOT NULL DEFAULT 0,
    unrealized_pnl_pct       REAL     NOT NULL DEFAULT 0,
    income                   INTEGER  NOT NULL DEFAULT 0,
    fees                     INTEGER  NOT NULL DEFAULT 0,
    taxes                    INTEGER  NOT NULL DEFAULT 0,
    cash_balance             INTEGER  NOT NULL DEFAULT 0,
    total_pnl                INTEGER  NOT NULL DEFAULT 0,
    total_pnl_pct            REAL     NOT NULL DEFAULT 0,
    daily_delta_pnl          INTEGER  NOT NULL DEFAULT 0,
    daily_delta_pnl_pct      REAL     NOT NULL DEFAULT 0,
    time_weighted_return_pct REAL     NOT NULL DEFAULT 0,
    net_cashflow             INTEGER  NOT NULL DEFAULT 0,
    cumulative_net_cashflow  INTEGER  NOT NULL DEFAULT 0,
    created_at               DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP,
    updated_at               DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP,
    CONSTRAINT uq_portfolio_snapshots_account_day UNIQUE (account_id, occurred_at)
);

-- ---------------------------------------------------------------------------
-- Market data: end-of-day prices + upload tracking.
-- ---------------------------------------------------------------------------

CREATE TABLE dailies (
    id          TEXT PRIMARY KEY,
    listing_id  TEXT     NOT NULL REFERENCES listings(id) ON DELETE CASCADE,
    symbol      TEXT     NOT NULL,
    date        DATE     NOT NULL,
    open_cents  INTEGER  NOT NULL,
    high_cents  INTEGER  NOT NULL,
    low_cents   INTEGER  NOT NULL,
    close_cents INTEGER  NOT NULL,
    volume      INTEGER  NOT NULL,
    created_at  DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP,
    updated_at  DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP,
    CONSTRAINT uq_dailies_listing_date UNIQUE (listing_id, date)
);

CREATE TABLE daily_uploads (
    id                TEXT PRIMARY KEY,
    listing_id        TEXT     NOT NULL REFERENCES listings(id) ON DELETE CASCADE,
    source            TEXT     NOT NULL,
    status            TEXT     NOT NULL CHECK (status IN ('PENDING', 'PROCESSING', 'SUCCEEDED', 'PARTIAL', 'FAILED')),
    stored_filename   TEXT     NOT NULL,
    original_filename TEXT     NOT NULL,
    status_message    TEXT     NOT NULL DEFAULT '',
    total_rows        INTEGER  NOT NULL DEFAULT 0,
    inserted_rows     INTEGER  NOT NULL DEFAULT 0,
    duplicate_rows    INTEGER  NOT NULL DEFAULT 0,
    error_rows        INTEGER  NOT NULL DEFAULT 0,
    row_errors_json   TEXT     NOT NULL DEFAULT '[]',
    started_at        DATETIME,
    finished_at       DATETIME,
    created_at        DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP,
    updated_at        DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP
);

-- ---------------------------------------------------------------------------
-- Assets.
-- ---------------------------------------------------------------------------

CREATE TABLE assets_accounts (
    id         TEXT PRIMARY KEY,
    account_id TEXT     NOT NULL UNIQUE REFERENCES accounts(id) ON DELETE CASCADE,
    created_at DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP,
    updated_at DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP
);

CREATE TABLE asset_classes (
    id         TEXT PRIMARY KEY,
    account_id TEXT     NOT NULL REFERENCES assets_accounts(account_id) ON DELETE CASCADE,
    name       TEXT     NOT NULL,
    source     TEXT     NOT NULL CHECK (source IN ('MANUAL', 'PORTFOLIO')),
    archived   INTEGER  NOT NULL DEFAULT 0,
    created_at DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP,
    updated_at DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP,
    CONSTRAINT uq_asset_classes_account_name UNIQUE (account_id, name)
);

CREATE TABLE asset_items (
    id            TEXT PRIMARY KEY,
    class_id      TEXT     NOT NULL REFERENCES asset_classes(id) ON DELETE CASCADE,
    account_id    TEXT     NOT NULL REFERENCES assets_accounts(account_id) ON DELETE CASCADE,
    name          TEXT     NOT NULL,
    current_worth INTEGER  NOT NULL DEFAULT 0,
    archived      INTEGER  NOT NULL DEFAULT 0,
    created_at    DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP,
    updated_at    DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP,
    CONSTRAINT uq_asset_items_class_name UNIQUE (class_id, name)
);

CREATE TABLE asset_histories (
    id                TEXT PRIMARY KEY,
    account_id        TEXT     NOT NULL REFERENCES assets_accounts(account_id) ON DELETE CASCADE,
    class_id          TEXT     NOT NULL REFERENCES asset_classes(id) ON DELETE CASCADE,
    item_id           TEXT     NOT NULL REFERENCES asset_items(id) ON DELETE CASCADE,
    change_type       TEXT     NOT NULL CHECK (change_type IN ('SET', 'ADJUST')),
    direction         TEXT     CHECK (direction IS NULL OR direction IN ('INCREASE', 'DECREASE')),
    amount            INTEGER  NOT NULL,
    previous_worth    INTEGER  NOT NULL,
    new_worth         INTEGER  NOT NULL,
    class_total_worth INTEGER  NOT NULL,
    effective_date    DATE     NOT NULL,
    note              TEXT,
    created_at        DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP
);

CREATE TABLE asset_snapshots (
    id          TEXT PRIMARY KEY,
    account_id  TEXT     NOT NULL REFERENCES assets_accounts(account_id) ON DELETE CASCADE,
    occurred_at DATE     NOT NULL,
    total_worth INTEGER  NOT NULL,
    created_at  DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP,
    updated_at  DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP,
    CONSTRAINT uq_asset_snapshots_account_day UNIQUE (account_id, occurred_at)
);

-- ---------------------------------------------------------------------------
-- Indexes (mirrors the Postgres set; justified by internal/storage queries).
-- ---------------------------------------------------------------------------

CREATE INDEX idx_vendors_active_name ON vendors (active, name);

CREATE INDEX idx_imports_status_created ON imports (status, created_at);
CREATE INDEX idx_imports_vendor_id ON imports (vendor_id);

CREATE INDEX idx_cashflow_transactions_account_date ON transactions (account_id, date DESC);
CREATE INDEX idx_cashflow_transactions_import_id ON transactions (import_id);
CREATE INDEX idx_cashflow_transactions_analytics ON transactions (date, direction, tag) WHERE ignored = 0;

CREATE INDEX idx_portfolio_transactions_account_listing ON portfolio_transactions (account_id, occurred_at DESC, created_at DESC, id DESC);
CREATE INDEX idx_portfolio_transactions_position_id ON portfolio_transactions (position_id, occurred_at);
CREATE INDEX idx_portfolio_transactions_import_id ON portfolio_transactions (import_id);

CREATE INDEX idx_portfolio_positions_account_open ON positions (account_id, open_date);
CREATE INDEX idx_portfolio_positions_listing_id ON positions (listing_id);

CREATE INDEX idx_position_snapshots_account_day ON position_snapshots (account_id, occurred_at);
CREATE INDEX idx_position_snapshots_listing_id ON position_snapshots (listing_id);

CREATE INDEX idx_daily_uploads_status_created ON daily_uploads (status, created_at);
CREATE INDEX idx_daily_uploads_listing_id ON daily_uploads (listing_id);

CREATE INDEX idx_providers_name_mode ON providers (name, ingestion_mode);

CREATE INDEX idx_asset_classes_account_source ON asset_classes (account_id, source);

CREATE INDEX idx_asset_items_account_class_archived ON asset_items (account_id, class_id, archived);

CREATE INDEX idx_asset_histories_account_class_date ON asset_histories (account_id, class_id, effective_date DESC);
CREATE INDEX idx_asset_histories_account_date ON asset_histories (account_id, effective_date DESC);
CREATE INDEX idx_asset_histories_item_id ON asset_histories (item_id);

-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin

DROP TABLE IF EXISTS asset_snapshots;
DROP TABLE IF EXISTS asset_histories;
DROP TABLE IF EXISTS asset_items;
DROP TABLE IF EXISTS asset_classes;
DROP TABLE IF EXISTS assets_accounts;
DROP TABLE IF EXISTS daily_uploads;
DROP TABLE IF EXISTS dailies;
DROP TABLE IF EXISTS portfolio_snapshots;
DROP TABLE IF EXISTS position_snapshots;
DROP TABLE IF EXISTS portfolio_transactions;
DROP TABLE IF EXISTS positions;
DROP TABLE IF EXISTS portfolio_accounts;
DROP TABLE IF EXISTS transactions;
DROP TABLE IF EXISTS cashflow_accounts;
DROP TABLE IF EXISTS imports;
DROP TABLE IF EXISTS listings;
DROP TABLE IF EXISTS providers;
DROP TABLE IF EXISTS accounts;
DROP TABLE IF EXISTS vendors;

-- +goose StatementEnd
