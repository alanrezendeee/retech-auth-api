package usecase

import (
	"context"
	"errors"

	"github.com/theretech/retechauth-api/internal/application/service"
	"github.com/theretech/retechauth-api/internal/domain/dto"
	"github.com/theretech/retechauth-api/internal/domain/repository"
)

var (
	ErrInvalidCredentials = errors.New("credenciais inválidas")
	ErrInactiveUser       = errors.New("usuário inativo")
	ErrInactiveApp        = errors.New("aplicação inativa")
)

// AuthenticateUseCase representa o caso de uso de autenticação
type AuthenticateUseCase struct {
	authRepo    repository.AuthRepository
	hashService service.HashService
	jwtService  service.JWTService
}

// NewAuthenticateUseCase cria uma nova instância de AuthenticateUseCase
func NewAuthenticateUseCase(
	authRepo repository.AuthRepository,
	hashService service.HashService,
	jwtService service.JWTService,
) *AuthenticateUseCase {
	return &AuthenticateUseCase{
		authRepo:    authRepo,
		hashService: hashService,
		jwtService:  jwtService,
	}
}

// Execute executa a autenticação do usuário
func (uc *AuthenticateUseCase) Execute(ctx context.Context, req dto.AuthenticateRequest) (*dto.AuthenticateResponse, error) {
	// Busca o usuário por email e aplicação
	user, app, err := uc.authRepo.FindUserByEmailAndApplication(ctx, req.Email, req.ApplicationCode)
	if err != nil {
		return nil, ErrInvalidCredentials
	}

	// Verifica se o usuário está ativo
	if !user.Active {
		return nil, ErrInactiveUser
	}

	// Verifica se a aplicação está ativa
	if !app.Active {
		return nil, ErrInactiveApp
	}

	// Verifica a senha
	if err := uc.hashService.CheckPassword(user.Password, req.Password); err != nil {
		return nil, ErrInvalidCredentials
	}

	// Busca as roles do usuário para incluir no JWT
	roles, err := uc.authRepo.GetUserRoles(ctx, user.ID, app.ID)
	if err != nil {
		return nil, err
	}

	// Extrai apenas os codes das roles (formato leve para o JWT)
	roleCodes := make([]string, 0, len(roles))
	for _, role := range roles {
		if role.Code != "" {
			roleCodes = append(roleCodes, role.Code)
		}
	}

	// Gera os tokens incluindo tenant_id, roles e name (carregados do banco)
	accessToken, err := uc.jwtService.GenerateAccessToken(user.ID, app.ID, user.Email, user.Name, user.TenantID, roleCodes)
	if err != nil {
		return nil, err
	}

	refreshToken, err := uc.jwtService.GenerateRefreshToken(user.ID, app.ID, user.Email, user.Name, user.TenantID, roleCodes)
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
