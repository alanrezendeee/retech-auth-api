package repository

import (
	"context"
	"database/sql"
	"errors"

	"github.com/google/uuid"
	"github.com/theretech/retech-auth-api/internal/domain/entity"
	"github.com/theretech/retech-auth-api/internal/domain/repository"
)

type postgresApplicationRepository struct {
	db *sql.DB
}

// NewPostgresApplicationRepository cria uma nova instância de PostgresApplicationRepository
func NewPostgresApplicationRepository(db *sql.DB) repository.ApplicationRepository {
	return &postgresApplicationRepository{db: db}
}

func (r *postgresApplicationRepository) Create(ctx context.Context, app *entity.Application) error {
	query := `
		INSERT INTO applications (id, name, code, description, active, created_at, updated_at)
		VALUES ($1, $2, $3, $4, $5, $6, $7)
	`
	_, err := r.db.ExecContext(
		ctx, query,
		app.ID, app.Name, app.Code, app.Description, app.Active,
		app.CreatedAt, app.UpdatedAt,
	)
	return err
}

func (r *postgresApplicationRepository) FindByID(ctx context.Context, id uuid.UUID) (*entity.Application, error) {
	query := `
		SELECT id, name, code, description, active, created_at, updated_at
		FROM applications
		WHERE id = $1
	`
	app := &entity.Application{}
	err := r.db.QueryRowContext(ctx, query, id).Scan(
		&app.ID, &app.Name, &app.Code, &app.Description, &app.Active,
		&app.CreatedAt, &app.UpdatedAt,
	)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, errors.New("aplicação não encontrada")
		}
		return nil, err
	}
	return app, nil
}

func (r *postgresApplicationRepository) FindByCode(ctx context.Context, code string) (*entity.Application, error) {
	query := `
		SELECT id, name, code, description, active, created_at, updated_at
		FROM applications
		WHERE code = $1
	`
	app := &entity.Application{}
	err := r.db.QueryRowContext(ctx, query, code).Scan(
		&app.ID, &app.Name, &app.Code, &app.Description, &app.Active,
		&app.CreatedAt, &app.UpdatedAt,
	)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, errors.New("aplicação não encontrada")
		}
		return nil, err
	}
	return app, nil
}

func (r *postgresApplicationRepository) Update(ctx context.Context, app *entity.Application) error {
	query := `
		UPDATE applications
		SET name = $2, code = $3, description = $4, active = $5, updated_at = $6
		WHERE id = $1
	`
	_, err := r.db.ExecContext(
		ctx, query,
		app.ID, app.Name, app.Code, app.Description, app.Active, app.UpdatedAt,
	)
	return err
}

func (r *postgresApplicationRepository) UpsertByCode(ctx context.Context, app *entity.Application) error {
	// Verifica se a aplicação já existe
	existingApp, err := r.FindByCode(ctx, app.Code)
	if err != nil {
		// Não existe, cria nova
		if app.ID == uuid.Nil {
			app.ID = uuid.New()
		}
		return r.Create(ctx, app)
	}

	// Existe, atualiza
	app.ID = existingApp.ID
	app.Active = existingApp.Active
	return r.Update(ctx, app)
}

func (r *postgresApplicationRepository) Delete(ctx context.Context, id uuid.UUID) error {
	query := `DELETE FROM applications WHERE id = $1`
	_, err := r.db.ExecContext(ctx, query, id)
	return err
}

func (r *postgresApplicationRepository) List(ctx context.Context, limit, offset int) ([]*entity.Application, error) {
	query := `
		SELECT id, name, code, description, active, created_at, updated_at
		FROM applications
		ORDER BY name
		LIMIT $1 OFFSET $2
	`
	rows, err := r.db.QueryContext(ctx, query, limit, offset)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var apps []*entity.Application
	for rows.Next() {
		app := &entity.Application{}
		if err := rows.Scan(
			&app.ID, &app.Name, &app.Code, &app.Description, &app.Active,
			&app.CreatedAt, &app.UpdatedAt,
		); err != nil {
			return nil, err
		}
		apps = append(apps, app)
	}

	return apps, rows.Err()
}
