package usecase

import (
	"context"
	"errors"
	"fmt"
	"regexp"
	"strings"
	"time"

	"github.com/google/uuid"
	"github.com/theretech/retech-auth-api/internal/application/service"
	"github.com/theretech/retech-auth-api/internal/domain/dto"
	"github.com/theretech/retech-auth-api/internal/domain/entity"
	"github.com/theretech/retech-auth-api/internal/domain/repository"
)

// ManagementUseCase consolida operações de gerenciamento
type ManagementUseCase struct {
	appRepo     repository.ApplicationRepository
	authRepo    repository.AuthRepository
	userRepo    repository.UserRepository
	hashService service.HashService
}

func NewManagementUseCase(
	appRepo repository.ApplicationRepository,
	authRepo repository.AuthRepository,
	userRepo repository.UserRepository,
	hashService service.HashService,
) *ManagementUseCase {
	return &ManagementUseCase{
		appRepo:     appRepo,
		authRepo:    authRepo,
		userRepo:    userRepo,
		hashService: hashService,
	}
}

// ==================== APPLICATIONS ====================

func (uc *ManagementUseCase) ListApplications(ctx context.Context) (*dto.ListApplicationsResponse, error) {
	apps, err := uc.appRepo.List(ctx, 1000, 0) // TODO: adicionar paginação
	if err != nil {
		return nil, err
	}

	items := make([]dto.ApplicationItemDTO, len(apps))
	for i, app := range apps {
		items[i] = dto.ApplicationItemDTO{
			ID:          app.ID,
			Name:        app.Name,
			Code:        app.Code,
			Description: app.Description,
			Active:      app.Active,
			CreatedAt:   app.CreatedAt,
			UpdatedAt:   app.UpdatedAt,
		}
	}

	return &dto.ListApplicationsResponse{
		Applications: items,
		Total:        len(items),
	}, nil
}

func (uc *ManagementUseCase) GetApplication(ctx context.Context, id uuid.UUID) (*dto.ApplicationItemDTO, error) {
	app, err := uc.appRepo.FindByID(ctx, id)
	if err != nil {
		return nil, err
	}

	return &dto.ApplicationItemDTO{
		ID:          app.ID,
		Name:        app.Name,
		Code:        app.Code,
		Description: app.Description,
		Active:      app.Active,
		CreatedAt:   app.CreatedAt,
		UpdatedAt:   app.UpdatedAt,
	}, nil
}

func (uc *ManagementUseCase) CreateApplication(ctx context.Context, req dto.CreateApplicationRequest) (*dto.ApplicationItemDTO, error) {
	if req.Name == "" || req.Code == "" {
		return nil, errors.New("name e code são obrigatórios")
	}

	app := entity.NewApplication(req.Name, req.Code, req.Description)
	if err := uc.appRepo.Create(ctx, app); err != nil {
		return nil, err
	}

	return &dto.ApplicationItemDTO{
		ID:          app.ID,
		Name:        app.Name,
		Code:        app.Code,
		Description: app.Description,
		Active:      app.Active,
		CreatedAt:   app.CreatedAt,
		UpdatedAt:   app.UpdatedAt,
	}, nil
}

func (uc *ManagementUseCase) UpdateApplication(ctx context.Context, id uuid.UUID, req dto.UpdateApplicationRequest) (*dto.ApplicationItemDTO, error) {
	app, err := uc.appRepo.FindByID(ctx, id)
	if err != nil {
		return nil, err
	}

	if req.Name != nil {
		app.Name = *req.Name
	}
	if req.Description != nil {
		app.Description = *req.Description
	}
	if req.Active != nil {
		app.Active = *req.Active
	}
	app.UpdatedAt = time.Now()

	if err := uc.appRepo.Update(ctx, app); err != nil {
		return nil, err
	}

	return &dto.ApplicationItemDTO{
		ID:          app.ID,
		Name:        app.Name,
		Code:        app.Code,
		Description: app.Description,
		Active:      app.Active,
		CreatedAt:   app.CreatedAt,
		UpdatedAt:   app.UpdatedAt,
	}, nil
}

func (uc *ManagementUseCase) DeleteApplication(ctx context.Context, id uuid.UUID) error {
	// Soft delete: apenas marca como inativa
	app, err := uc.appRepo.FindByID(ctx, id)
	if err != nil {
		return err
	}

	app.Active = false
	app.UpdatedAt = time.Now()

	return uc.appRepo.Update(ctx, app)
}

// ==================== ROLES ====================

