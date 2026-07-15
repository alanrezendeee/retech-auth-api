package usecase

import (
	"context"
	"crypto/rand"
	"crypto/sha256"
	"encoding/base64"
	"encoding/hex"
	"errors"
	"strings"
	"time"

	"github.com/google/uuid"
	"github.com/theretech/retech-auth-api/internal/application/service"
	"github.com/theretech/retech-auth-api/internal/domain/entity"
	"github.com/theretech/retech-auth-api/internal/domain/repository"
)

// Erros sentinela do fluxo de reset — o chamador (serviço interno via HMAC)
// decide como traduzi-los; a API pública do produto nunca vaza existência de e-mail.
var (
	ErrResetUserNotFound = errors.New("usuário não encontrado ou inativo")
	ErrResetTokenInvalid = errors.New("token inválido, expirado ou já utilizado")
	ErrResetWeakPassword = errors.New("a nova senha deve ter no mínimo 8 caracteres")
)

const resetTokenTTL = 60 * time.Minute

// PasswordResetUseCase implementa o fluxo "esqueci a senha":
// Request emite um token de uso único (retornado ao serviço chamador, que envia
// o e-mail) e Confirm consome o token trocando a senha.
type PasswordResetUseCase struct {
	userRepo  repository.UserRepository
	resetRepo repository.PasswordResetRepository
	hash      service.HashService
}

func NewPasswordResetUseCase(
	userRepo repository.UserRepository,
	resetRepo repository.PasswordResetRepository,
	hash service.HashService,
) *PasswordResetUseCase {
	return &PasswordResetUseCase{userRepo: userRepo, resetRepo: resetRepo, hash: hash}
}

// ResetRequestResult devolve o token em claro (única vez que ele existe fora
// do e-mail) e os dados necessários para compor a mensagem.
type ResetRequestResult struct {
	Token     string
	UserName  string
	Email     string
	ExpiresAt time.Time
}

func hashToken(token string) string {
	sum := sha256.Sum256([]byte(token))
	return hex.EncodeToString(sum[:])
}

// Request emite um novo token para o e-mail informado. Tokens pendentes
// anteriores são invalidados (só o link mais recente funciona).
func (u *PasswordResetUseCase) Request(ctx context.Context, email string) (*ResetRequestResult, error) {
	email = strings.ToLower(strings.TrimSpace(email))
	if email == "" {
		return nil, ErrResetUserNotFound
	}

	user, err := u.userRepo.FindByEmail(ctx, email)
	if err != nil || user == nil || !user.Active {
		return nil, ErrResetUserNotFound
	}

	raw := make([]byte, 32)
	if _, err := rand.Read(raw); err != nil {
		return nil, err
	}
	token := base64.RawURLEncoding.EncodeToString(raw)

	if err := u.resetRepo.InvalidateByUser(ctx, user.ID); err != nil {
		return nil, err
	}

	now := time.Now().UTC()
	rec := &entity.PasswordResetToken{
		ID:        uuid.New(),
		UserID:    user.ID,
		TokenHash: hashToken(token),
		ExpiresAt: now.Add(resetTokenTTL),
		CreatedAt: now,
	}
	if err := u.resetRepo.Create(ctx, rec); err != nil {
		return nil, err
	}

	return &ResetRequestResult{
		Token:     token,
		UserName:  user.Name,
		Email:     user.Email,
		ExpiresAt: rec.ExpiresAt,
	}, nil
}

// Confirm consome o token e define a nova senha. A versão do usuário é
// incrementada, invalidando refresh tokens emitidos com a senha antiga.
func (u *PasswordResetUseCase) Confirm(ctx context.Context, token, newPassword string) error {
	if len(strings.TrimSpace(newPassword)) < 8 {
		return ErrResetWeakPassword
	}
	token = strings.TrimSpace(token)
	if token == "" {
		return ErrResetTokenInvalid
	}

	rec, err := u.resetRepo.FindByTokenHash(ctx, hashToken(token))
	if err != nil || rec == nil || !rec.Valid(time.Now().UTC()) {
		return ErrResetTokenInvalid
	}

	user, err := u.userRepo.FindByID(ctx, rec.UserID)
	if err != nil || user == nil || !user.Active {
		return ErrResetTokenInvalid
	}

	hashed, err := u.hash.HashPassword(newPassword)
	if err != nil {
		return err
	}

	if err := u.userRepo.UpdatePasswordWithVersion(ctx, user.ID, hashed, user.Version); err != nil {
		return err
	}

	if err := u.resetRepo.MarkUsed(ctx, rec.ID); err != nil {
		return err
	}
	// Garante que nenhum outro token pendente do usuário sobreviva ao reset.
	return u.resetRepo.InvalidateByUser(ctx, rec.UserID)
}
