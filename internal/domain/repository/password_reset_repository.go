package repository

import (
	"context"

	"github.com/google/uuid"
	"github.com/theretech/retech-auth-api/internal/domain/entity"
)

// PasswordResetRepository persiste tokens de redefinição de senha.
type PasswordResetRepository interface {
	Create(ctx context.Context, token *entity.PasswordResetToken) error
	FindByTokenHash(ctx context.Context, tokenHash string) (*entity.PasswordResetToken, error)
	MarkUsed(ctx context.Context, id uuid.UUID) error
	// InvalidateByUser marca como usados todos os tokens pendentes do usuário
	// (chamado ao emitir um novo token e após um reset bem-sucedido).
	InvalidateByUser(ctx context.Context, userID uuid.UUID) error
}