func (uc *ManagementUseCase) ListRoles(ctx context.Context, applicationID uuid.UUID, req dto.ListRolesRequest) (*dto.ListRolesResponse, error) {
	roles, err := uc.authRepo.GetRolesByApplication(ctx, applicationID, req.Active)
	if err != nil {
		return nil, err
	}

	items := make([]dto.RoleItemDTO, len(roles))
	for i, role := range roles {
		// Buscar count de permissions por role (sempre busca para count)
		permissions, err := uc.authRepo.GetRolePermissions(ctx, role.ID)
		permissionCount := 0
		if err == nil {
			permissionCount = len(permissions)
		}

		roleItem := dto.RoleItemDTO{
			ID:              role.ID,
			Name:            role.Name,
			Code:            role.Code,
			Description:     role.Description,
			System:          role.System,
			Active:          role.Active,
			PermissionCount: permissionCount,
			CreatedAt:       role.CreatedAt,
			UpdatedAt:       role.UpdatedAt,
		}

		// Se include_permissions=true, incluir permissions no item
		if req.IncludePermissions {
			permissionsDTO := make([]dto.PermissionDTO, len(permissions))
			for j, perm := range permissions {
				permissionsDTO[j] = dto.PermissionDTO{
					ID:          perm.ID,
					Code:        perm.Code,
					Subject:     perm.Subject,
					Action:      perm.Action,
					Conditions:  perm.Conditions,
					Description: perm.Description,
				}
			}
			roleItem.Permissions = permissionsDTO
		}

		items[i] = roleItem
	}

	return &dto.ListRolesResponse{
		Roles: items,
		Total: len(items),
	}, nil
}

func (uc *ManagementUseCase) GetRole(ctx context.Context, roleID, applicationID uuid.UUID) (*dto.GetRoleResponse, error) {
	role, err := uc.authRepo.GetRole(ctx, roleID)
	if err != nil {
		return nil, err
	}

	if role.ApplicationID != applicationID {
		return nil, errors.New("role não pertence à aplicação")
	}

	// Buscar permissions da role
	permissions, err := uc.authRepo.GetRolePermissions(ctx, roleID)
	if err != nil {
		return nil, err
	}

	permsDTO := make([]dto.PermissionDTO, len(permissions))
	for i, perm := range permissions {
		permsDTO[i] = dto.PermissionDTO{
			ID:          perm.ID,
			Code:        perm.Code,
			Subject:     perm.Subject,
			Action:      perm.Action,
			Conditions:  perm.Conditions,
			Description: perm.Description,
		}
	}

	return &dto.GetRoleResponse{
		ID:          role.ID,
		Name:        role.Name,
		Code:        role.Code,
		Description: role.Description,
		System:      role.System,
		Active:      role.Active,
		Permissions: permsDTO,
		CreatedAt:   role.CreatedAt,
		UpdatedAt:   role.UpdatedAt,
	}, nil
}

func (uc *ManagementUseCase) CreateRole(ctx context.Context, applicationID uuid.UUID, req dto.CreateRoleRequest) (*dto.RoleItemDTO, error) {
	if req.Name == "" || req.Code == "" {
		return nil, errors.New("name e code são obrigatórios")
	}

	// Validar formato do code (slug-like: minúsculas, letras, números, hífens, underscores, pontos)
	if err := validateRoleCode(req.Code); err != nil {
		return nil, err
	}

	role := entity.NewRoleWithSystem(applicationID, req.Name, req.Code, req.Description, req.System)
	if err := uc.authRepo.CreateRole(ctx, role); err != nil {
		return nil, err
	}

	// Se permission_ids foram fornecidos, vincular permissões à role
	if len(req.PermissionIDs) > 0 {
		// Validar que as permissões pertencem à aplicação
		allPermissions, err := uc.authRepo.GetPermissionsByApplication(ctx, applicationID)
		if err != nil {
			return nil, err
		}

		// Criar mapa de permissions para validação
		permissionMap := make(map[uuid.UUID]bool)
		for _, perm := range allPermissions {
			permissionMap[perm.ID] = true
		}

		// Validar que todos os permission_ids pertencem à aplicação
		var validPermissionIDs []uuid.UUID
		for _, permID := range req.PermissionIDs {
			if permissionMap[permID] {
				validPermissionIDs = append(validPermissionIDs, permID)
			}
		}

		// Vincular permissões válidas
		if len(validPermissionIDs) > 0 {
			if err := uc.authRepo.UpsertRolePermissions(ctx, role.ID, validPermissionIDs); err != nil {
				return nil, fmt.Errorf("erro ao vincular permissões à role: %w", err)
			}
		}
	}

	// Buscar count de permissões para retornar na resposta
	permissions, err := uc.authRepo.GetRolePermissions(ctx, role.ID)
	if err != nil {
		// Se der erro, retorna 0 mas não falha a criação
		permissions = []*entity.Permission{}
	}

	return &dto.RoleItemDTO{
		ID:              role.ID,
		Name:            role.Name,
		Code:            role.Code,
		Description:     role.Description,
		System:          role.System,
		Active:          role.Active,
		PermissionCount: len(permissions),
		CreatedAt:       role.CreatedAt,
		UpdatedAt:       role.UpdatedAt,
	}, nil
}

