package repository

import (
	"context"
	"database/sql"
	"errors"
	"fmt"

	"github.com/google/uuid"
	"github.com/theretech/retech-auth-api/internal/domain/entity"
	"github.com/theretech/retech-auth-api/internal/domain/repository"
)

type postgresUserRepository struct {
	db *sql.DB
}

// NewPostgresUserRepository cria uma nova instância de PostgresUserRepository
func NewPostgresUserRepository(db *sql.DB) repository.UserRepository {
	return &postgresUserRepository{db: db}
}

func (r *postgresUserRepository) Create(ctx context.Context, user *entity.User) error {
	query := `
		INSERT INTO users (id, email, password, name, tenant_id, active, version, created_at, updated_at)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9)
	`
	_, err := r.db.ExecContext(
		ctx, query,
		user.ID, user.Email, user.Password, user.Name, user.TenantID, user.Active, user.Version,
		user.CreatedAt, user.UpdatedAt,
	)
	return err
}

func (r *postgresUserRepository) FindByID(ctx context.Context, id uuid.UUID) (*entity.User, error) {
	query := `
		SELECT id, email, password, name, tenant_id, active, COALESCE(version, 1) as version, created_at, updated_at
		FROM users
		WHERE id = $1
	`
	user := &entity.User{}
	err := r.db.QueryRowContext(ctx, query, id).Scan(
		&user.ID, &user.Email, &user.Password, &user.Name, &user.TenantID, &user.Active,
		&user.Version, &user.CreatedAt, &user.UpdatedAt,
	)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, errors.New("usuário não encontrado")
		}
		return nil, err
	}
	return user, nil
}

func (r *postgresUserRepository) FindByEmail(ctx context.Context, email string) (*entity.User, error) {
	query := `
		SELECT id, email, password, name, tenant_id, active, COALESCE(version, 1) as version, created_at, updated_at
		FROM users
		WHERE email = $1
	`
	user := &entity.User{}
	err := r.db.QueryRowContext(ctx, query, email).Scan(
		&user.ID, &user.Email, &user.Password, &user.Name, &user.TenantID, &user.Active,
		&user.Version, &user.CreatedAt, &user.UpdatedAt,
	)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, errors.New("usuário não encontrado")
		}
		return nil, err
	}
	return user, nil
}

func (r *postgresUserRepository) Update(ctx context.Context, user *entity.User) error {
	query := `
		UPDATE users
		SET email = $2, password = $3, name = $4, tenant_id = $5, active = $6, version = COALESCE(version, 1) + 1, updated_at = $7
		WHERE id = $1
	`
	_, err := r.db.ExecContext(
		ctx, query,
		user.ID, user.Email, user.Password, user.Name, user.TenantID, user.Active, user.UpdatedAt,
	)
	return err
}

func (r *postgresUserRepository) UpdateWithVersion(ctx context.Context, user *entity.User, expectedVersion int) error {
	query := `
		UPDATE users
		SET email = $2, password = $3, name = $4, tenant_id = $5, active = $6, version = COALESCE(version, 1) + 1, updated_at = $7
		WHERE id = $1 AND COALESCE(version, 1) = $8
	`
	result, err := r.db.ExecContext(
		ctx, query,
		user.ID, user.Email, user.Password, user.Name, user.TenantID, user.Active, user.UpdatedAt, expectedVersion,
	)
	if err != nil {
		return err
	}
	rows, err := result.RowsAffected()
	if err != nil {
		return err
	}
	if rows == 0 {
		return errors.New("versão do usuário está desatualizada (409 Conflict)")
	}
	return nil
}

func (r *postgresUserRepository) UpdateStatus(ctx context.Context, userID uuid.UUID, active bool, expectedVersion int) error {
	query := `
		UPDATE users
		SET active = $2, version = COALESCE(version, 1) + 1, updated_at = NOW()
		WHERE id = $1 AND COALESCE(version, 1) = $3
	`
	result, err := r.db.ExecContext(ctx, query, userID, active, expectedVersion)
	if err != nil {
		return err
	}
	rows, err := result.RowsAffected()
	if err != nil {
		return err
	}
	if rows == 0 {
		return errors.New("versão do usuário está desatualizada (409 Conflict)")
	}
	return nil
}

