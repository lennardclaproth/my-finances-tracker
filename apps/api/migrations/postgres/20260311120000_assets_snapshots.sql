-- +goose Up
-- +goose StatementBegin

CREATE TABLE IF NOT EXISTS assets.snapshots (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    account_id UUID NOT NULL REFERENCES assets.accounts(account_id) ON DELETE CASCADE,
    occurred_at DATE NOT NULL,
    total_worth BIGINT NOT NULL,
    created_at TIMESTAMPTZ NOT NULL DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMPTZ NOT NULL DEFAULT CURRENT_TIMESTAMP,
    CONSTRAINT uq_asset_snapshots_account_day UNIQUE (account_id, occurred_at)
);

CREATE INDEX IF NOT EXISTS idx_asset_snapshots_account_day ON assets.snapshots(account_id, occurred_at);

-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin
DROP TABLE IF EXISTS assets.snapshots;
-- +goose StatementEnd
