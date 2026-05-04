-- +goose Up
CREATE TABLE IF NOT EXISTS users (
	id UUID PRIMARY KEY,
	email VARCHAR(255) NOT NULL UNIQUE,
	password VARCHAR(255) NOT NULL,
	name VARCHAR(255) NOT NULL,
	active BOOLEAN NOT NULL DEFAULT true,
	created_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
	updated_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP
);

CREATE TABLE IF NOT EXISTS applications (
	id UUID PRIMARY KEY,
	name VARCHAR(255) NOT NULL,
	code VARCHAR(100) NOT NULL UNIQUE,
	description TEXT,
	active BOOLEAN NOT NULL DEFAULT true,
	created_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
	updated_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP
);

CREATE TABLE IF NOT EXISTS user_applications (
	id UUID PRIMARY KEY,
	user_id UUID NOT NULL REFERENCES users(id) ON DELETE CASCADE,
	application_id UUID NOT NULL REFERENCES applications(id) ON DELETE CASCADE,
	tenant_id VARCHAR(255),
	active BOOLEAN NOT NULL DEFAULT true,
	created_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
	updated_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
	UNIQUE(user_id, application_id)
);

CREATE TABLE IF NOT EXISTS roles (
	id UUID PRIMARY KEY,
	application_id UUID NOT NULL REFERENCES applications(id) ON DELETE CASCADE,
	name VARCHAR(255) NOT NULL,
	code VARCHAR(100) NOT NULL,
	description TEXT,
	active BOOLEAN NOT NULL DEFAULT true,
	created_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
	updated_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
	UNIQUE(application_id, code)
);

CREATE TABLE IF NOT EXISTS permissions (
	id UUID PRIMARY KEY,
	application_id UUID NOT NULL REFERENCES applications(id) ON DELETE CASCADE,
	subject VARCHAR(255) NOT NULL,
	action VARCHAR(100) NOT NULL,
	conditions TEXT,
	description TEXT,
	active BOOLEAN NOT NULL DEFAULT true,
	created_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
	updated_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
	UNIQUE(application_id, subject, action)
);

CREATE TABLE IF NOT EXISTS role_permissions (
	id UUID PRIMARY KEY,
	role_id UUID NOT NULL REFERENCES roles(id) ON DELETE CASCADE,
	permission_id UUID NOT NULL REFERENCES permissions(id) ON DELETE CASCADE,
	active BOOLEAN NOT NULL DEFAULT true,
	created_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
	updated_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
	UNIQUE(role_id, permission_id)
);

CREATE TABLE IF NOT EXISTS user_roles (
	id UUID PRIMARY KEY,
	user_application_id UUID NOT NULL REFERENCES user_applications(id) ON DELETE CASCADE,
	role_id UUID NOT NULL REFERENCES roles(id) ON DELETE CASCADE,
	active BOOLEAN NOT NULL DEFAULT true,
	created_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
	updated_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
	UNIQUE(user_application_id, role_id)
);

CREATE INDEX IF NOT EXISTS idx_users_email ON users(email);
CREATE INDEX IF NOT EXISTS idx_applications_code ON applications(code);
CREATE INDEX IF NOT EXISTS idx_user_applications_user_id ON user_applications(user_id);
CREATE INDEX IF NOT EXISTS idx_user_applications_application_id ON user_applications(application_id);
CREATE INDEX IF NOT EXISTS idx_user_applications_tenant_id ON user_applications(tenant_id) WHERE tenant_id IS NOT NULL;
CREATE INDEX IF NOT EXISTS idx_roles_application_id ON roles(application_id);
CREATE INDEX IF NOT EXISTS idx_permissions_application_id ON permissions(application_id);
CREATE INDEX IF NOT EXISTS idx_user_roles_user_application_id ON user_roles(user_application_id);
CREATE INDEX IF NOT EXISTS idx_user_roles_role_id ON user_roles(role_id);
CREATE INDEX IF NOT EXISTS idx_role_permissions_role_id ON role_permissions(role_id);
CREATE INDEX IF NOT EXISTS idx_role_permissions_permission_id ON role_permissions(permission_id);

-- +goose Down
DROP TABLE IF EXISTS user_roles CASCADE;
DROP TABLE IF EXISTS role_permissions CASCADE;
DROP TABLE IF EXISTS permissions CASCADE;
DROP TABLE IF EXISTS roles CASCADE;
DROP TABLE IF EXISTS user_applications CASCADE;
DROP TABLE IF EXISTS applications CASCADE;
DROP TABLE IF EXISTS users CASCADE;

