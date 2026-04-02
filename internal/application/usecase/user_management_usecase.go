package usecase

import (
	"context"
	"encoding/json"
	"errors"
	"time"

	"github.com/theretech/retechauth-api/internal/application/service"
	"github.com/theretech/retechauth-api/internal/domain/dto"
	"github.com/theretech/retechauth-api/internal/domain/entity"
	"github.com/theretech/retechauth-api/internal/domain/repository"
	"github.com/theretech/retechauth-api/internal/infrastructure/http/middleware"
	"github.com/google/uuid"
)

var (
	ErrUnauthorized = errors.New("não autorizado")
	ErrInvalidData  = errors.New("dados inválidos")
)

// UserManagementUseCase agrupa operações de gerenciamento de usuários
type UserManagementUseCase struct {
	userRepo repository.UserRepository
	authRepo repository.AuthRepository
	appRepo  repository.ApplicationRepository
	hashSvc  service.HashService
}

// NewUserManagementUseCase cria uma nova instância
func NewUserManagementUseCase(
	userRepo repository.UserRepository,
	authRepo repository.AuthRepository,
	appRepo repository.ApplicationRepository,
	hashSvc service.HashService,
) *UserManagementUseCase {
	return &UserManagementUseCase{
		userRepo: userRepo,
		authRepo: authRepo,
		appRepo:  appRepo,
		hashSvc:  hashSvc,
	}
}

// GetUser busca detalhes de um usuário específico
func (uc *UserManagementUseCase) GetUser(ctx context.Context, userID, applicationID uuid.UUID) (*dto.GetUserResponse, error) {
	// Busca usuário verificando se pertence à aplicação
	user, err := uc.userRepo.FindByIDAndApplication(ctx, userID, applicationID)
	if err != nil {
		return nil, err
	}

	// Busca roles
	roles, err := uc.authRepo.GetUserRoles(ctx, userID, applicationID)
	if err != nil {
		return nil, err
	}

	// Busca permissions
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
		if role.Code == "master" {
			isMaster = true
		}
	}

	var abilitiesDTO []dto.AbilityDTO
	var permissionsDTO []dto.PermissionDTO

	if isMaster {
		abilitiesDTO = []dto.AbilityDTO{{Action: "manage", Subject: "all"}}
		permissionsDTO = []dto.PermissionDTO{}
	} else {
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

			ability := dto.AbilityDTO{
				Action:  perm.Action,
				Subject: perm.Subject,
			}

			if perm.Conditions != nil && *perm.Conditions != "" {
				var conditions map[string]interface{}
				if err := json.Unmarshal([]byte(*perm.Conditions), &conditions); err == nil {
					ability.Conditions = conditions
				}
			}

			abilitiesDTO[i] = ability
		}
	}

	return &dto.GetUserResponse{
		ID:          user.ID,
		Email:       user.Email,
		Name:        user.Name,
		Active:      user.Active,
		Roles:       rolesDTO,
		Permissions: permissionsDTO,
		Abilities:   abilitiesDTO,
		Version:     user.Version,
		CreatedAt:   user.CreatedAt,
		UpdatedAt:   user.UpdatedAt,
	}, nil
}

// UpdateUser atualiza apenas o nome do usuário
func (uc *UserManagementUseCase) UpdateUser(ctx context.Context, userID, applicationID uuid.UUID, req dto.UpdateUserRequest) (*dto.UserResponse, error) {
	user, err := uc.userRepo.FindByIDAndApplication(ctx, userID, applicationID)
	if err != nil {
		return nil, err
	}

	if req.Name == nil {
		return nil, errors.New("nome é obrigatório")
	}

	expectedVersion := user.Version
	if req.Version != nil {
		expectedVersion = *req.Version
	}

		user.Name = *req.Name
	user.IncrementVersion()
	user.UpdatedAt = time.Now()

	if err := uc.userRepo.UpdateWithVersion(ctx, user, expectedVersion); err != nil {
		if err.Error() == "versão do usuário está desatualizada (409 Conflict)" {
			return nil, errors.New("versão desatualizada: o usuário foi modificado por outro processo")
		}
		return nil, err
	}

	user, _ = uc.userRepo.FindByIDAndApplication(ctx, userID, applicationID)
	return &dto.UserResponse{
		ID:        user.ID,
		Email:     user.Email,
		Name:      user.Name,
		Active:    user.Active,
		Version:   user.Version,
		CreatedAt: user.CreatedAt,
		UpdatedAt: user.UpdatedAt,
	}, nil
}

