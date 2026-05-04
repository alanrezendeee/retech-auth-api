-- +goose Up
-- Move tenant_id from users to user_applications (correct multi-tenant model)
-- tenant_id on users was global per user; on user_applications it is per user per app
ALTER TABLE users DROP COLUMN IF EXISTS tenant_id;

-- +goose Down
ALTER TABLE users ADD COLUMN IF NOT EXISTS tenant_id VARCHAR(255);
