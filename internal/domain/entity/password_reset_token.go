package entity

import (
	"time"

	"github.com/google/uuid"
)

// PasswordResetToken representa um token de redefinição de senha.
// O campo TokenHash guarda o SHA-256 hex do token — o token em claro só
// trafega no e-mail enviado ao usuário.
type PasswordResetToken struct {
	ID        uuid.UUID  `json:"id"`
	UserID    uuid.UUID  `json:"user_id"`
	TokenHash string     `json:"-"`
	ExpiresAt time.Time  `json:"expires_at"`
	UsedAt    *time.Time `json:"used_at,omitempty"`
	CreatedAt time.Time  `json:"created_at"`
}

// Valid informa se o token ainda pode ser consumido.
func (t *PasswordResetToken) Valid(now time.Time) bool {
	return t.UsedAt == nil && now.Before(t.ExpiresAt)
}