// UpdatePassword atualiza a senha do usuário
func (uc *UserManagementUseCase) UpdatePassword(ctx context.Context, userID, applicationID, currentUserID uuid.UUID, isMaster bool, req dto.UpdatePasswordRequest) error {
	// Busca usuário verificando se pertence à aplicação
	user, err := uc.userRepo.FindByIDAndApplication(ctx, userID, applicationID)
	if err != nil {
		return err
	}

	// Se não é Master E não é o próprio usuário, precisa validar senha atual
	if !isMaster && currentUserID != userID {
		return ErrUnauthorized
	}

	// Se não é Master, valida senha atual
	if !isMaster && req.CurrentPassword != "" {
		if err := uc.hashSvc.CheckPassword(user.Password, req.CurrentPassword); err != nil {
			return errors.New("senha atual incorreta")
		}
	}

	// Valida nova senha
	if len(req.NewPassword) < 6 {
		return errors.New("senha deve ter no mínimo 6 caracteres")
	}

	// Gera hash da nova senha
	passwordHash, err := uc.hashSvc.HashPassword(req.NewPassword)
	if err != nil {
		return err
	}

	// Atualiza senha
	return uc.userRepo.UpdatePassword(ctx, userID, passwordHash)
}

// DeleteUser soft delete do usuário (remove apenas da aplicação)
func (uc *UserManagementUseCase) DeleteUser(ctx context.Context, userID, applicationID, currentUserID uuid.UUID) error {
	// Verifica se o usuário existe e pertence à aplicação
	_, err := uc.userRepo.FindByIDAndApplication(ctx, userID, applicationID)
	if err != nil {
		return err
	}

	// Busca roles do usuário alvo para validações
	roles, err := uc.authRepo.GetUserRoles(ctx, userID, applicationID)
	if err != nil {
		// Se não conseguir buscar roles, continua (pode não ter roles)
		roles = []*entity.Role{}
	}

	// Verifica se o usuário alvo é master
	for _, role := range roles {
		if role.Code == "master" {
			if userID == currentUserID {
				return errors.New("não é possível excluir um usuário master, incluindo a si mesmo")
			}
			return errors.New("não é possível excluir um usuário master")
		}
	}

	// Verifica se está tentando excluir a si mesmo (após verificar master)
	if userID == currentUserID {
		return errors.New("não é possível se excluir a si mesmo")
	}

	// Soft delete: apenas desativa o vínculo com a aplicação
	return uc.userRepo.SoftDeleteFromApplication(ctx, userID, applicationID)
}

// CreateUser cria um novo usuário e vincula à aplicação
// O tenant_id do novo usuário é definido automaticamente pelo AUTH usando o tenant_id do JWT do criador
func (uc *UserManagementUseCase) CreateUser(ctx context.Context, applicationID uuid.UUID, req dto.CreateUserRequest) (*dto.UserResponse, error) {
	// Validações
	if req.Email == "" || req.Password == "" || req.Name == "" {
		return nil, ErrInvalidData
	}

	if len(req.Password) < 6 {
		return nil, errors.New("senha deve ter no mínimo 6 caracteres")
	}

	// Verifica se app existe
	_, err := uc.appRepo.FindByID(ctx, applicationID)
	if err != nil {
		return nil, errors.New("aplicação não encontrada")
	}

	// Extrai tenant_id do contexto (do JWT do criador)
	// O tenant_id é definido pelo AUTH, não pelo frontend
	// Se o criador não tiver tenant_id, o novo usuário também não terá
	var tenantID *string
	if tenantIDStr, ok := middleware.GetTenantID(ctx); ok && tenantIDStr != "" {
		tenantID = &tenantIDStr
	}

	// Gera hash da senha
	passwordHash, err := uc.hashSvc.HashPassword(req.Password)
	if err != nil {
		return nil, err
	}

	// Cria usuário com tenant_id do criador
	// NOTA: Qualquer tenant_id enviado no request é ignorado (segurança)
	user := entity.NewUser(req.Email, passwordHash, req.Name)
	user.TenantID = tenantID // Define tenant_id do criador

	if err := uc.userRepo.Create(ctx, user); err != nil {
		return nil, err
	}

	// Cria vínculo user_application
	userApp := entity.NewUserApplication(user.ID, applicationID)
	if err := uc.authRepo.CreateUserApplication(ctx, userApp); err != nil {
		return nil, errors.New("erro ao vincular usuário à aplicação: " + err.Error())
	}

	// Atribui roles ao usuário (se fornecidas)
	if len(req.RoleIDs) > 0 {
		for _, roleID := range req.RoleIDs {
			// Verifica se a role existe e pertence à aplicação
			role, err := uc.authRepo.GetRole(ctx, roleID)
			if err != nil {
				continue // Ignora roles inválidas
			}
			if role.ApplicationID != applicationID {
				continue // Ignora roles de outras aplicações
			}

			// Cria vínculo user_role
			userRole := entity.NewUserRole(userApp.ID, roleID)
			if err := uc.authRepo.AssignRoleToUser(ctx, userRole); err != nil {
				// Log erro mas continua (não falha a criação do usuário)
				continue
			}
		}
	}

	return &dto.UserResponse{
		ID:        user.ID,
		Email:     user.Email,
		Name:      user.Name,
		Active:    user.Active,
		Version:   user.Version,
		CreatedAt: user.CreatedAt,
		UpdatedAt: user.UpdatedAt,
	}, nil
}