func (uc *ManagementUseCase) UpdateRole(ctx context.Context, roleID, applicationID uuid.UUID, req dto.UpdateRoleRequest) (*dto.RoleItemDTO, error) {
	role, err := uc.authRepo.GetRole(ctx, roleID)
	if err != nil {
		return nil, err
	}

	if role.ApplicationID != applicationID {
		return nil, errors.New("role não pertence à aplicação")
	}

	// Protege roles base (system: true) de edição
	if role.System {
		return nil, errors.New("não é possível editar uma role base do sistema (system: true)")
	}

	if req.Name != nil {
		role.Name = *req.Name
	}
	if req.Description != nil {
		role.Description = *req.Description
	}
	if req.Active != nil {
		role.Active = *req.Active
	}
	role.UpdatedAt = time.Now()

	if err := uc.authRepo.UpdateRole(ctx, role); err != nil {
		return nil, err
	}

	// Se permission_ids foram fornecidos, atualizar vínculos
	if req.PermissionIDs != nil {
		// Buscar todas as permissions da aplicação para validação
		allPermissions, err := uc.authRepo.GetPermissionsByApplication(ctx, applicationID)
		if err != nil {
			return nil, err
		}

		// Criar mapa de permissions para validação
		permissionMap := make(map[uuid.UUID]bool)
		for _, perm := range allPermissions {
			permissionMap[perm.ID] = true
		}

		// Validar que todos os permission_ids pertencem à aplicação
		var validPermissionIDs []uuid.UUID
		for _, permID := range req.PermissionIDs {
			if permissionMap[permID] {
				validPermissionIDs = append(validPermissionIDs, permID)
			}
		}

		// Atualizar vínculos (passar array vazio remove todas as permissões)
		if err := uc.authRepo.UpsertRolePermissions(ctx, roleID, validPermissionIDs); err != nil {
			return nil, fmt.Errorf("erro ao atualizar permissões da role: %w", err)
		}
	}

	// Buscar count de permissões para retornar na resposta
	permissions, err := uc.authRepo.GetRolePermissions(ctx, roleID)
	permissionCount := 0
	if err == nil {
		permissionCount = len(permissions)
	}

	return &dto.RoleItemDTO{
		ID:              role.ID,
		Name:            role.Name,
		Code:            role.Code,
		Description:     role.Description,
		System:          role.System,
		Active:          role.Active,
		PermissionCount: permissionCount,
		CreatedAt:       role.CreatedAt,
		UpdatedAt:       role.UpdatedAt,
	}, nil
}

func (uc *ManagementUseCase) DeleteRole(ctx context.Context, roleID, applicationID uuid.UUID) error {
	role, err := uc.authRepo.GetRole(ctx, roleID)
	if err != nil {
		return err
	}

	if role.ApplicationID != applicationID {
		return errors.New("role não pertence à aplicação")
	}

	// Protege roles base (system: true) de deleção
	if role.System {
		return errors.New("não é possível deletar uma role base do sistema (system: true)")
	}

	// Verifica se há usuários ativos com esta role
	activeUsers, err := uc.authRepo.GetActiveUsersByRole(ctx, roleID, applicationID)
	if err != nil {
		return err
	}

	if len(activeUsers) > 0 {
		// Lista emails dos usuários ativos
		userEmails := make([]string, len(activeUsers))
		for i, user := range activeUsers {
			userEmails[i] = user.Email
		}
		return fmt.Errorf(
			"não é possível deletar a role pois ela está vinculada a %d usuário(s) ativo(s): %v",
			len(activeUsers),
			userEmails,
		)
	}

	role.Active = false
	role.UpdatedAt = time.Now()

	if err := uc.authRepo.UpdateRole(ctx, role); err != nil {
		return err
	}

	// Ao deletar role (soft delete), desvincular todas as permissões
	// Remove todos os vínculos role_permissions para esta role
	if err := uc.authRepo.UpsertRolePermissions(ctx, roleID, []uuid.UUID{}); err != nil {
		// Log erro mas não falha a deleção (role já foi desativada)
		// Se der erro aqui, a role fica desativada mas com permissões vinculadas
		// Isso é aceitável pois role inativa não é retornada em queries normais
	}

	return nil
}

