package usecase

import (
	"context"
	"encoding/json"
	"errors"

	"github.com/theretech/retechauth-api/internal/domain/dto"
	"github.com/theretech/retechauth-api/internal/domain/repository"
	"github.com/google/uuid"
)

var (
	ErrUserNotFound = errors.New("usuário não encontrado")
)

// GetUserInfoUseCase representa o caso de uso para obter informações do usuário
type GetUserInfoUseCase struct {
	userRepo repository.UserRepository
	authRepo repository.AuthRepository
	appRepo  repository.ApplicationRepository
}

// NewGetUserInfoUseCase cria uma nova instância de GetUserInfoUseCase
func NewGetUserInfoUseCase(
	userRepo repository.UserRepository,
	authRepo repository.AuthRepository,
	appRepo repository.ApplicationRepository,
) *GetUserInfoUseCase {
	return &GetUserInfoUseCase{
		userRepo: userRepo,
		authRepo: authRepo,
		appRepo:  appRepo,
	}
}

// Execute executa a busca de informações do usuário
func (uc *GetUserInfoUseCase) Execute(ctx context.Context, userID, applicationID uuid.UUID) (*dto.MeResponse, error) {
	// Busca o usuário
	user, err := uc.userRepo.FindByID(ctx, userID)
	if err != nil {
		return nil, ErrUserNotFound
	}

	// Busca a aplicação
	app, err := uc.appRepo.FindByID(ctx, applicationID)
	if err != nil {
		return nil, errors.New("aplicação não encontrada")
	}

	// Busca as roles do usuário
	roles, err := uc.authRepo.GetUserRoles(ctx, userID, applicationID)
	if err != nil {
		return nil, err
	}

	// Busca as permissões do usuário
	permissions, err := uc.authRepo.GetUserPermissions(ctx, userID, applicationID)
	if err != nil {
		return nil, err
	}

	// Converte para DTOs
	rolesDTO := make([]dto.RoleDTO, len(roles))
	isMaster := false
	
	for i, role := range roles {
		rolesDTO[i] = dto.RoleDTO{
			ID:          role.ID,
			Name:        role.Name,
			Code:        role.Code,
			Description: role.Description,
		}
		
		// Detecta se o usuário tem a role "master"
		if role.Code == "master" {
			isMaster = true
		}
	}

	// Se o usuário é Master, retorna apenas { action: "manage", subject: "all" }
	// Isso dá acesso total automático a todos os recursos (padrão CASL.js)
	var abilitiesDTO []dto.AbilityDTO
	var permissionsDTO []dto.PermissionDTO
	
	if isMaster {
		// Master tem acesso total - apenas uma ability "manage all"
		abilitiesDTO = []dto.AbilityDTO{
			{
				Action:  "manage",
				Subject: "all",
			},
		}
		// Permissions fica vazio para Master (abilities é o que importa)
		permissionsDTO = []dto.PermissionDTO{}
	} else {
		// Usuários comuns: converte permissions para abilities
		permissionsDTO = make([]dto.PermissionDTO, len(permissions))
		abilitiesDTO = make([]dto.AbilityDTO, len(permissions))
		
		for i, permInfo := range permissions {
			perm := permInfo.Permission
			permissionsDTO[i] = dto.PermissionDTO{
				ID:          perm.ID,
				Code:        perm.Code,
				Subject:     perm.Subject,
				Action:      perm.Action,
				Conditions:  perm.Conditions,
				Description: perm.Description,
			}

			// Converte para formato CASL
			ability := dto.AbilityDTO{
				Action:  perm.Action,
				Subject: perm.Subject,
			}

			// Parse das condições se existirem
			if perm.Conditions != nil && *perm.Conditions != "" {
				var conditions map[string]interface{}
				if err := json.Unmarshal([]byte(*perm.Conditions), &conditions); err == nil {
					ability.Conditions = conditions
				}
			}

			abilitiesDTO[i] = ability
		}
	}

	return &dto.MeResponse{
		User: dto.UserDTO{
			ID:    user.ID,
			Email: user.Email,
			Name:  user.Name,
		},
		Application: dto.ApplicationDTO{
			ID:          app.ID,
			Name:        app.Name,
			Code:        app.Code,
			Description: app.Description,
		},
		Roles:       rolesDTO,
		Permissions: permissionsDTO,
		Abilities:   abilitiesDTO,
	}, nil
}

