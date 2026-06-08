-- +goose Up
-- +goose StatementBegin

-- gen_random_uuid() is core in PostgreSQL 13+, but pgcrypto keeps this portable
-- for older servers. Application code always supplies its own UUIDs; the default
-- only matters for ad-hoc manual inserts.
CREATE EXTENSION IF NOT EXISTS pgcrypto;

CREATE SCHEMA IF NOT EXISTS vendor;
CREATE SCHEMA IF NOT EXISTS account;
CREATE SCHEMA IF NOT EXISTS import;
CREATE SCHEMA IF NOT EXISTS cashflow;
CREATE SCHEMA IF NOT EXISTS marketdata;
CREATE SCHEMA IF NOT EXISTS portfolio;
CREATE SCHEMA IF NOT EXISTS assets;

-- ---------------------------------------------------------------------------
-- Roots: vendors and accounts have no dependencies.
-- ---------------------------------------------------------------------------

CREATE TABLE vendor.vendors (
    id              UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    name            VARCHAR(64)  NOT NULL UNIQUE,
    type            VARCHAR(16)  NOT NULL CHECK (type IN ('brokerage', 'bank')),
    active          BOOLEAN      NOT NULL DEFAULT TRUE,
    import_disabled BOOLEAN      NOT NULL DEFAULT FALSE,
    created_at      TIMESTAMPTZ  NOT NULL DEFAULT CURRENT_TIMESTAMP,
    updated_at      TIMESTAMPTZ  NOT NULL DEFAULT CURRENT_TIMESTAMP
);

CREATE TABLE account.accounts (
    id          UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    external_id VARCHAR(255),
    name        VARCHAR(255) NOT NULL UNIQUE,
    created_at  TIMESTAMPTZ  NOT NULL DEFAULT CURRENT_TIMESTAMP,
    updated_at  TIMESTAMPTZ  NOT NULL DEFAULT CURRENT_TIMESTAMP
);

-- ---------------------------------------------------------------------------
-- Market data: providers and listings underpin portfolio + EOD data.
-- ---------------------------------------------------------------------------

CREATE TABLE marketdata.providers (
    id             UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    name           VARCHAR(16) NOT NULL CHECK (name IN ('marketstack', 'alphavantage', 'brandnewday')),
    ingestion_mode VARCHAR(8)  NOT NULL CHECK (ingestion_mode IN ('API', 'MANUAL')),
    api_key        VARCHAR(255),
    base_uri       VARCHAR(255),
    remaining      INTEGER     NOT NULL DEFAULT 0,
    used           INTEGER     NOT NULL DEFAULT 0,
    total          INTEGER     NOT NULL DEFAULT 0,
    resets_at      VARCHAR(64),
    CONSTRAINT ck_providers_ingestion_fields CHECK (
        (ingestion_mode = 'MANUAL' AND api_key IS NULL AND base_uri IS NULL AND resets_at IS NULL)
        OR
        (ingestion_mode = 'API' AND api_key IS NOT NULL AND base_uri IS NOT NULL AND resets_at IS NOT NULL)
    )
);

CREATE TABLE marketdata.listings (
    id                UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    symbol            VARCHAR(50)  NOT NULL,
    name              VARCHAR(255) NOT NULL,
    source            VARCHAR(20)  NOT NULL CHECK (source IN ('alpha_vantage', 'market_stack', 'brandnewday')),
    type              VARCHAR(50),
    ticker            VARCHAR(50),
    isin              VARCHAR(20),
    description       TEXT,
    exchange          VARCHAR(50),
    region            VARCHAR(50),
    currency          VARCHAR(3),
    provider          UUID REFERENCES marketdata.providers(id) ON DELETE SET NULL,
    active            BOOLEAN      NOT NULL DEFAULT TRUE,
    should_accumulate BOOLEAN      NOT NULL DEFAULT FALSE,
    syncing           BOOLEAN      NOT NULL DEFAULT FALSE,
    accumulated_start DATE,
    accumulated_end   DATE,
    created_at        TIMESTAMPTZ  NOT NULL DEFAULT CURRENT_TIMESTAMP,
    updated_at        TIMESTAMPTZ  NOT NULL DEFAULT CURRENT_TIMESTAMP,
    CONSTRAINT uq_listings_symbol_source UNIQUE (symbol, source)
);

