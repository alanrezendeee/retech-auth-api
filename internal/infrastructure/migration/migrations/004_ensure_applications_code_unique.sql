-- +goose Up
-- Garantir que a constraint UNIQUE em applications.code existe explicitamente
-- Remove constraint antiga se existir (pode ter nome diferente gerado pelo PostgreSQL)
DO $$
DECLARE
    constraint_name text;
BEGIN
    -- Encontra o nome da constraint UNIQUE na coluna code
    SELECT conname INTO constraint_name
    FROM pg_constraint
    WHERE conrelid = 'applications'::regclass
      AND contype = 'u'
      AND array_length(conkey, 1) = 1
      AND conkey[1] = (SELECT attnum FROM pg_attribute WHERE attrelid = 'applications'::regclass AND attname = 'code');
    
    IF constraint_name IS NOT NULL THEN
        EXECUTE 'ALTER TABLE applications DROP CONSTRAINT ' || quote_ident(constraint_name);
    END IF;
END $$;
-- Cria a constraint com nome explícito
ALTER TABLE applications ADD CONSTRAINT applications_code_key UNIQUE (code);

-- +goose Down
-- Remove a constraint se existir
ALTER TABLE applications DROP CONSTRAINT IF EXISTS applications_code_key;

