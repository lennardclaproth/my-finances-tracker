-- +goose Up
-- +goose StatementBegin
ALTER TABLE vendor.vendors
    ADD COLUMN IF NOT EXISTS import_disabled BOOLEAN NOT NULL DEFAULT FALSE;

ALTER TABLE portfolio.transactions
    ADD COLUMN IF NOT EXISTS origin TEXT NOT NULL DEFAULT 'IMPORT';

DO $$
BEGIN
    IF NOT EXISTS (
        SELECT 1
        FROM pg_constraint
        WHERE conname = 'ck_portfolio_transactions_origin'
    ) THEN
        ALTER TABLE portfolio.transactions
            ADD CONSTRAINT ck_portfolio_transactions_origin
            CHECK (origin IN ('IMPORT', 'MANUAL'));
    END IF;
END;
$$;

ALTER TABLE portfolio.transactions
    ALTER COLUMN import_id DROP NOT NULL;

DO $$
BEGIN
    IF NOT EXISTS (
        SELECT 1
        FROM pg_constraint
        WHERE conname = 'ck_portfolio_transactions_import_origin'
    ) THEN
        ALTER TABLE portfolio.transactions
            ADD CONSTRAINT ck_portfolio_transactions_import_origin
            CHECK (
                (origin = 'IMPORT' AND import_id IS NOT NULL) OR
                (origin = 'MANUAL' AND import_id IS NULL)
            );
    END IF;
END;
$$;
-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin
DELETE FROM portfolio.transactions WHERE origin = 'MANUAL';

ALTER TABLE portfolio.transactions
    DROP CONSTRAINT IF EXISTS ck_portfolio_transactions_import_origin;

ALTER TABLE portfolio.transactions
    DROP CONSTRAINT IF EXISTS ck_portfolio_transactions_origin;

ALTER TABLE portfolio.transactions
    ALTER COLUMN import_id SET NOT NULL;

ALTER TABLE portfolio.transactions
    DROP COLUMN IF EXISTS origin;

ALTER TABLE vendor.vendors
    DROP COLUMN IF EXISTS import_disabled;
-- +goose StatementEnd
