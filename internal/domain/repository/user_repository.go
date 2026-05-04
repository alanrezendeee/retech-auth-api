package repository

import (
	"context"

	"github.com/google/uuid"
	"github.com/theretech/retech-auth-api/internal/domain/entity"
)

// UserFilters define os filtros para listagem de usuários
type UserFilters struct {
	ApplicationID uuid.UUID
	TenantID      *string
	Email         string
	Name          string
	Active        *bool
	RoleCode      string
	Limit         int
	Offset        int
}

// UserRepository define a interface para operações com usuários
type UserRepository interface {
	Create(ctx context.Context, user *entity.User) error
	FindByID(ctx context.Context, id uuid.UUID) (*entity.User, error)
	FindByEmail(ctx context.Context, email string) (*entity.User, error)
	FindByIDAndApplication(ctx context.Context, userID, applicationID uuid.UUID) (*entity.User, error)
	Update(ctx context.Context, user *entity.User) error
	UpdateWithVersion(ctx context.Context, user *entity.User, expectedVersion int) error
	UpdatePassword(ctx context.Context, userID uuid.UUID, passwordHash string) error
	UpdatePasswordWithVersion(ctx context.Context, userID uuid.UUID, passwordHash string, expectedVersion int) error
	UpdateStatus(ctx context.Context, userID uuid.UUID, active bool, expectedVersion int) error
	UpdateUserApplicationStatus(ctx context.Context, userID, applicationID uuid.UUID, active bool, expectedVersion int) error
	Delete(ctx context.Context, id uuid.UUID) error
	SoftDeleteFromApplication(ctx context.Context, userID, applicationID uuid.UUID) error
	List(ctx context.Context, limit, offset int) ([]*entity.User, error)
	ListByApplication(ctx context.Context, filters UserFilters) ([]*entity.User, int, error)
}
