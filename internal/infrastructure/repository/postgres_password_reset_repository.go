package repository

import (
	"context"
	"database/sql"
	"errors"

	"github.com/google/uuid"
	"github.com/theretech/retech-auth-api/internal/domain/entity"
	"github.com/theretech/retech-auth-api/internal/domain/repository"
)

type postgresPasswordResetRepository struct {
	db *sql.DB
}

// NewPostgresPasswordResetRepository cria uma nova instância de PasswordResetRepository
func NewPostgresPasswordResetRepository(db *sql.DB) repository.PasswordResetRepository {
	return &postgresPasswordResetRepository{db: db}
}

func (r *postgresPasswordResetRepository) Create(ctx context.Context, token *entity.PasswordResetToken) error {
	query := `
		INSERT INTO password_reset_tokens (id, user_id, token_hash, expires_at, created_at)
		VALUES ($1, $2, $3, $4, $5)
	`
	_, err := r.db.ExecContext(ctx, query,
		token.ID, token.UserID, token.TokenHash, token.ExpiresAt, token.CreatedAt,
	)
	return err
}

func (r *postgresPasswordResetRepository) FindByTokenHash(ctx context.Context, tokenHash string) (*entity.PasswordResetToken, error) {
	query := `
		SELECT id, user_id, token_hash, expires_at, used_at, created_at
		FROM password_reset_tokens
		WHERE token_hash = $1
	`
	t := &entity.PasswordResetToken{}
	err := r.db.QueryRowContext(ctx, query, tokenHash).Scan(
		&t.ID, &t.UserID, &t.TokenHash, &t.ExpiresAt, &t.UsedAt, &t.CreatedAt,
	)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, errors.New("token não encontrado")
		}
		return nil, err
	}
	return t, nil
}

func (r *postgresPasswordResetRepository) MarkUsed(ctx context.Context, id uuid.UUID) error {
	_, err := r.db.ExecContext(ctx,
		`UPDATE password_reset_tokens SET used_at = NOW() WHERE id = $1 AND used_at IS NULL`, id)
	return err
}

func (r *postgresPasswordResetRepository) InvalidateByUser(ctx context.Context, userID uuid.UUID) error {
	_, err := r.db.ExecContext(ctx,
		`UPDATE password_reset_tokens SET used_at = NOW() WHERE user_id = $1 AND used_at IS NULL`, userID)
	return err
}