-- ---------------------------------------------------------------------------
-- Imports: durable record of an uploaded file and its processing outcome.
-- type/source/listing_id are modelled on importer.Import; the store does not
-- yet persist them (see CHANGELOG), so they are nullable to stay compatible.
-- ---------------------------------------------------------------------------

CREATE TABLE import.imports (
    id          UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    vendor_id   UUID         NOT NULL REFERENCES vendor.vendors(id) ON DELETE CASCADE,
    account_id  UUID         REFERENCES account.accounts(id) ON DELETE SET NULL,
    listing_id  UUID         REFERENCES marketdata.listings(id) ON DELETE SET NULL,
    type        VARCHAR(16)  CHECK (type IS NULL OR type IN ('cashflow', 'portfolio', 'eod')),
    source      VARCHAR(64)  NOT NULL DEFAULT '',
    path        VARCHAR(512) NOT NULL,
    status      VARCHAR(16)  NOT NULL DEFAULT 'pending'
                    CHECK (status IN ('pending', 'processing', 'in_progress', 'completed', 'failed')),
    status_msg  TEXT         NOT NULL DEFAULT '',
    duplicates  INTEGER      NOT NULL DEFAULT 0,
    total_rows  INTEGER      NOT NULL DEFAULT 0,
    imported    INTEGER      NOT NULL DEFAULT 0,
    failed      INTEGER      NOT NULL DEFAULT 0,
    created_at  TIMESTAMPTZ  NOT NULL DEFAULT CURRENT_TIMESTAMP,
    updated_at  TIMESTAMPTZ  NOT NULL DEFAULT CURRENT_TIMESTAMP
);

-- ---------------------------------------------------------------------------
-- Cashflow: per-account projection + bank transactions.
-- ---------------------------------------------------------------------------

CREATE TABLE cashflow.accounts (
    id         UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    account_id UUID        NOT NULL UNIQUE REFERENCES account.accounts(id) ON DELETE CASCADE,
    created_at TIMESTAMPTZ NOT NULL DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMPTZ NOT NULL DEFAULT CURRENT_TIMESTAMP
);

CREATE TABLE cashflow.transactions (
    id           UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    account_id   UUID        NOT NULL REFERENCES cashflow.accounts(account_id) ON DELETE CASCADE,
    import_id    UUID        REFERENCES import.imports(id) ON DELETE CASCADE,
    description  TEXT        NOT NULL,
    note         TEXT        NOT NULL DEFAULT '',
    source       VARCHAR(64) NOT NULL,
    amount_cents BIGINT      NOT NULL,
    direction    VARCHAR(8)  NOT NULL CHECK (direction IN ('in', 'out')),
    date         DATE        NOT NULL,
    tag          VARCHAR(255) NOT NULL DEFAULT '',
    account_type VARCHAR(16) CHECK (account_type IS NULL OR account_type IN ('checking', 'savings', 'credit', 'brokerage')),
    ignored      BOOLEAN     NOT NULL DEFAULT FALSE,
    row_number   INTEGER     NOT NULL,
    checksum     VARCHAR(64) NOT NULL UNIQUE,
    created_at   TIMESTAMPTZ NOT NULL DEFAULT CURRENT_TIMESTAMP,
    updated_at   TIMESTAMPTZ NOT NULL DEFAULT CURRENT_TIMESTAMP
);

-- ---------------------------------------------------------------------------
-- Portfolio: per-account projection, transactions, positions, snapshots.
-- ---------------------------------------------------------------------------

CREATE TABLE portfolio.accounts (
    id         UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    account_id UUID        NOT NULL UNIQUE REFERENCES account.accounts(id) ON DELETE CASCADE,
    building   BOOLEAN     NOT NULL DEFAULT FALSE,
    created_at TIMESTAMPTZ NOT NULL DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMPTZ NOT NULL DEFAULT CURRENT_TIMESTAMP
);

