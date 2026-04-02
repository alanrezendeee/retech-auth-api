package dto

import (
	"time"

	"github.com/google/uuid"
)

// ==================== APPLICATIONS ====================

type ListApplicationsResponse struct {
	Applications []ApplicationItemDTO `json:"applications"`
	Total        int                  `json:"total"`
}

type ApplicationItemDTO struct {
	ID          uuid.UUID `json:"id"`
	Name        string    `json:"name"`
	Code        string    `json:"code"`
	Description string    `json:"description"`
	Active      bool      `json:"active"`
	CreatedAt   time.Time `json:"created_at"`
	UpdatedAt   time.Time `json:"updated_at"`
}

type CreateApplicationRequest struct {
	Name        string `json:"name"`
	Code        string `json:"code"`
	Description string `json:"description,omitempty"`
}

type UpdateApplicationRequest struct {
	Name        *string `json:"name,omitempty"`
	Description *string `json:"description,omitempty"`
	Active      *bool   `json:"active,omitempty"`
}

// ==================== ROLES ====================

type ListRolesRequest struct {
	Active            *bool `json:"active,omitempty"`            // Filtro: true=ativas, false=inativas, nil=todas
	IncludePermissions bool  `json:"include_permissions,omitempty"` // Se true, inclui permissions em cada role
}

type ListRolesResponse struct {
	Roles  []RoleItemDTO `json:"roles"`
	Total  int           `json:"total"`
}

type RoleItemDTO struct {
	ID              uuid.UUID       `json:"id"`
	Name            string          `json:"name"`
	Code            string          `json:"code"`
	Description     string          `json:"description"`
	System          bool            `json:"system"` // true = role base do sistema (não editável), false = role customizada
	Active          bool            `json:"active"`
	PermissionCount int             `json:"permission_count"`
	Permissions     []PermissionDTO `json:"permissions,omitempty"` // Incluído quando include_permissions=true
	CreatedAt       time.Time       `json:"created_at"`
	UpdatedAt       time.Time       `json:"updated_at"`
}

type GetRoleResponse struct {
	ID          uuid.UUID       `json:"id"`
	Name        string          `json:"name"`
	Code        string          `json:"code"`
	Description string          `json:"description"`
	System      bool            `json:"system"` // true = role base do sistema (não editável), false = role customizada
	Active      bool            `json:"active"`
	Permissions []PermissionDTO `json:"permissions"`
	CreatedAt   time.Time       `json:"created_at"`
	UpdatedAt   time.Time       `json:"updated_at"`
}

type CreateRoleRequest struct {
	Name          string      `json:"name"`
	Code          string      `json:"code"`
	Description   string      `json:"description,omitempty"`
	System        bool        `json:"system"`         // true = role base do sistema (não editável), false = role customizada
	PermissionIDs []uuid.UUID `json:"permission_ids,omitempty"` // IDs de permissões para vincular à role na criação (opcional)
}

type UpdateRoleRequest struct {
	Name          *string    `json:"name,omitempty"`
	Description   *string    `json:"description,omitempty"`
	Active        *bool      `json:"active,omitempty"`
	PermissionIDs []uuid.UUID `json:"permission_ids,omitempty"` // IDs de permissões para atualizar vínculos (opcional)
}

type AssignPermissionsRequest struct {
	PermissionIDs []uuid.UUID `json:"permission_ids"`
}

type UpdateRolePermissionsRequest struct {
	PermissionCodes []string `json:"permission_codes"` // Lista de codes de permissions ou wildcards (ex: ["admin.*"])
}

type UpdateRolePermissionsResponse struct {
	RoleID      uuid.UUID       `json:"role_id"`
	Permissions []PermissionDTO `json:"permissions"`
}

// ==================== PERMISSIONS ====================

type ListPermissionsResponse struct {
	Permissions []PermissionItemDTO `json:"permissions"`
	Total       int                 `json:"total"`
}

type PermissionItemDTO struct {
	ID          uuid.UUID `json:"id"`
	Code        string    `json:"code"`        // Código único da permissão (ex: "admin.users.view")
	Subject     string    `json:"subject"`
	Action      string    `json:"action"`
	Description string    `json:"description"`
	Active      bool      `json:"active"`
	CreatedAt   time.Time `json:"created_at"`
	UpdatedAt   time.Time `json:"updated_at"`
}

type CreatePermissionRequest struct {
	Subject     string  `json:"subject"`
	Action      string  `json:"action"`
	Conditions  *string `json:"conditions,omitempty"`
	Description string  `json:"description,omitempty"`
}

type UpdatePermissionRequest struct {
	Description *string `json:"description,omitempty"`
	Conditions  *string `json:"conditions,omitempty"`
	Active      *bool   `json:"active,omitempty"`
}

// ==================== USER-ROLE ====================

type AssignRolesRequest struct {
	RoleIDs []uuid.UUID `json:"role_ids"`
}

// ==================== SYNC / MANIFEST ====================

type SyncManifestRequest struct {
	Application  SyncApplicationDTO   `json:"application"`
	Permissions  []SyncPermissionDTO  `json:"permissions"`
	Roles        []SyncRoleDTO        `json:"roles"`
	Users        []SyncUserDTO        `json:"users,omitempty"`
}

type SyncApplicationDTO struct {
	Code        string `json:"code"`
	Name        string `json:"name"`
	Description string `json:"description,omitempty"`
}

type SyncPermissionDTO struct {
	Code        string  `json:"code"`
	Description string  `json:"description,omitempty"`
	Subject     string  `json:"subject,omitempty"`    // Mantido para compatibilidade CASL
	Action      string  `json:"action,omitempty"`     // Mantido para compatibilidade CASL
	Conditions  *string `json:"conditions,omitempty"`
}

type SyncRoleDTO struct {
	Code        string   `json:"code"`
	Name        string   `json:"name"`
	System      bool     `json:"system"`       // true = role base (não editável), false = role customizada
	Description string   `json:"description,omitempty"`
	Permissions []string `json:"permissions"`  // Lista de codes de permissions ou wildcards (ex: "admin.*")
}

type SyncUserDTO struct {
	Email    string   `json:"email"`
	Password string   `json:"password"`  // Senha em texto claro (será hasheada pelo servidor)
	Name     string   `json:"name"`
	TenantID *string  `json:"tenant_id,omitempty"`  // ID da unidade (tenant). Opcional, usado apenas na criação do primeiro usuário (bootstrap)
	Active   bool     `json:"active,omitempty"`  // default: true
	Roles    []string `json:"roles"`  // Lista de codes de roles (ex: ["master", "admin.operator"])
}

type SyncManifestResponse struct {
	Application  SyncResultDTO       `json:"application"`
	Permissions  []SyncResultDTO     `json:"permissions"`
	Roles        []SyncRoleResultDTO `json:"roles"`
	Users        []SyncResultDTO     `json:"users,omitempty"`
}

type SyncResultDTO struct {
	Code      string `json:"code"`
	Action    string `json:"action"`    // "created" ou "updated"
	ID        uuid.UUID `json:"id,omitempty"`
}

type SyncRoleResultDTO struct {
	Code        string   `json:"code"`
	Action      string   `json:"action"`      // "created" ou "updated"
	ID          uuid.UUID `json:"id,omitempty"`
	Permissions []SyncResultDTO `json:"permissions"`
}

