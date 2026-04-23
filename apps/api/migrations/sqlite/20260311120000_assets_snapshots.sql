-- +goose Up
-- +goose StatementBegin

CREATE TABLE IF NOT EXISTS asset_snapshots (
    id TEXT PRIMARY KEY,
    account_id TEXT NOT NULL REFERENCES assets_accounts(account_id) ON DELETE CASCADE,
    occurred_at DATE NOT NULL,
    total_worth BIGINT NOT NULL,
    created_at DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP,
    updated_at DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP,
    CONSTRAINT uq_asset_snapshots_account_day UNIQUE (account_id, occurred_at)
);

CREATE INDEX IF NOT EXISTS idx_asset_snapshots_account_day ON asset_snapshots(account_id, occurred_at);

-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin
DROP TABLE IF EXISTS asset_snapshots;
-- +goose StatementEnd
