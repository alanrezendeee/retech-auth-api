-- +goose Up
-- Adiciona campo tenant_id na tabela users para suporte a multitenancy por unidade
-- O tenant_id é definido pelo sistema de gestão de usuários/onboarding
-- O AUTH apenas armazena, carrega e inclui no token JWT
ALTER TABLE users ADD COLUMN IF NOT EXISTS tenant_id VARCHAR(255);

-- Índice para melhorar performance de queries por tenant
CREATE INDEX IF NOT EXISTS idx_users_tenant_id ON users(tenant_id) WHERE tenant_id IS NOT NULL;

-- +goose Down
DROP INDEX IF EXISTS idx_users_tenant_id;
ALTER TABLE users DROP COLUMN IF EXISTS tenant_id;