// UpdateRolePermissions atualiza as permissões de uma role (apenas roles customizadas, system=false)
func (uc *ManagementUseCase) UpdateRolePermissions(ctx context.Context, roleID, applicationID uuid.UUID, req dto.UpdateRolePermissionsRequest) (*dto.UpdateRolePermissionsResponse, error) {
	role, err := uc.authRepo.GetRole(ctx, roleID)
	if err != nil {
		return nil, err
	}

	if role.ApplicationID != applicationID {
		return nil, errors.New("role não pertence à aplicação")
	}

	// Protege roles base (system: true) de edição de permissões
	if role.System {
		return nil, errors.New("não é possível atualizar permissões de uma role base do sistema (system: true)")
	}

	// Buscar todas as permissions da aplicação (para wildcards e validação)
	allPermissions, err := uc.authRepo.GetPermissionsByApplication(ctx, applicationID)
	if err != nil {
		return nil, err
	}

	// Criar map de permissions por code para facilitar busca
	permissionMap := make(map[string]*entity.Permission)
	for _, perm := range allPermissions {
		permissionMap[perm.Code] = perm
	}

	// Resolver permission IDs (suporta wildcards)
	permissionIDs := uc.resolvePermissionIDs(req.PermissionCodes, permissionMap, allPermissions)

	// Atualizar vínculos role_permissions
	if err := uc.authRepo.UpsertRolePermissions(ctx, roleID, permissionIDs); err != nil {
		return nil, err
	}

	// Buscar permissions atualizadas para retornar
	updatedPermissions, err := uc.authRepo.GetRolePermissions(ctx, roleID)
	if err != nil {
		return nil, err
	}

	permissionsDTO := make([]dto.PermissionDTO, len(updatedPermissions))
	for i, perm := range updatedPermissions {
		permissionsDTO[i] = dto.PermissionDTO{
			ID:          perm.ID,
			Code:        perm.Code,
			Subject:     perm.Subject,
			Action:      perm.Action,
			Conditions:  perm.Conditions,
			Description: perm.Description,
		}
	}

	// NOTA: Revogação de tokens
	// JWT é stateless e não há blacklist implementada.
	// Tokens ativos continuarão válidos até expirarem.
	// Para revogação imediata, seria necessário implementar blacklist de tokens
	// ou adicionar version/timestamp nos tokens (requer mudança na arquitetura JWT).

	return &dto.UpdateRolePermissionsResponse{
		RoleID:      roleID,
		Permissions: permissionsDTO,
	}, nil
}

// ==================== PERMISSIONS ====================

func (uc *ManagementUseCase) ListPermissions(ctx context.Context, applicationID uuid.UUID) (*dto.ListPermissionsResponse, error) {
	permissions, err := uc.authRepo.GetPermissionsByApplication(ctx, applicationID)
	if err != nil {
		return nil, err
	}

	items := make([]dto.PermissionItemDTO, len(permissions))
	for i, perm := range permissions {
		items[i] = dto.PermissionItemDTO{
			ID:          perm.ID,
			Code:        perm.Code,
			Subject:     perm.Subject,
			Action:      perm.Action,
			Description: perm.Description,
			Active:      perm.Active,
			CreatedAt:   perm.CreatedAt,
			UpdatedAt:   perm.UpdatedAt,
		}
	}

	return &dto.ListPermissionsResponse{
		Permissions: items,
		Total:       len(items),
	}, nil
}

func (uc *ManagementUseCase) CreatePermission(ctx context.Context, applicationID uuid.UUID, req dto.CreatePermissionRequest) (*dto.PermissionItemDTO, error) {
	// Permissions só podem ser criadas via sync (manifest)
	return nil, errors.New("permissions devem ser criadas apenas via sync de manifest (POST /applications/sync). Permissions são sempre definidas pela aplicação, nunca pelo cliente")
}