CREATE TABLE portfolio.positions (
    id           UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    account_id   UUID             NOT NULL REFERENCES portfolio.accounts(account_id) ON DELETE CASCADE,
    listing_id   UUID             REFERENCES marketdata.listings(id) ON DELETE SET NULL,
    isin         VARCHAR(20),
    symbol       VARCHAR(50),
    open_date    TIMESTAMPTZ      NOT NULL,
    close_date   TIMESTAMPTZ,
    quantity     DOUBLE PRECISION NOT NULL DEFAULT 0,
    cost_basis   BIGINT           NOT NULL DEFAULT 0,
    fees         BIGINT           NOT NULL DEFAULT 0,
    income       BIGINT           NOT NULL DEFAULT 0,
    taxes        BIGINT           NOT NULL DEFAULT 0,
    realized_pnl BIGINT           NOT NULL DEFAULT 0,
    created_at   TIMESTAMPTZ      NOT NULL DEFAULT CURRENT_TIMESTAMP,
    updated_at   TIMESTAMPTZ      NOT NULL DEFAULT CURRENT_TIMESTAMP
);

CREATE TABLE portfolio.transactions (
    id           UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    account_id   UUID             REFERENCES portfolio.accounts(account_id) ON DELETE SET NULL,
    import_id    UUID             REFERENCES import.imports(id) ON DELETE CASCADE,
    position_id  UUID             REFERENCES portfolio.positions(id) ON DELETE SET NULL,
    origin       VARCHAR(8)       NOT NULL DEFAULT 'IMPORT' CHECK (origin IN ('IMPORT', 'MANUAL')),
    source       VARCHAR(64)      NOT NULL,
    occurred_at  DATE             NOT NULL,
    isin         VARCHAR(20),
    symbol       VARCHAR(50),
    description  TEXT             NOT NULL DEFAULT '',
    type         VARCHAR(16)      NOT NULL CHECK (type IN ('BUY', 'SELL', 'DIVIDEND', 'TAX', 'FEE', 'CASH')),
    quantity     DOUBLE PRECISION NOT NULL DEFAULT 0,
    unit_price   BIGINT           NOT NULL DEFAULT 0,
    amount_cents BIGINT           NOT NULL DEFAULT 0,
    checksum     VARCHAR(64)      NOT NULL UNIQUE,
    row_number   INTEGER          NOT NULL,
    created_at   TIMESTAMPTZ      NOT NULL DEFAULT CURRENT_TIMESTAMP,
    updated_at   TIMESTAMPTZ      NOT NULL DEFAULT CURRENT_TIMESTAMP,
    CONSTRAINT ck_portfolio_transactions_import_origin CHECK (
        (origin = 'IMPORT' AND import_id IS NOT NULL) OR
        (origin = 'MANUAL' AND import_id IS NULL)
    )
);

CREATE TABLE portfolio.position_snapshots (
    id                  UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    account_id          UUID             NOT NULL REFERENCES portfolio.accounts(account_id) ON DELETE CASCADE,
    position_id         UUID             NOT NULL REFERENCES portfolio.positions(id) ON DELETE CASCADE,
    listing_id          UUID             NOT NULL REFERENCES marketdata.listings(id) ON DELETE CASCADE,
    symbol              VARCHAR(50)      NOT NULL,
    name                VARCHAR(255),
    occurred_at         DATE             NOT NULL,
    quantity            DOUBLE PRECISION NOT NULL DEFAULT 0,
    unit_price          BIGINT           NOT NULL DEFAULT 0,
    average_price       BIGINT           NOT NULL DEFAULT 0,
    market_value        BIGINT           NOT NULL DEFAULT 0,
    cost_basis          BIGINT           NOT NULL DEFAULT 0,
    income              BIGINT           NOT NULL DEFAULT 0,
    fees                BIGINT           NOT NULL DEFAULT 0,
    taxes               BIGINT           NOT NULL DEFAULT 0,
    total_pnl           BIGINT           NOT NULL DEFAULT 0,
    total_pnl_pct       DOUBLE PRECISION NOT NULL DEFAULT 0,
    realized_pnl        BIGINT           NOT NULL DEFAULT 0,
    unrealized_pnl      BIGINT           NOT NULL DEFAULT 0,
    unrealized_pnl_pct  DOUBLE PRECISION NOT NULL DEFAULT 0,
    daily_delta_pnl     BIGINT           NOT NULL DEFAULT 0,
    daily_delta_pnl_pct DOUBLE PRECISION NOT NULL DEFAULT 0,
    created_at          TIMESTAMPTZ      NOT NULL DEFAULT CURRENT_TIMESTAMP,
    updated_at          TIMESTAMPTZ      NOT NULL DEFAULT CURRENT_TIMESTAMP,
    CONSTRAINT uq_position_snapshots_position_day UNIQUE (position_id, occurred_at)
);