func (r *postgresUserRepository) UpdateUserApplicationStatus(ctx context.Context, userID, applicationID uuid.UUID, active bool, expectedVersion int) error {
	query := `
		UPDATE user_applications
		SET active = $3, updated_at = NOW()
		WHERE user_id = $1 
			AND application_id = $2
			AND EXISTS (
				SELECT 1 FROM users 
				WHERE id = $1 
				AND COALESCE(version, 1) = $4
			)
	`
	result, err := r.db.ExecContext(ctx, query, userID, applicationID, active, expectedVersion)
	if err != nil {
		return err
	}
	rows, err := result.RowsAffected()
	if err != nil {
		return err
	}
	if rows == 0 {
		user, _ := r.FindByID(ctx, userID)
		if user != nil && user.Version != expectedVersion {
			return errors.New("versão do usuário está desatualizada (409 Conflict)")
		}
		return errors.New("vínculo usuário-aplicação não encontrado")
	}

	query = `
		UPDATE users
		SET version = COALESCE(version, 1) + 1, updated_at = NOW()
		WHERE id = $1
	`
	_, err = r.db.ExecContext(ctx, query, userID)
	return err
}

func (r *postgresUserRepository) Delete(ctx context.Context, id uuid.UUID) error {
	query := `DELETE FROM users WHERE id = $1`
	_, err := r.db.ExecContext(ctx, query, id)
	return err
}

func (r *postgresUserRepository) List(ctx context.Context, limit, offset int) ([]*entity.User, error) {
	query := `
		SELECT id, email, password, name, tenant_id, active, COALESCE(version, 1) as version, created_at, updated_at
		FROM users
		ORDER BY created_at DESC
		LIMIT $1 OFFSET $2
	`
	rows, err := r.db.QueryContext(ctx, query, limit, offset)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var users []*entity.User
	for rows.Next() {
		user := &entity.User{}
		if err := rows.Scan(
			&user.ID, &user.Email, &user.Password, &user.Name, &user.TenantID, &user.Active,
			&user.Version, &user.CreatedAt, &user.UpdatedAt,
		); err != nil {
			return nil, err
		}
		users = append(users, user)
	}

	return users, rows.Err()
}

func (r *postgresUserRepository) FindByIDAndApplication(ctx context.Context, userID, applicationID uuid.UUID) (*entity.User, error) {
	query := `
		SELECT u.id, u.email, u.password, u.name, u.tenant_id, ua.active, COALESCE(u.version, 1) as version, u.created_at, u.updated_at
		FROM users u
		INNER JOIN user_applications ua ON u.id = ua.user_id
		WHERE u.id = $1 AND ua.application_id = $2
	`
	user := &entity.User{}
	err := r.db.QueryRowContext(ctx, query, userID, applicationID).Scan(
		&user.ID, &user.Email, &user.Password, &user.Name, &user.TenantID, &user.Active,
		&user.Version, &user.CreatedAt, &user.UpdatedAt,
	)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, errors.New("usuário não encontrado nesta aplicação")
		}
		return nil, err
	}
	return user, nil
}

func (r *postgresUserRepository) UpdatePassword(ctx context.Context, userID uuid.UUID, passwordHash string) error {
	query := `
		UPDATE users
		SET password = $2, updated_at = NOW()
		WHERE id = $1
	`
	result, err := r.db.ExecContext(ctx, query, userID, passwordHash)
	if err != nil {
		return err
	}

	rows, err := result.RowsAffected()
	if err != nil {
		return err
	}
	if rows == 0 {
		return errors.New("usuário não encontrado")
	}

	return nil
}