func (uc *ManagementUseCase) UpdatePermission(ctx context.Context, permID, applicationID uuid.UUID, req dto.UpdatePermissionRequest) (*dto.PermissionItemDTO, error) {
	perm, err := uc.authRepo.GetPermission(ctx, permID)
	if err != nil {
		return nil, err
	}

	if perm.ApplicationID != applicationID {
		return nil, errors.New("permission não pertence à aplicação")
	}

	if req.Description != nil {
		perm.Description = *req.Description
	}
	if req.Conditions != nil {
		perm.Conditions = req.Conditions
	}
	if req.Active != nil {
		perm.Active = *req.Active
	}
	perm.UpdatedAt = time.Now()

	if err := uc.authRepo.UpdatePermission(ctx, perm); err != nil {
		return nil, err
	}

	return &dto.PermissionItemDTO{
		ID:          perm.ID,
		Code:        perm.Code,
		Subject:     perm.Subject,
		Action:      perm.Action,
		Description: perm.Description,
		Active:      perm.Active,
		CreatedAt:   perm.CreatedAt,
		UpdatedAt:   perm.UpdatedAt,
	}, nil
}

func (uc *ManagementUseCase) DeletePermission(ctx context.Context, permID, applicationID uuid.UUID) error {
	perm, err := uc.authRepo.GetPermission(ctx, permID)
	if err != nil {
		return err
	}

	if perm.ApplicationID != applicationID {
		return errors.New("permission não pertence à aplicação")
	}

	perm.Active = false
	perm.UpdatedAt = time.Now()

	return uc.authRepo.UpdatePermission(ctx, perm)
}

// ==================== SYNC / MANIFEST ====================

