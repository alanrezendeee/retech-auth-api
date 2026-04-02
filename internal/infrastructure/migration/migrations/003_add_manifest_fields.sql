-- +goose Up
-- Adiciona campo 'code' em permissions (identificador único por aplicação)
ALTER TABLE permissions ADD COLUMN IF NOT EXISTS code VARCHAR(255);
-- Adiciona índice único para (application_id, code) se não existir
CREATE UNIQUE INDEX IF NOT EXISTS idx_permissions_application_code ON permissions(application_id, code) WHERE code IS NOT NULL;
-- Migra dados existentes: se não tem code, cria a partir de subject.action
UPDATE permissions SET code = subject || '.' || action WHERE code IS NULL;

-- Adiciona campo 'system' em roles (indica se é role base do sistema)
ALTER TABLE roles ADD COLUMN IF NOT EXISTS system BOOLEAN NOT NULL DEFAULT false;
-- Atualiza role master para system = true
UPDATE roles SET system = true WHERE code = 'master';

-- +goose Down
-- Remove campo system
ALTER TABLE roles DROP COLUMN IF EXISTS system;
-- Remove índice e campo code
DROP INDEX IF EXISTS idx_permissions_application_code;
ALTER TABLE permissions DROP COLUMN IF EXISTS code;

