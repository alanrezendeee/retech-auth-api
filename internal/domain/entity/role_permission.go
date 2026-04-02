package entity

import (
	"time"

	"github.com/google/uuid"
)

// RolePermission representa a associação entre uma role e uma permission
type RolePermission struct {
	ID           uuid.UUID `json:"id"`
	RoleID       uuid.UUID `json:"role_id"`
	PermissionID uuid.UUID `json:"permission_id"`
	Active       bool      `json:"active"`
	CreatedAt    time.Time `json:"created_at"`
	UpdatedAt    time.Time `json:"updated_at"`
}

// NewRolePermission cria uma nova instância de RolePermission
func NewRolePermission(roleID, permissionID uuid.UUID) *RolePermission {
	now := time.Now()
	return &RolePermission{
		ID:           uuid.New(),
		RoleID:       roleID,
		PermissionID: permissionID,
		Active:       true,
		CreatedAt:    now,
		UpdatedAt:    now,
	}
}