func (uc *ManagementUseCase) SyncManifest(ctx context.Context, req dto.SyncManifestRequest) (*dto.SyncManifestResponse, error) {
	response := &dto.SyncManifestResponse{
		Permissions: []dto.SyncResultDTO{},
		Roles:       []dto.SyncRoleResultDTO{},
		Users:       []dto.SyncResultDTO{},
	}

	// 1. Upsert Application
	app := entity.NewApplication(req.Application.Name, req.Application.Code, req.Application.Description)
	existingApp, err := uc.appRepo.FindByCode(ctx, req.Application.Code)
	if err != nil {
		app.ID = uuid.New()
		app.CreatedAt = time.Now()
		app.UpdatedAt = time.Now()
		if err := uc.appRepo.UpsertByCode(ctx, app); err != nil {
			return nil, err
		}
		response.Application = dto.SyncResultDTO{
			Code:   app.Code,
			Action: "created",
			ID:     app.ID,
		}
	} else {
		app.ID = existingApp.ID
		app.Active = existingApp.Active
		app.UpdatedAt = time.Now()
		if err := uc.appRepo.UpsertByCode(ctx, app); err != nil {
			return nil, err
		}
		response.Application = dto.SyncResultDTO{
			Code:   app.Code,
			Action: "updated",
			ID:     app.ID,
		}
	}

	// 2. Upsert Permissions
	// Extrair prefixo da aplicação (ex: "retech-fin-admin" → "admin")
	appPrefix := extractAppPrefix(app.Code)

	permissionMap := make(map[string]*entity.Permission)
	for _, permDTO := range req.Permissions {
		// Parse code para subject.action se não vier separado
		// Valida e infere subject/action para compatibilidade com CASL.js
		// Subject deve incluir namespace da aplicação (ex: "admin.users")
		subject, action, err := uc.parsePermissionCode(permDTO.Code, permDTO.Subject, permDTO.Action, appPrefix)
		if err != nil {
			return nil, fmt.Errorf("erro ao processar permission '%s': %w", permDTO.Code, err)
		}

		perm := entity.NewPermissionWithCode(
			app.ID,
			permDTO.Code,
			subject,
			action,
			permDTO.Description,
			permDTO.Conditions,
		)

		existingPerm, err := uc.authRepo.GetPermissionByCode(ctx, app.ID, permDTO.Code)
		if err != nil {
			// Fallback: rows criadas fora do sync (SQL manual) não têm code —
			// busca pela chave natural para não violar (application_id, subject, action).
			existingPerm, err = uc.authRepo.GetPermissionBySubjectAction(ctx, app.ID, subject, action)
		}
		if err != nil {
			perm.ID = uuid.New()
			perm.CreatedAt = time.Now()
			perm.UpdatedAt = time.Now()
			if err := uc.authRepo.UpsertPermission(ctx, perm); err != nil {
				return nil, err
			}
			response.Permissions = append(response.Permissions, dto.SyncResultDTO{
				Code:   perm.Code,
				Action: "created",
				ID:     perm.ID,
			})
		} else {
			perm.ID = existingPerm.ID
			perm.Active = existingPerm.Active
			perm.UpdatedAt = time.Now()
			if err := uc.authRepo.UpsertPermission(ctx, perm); err != nil {
				return nil, err
			}
			response.Permissions = append(response.Permissions, dto.SyncResultDTO{
				Code:   perm.Code,
				Action: "updated",
				ID:     perm.ID,
			})
		}

		permissionMap[perm.Code] = perm
	}

	// 3. Buscar todas as permissions da aplicação (para wildcards)
	allPermissions, err := uc.authRepo.GetPermissionsByApplication(ctx, app.ID)
	if err != nil {
		return nil, err
	}

	// 4. Upsert Roles e seus vínculos
	for _, roleDTO := range req.Roles {
		role := entity.NewRoleWithSystem(app.ID, roleDTO.Name, roleDTO.Code, roleDTO.Description, roleDTO.System)

		existingRole, err := uc.authRepo.GetRoleByCode(ctx, app.ID, roleDTO.Code)
		if err != nil {
			role.ID = uuid.New()
			role.CreatedAt = time.Now()
			role.UpdatedAt = time.Now()
			if err := uc.authRepo.UpsertRole(ctx, role); err != nil {
				return nil, err
			}
			response.Roles = append(response.Roles, dto.SyncRoleResultDTO{
				Code:        role.Code,
				Action:      "created",
				ID:          role.ID,
				Permissions: []dto.SyncResultDTO{},
			})
		} else {
			role.ID = existingRole.ID
			role.Active = existingRole.Active
			role.UpdatedAt = time.Now()
			if err := uc.authRepo.UpsertRole(ctx, role); err != nil {
				return nil, err
			}
			response.Roles = append(response.Roles, dto.SyncRoleResultDTO{
				Code:        role.Code,
				Action:      "updated",
				ID:          role.ID,
				Permissions: []dto.SyncResultDTO{},
			})
		}

		// 5. Expandir wildcards e resolver permission IDs
		permissionIDs := uc.resolvePermissionIDs(roleDTO.Permissions, permissionMap, allPermissions)

		// 6. Regenerar role_permissions
		if err := uc.authRepo.UpsertRolePermissions(ctx, role.ID, permissionIDs); err != nil {
			return nil, err
		}

		// Adicionar informações de permissions no response
		roleResult := &response.Roles[len(response.Roles)-1]
		for _, permID := range permissionIDs {
			for _, perm := range allPermissions {
				if perm.ID == permID {
					roleResult.Permissions = append(roleResult.Permissions, dto.SyncResultDTO{
						Code:   perm.Code,
						Action: "linked",
					})
					break
				}
			}
		}
	}

	// 7. Upsert Users (se fornecido no manifest)
	if len(req.Users) > 0 {
		// Criar mapa de roles por código para buscar IDs rapidamente
		// Buscar todas as roles da aplicação (incluindo as que já existiam antes do sync)
		// Passa nil para active para buscar todas (ativas e inativas)
		allRoles, err := uc.authRepo.GetRolesByApplication(ctx, app.ID, nil)
		if err != nil {
			// Log erro mas continua sem processar usuários
			return response, nil
		}

		roleMap := make(map[string]uuid.UUID)
		for _, role := range allRoles {
			roleMap[role.Code] = role.ID
		}

		for _, userDTO := range req.Users {
			// Verificar se usuário já existe
			existingUser, err := uc.userRepo.FindByEmail(ctx, userDTO.Email)

			var user *entity.User
			var action string

			if err != nil {
				// Não existe, criar novo
				passwordHash, err := uc.hashService.HashPassword(userDTO.Password)
				if err != nil {
					// Log erro mas continua com próximo usuário
					continue
				}

				user = entity.NewUser(userDTO.Email, passwordHash, userDTO.Name)
				if userDTO.Active {
					user.Active = true
				}
				user.ID = uuid.New()
				user.CreatedAt = time.Now()
				user.UpdatedAt = time.Now()

				if err := uc.userRepo.Create(ctx, user); err != nil {
					// Log erro mas continua
					continue
				}
				action = "created"
			} else {
				// Existe, atualizar apenas dados básicos (NÃO atualiza senha)
				user = existingUser
				user.Name = userDTO.Name
				if userDTO.Active {
					user.Active = true
				}
				user.UpdatedAt = time.Now()

				if err := uc.userRepo.Update(ctx, user); err != nil {
					// Log erro mas continua
					continue
				}
				action = "updated"
			}

			// 8. Criar/atualizar vínculo user_application
			var userAppID uuid.UUID
			existingUserAppID, err := uc.authRepo.GetUserApplicationID(ctx, user.ID, app.ID)
			if err != nil {
				// Não existe, criar
				userApp := entity.NewUserApplication(user.ID, app.ID)
				userApp.TenantID = userDTO.TenantID // tenant_id no vínculo, não no usuário
				if userDTO.Active {
					userApp.Active = true
				}
				if err := uc.authRepo.CreateUserApplication(ctx, userApp); err != nil {
					// Log erro mas continua
					continue
				}
				userAppID = userApp.ID
			} else {
				userAppID = existingUserAppID
			}

			// 9. Resolver role codes para IDs e atualizar user_roles
			var roleIDs []uuid.UUID
			for _, roleCode := range userDTO.Roles {
				if roleID, exists := roleMap[roleCode]; exists {
					roleIDs = append(roleIDs, roleID)
				}
			}

			if len(roleIDs) > 0 {
				if err := uc.authRepo.UpdateUserRoles(ctx, userAppID, roleIDs); err != nil {
					// Log erro mas continua
					continue
				}
			}

			response.Users = append(response.Users, dto.SyncResultDTO{
				Code:   user.Email,
				Action: action,
				ID:     user.ID,
			})
		}
	}

	return response, nil
}