CREATE TABLE portfolio.portfolio_snapshots (
    id                      UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    account_id              UUID             NOT NULL REFERENCES portfolio.accounts(account_id) ON DELETE CASCADE,
    occurred_at             DATE             NOT NULL,
    market_value            BIGINT           NOT NULL DEFAULT 0,
    cost_basis              BIGINT           NOT NULL DEFAULT 0,
    realized_pnl            BIGINT           NOT NULL DEFAULT 0,
    unrealized_pnl          BIGINT           NOT NULL DEFAULT 0,
    unrealized_pnl_pct      DOUBLE PRECISION NOT NULL DEFAULT 0,
    income                  BIGINT           NOT NULL DEFAULT 0,
    fees                    BIGINT           NOT NULL DEFAULT 0,
    taxes                   BIGINT           NOT NULL DEFAULT 0,
    cash_balance            BIGINT           NOT NULL DEFAULT 0,
    total_pnl               BIGINT           NOT NULL DEFAULT 0,
    total_pnl_pct           DOUBLE PRECISION NOT NULL DEFAULT 0,
    daily_delta_pnl         BIGINT           NOT NULL DEFAULT 0,
    daily_delta_pnl_pct     DOUBLE PRECISION NOT NULL DEFAULT 0,
    time_weighted_return_pct DOUBLE PRECISION NOT NULL DEFAULT 0,
    net_cashflow            BIGINT           NOT NULL DEFAULT 0,
    cumulative_net_cashflow BIGINT           NOT NULL DEFAULT 0,
    created_at              TIMESTAMPTZ      NOT NULL DEFAULT CURRENT_TIMESTAMP,
    updated_at              TIMESTAMPTZ      NOT NULL DEFAULT CURRENT_TIMESTAMP,
    CONSTRAINT uq_portfolio_snapshots_account_day UNIQUE (account_id, occurred_at)
);

-- ---------------------------------------------------------------------------
-- Market data: end-of-day prices + upload tracking.
-- ---------------------------------------------------------------------------

CREATE TABLE marketdata.eods (
    id          UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    listing_id  UUID        NOT NULL REFERENCES marketdata.listings(id) ON DELETE CASCADE,
    symbol      VARCHAR(50) NOT NULL,
    date        DATE        NOT NULL,
    open_cents  BIGINT      NOT NULL,
    high_cents  BIGINT      NOT NULL,
    low_cents   BIGINT      NOT NULL,
    close_cents BIGINT      NOT NULL,
    volume      BIGINT      NOT NULL,
    created_at  TIMESTAMPTZ NOT NULL DEFAULT CURRENT_TIMESTAMP,
    updated_at  TIMESTAMPTZ NOT NULL DEFAULT CURRENT_TIMESTAMP,
    CONSTRAINT uq_eods_listing_date UNIQUE (listing_id, date)
);

CREATE TABLE marketdata.eod_uploads (
    id                UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    listing_id        UUID         NOT NULL REFERENCES marketdata.listings(id) ON DELETE CASCADE,
    source            VARCHAR(20)  NOT NULL,
    status            VARCHAR(16)  NOT NULL CHECK (status IN ('PENDING', 'PROCESSING', 'SUCCEEDED', 'PARTIAL', 'FAILED')),
    stored_filename   VARCHAR(255) NOT NULL,
    original_filename VARCHAR(255) NOT NULL,
    status_message    TEXT         NOT NULL DEFAULT '',
    total_rows        INTEGER      NOT NULL DEFAULT 0,
    inserted_rows     INTEGER      NOT NULL DEFAULT 0,
    duplicate_rows    INTEGER      NOT NULL DEFAULT 0,
    error_rows        INTEGER      NOT NULL DEFAULT 0,
    row_errors_json   JSONB        NOT NULL DEFAULT '[]'::jsonb,
    started_at        TIMESTAMPTZ,
    finished_at       TIMESTAMPTZ,
    created_at        TIMESTAMPTZ  NOT NULL DEFAULT CURRENT_TIMESTAMP,
    updated_at        TIMESTAMPTZ  NOT NULL DEFAULT CURRENT_TIMESTAMP
);

