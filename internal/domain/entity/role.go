package entity

import (
	"time"

	"github.com/google/uuid"
)

// Role representa um papel/função no sistema (ex: admin, user, master)
type Role struct {
	ID            uuid.UUID `json:"id"`
	ApplicationID uuid.UUID `json:"application_id"`
	Name          string    `json:"name"`
	Code          string    `json:"code"` // Ex: "master", "admin", "user"
	Description   string    `json:"description"`
	System        bool      `json:"system"` // true = role base do sistema (não editável), false = role customizada
	Active        bool      `json:"active"`
	CreatedAt     time.Time `json:"created_at"`
	UpdatedAt     time.Time `json:"updated_at"`
}

// NewRole cria uma nova instância de Role
func NewRole(applicationID uuid.UUID, name, code, description string) *Role {
	return NewRoleWithSystem(applicationID, name, code, description, false)
}

// NewRoleWithSystem cria uma nova instância de Role com flag system
func NewRoleWithSystem(applicationID uuid.UUID, name, code, description string, system bool) *Role {
	now := time.Now()
	return &Role{
		ID:            uuid.New(),
		ApplicationID: applicationID,
		Name:          name,
		Code:          code,
		Description:   description,
		System:        system,
		Active:        true,
		CreatedAt:     now,
		UpdatedAt:     now,
	}
}

