package entity

import (
	"time"

	"github.com/google/uuid"
)

// UserRole representa a associação entre um usuário e uma role em uma aplicação
type UserRole struct {
	ID                uuid.UUID `json:"id"`
	UserApplicationID uuid.UUID `json:"user_application_id"`
	RoleID            uuid.UUID `json:"role_id"`
	Active            bool      `json:"active"`
	CreatedAt         time.Time `json:"created_at"`
	UpdatedAt         time.Time `json:"updated_at"`
}

// NewUserRole cria uma nova instância de UserRole
func NewUserRole(userApplicationID, roleID uuid.UUID) *UserRole {
	now := time.Now()
	return &UserRole{
		ID:                uuid.New(),
		UserApplicationID: userApplicationID,
		RoleID:            roleID,
		Active:            true,
		CreatedAt:         now,
		UpdatedAt:         now,
	}
}

