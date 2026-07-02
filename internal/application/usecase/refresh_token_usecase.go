package usecase

import (
	"context"
	"errors"

	"github.com/theretech/retech-auth-api/internal/application/service"
	"github.com/theretech/retech-auth-api/internal/domain/dto"
	"github.com/theretech/retech-auth-api/internal/domain/repository"
)

var (
	ErrInvalidRefreshToken = errors.New("refresh token inválido")
)

// RefreshTokenUseCase representa o caso de uso de renovação de token
type RefreshTokenUseCase struct {
	userRepo   repository.UserRepository
	authRepo   repository.AuthRepository
	jwtService service.JWTService
}

// NewRefreshTokenUseCase cria uma nova instância de RefreshTokenUseCase
func NewRefreshTokenUseCase(
	userRepo repository.UserRepository,
	authRepo repository.AuthRepository,
	jwtService service.JWTService,
) *RefreshTokenUseCase {
	return &RefreshTokenUseCase{
		userRepo:   userRepo,
		authRepo:   authRepo,
		jwtService: jwtService,
	}
}

// Execute executa a renovação do token
func (uc *RefreshTokenUseCase) Execute(ctx context.Context, req dto.RefreshTokenRequest) (*dto.AuthenticateResponse, error) {
	// Valida o refresh token
	claims, err := uc.jwtService.ValidateToken(req.RefreshToken)
	if err != nil {
		return nil, ErrInvalidRefreshToken
	}

	// Busca o usuário para verificar se ainda está ativo
	user, err := uc.userRepo.FindByID(ctx, claims.UserID)
	if err != nil {
		return nil, ErrInvalidRefreshToken
	}

	if !user.Active {
		return nil, ErrInactiveUser
	}

	// Busca tenant_id atual do user_application (reflete mudanças sem precisar relogar)
	// Fallback para claims se não encontrar o vínculo
	tenantID := claims.TenantID
	if userApp, err := uc.authRepo.FindUserApplication(ctx, user.ID, claims.ApplicationID); err == nil {
		tenantID = userApp.TenantID
	}

	// Busca as roles do usuário para incluir no novo token
	// Se falhar, usa roles do token antigo (fallback para compatibilidade)
	var roleCodes []string
	roles, err := uc.authRepo.GetUserRoles(ctx, user.ID, claims.ApplicationID)
	if err == nil && len(roles) > 0 {
		// Extrai apenas os codes das roles
		roleCodes = make([]string, 0, len(roles))
		for _, role := range roles {
			if role.Code != "" {
				roleCodes = append(roleCodes, role.Code)
			}
		}
	} else {
		// Fallback: usa roles do token antigo se não conseguir buscar do banco
		roleCodes = claims.Roles
	}

	// Permissions efetivas recalculadas a cada refresh — é assim que mudança de
	// grupo/permissão se propaga sem relogin. Fallback: perms do token antigo.
	var permCodes []string
	if permissions, permErr := uc.authRepo.GetUserPermissions(ctx, user.ID, claims.ApplicationID); permErr == nil {
		permCodes = buildPermCodes(roleCodes, permissions)
	} else {
		permCodes = claims.Perms
	}

	// Gera novos tokens incluindo tenant_id, roles, perms e name
	accessToken, err := uc.jwtService.GenerateAccessToken(user.ID, claims.ApplicationID, user.Email, user.Name, tenantID, roleCodes, permCodes)
	if err != nil {
		return nil, err
	}

	refreshToken, err := uc.jwtService.GenerateRefreshToken(user.ID, claims.ApplicationID, user.Email, user.Name, tenantID, roleCodes)
	if err != nil {
		return nil, err
	}

	return &dto.AuthenticateResponse{
		AccessToken:  accessToken,
		RefreshToken: refreshToken,
		TokenType:    "Bearer",
		ExpiresIn:    uc.jwtService.GetExpirationTime(),
		User: dto.UserDTO{
			ID:    user.ID,
			Email: user.Email,
			Name:  user.Name,
		},
	}, nil
}