// UpdateUserRoles atualiza as roles de um usuário (calcula diff e aplica regras)
func (uc *UserManagementUseCase) UpdateUserRoles(ctx context.Context, userID, applicationID, currentUserID uuid.UUID, isMaster bool, req dto.UpdateUserRolesRequest) (*dto.UpdateUserRolesResponse, error) {
	user, err := uc.userRepo.FindByIDAndApplication(ctx, userID, applicationID)
	if err != nil {
		return nil, err
	}

	expectedVersion := user.Version
	if req.Version != nil {
		expectedVersion = *req.Version
	}

	currentRoles, err := uc.authRepo.GetUserRoles(ctx, userID, applicationID)
	if err != nil {
		return nil, err
	}

	currentRoleIDs := make(map[uuid.UUID]bool)
	for _, role := range currentRoles {
		currentRoleIDs[role.ID] = true
	}

	desiredRoleIDs := make(map[uuid.UUID]bool)
	for _, roleID := range req.RoleIDs {
		desiredRoleIDs[roleID] = true
	}

	var toAdd []uuid.UUID
	var toRemove []uuid.UUID

	for roleID := range desiredRoleIDs {
		if !currentRoleIDs[roleID] {
			toAdd = append(toAdd, roleID)
		}
	}

	for roleID := range currentRoleIDs {
		if !desiredRoleIDs[roleID] {
			toRemove = append(toRemove, roleID)
		}
	}

	if userID == currentUserID {
		for _, roleID := range toRemove {
			role, _ := uc.authRepo.GetRole(ctx, roleID)
			if role != nil && role.Code == "master" {
				return nil, errors.New("não é possível remover a role master de si mesmo")
			}
		}
	}

	for _, roleID := range req.RoleIDs {
		role, err := uc.authRepo.GetRole(ctx, roleID)
		if err != nil {
			return nil, errors.New("role não encontrada: " + roleID.String())
		}
		if role.ApplicationID != applicationID {
			return nil, errors.New("role não pertence à aplicação")
		}
	}

	userApplicationID, err := uc.authRepo.GetUserApplicationID(ctx, userID, applicationID)
	if err != nil {
		return nil, err
	}

	if err := uc.authRepo.UpdateUserRoles(ctx, userApplicationID, req.RoleIDs); err != nil {
		return nil, err
	}

	user.IncrementVersion()
	if err := uc.userRepo.UpdateWithVersion(ctx, user, expectedVersion); err != nil {
		if err.Error() == "versão do usuário está desatualizada (409 Conflict)" {
			return nil, errors.New("versão desatualizada: o usuário foi modificado por outro processo")
		}
		return nil, err
	}

	updatedRoles, _ := uc.authRepo.GetUserRoles(ctx, userID, applicationID)
	rolesDTO := make([]dto.RoleDTO, len(updatedRoles))
	for i, role := range updatedRoles {
		rolesDTO[i] = dto.RoleDTO{
			ID:          role.ID,
			Name:        role.Name,
			Code:        role.Code,
			Description: role.Description,
		}
	}

	user, _ = uc.userRepo.FindByIDAndApplication(ctx, userID, applicationID)
	return &dto.UpdateUserRolesResponse{
		UserID:    userID,
		Roles:     rolesDTO,
		Version:   user.Version,
		UpdatedAt: user.UpdatedAt,
	}, nil
}