// extractAppPrefix extrai o prefixo da aplicação (ex: "retech-fin-admin" → "admin")
func extractAppPrefix(appCode string) string {
	parts := strings.Split(appCode, "-")
	if len(parts) > 1 {
		// Se tem hífen, pegar última parte
		return parts[len(parts)-1]
	}
	// Se não tem hífen, usar como está
	return appCode
}

// parsePermissionCode extrai subject e action de um code para compatibilidade com CASL.js
// Prioridade: 1) subject/action fornecidos explicitamente, 2) inferência do code, 3) fallback
// Actions válidos CASL.js: read, create, update, delete, manage, view
// Subject deve incluir namespace da aplicação (ex: "admin.users")
func (uc *ManagementUseCase) parsePermissionCode(code, providedSubject, providedAction, appPrefix string) (string, string, error) {
	// Actions válidos para CASL.js
	validActions := map[string]bool{
		"read":   true,
		"create": true,
		"update": true,
		"delete": true,
		"manage": true,
		"view":   true,
	}

	// 1. Se fornecido explicitamente, validar e usar (prioridade máxima)
	if providedSubject != "" && providedAction != "" {
		if !validActions[providedAction] {
			return "", "", fmt.Errorf(
				"action '%s' inválido para CASL.js. Actions válidos: read, create, update, delete, manage, view",
				providedAction,
			)
		}
		// Garantir que subject inclui namespace se não tiver
		// Ex: "users" → "admin.users"
		// Ex: "admin.users" → mantém como está
		finalSubject := providedSubject
		if appPrefix != "" && !strings.Contains(finalSubject, ".") && !strings.HasPrefix(finalSubject, "Menu:") {
			finalSubject = appPrefix + "." + finalSubject
		}
		return finalSubject, providedAction, nil
	}

	// 2. Tentar inferir do code
	// Padrão Menu:{Nome} → subject="Menu:{Nome}", action="view"
	if strings.HasPrefix(code, "Menu:") {
		return code, "view", nil
	}

	// 3. Padrão simples: {subject}.{action} (ex: "Device.read")
	// Neste caso, se não tem namespace, adicionar
	parts := strings.SplitN(code, ".", 2)
	if len(parts) == 2 {
		possibleSubject := parts[0]
		possibleAction := parts[1]

		// Se action é válido CASL, usar diretamente
		if validActions[possibleAction] {
			// Se não tem namespace e não é Menu, adicionar
			if appPrefix != "" && !strings.HasPrefix(possibleSubject, "Menu:") {
				possibleSubject = appPrefix + "." + possibleSubject
			}
			return possibleSubject, possibleAction, nil
		}
	}

	// 4. Padrão complexo: {prefixo}.{recurso}.{action} (ex: "admin.users.read")
	// Tentar extrair da última parte do code
	lastDotIndex := strings.LastIndex(code, ".")
	if lastDotIndex > 0 {
		possibleAction := code[lastDotIndex+1:]

		// Se a última parte é um action válido, extrair subject
		if validActions[possibleAction] {
			// Pegar tudo antes do último ponto (ex: "admin.users" → subject completo)
			beforeLastDot := code[:lastDotIndex]

			// Verificar se já tem namespace da aplicação
			if strings.HasPrefix(beforeLastDot, appPrefix+".") {
				// Já tem namespace, usar como está (ex: "admin.users")
				return beforeLastDot, possibleAction, nil
			}

			// Se não tem namespace, adicionar
			// Ex: "users" → "admin.users"
			if appPrefix != "" {
				return appPrefix + "." + beforeLastDot, possibleAction, nil
			}

			// Fallback: usar tudo antes do último ponto
			return beforeLastDot, possibleAction, nil
		}
	}

	// 5. Fallback: se code não tem ".", usar code como subject e "manage" como action
	if !strings.Contains(code, ".") {
		// Capitalizar primeira letra
		subject := code
		if len(subject) > 0 {
			subject = strings.ToUpper(subject[:1]) + subject[1:]
		}
		return subject, "manage", nil
	}

	// 6. Se não conseguiu inferir, retornar erro pedindo subject/action explícitos
	return "", "", fmt.Errorf(
		"não foi possível inferir subject e action do code '%s'. "+
			"Para compatibilidade com CASL.js, forneça 'subject' e 'action' explicitamente no manifest. "+
			"Actions válidos: read, create, update, delete, manage, view",
		code,
	)
}

