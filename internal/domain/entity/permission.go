package entity

import (
	"time"

	"github.com/google/uuid"
)

// Permission representa uma permissão no sistema
// Segue o padrão CASL: subject (recurso) e action (ação)
type Permission struct {
	ID            uuid.UUID `json:"id"`
	ApplicationID uuid.UUID `json:"application_id"`
	Code          string    `json:"code"`        // Identificador único da permissão (ex: "admin.users.read")
	Subject       string    `json:"subject"`     // Ex: "User", "Product", "Order" (mantido para compatibilidade CASL)
	Action        string    `json:"action"`      // Ex: "create", "read", "update", "delete", "manage" (mantido para compatibilidade CASL)
	Conditions    *string   `json:"conditions"`  // JSON opcional com condições (ex: {"userId": "${user.id}"})
	Description   string    `json:"description"` // Descrição amigável da permissão
	Active        bool      `json:"active"`
	CreatedAt     time.Time `json:"created_at"`
	UpdatedAt     time.Time `json:"updated_at"`
}

// NewPermission cria uma nova instância de Permission (compatibilidade)
func NewPermission(applicationID uuid.UUID, subject, action, description string, conditions *string) *Permission {
	code := subject + "." + action
	return NewPermissionWithCode(applicationID, code, subject, action, description, conditions)
}

// NewPermissionWithCode cria uma nova instância de Permission com code
func NewPermissionWithCode(applicationID uuid.UUID, code, subject, action, description string, conditions *string) *Permission {
	now := time.Now()
	return &Permission{
		ID:            uuid.New(),
		ApplicationID: applicationID,
		Code:          code,
		Subject:       subject,
		Action:        action,
		Conditions:    conditions,
		Description:   description,
		Active:        true,
		CreatedAt:     now,
		UpdatedAt:     now,
	}
}

