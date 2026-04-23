-- +goose Up
-- +goose StatementBegin

CREATE SCHEMA IF NOT EXISTS assets;

CREATE TABLE IF NOT EXISTS assets.accounts (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    account_id UUID NOT NULL UNIQUE REFERENCES account.accounts(id) ON DELETE CASCADE,
    created_at TIMESTAMPTZ NOT NULL DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMPTZ NOT NULL DEFAULT CURRENT_TIMESTAMP
);

CREATE TABLE IF NOT EXISTS assets.classes (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    account_id UUID NOT NULL REFERENCES assets.accounts(account_id) ON DELETE CASCADE,
    name TEXT NOT NULL,
    source TEXT NOT NULL CHECK (source IN ('MANUAL', 'PORTFOLIO')),
    archived BOOLEAN NOT NULL DEFAULT FALSE,
    created_at TIMESTAMPTZ NOT NULL DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMPTZ NOT NULL DEFAULT CURRENT_TIMESTAMP,
    CONSTRAINT uq_asset_classes_account_name UNIQUE (account_id, name)
);

CREATE TABLE IF NOT EXISTS assets.items (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    class_id UUID NOT NULL REFERENCES assets.classes(id) ON DELETE CASCADE,
    account_id UUID NOT NULL REFERENCES assets.accounts(account_id) ON DELETE CASCADE,
    name TEXT NOT NULL,
    current_worth BIGINT NOT NULL DEFAULT 0,
    archived BOOLEAN NOT NULL DEFAULT FALSE,
    created_at TIMESTAMPTZ NOT NULL DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMPTZ NOT NULL DEFAULT CURRENT_TIMESTAMP,
    CONSTRAINT uq_asset_items_class_name UNIQUE (class_id, name)
);

CREATE TABLE IF NOT EXISTS assets.histories (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    account_id UUID NOT NULL REFERENCES assets.accounts(account_id) ON DELETE CASCADE,
    class_id UUID NOT NULL REFERENCES assets.classes(id) ON DELETE CASCADE,
    item_id UUID NOT NULL REFERENCES assets.items(id) ON DELETE CASCADE,
    change_type TEXT NOT NULL CHECK (change_type IN ('SET', 'ADJUST')),
    direction TEXT CHECK (direction IN ('INCREASE', 'DECREASE')),
    amount BIGINT NOT NULL,
    previous_worth BIGINT NOT NULL,
    new_worth BIGINT NOT NULL,
    class_total_worth BIGINT NOT NULL,
    effective_date DATE NOT NULL,
    note TEXT NOT NULL DEFAULT '',
    created_at TIMESTAMPTZ NOT NULL DEFAULT CURRENT_TIMESTAMP
);

CREATE INDEX IF NOT EXISTS idx_assets_accounts_account_id ON assets.accounts(account_id);
CREATE INDEX IF NOT EXISTS idx_asset_classes_account_id ON assets.classes(account_id);
CREATE INDEX IF NOT EXISTS idx_asset_classes_source ON assets.classes(source);
CREATE INDEX IF NOT EXISTS idx_asset_items_class_id ON assets.items(class_id);
CREATE INDEX IF NOT EXISTS idx_asset_items_account_id ON assets.items(account_id);
CREATE INDEX IF NOT EXISTS idx_asset_histories_account_id ON assets.histories(account_id);
CREATE INDEX IF NOT EXISTS idx_asset_histories_class_id ON assets.histories(class_id);
CREATE INDEX IF NOT EXISTS idx_asset_histories_item_id ON assets.histories(item_id);
CREATE INDEX IF NOT EXISTS idx_asset_histories_effective_date ON assets.histories(effective_date);

-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin
DROP TABLE IF EXISTS assets.histories;
DROP TABLE IF EXISTS assets.items;
DROP TABLE IF EXISTS assets.classes;
DROP TABLE IF EXISTS assets.accounts;
DROP SCHEMA IF EXISTS assets;
-- +goose StatementEnd
