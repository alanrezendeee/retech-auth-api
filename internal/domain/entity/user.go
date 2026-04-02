package entity

import (
	"time"

	"github.com/google/uuid"
)

// User representa um usuário no sistema
type User struct {
	ID        uuid.UUID `json:"id"`
	Email     string    `json:"email"`
	Password  string    `json:"-"`
	Name      string    `json:"name"`
	TenantID  *string   `json:"tenant_id,omitempty"` // ID da unidade (tenant). Definido pelo sistema de gestão de usuários/onboarding. O AUTH apenas armazena e inclui no token.
	Active    bool      `json:"active"`
	Version   int       `json:"version"`
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
}

// NewUser cria uma nova instância de User
func NewUser(email, password, name string) *User {
	now := time.Now()
	return &User{
		ID:        uuid.New(),
		Email:     email,
		Password:  password,
		Name:      name,
		Active:    true,
		Version:   1,
		CreatedAt: now,
		UpdatedAt: now,
	}
}

// IncrementVersion incrementa a versão do usuário
func (u *User) IncrementVersion() {
	u.Version++
	u.UpdatedAt = time.Now()
}

// Deactivate desativa o usuário
func (u *User) Deactivate() {
	u.Active = false
	u.UpdatedAt = time.Now()
}

// Activate ativa o usuário
func (u *User) Activate() {
	u.Active = true
	u.UpdatedAt = time.Now()
}

// UpdatePassword atualiza a senha do usuário
func (u *User) UpdatePassword(newPassword string) {
	u.Password = newPassword
	u.UpdatedAt = time.Now()
}