// resolvePermissionIDs resolve uma lista de codes de permissions (incluindo wildcards) para UUIDs
// Genérico: funciona com qualquer formato de código de permission e qualquer prefixo de wildcard
func (uc *ManagementUseCase) resolvePermissionIDs(
	permissionCodes []string,
	permissionMap map[string]*entity.Permission,
	allPermissions []*entity.Permission,
) []uuid.UUID {
	var permissionIDs []uuid.UUID
	added := make(map[uuid.UUID]bool)

	for _, permCode := range permissionCodes {
		if strings.HasSuffix(permCode, ".*") {
			// Wildcard: expande para todas as permissions que começam com o prefixo (qualquer prefixo)
			prefix := strings.TrimSuffix(permCode, ".*")
			for _, perm := range allPermissions {
				if strings.HasPrefix(perm.Code, prefix+".") && !added[perm.ID] {
					permissionIDs = append(permissionIDs, perm.ID)
					added[perm.ID] = true
				}
			}
		} else {
			// Permission específica (qualquer formato)
			if perm, exists := permissionMap[permCode]; exists {
				if !added[perm.ID] {
					permissionIDs = append(permissionIDs, perm.ID)
					added[perm.ID] = true
				}
			} else {
				// Tenta buscar nas permissions existentes (busca exata por code)
				for _, perm := range allPermissions {
					if perm.Code == permCode && !added[perm.ID] {
						permissionIDs = append(permissionIDs, perm.ID)
						added[perm.ID] = true
						break
					}
				}
			}
		}
	}

	return permissionIDs
}

// validateRoleCode valida o formato do code de uma role
// Formato esperado: slug-like (minúsculas, letras, números, hífens, underscores, pontos)
// Exemplos válidos: "master", "admin", "admin.operator", "ops-manager"
// Não permite: espaços, caracteres especiais, acentos, maiúsculas
func validateRoleCode(code string) error {
	if code == "" {
		return errors.New("code não pode ser vazio")
	}

	// Comprimento mínimo e máximo
	if len(code) < 2 {
		return errors.New("code deve ter no mínimo 2 caracteres")
	}
	if len(code) > 100 {
		return errors.New("code deve ter no máximo 100 caracteres")
	}

	// Verificar se tem espaços
	if strings.Contains(code, " ") {
		return errors.New("code não pode conter espaços")
	}

	// Verificar se começa ou termina com hífen, underscore ou ponto
	if strings.HasPrefix(code, "-") || strings.HasPrefix(code, "_") || strings.HasPrefix(code, ".") {
		return errors.New("code não pode começar com hífen, underscore ou ponto")
	}
	if strings.HasSuffix(code, "-") || strings.HasSuffix(code, "_") || strings.HasSuffix(code, ".") {
		return errors.New("code não pode terminar com hífen, underscore ou ponto")
	}

	// Regex para validar formato: letras minúsculas, números, hífens, underscores, pontos
	// Não permite maiúsculas, espaços, caracteres especiais, acentos
	validPattern := regexp.MustCompile(`^[a-z0-9._-]+$`)
	if !validPattern.MatchString(code) {
		return errors.New("code deve conter apenas letras minúsculas, números, hífens (-), underscores (_) e pontos (.)")
	}

	// Verificar se não tem sequências duplicadas de hífens, underscores ou pontos
	if strings.Contains(code, "--") || strings.Contains(code, "__") || strings.Contains(code, "..") {
		return errors.New("code não pode conter sequências duplicadas de hífens, underscores ou pontos")
	}

	return nil
}
