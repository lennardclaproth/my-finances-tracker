-- +goose Up
-- +goose StatementBegin

CREATE TABLE IF NOT EXISTS assets_accounts (
    id TEXT PRIMARY KEY,
    account_id TEXT NOT NULL UNIQUE REFERENCES accounts(id) ON DELETE CASCADE,
    created_at DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP,
    updated_at DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP
);

CREATE TABLE IF NOT EXISTS asset_classes (
    id TEXT PRIMARY KEY,
    account_id TEXT NOT NULL REFERENCES assets_accounts(account_id) ON DELETE CASCADE,
    name TEXT NOT NULL,
    source TEXT NOT NULL CHECK (source IN ('MANUAL', 'PORTFOLIO')),
    archived INTEGER NOT NULL DEFAULT 0,
    created_at DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP,
    updated_at DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP,
    CONSTRAINT uq_asset_classes_account_name UNIQUE (account_id, name)
);

CREATE TABLE IF NOT EXISTS asset_items (
    id TEXT PRIMARY KEY,
    class_id TEXT NOT NULL REFERENCES asset_classes(id) ON DELETE CASCADE,
    account_id TEXT NOT NULL REFERENCES assets_accounts(account_id) ON DELETE CASCADE,
    name TEXT NOT NULL,
    current_worth BIGINT NOT NULL DEFAULT 0,
    archived INTEGER NOT NULL DEFAULT 0,
    created_at DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP,
    updated_at DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP,
    CONSTRAINT uq_asset_items_class_name UNIQUE (class_id, name)
);

CREATE TABLE IF NOT EXISTS asset_histories (
    id TEXT PRIMARY KEY,
    account_id TEXT NOT NULL REFERENCES assets_accounts(account_id) ON DELETE CASCADE,
    class_id TEXT NOT NULL REFERENCES asset_classes(id) ON DELETE CASCADE,
    item_id TEXT NOT NULL REFERENCES asset_items(id) ON DELETE CASCADE,
    change_type TEXT NOT NULL CHECK (change_type IN ('SET', 'ADJUST')),
    direction TEXT CHECK (direction IN ('INCREASE', 'DECREASE')),
    amount BIGINT NOT NULL,
    previous_worth BIGINT NOT NULL,
    new_worth BIGINT NOT NULL,
    class_total_worth BIGINT NOT NULL,
    effective_date DATE NOT NULL,
    note TEXT NOT NULL DEFAULT '',
    created_at DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP
);

CREATE INDEX IF NOT EXISTS idx_assets_accounts_account_id ON assets_accounts(account_id);
CREATE INDEX IF NOT EXISTS idx_asset_classes_account_id ON asset_classes(account_id);
CREATE INDEX IF NOT EXISTS idx_asset_classes_source ON asset_classes(source);
CREATE INDEX IF NOT EXISTS idx_asset_items_class_id ON asset_items(class_id);
CREATE INDEX IF NOT EXISTS idx_asset_items_account_id ON asset_items(account_id);
CREATE INDEX IF NOT EXISTS idx_asset_histories_account_id ON asset_histories(account_id);
CREATE INDEX IF NOT EXISTS idx_asset_histories_class_id ON asset_histories(class_id);
CREATE INDEX IF NOT EXISTS idx_asset_histories_item_id ON asset_histories(item_id);
CREATE INDEX IF NOT EXISTS idx_asset_histories_effective_date ON asset_histories(effective_date);

-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin
DROP TABLE IF EXISTS asset_histories;
DROP TABLE IF EXISTS asset_items;
DROP TABLE IF EXISTS asset_classes;
DROP TABLE IF EXISTS assets_accounts;
-- +goose StatementEnd