func (r *postgresUserRepository) UpdatePasswordWithVersion(ctx context.Context, userID uuid.UUID, passwordHash string, expectedVersion int) error {
	query := `
		UPDATE users
		SET password = $2, version = COALESCE(version, 1) + 1, updated_at = NOW()
		WHERE id = $1 AND COALESCE(version, 1) = $3
	`
	result, err := r.db.ExecContext(ctx, query, userID, passwordHash, expectedVersion)
	if err != nil {
		return err
	}

	rows, err := result.RowsAffected()
	if err != nil {
		return err
	}
	if rows == 0 {
		user, _ := r.FindByID(ctx, userID)
		if user != nil && user.Version != expectedVersion {
			return errors.New("versão do usuário está desatualizada (409 Conflict)")
		}
		return errors.New("usuário não encontrado")
	}

	return nil
}

func (r *postgresUserRepository) SoftDeleteFromApplication(ctx context.Context, userID, applicationID uuid.UUID) error {
	query := `
		UPDATE user_applications
		SET active = false, updated_at = NOW()
		WHERE user_id = $1 AND application_id = $2
	`
	result, err := r.db.ExecContext(ctx, query, userID, applicationID)
	if err != nil {
		return err
	}

	rows, err := result.RowsAffected()
	if err != nil {
		return err
	}
	if rows == 0 {
		return errors.New("vínculo usuário-aplicação não encontrado")
	}

	return nil
}

func (r *postgresUserRepository) ListByApplication(ctx context.Context, filters repository.UserFilters) ([]*entity.User, int, error) {
	// Query base
	baseQuery := `
		FROM users u
		INNER JOIN user_applications ua ON u.id = ua.user_id
	`

	// Adiciona JOIN com roles se filtrar por role
	if filters.RoleCode != "" {
		baseQuery += `
			INNER JOIN user_roles ur ON ua.id = ur.user_application_id
			INNER JOIN roles r ON ur.role_id = r.id
		`
	}

	// WHERE clause
	where := ` WHERE ua.application_id = $1 AND ua.active = true`
	args := []interface{}{filters.ApplicationID}
	argCount := 1

	if filters.Email != "" {
		argCount++
		where += fmt.Sprintf(" AND u.email ILIKE $%d", argCount)
		args = append(args, "%"+filters.Email+"%")
	}

	if filters.Name != "" {
		argCount++
		where += fmt.Sprintf(" AND u.name ILIKE $%d", argCount)
		args = append(args, "%"+filters.Name+"%")
	}

	if filters.Active != nil {
		argCount++
		where += fmt.Sprintf(" AND u.active = $%d", argCount)
		args = append(args, *filters.Active)
	}

	if filters.RoleCode != "" {
		argCount++
		where += fmt.Sprintf(" AND r.code = $%d", argCount)
		args = append(args, filters.RoleCode)
	}

	// Count total
	countQuery := "SELECT COUNT(DISTINCT u.id) " + baseQuery + where
	var total int
	err := r.db.QueryRowContext(ctx, countQuery, args...).Scan(&total)
	if err != nil {
		return nil, 0, err
	}

	// Select com paginação
	selectQuery := fmt.Sprintf(`
		SELECT DISTINCT u.id, u.email, u.password, u.name, u.tenant_id, u.active, COALESCE(u.version, 1) as version, u.created_at, u.updated_at
		%s%s
		ORDER BY u.created_at DESC
		LIMIT $%d OFFSET $%d
	`, baseQuery, where, argCount+1, argCount+2)

	args = append(args, filters.Limit, filters.Offset)

	rows, err := r.db.QueryContext(ctx, selectQuery, args...)
	if err != nil {
		return nil, 0, err
	}
	defer rows.Close()

	var users []*entity.User
	for rows.Next() {
		user := &entity.User{}
		if err := rows.Scan(
			&user.ID, &user.Email, &user.Password, &user.Name, &user.TenantID, &user.Active,
			&user.Version, &user.CreatedAt, &user.UpdatedAt,
		); err != nil {
			return nil, 0, err
		}
		users = append(users, user)
	}

	return users, total, rows.Err()
}
