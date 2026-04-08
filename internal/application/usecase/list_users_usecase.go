package usecase

import (
	"context"

	"github.com/google/uuid"
	"github.com/theretech/retech-auth-api/internal/domain/dto"
	"github.com/theretech/retech-auth-api/internal/domain/entity"
	"github.com/theretech/retech-auth-api/internal/domain/repository"
)

// ListUsersUseCase representa o caso de uso para listar usuários
type ListUsersUseCase struct {
	userRepo repository.UserRepository
	authRepo repository.AuthRepository
}

// NewListUsersUseCase cria uma nova instância de ListUsersUseCase
func NewListUsersUseCase(
	userRepo repository.UserRepository,
	authRepo repository.AuthRepository,
) *ListUsersUseCase {
	return &ListUsersUseCase{
		userRepo: userRepo,
		authRepo: authRepo,
	}
}

// Execute lista os usuários da aplicação com filtros
func (uc *ListUsersUseCase) Execute(ctx context.Context, applicationID uuid.UUID, req dto.ListUsersRequest) (*dto.ListUsersResponse, error) {
	// Define defaults para paginação
	if req.Limit == 0 {
		req.Limit = 20
	}
	if req.Limit > 100 {
		req.Limit = 100 // Máximo de 100 por página
	}

	// Monta os filtros
	filters := repository.UserFilters{
		ApplicationID: applicationID,
		Email:         req.Email,
		Name:          req.Name,
		Active:        req.Active,
		RoleCode:      req.Role,
		Limit:         req.Limit,
		Offset:        req.Offset,
	}

	// Busca usuários
	users, total, err := uc.userRepo.ListByApplication(ctx, filters)
	if err != nil {
		return nil, err
	}

	// Converte para DTOs e busca roles de cada usuário
	userItems := make([]dto.UserItemDTO, len(users))
	for i, user := range users {
		// Busca roles do usuário
		roles, err := uc.authRepo.GetUserRoles(ctx, user.ID, applicationID)
		if err != nil {
			// Se erro ao buscar roles, continua sem roles
			roles = []*entity.Role{}
		}

		rolesDTO := make([]dto.RoleDTO, len(roles))
		for j, role := range roles {
			rolesDTO[j] = dto.RoleDTO{
				ID:          role.ID,
				Name:        role.Name,
				Code:        role.Code,
				Description: role.Description,
			}
		}

		userItems[i] = dto.UserItemDTO{
			ID:        user.ID,
			Email:     user.Email,
			Name:      user.Name,
			Active:    user.Active,
			Roles:     rolesDTO,
			CreatedAt: user.CreatedAt,
			UpdatedAt: user.UpdatedAt,
		}
	}

	return &dto.ListUsersResponse{
		Users:  userItems,
		Total:  total,
		Limit:  req.Limit,
		Offset: req.Offset,
	}, nil
}
