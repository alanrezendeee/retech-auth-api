package handler

import (
	"net/http"

	"github.com/theretech/retechauth-api/internal/application/service"
	"github.com/theretech/retechauth-api/internal/application/usecase"
	"github.com/theretech/retechauth-api/internal/domain/dto"
	"github.com/theretech/retechauth-api/internal/infrastructure/http/middleware"
	"github.com/gin-gonic/gin"
)

// AuthHandler gerencia as requisições de autenticação
type AuthHandler struct {
	authenticateUseCase *usecase.AuthenticateUseCase
	refreshTokenUseCase *usecase.RefreshTokenUseCase
	getUserInfoUseCase  *usecase.GetUserInfoUseCase
	jwtService          service.JWTService
}

// NewAuthHandler cria uma nova instância de AuthHandler
func NewAuthHandler(
	authenticateUseCase *usecase.AuthenticateUseCase,
	refreshTokenUseCase *usecase.RefreshTokenUseCase,
	getUserInfoUseCase *usecase.GetUserInfoUseCase,
	jwtService service.JWTService,
) *AuthHandler {
	return &AuthHandler{
		authenticateUseCase: authenticateUseCase,
		refreshTokenUseCase: refreshTokenUseCase,
		getUserInfoUseCase:  getUserInfoUseCase,
		jwtService:          jwtService,
	}
}

// Authenticate manipula a requisição de autenticação
func (h *AuthHandler) Authenticate(c *gin.Context) {
	var req dto.AuthenticateRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		respondWithError(c, http.StatusBadRequest, "Corpo da requisição inválido")
		return
	}

	if req.Email == "" || req.Password == "" || req.ApplicationCode == "" {
		respondWithError(c, http.StatusBadRequest, "Email, senha e código da aplicação são obrigatórios")
		return
	}

	response, err := h.authenticateUseCase.Execute(c.Request.Context(), req)
	if err != nil {
		switch err {
		case usecase.ErrInvalidCredentials:
			respondWithError(c, http.StatusUnauthorized, "Credenciais inválidas")
		case usecase.ErrInactiveUser:
			respondWithError(c, http.StatusForbidden, "Usuário inativo")
		case usecase.ErrInactiveApp:
			respondWithError(c, http.StatusForbidden, "Aplicação inativa")
		default:
			respondWithError(c, http.StatusInternalServerError, "Erro ao autenticar usuário")
		}
		return
	}

	respondWithJSON(c, http.StatusOK, response)
}

// RefreshToken manipula a requisição de renovação de token
func (h *AuthHandler) RefreshToken(c *gin.Context) {
	var req dto.RefreshTokenRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		respondWithError(c, http.StatusBadRequest, "Corpo da requisição inválido")
		return
	}

	if req.RefreshToken == "" {
		respondWithError(c, http.StatusBadRequest, "Refresh token é obrigatório")
		return
	}

	response, err := h.refreshTokenUseCase.Execute(c.Request.Context(), req)
	if err != nil {
		switch err {
		case usecase.ErrInvalidRefreshToken:
			respondWithError(c, http.StatusUnauthorized, "Refresh token inválido")
		case usecase.ErrInactiveUser:
			respondWithError(c, http.StatusForbidden, "Usuário inativo")
		default:
			respondWithError(c, http.StatusInternalServerError, "Erro ao renovar token")
		}
		return
	}

	respondWithJSON(c, http.StatusOK, response)
}

// Me manipula a requisição de informações do usuário autenticado
func (h *AuthHandler) Me(c *gin.Context) {
	userID, ok := middleware.GetUserID(c.Request.Context())
	if !ok {
		respondWithError(c, http.StatusUnauthorized, "Usuário não autenticado")
		return
	}

	applicationID, ok := middleware.GetApplicationID(c.Request.Context())
	if !ok {
		respondWithError(c, http.StatusUnauthorized, "Aplicação não identificada")
		return
	}

	response, err := h.getUserInfoUseCase.Execute(c.Request.Context(), userID, applicationID)
	if err != nil {
		switch err {
		case usecase.ErrUserNotFound:
			respondWithError(c, http.StatusNotFound, "Usuário não encontrado")
		default:
			respondWithError(c, http.StatusInternalServerError, "Erro ao buscar informações do usuário")
		}
		return
	}

	respondWithJSON(c, http.StatusOK, response)
}

// Health retorna o status de saúde da API
func (h *AuthHandler) Health(c *gin.Context) {
	respondWithJSON(c, http.StatusOK, map[string]string{
		"status":  "ok",
		"service": "retechauth-api",
		"version": "1.0.0",
	})
}

// JWKS retorna o JSON Web Key Set (JWKS) para validação de tokens JWT
// Endpoint padrão da indústria: GET /.well-known/jwks.json
func (h *AuthHandler) JWKS(c *gin.Context) {
	jwks, err := h.jwtService.GetJWKS()
	if err != nil {
		respondWithError(c, http.StatusInternalServerError, "Erro ao gerar JWKS")
		return
	}

	// JWKS deve retornar Content-Type: application/json
	c.Header("Content-Type", "application/json")
	c.JSON(http.StatusOK, jwks)
}