-- ---------------------------------------------------------------------------
-- Assets: per-account projection, classes, items, mutations, snapshots.
-- ---------------------------------------------------------------------------

CREATE TABLE assets.accounts (
    id         UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    account_id UUID        NOT NULL UNIQUE REFERENCES account.accounts(id) ON DELETE CASCADE,
    created_at TIMESTAMPTZ NOT NULL DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMPTZ NOT NULL DEFAULT CURRENT_TIMESTAMP
);

CREATE TABLE assets.classes (
    id         UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    account_id UUID         NOT NULL REFERENCES assets.accounts(account_id) ON DELETE CASCADE,
    name       VARCHAR(255) NOT NULL,
    source     VARCHAR(16)  NOT NULL CHECK (source IN ('MANUAL', 'PORTFOLIO')),
    archived   BOOLEAN      NOT NULL DEFAULT FALSE,
    created_at TIMESTAMPTZ  NOT NULL DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMPTZ  NOT NULL DEFAULT CURRENT_TIMESTAMP,
    CONSTRAINT uq_asset_classes_account_name UNIQUE (account_id, name)
);

CREATE TABLE assets.items (
    id            UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    class_id      UUID         NOT NULL REFERENCES assets.classes(id) ON DELETE CASCADE,
    account_id    UUID         NOT NULL REFERENCES assets.accounts(account_id) ON DELETE CASCADE,
    name          VARCHAR(255) NOT NULL,
    current_worth BIGINT       NOT NULL DEFAULT 0,
    archived      BOOLEAN      NOT NULL DEFAULT FALSE,
    created_at    TIMESTAMPTZ  NOT NULL DEFAULT CURRENT_TIMESTAMP,
    updated_at    TIMESTAMPTZ  NOT NULL DEFAULT CURRENT_TIMESTAMP,
    CONSTRAINT uq_asset_items_class_name UNIQUE (class_id, name)
);

CREATE TABLE assets.mutations (
    id                UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    account_id        UUID        NOT NULL REFERENCES assets.accounts(account_id) ON DELETE CASCADE,
    class_id          UUID        NOT NULL REFERENCES assets.classes(id) ON DELETE CASCADE,
    item_id           UUID        NOT NULL REFERENCES assets.items(id) ON DELETE CASCADE,
    change_type       VARCHAR(16) NOT NULL CHECK (change_type IN ('SET', 'ADJUST')),
    direction         VARCHAR(16) CHECK (direction IS NULL OR direction IN ('INCREASE', 'DECREASE')),
    amount            BIGINT      NOT NULL,
    previous_worth    BIGINT      NOT NULL,
    new_worth         BIGINT      NOT NULL,
    class_total_worth BIGINT      NOT NULL,
    effective_date    DATE        NOT NULL,
    note              TEXT,
    created_at        TIMESTAMPTZ NOT NULL DEFAULT CURRENT_TIMESTAMP
);

CREATE TABLE assets.snapshots (
    id          UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    account_id  UUID        NOT NULL REFERENCES assets.accounts(account_id) ON DELETE CASCADE,
    occurred_at DATE        NOT NULL,
    total_worth BIGINT      NOT NULL,
    created_at  TIMESTAMPTZ NOT NULL DEFAULT CURRENT_TIMESTAMP,
    updated_at  TIMESTAMPTZ NOT NULL DEFAULT CURRENT_TIMESTAMP,
    CONSTRAINT uq_asset_snapshots_account_day UNIQUE (account_id, occurred_at)
);

-- ---------------------------------------------------------------------------
-- Indexes (beyond the implicit PK / UNIQUE indexes above).
-- Each is justified by a real query pattern in internal/storage.
-- ---------------------------------------------------------------------------

-- vendors: list active vendors ordered by name.
CREATE INDEX idx_vendors_active_name ON vendor.vendors (active, name);

-- imports: worker queue (oldest pending first) + vendor cascade.
CREATE INDEX idx_imports_status_created ON import.imports (status, created_at);
CREATE INDEX idx_imports_vendor_id ON import.imports (vendor_id);

