package dto

import (
	"time"

	"github.com/google/uuid"
)

// ListUsersRequest representa os filtros para listar usuários
type ListUsersRequest struct {
	Limit  int    `json:"limit"`
	Offset int    `json:"offset"`
	Email  string `json:"email,omitempty"`
	Name   string `json:"name,omitempty"`
	Active *bool  `json:"active,omitempty"`
	Role   string `json:"role,omitempty"`
}

// ListUsersResponse representa a resposta de listagem de usuários
type ListUsersResponse struct {
	Users []UserItemDTO `json:"users"`
	Total int           `json:"total"`
	Limit int           `json:"limit"`
	Offset int          `json:"offset"`
}

// UserItemDTO representa um item na listagem de usuários
type UserItemDTO struct {
	ID        uuid.UUID   `json:"id"`
	Email     string      `json:"email"`
	Name      string      `json:"name"`
	Active    bool        `json:"active"`
	Roles     []RoleDTO   `json:"roles"`
	CreatedAt time.Time   `json:"created_at"`
	UpdatedAt time.Time   `json:"updated_at"`
}

// GetUserResponse representa a resposta detalhada de um usuário
type GetUserResponse struct {
	ID          uuid.UUID       `json:"id"`
	Email       string          `json:"email"`
	Name        string          `json:"name"`
	Active      bool            `json:"active"`
	Roles       []RoleDTO       `json:"roles"`
	Permissions []PermissionDTO `json:"permissions"`
	Abilities   []AbilityDTO    `json:"abilities"`
	Version     int             `json:"version"`
	CreatedAt   time.Time       `json:"created_at"`
	UpdatedAt   time.Time       `json:"updated_at"`
}

// CreateUserRequest representa a requisição para criar usuário
type CreateUserRequest struct {
	Email    string      `json:"email"`
	Password string      `json:"password"`
	Name     string      `json:"name"`
	RoleIDs  []uuid.UUID `json:"role_ids,omitempty"`
}

// UpdateUserRequest representa a requisição para atualizar nome do usuário
type UpdateUserRequest struct {
	Name    *string `json:"name,omitempty"`
	Version *int    `json:"version,omitempty"`
}

// UpdatePasswordRequest representa a requisição para alterar senha
type UpdatePasswordRequest struct {
	CurrentPassword string `json:"current_password,omitempty"`
	NewPassword     string `json:"new_password"`
}

// UpdateUserRolesRequest representa a requisição para atualizar roles
type UpdateUserRolesRequest struct {
	RoleIDs []uuid.UUID `json:"role_ids"`
	Version *int        `json:"version,omitempty"`
}

// UpdateUserStatusRequest representa a requisição para atualizar status
type UpdateUserStatusRequest struct {
	Active  bool `json:"active"`
	Version *int `json:"version,omitempty"`
}

// ChangePasswordRequest representa a requisição para alterar senha (com senha atual)
type ChangePasswordRequest struct {
	CurrentPassword string `json:"current_password"`
	NewPassword     string `json:"new_password"`
	Version         *int   `json:"version,omitempty"`
}

// ResetPasswordRequest representa a requisição para resetar senha (admin/master)
type ResetPasswordRequest struct {
	NewPassword string `json:"new_password"`
	Version     *int   `json:"version,omitempty"`
}

// UserResponse representa a resposta padrão de usuário
type UserResponse struct {
	ID        uuid.UUID `json:"id"`
	Email     string    `json:"email"`
	Name      string    `json:"name"`
	Active    bool      `json:"active"`
	Version   int       `json:"version"`
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
}

// UpdateUserRolesResponse representa a resposta de atualização de roles
type UpdateUserRolesResponse struct {
	UserID    uuid.UUID  `json:"user_id"`
	Roles     []RoleDTO  `json:"roles"`
	Version   int        `json:"version"`
	UpdatedAt time.Time  `json:"updated_at"`
}

