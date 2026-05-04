package repository

import (
	"context"

	"github.com/google/uuid"
	"github.com/theretech/retech-auth-api/internal/domain/entity"
)

// PermissionInfo contém informações completas de uma permissão
type PermissionInfo struct {
	Permission  *entity.Permission
	Role        *entity.Role
	Application *entity.Application
}

// AuthRepository define a interface para operações de autenticação e autorização
type AuthRepository interface {
	// GetUserApplications retorna todas as aplicações que um usuário tem acesso
	GetUserApplications(ctx context.Context, userID uuid.UUID) ([]*entity.Application, error)

	// GetUserPermissions retorna todas as permissões de um usuário em uma aplicação
	GetUserPermissions(ctx context.Context, userID, applicationID uuid.UUID) ([]*PermissionInfo, error)

	// GetUserRoles retorna todas as roles de um usuário em uma aplicação
	GetUserRoles(ctx context.Context, userID, applicationID uuid.UUID) ([]*entity.Role, error)

	// CreateUserApplication cria um vínculo entre usuário e aplicação
	CreateUserApplication(ctx context.Context, userApp *entity.UserApplication) error

	// AssignRoleToUser atribui uma role a um usuário em uma aplicação
	AssignRoleToUser(ctx context.Context, userRole *entity.UserRole) error

	// RemoveRoleFromUser remove uma role de um usuário (soft delete)
	RemoveRoleFromUser(ctx context.Context, userApplicationID, roleID uuid.UUID) error

	// GetUserApplicationID retorna o ID do vínculo user_application
	GetUserApplicationID(ctx context.Context, userID, applicationID uuid.UUID) (uuid.UUID, error)

	// UpdateUserRoles atualiza as roles de um usuário (calcula diff e aplica)
	UpdateUserRoles(ctx context.Context, userApplicationID uuid.UUID, roleIDs []uuid.UUID) error

	// CreateRole cria uma nova role
	CreateRole(ctx context.Context, role *entity.Role) error

	// CreatePermission cria uma nova permissão
	CreatePermission(ctx context.Context, permission *entity.Permission) error

	// AssignPermissionToRole atribui uma permissão a uma role
	AssignPermissionToRole(ctx context.Context, rolePermission *entity.RolePermission) error

	// FindUserByEmailAndApplication busca um usuário por email e aplicação
	// Retorna o user_application para acesso ao tenant_id correto (por app)
	FindUserByEmailAndApplication(ctx context.Context, email string, applicationCode string) (*entity.User, *entity.Application, *entity.UserApplication, error)

	// FindUserApplication busca o vínculo user_application por userID e applicationID
	FindUserApplication(ctx context.Context, userID, applicationID uuid.UUID) (*entity.UserApplication, error)

	// Roles
	GetRolesByApplication(ctx context.Context, applicationID uuid.UUID, active *bool) ([]*entity.Role, error)
	GetRole(ctx context.Context, roleID uuid.UUID) (*entity.Role, error)
	UpdateRole(ctx context.Context, role *entity.Role) error
	GetRolePermissions(ctx context.Context, roleID uuid.UUID) ([]*entity.Permission, error)

	// Permissions
	GetPermissionsByApplication(ctx context.Context, applicationID uuid.UUID) ([]*entity.Permission, error)
	GetPermission(ctx context.Context, permissionID uuid.UUID) (*entity.Permission, error)
	GetPermissionByCode(ctx context.Context, applicationID uuid.UUID, code string) (*entity.Permission, error)
	UpdatePermission(ctx context.Context, permission *entity.Permission) error
	UpsertPermission(ctx context.Context, permission *entity.Permission) error

	// Roles (sync)
	GetRoleByCode(ctx context.Context, applicationID uuid.UUID, code string) (*entity.Role, error)
	UpsertRole(ctx context.Context, role *entity.Role) error
	UpsertRolePermissions(ctx context.Context, roleID uuid.UUID, permissionIDs []uuid.UUID) error

	// GetActiveUsersByRole retorna usuários ativos que possuem uma role específica
	GetActiveUsersByRole(ctx context.Context, roleID, applicationID uuid.UUID) ([]*entity.User, error)
}