-- cashflow transactions: account scans, import cascade, and the analytics path
-- (monthly aggregation / tag distribution over non-ignored rows by date).
CREATE INDEX idx_cashflow_transactions_account_date ON cashflow.transactions (account_id, date DESC);
CREATE INDEX idx_cashflow_transactions_import_id ON cashflow.transactions (import_id);
CREATE INDEX idx_cashflow_transactions_analytics ON cashflow.transactions (date, direction, tag) WHERE ignored = FALSE;

-- portfolio transactions: default listing sort (account + occurred/created/id desc),
-- position history, and import cascade.
CREATE INDEX idx_portfolio_transactions_account_listing ON portfolio.transactions (account_id, occurred_at DESC, created_at DESC, id DESC);
CREATE INDEX idx_portfolio_transactions_position_id ON portfolio.transactions (position_id, occurred_at);
CREATE INDEX idx_portfolio_transactions_import_id ON portfolio.transactions (import_id);

-- positions: account listings and listing cascade.
CREATE INDEX idx_portfolio_positions_account_open ON portfolio.positions (account_id, open_date);
CREATE INDEX idx_portfolio_positions_listing_id ON portfolio.positions (listing_id);

-- position snapshots: account+day listings and listing cascade.
-- (latest-per-position lookups are served by uq_position_snapshots_position_day.)
CREATE INDEX idx_position_snapshots_account_day ON portfolio.position_snapshots (account_id, occurred_at);
CREATE INDEX idx_position_snapshots_listing_id ON portfolio.position_snapshots (listing_id);

-- eod uploads: pending queue ordering + listing cascade.
CREATE INDEX idx_eod_uploads_status_created ON marketdata.eod_uploads (status, created_at);
CREATE INDEX idx_eod_uploads_listing_id ON marketdata.eod_uploads (listing_id);

-- providers: name+mode is the lookup/ranking key for token selection.
CREATE INDEX idx_providers_name_mode ON marketdata.providers (name, ingestion_mode);

-- asset classes: source lookup (account_id+name unique already serves name scans).
CREATE INDEX idx_asset_classes_account_source ON assets.classes (account_id, source);

-- asset items: per-class/account aggregations filtered by archived.
CREATE INDEX idx_asset_items_account_class_archived ON assets.items (account_id, class_id, archived);

-- asset mutations: per-class and per-account mutations (newest first) + item cascade.
CREATE INDEX idx_asset_mutations_account_class_date ON assets.mutations (account_id, class_id, effective_date DESC);
CREATE INDEX idx_asset_mutations_account_date ON assets.mutations (account_id, effective_date DESC);
CREATE INDEX idx_asset_mutations_item_id ON assets.mutations (item_id);

-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin

DROP TABLE IF EXISTS assets.snapshots;
DROP TABLE IF EXISTS assets.mutations;
DROP TABLE IF EXISTS assets.items;
DROP TABLE IF EXISTS assets.classes;
DROP TABLE IF EXISTS assets.accounts;
DROP TABLE IF EXISTS marketdata.eod_uploads;
DROP TABLE IF EXISTS marketdata.eods;
DROP TABLE IF EXISTS portfolio.portfolio_snapshots;
DROP TABLE IF EXISTS portfolio.position_snapshots;
DROP TABLE IF EXISTS portfolio.transactions;
DROP TABLE IF EXISTS portfolio.positions;
DROP TABLE IF EXISTS portfolio.accounts;
DROP TABLE IF EXISTS cashflow.transactions;
DROP TABLE IF EXISTS cashflow.accounts;
DROP TABLE IF EXISTS import.imports;
DROP TABLE IF EXISTS marketdata.listings;
DROP TABLE IF EXISTS marketdata.providers;
DROP TABLE IF EXISTS account.accounts;
DROP TABLE IF EXISTS vendor.vendors;

DROP SCHEMA IF EXISTS assets;
DROP SCHEMA IF EXISTS portfolio;
DROP SCHEMA IF EXISTS marketdata;
DROP SCHEMA IF EXISTS cashflow;
DROP SCHEMA IF EXISTS import;
DROP SCHEMA IF EXISTS account;
DROP SCHEMA IF EXISTS vendor;

-- +goose StatementEnd
