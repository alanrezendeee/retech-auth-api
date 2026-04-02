-- +goose Up
-- Corrigir constraints para permissions e roles para que ON CONFLICT funcione

-- ============================================
-- PERMISSIONS: Converter INDEX para CONSTRAINT
-- ============================================
-- Remove o índice único se existir
DROP INDEX IF EXISTS idx_permissions_application_code;

-- Remove constraint antiga se existir
DO $$
DECLARE
    constraint_name text;
BEGIN
    -- Encontra constraint UNIQUE em (application_id, code) para permissions
    SELECT conname INTO constraint_name
    FROM pg_constraint
    WHERE conrelid = 'permissions'::regclass
      AND contype = 'u'
      AND array_length(conkey, 1) = 2
      AND conkey[1] = (SELECT attnum FROM pg_attribute WHERE attrelid = 'permissions'::regclass AND attname = 'application_id')
      AND conkey[2] = (SELECT attnum FROM pg_attribute WHERE attrelid = 'permissions'::regclass AND attname = 'code');
    
    IF constraint_name IS NOT NULL THEN
        EXECUTE 'ALTER TABLE permissions DROP CONSTRAINT ' || quote_ident(constraint_name);
    END IF;
END $$;

-- Cria constraint explícita para permissions
ALTER TABLE permissions ADD CONSTRAINT permissions_application_id_code_key UNIQUE (application_id, code);

-- ============================================
-- ROLES: Garantir CONSTRAINT existe
-- ============================================
-- Remove constraint antiga se existir
DO $$
DECLARE
    constraint_name text;
BEGIN
    -- Encontra constraint UNIQUE em (application_id, code) para roles
    SELECT conname INTO constraint_name
    FROM pg_constraint
    WHERE conrelid = 'roles'::regclass
      AND contype = 'u'
      AND array_length(conkey, 1) = 2
      AND conkey[1] = (SELECT attnum FROM pg_attribute WHERE attrelid = 'roles'::regclass AND attname = 'application_id')
      AND conkey[2] = (SELECT attnum FROM pg_attribute WHERE attrelid = 'roles'::regclass AND attname = 'code');
    
    IF constraint_name IS NOT NULL AND constraint_name != 'roles_application_id_code_key' THEN
        EXECUTE 'ALTER TABLE roles DROP CONSTRAINT ' || quote_ident(constraint_name);
    END IF;
END $$;

-- Cria constraint explícita para roles (se não existir)
DO $$
BEGIN
    IF NOT EXISTS (
        SELECT 1 FROM pg_constraint 
        WHERE conname = 'roles_application_id_code_key' 
        AND conrelid = 'roles'::regclass
    ) THEN
        ALTER TABLE roles ADD CONSTRAINT roles_application_id_code_key UNIQUE (application_id, code);
    END IF;
END $$;

-- +goose Down
-- Remove as constraints criadas
ALTER TABLE permissions DROP CONSTRAINT IF EXISTS permissions_application_id_code_key;
ALTER TABLE roles DROP CONSTRAINT IF EXISTS roles_application_id_code_key;
-- Recria o índice único (para compatibilidade reversa)
CREATE UNIQUE INDEX IF NOT EXISTS idx_permissions_application_code ON permissions(application_id, code) WHERE code IS NOT NULL;