// UpdateUserStatus atualiza o status do usuário na aplicação (ativa/desativa user_applications.active)
func (uc *UserManagementUseCase) UpdateUserStatus(ctx context.Context, userID, applicationID, currentUserID uuid.UUID, req dto.UpdateUserStatusRequest) (*dto.UserResponse, error) {
	user, err := uc.userRepo.FindByIDAndApplication(ctx, userID, applicationID)
	if err != nil {
		return nil, err
	}

	// Verifica se está tentando desativar
	if !req.Active {
		// Verifica se o usuário está tentando se desativar a si mesmo
		if userID == currentUserID {
			roles, err := uc.authRepo.GetUserRoles(ctx, userID, applicationID)
			if err == nil {
				for _, role := range roles {
					if role.Code == "master" {
						return nil, errors.New("não é possível desativar um usuário master, incluindo a si mesmo")
					}
				}
			}
			return nil, errors.New("não é possível se desativar a si mesmo")
		}
	}

	expectedVersion := user.Version
	if req.Version != nil {
		expectedVersion = *req.Version
	}

	if err := uc.userRepo.UpdateUserApplicationStatus(ctx, userID, applicationID, req.Active, expectedVersion); err != nil {
		if err.Error() == "versão do usuário está desatualizada (409 Conflict)" {
			return nil, errors.New("versão desatualizada: o usuário foi modificado por outro processo")
		}
		return nil, err
	}

	user, _ = uc.userRepo.FindByIDAndApplication(ctx, userID, applicationID)
	return &dto.UserResponse{
		ID:        user.ID,
		Email:     user.Email,
		Name:      user.Name,
		Active:    user.Active,
		Version:   user.Version,
		CreatedAt: user.CreatedAt,
		UpdatedAt: user.UpdatedAt,
	}, nil
}

// ChangePassword altera a senha do usuário (requer senha atual)
func (uc *UserManagementUseCase) ChangePassword(ctx context.Context, userID, applicationID uuid.UUID, req dto.ChangePasswordRequest) error {
	user, err := uc.userRepo.FindByIDAndApplication(ctx, userID, applicationID)
	if err != nil {
		return err
	}

	if err := uc.hashSvc.CheckPassword(user.Password, req.CurrentPassword); err != nil {
		return errors.New("senha atual incorreta")
	}

	if len(req.NewPassword) < 6 {
		return errors.New("senha deve ter no mínimo 6 caracteres")
	}

	expectedVersion := user.Version
	if req.Version != nil {
		expectedVersion = *req.Version
	}

	passwordHash, err := uc.hashSvc.HashPassword(req.NewPassword)
	if err != nil {
		return err
	}

	if err := uc.userRepo.UpdatePasswordWithVersion(ctx, userID, passwordHash, expectedVersion); err != nil {
		if err.Error() == "versão do usuário está desatualizada (409 Conflict)" {
			return errors.New("versão desatualizada: o usuário foi modificado por outro processo")
		}
		return err
	}

	return nil
}

// ResetPassword reseta a senha do usuário (admin/master)
func (uc *UserManagementUseCase) ResetPassword(ctx context.Context, userID, applicationID uuid.UUID, req dto.ResetPasswordRequest) error {
	user, err := uc.userRepo.FindByIDAndApplication(ctx, userID, applicationID)
	if err != nil {
		return err
	}

	if len(req.NewPassword) < 6 {
		return errors.New("senha deve ter no mínimo 6 caracteres")
	}

	expectedVersion := user.Version
	if req.Version != nil {
		expectedVersion = *req.Version
	}

	passwordHash, err := uc.hashSvc.HashPassword(req.NewPassword)
	if err != nil {
		return err
	}

	if err := uc.userRepo.UpdatePasswordWithVersion(ctx, userID, passwordHash, expectedVersion); err != nil {
		if err.Error() == "versão do usuário está desatualizada (409 Conflict)" {
			return errors.New("versão desatualizada: o usuário foi modificado por outro processo")
		}
		return err
	}

	return nil
}

