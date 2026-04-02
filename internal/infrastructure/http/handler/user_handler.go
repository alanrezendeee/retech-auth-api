package handler

import (
	"context"
	"net/http"
	"strconv"

	"github.com/theretech/retechauth-api/internal/application/usecase"
	"github.com/theretech/retechauth-api/internal/domain/dto"
	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
)

// UserHandler gerencia requisições relacionadas a usuários
type UserHandler struct {
	listUseCase       *usecase.ListUsersUseCase
	managementUseCase *usecase.UserManagementUseCase
}

// NewUserHandler cria uma nova instância de UserHandler
func NewUserHandler(
	listUseCase *usecase.ListUsersUseCase,
	managementUseCase *usecase.UserManagementUseCase,
) *UserHandler {
	return &UserHandler{
		listUseCase:       listUseCase,
		managementUseCase: managementUseCase,
	}
}

func (h *UserHandler) isMasterUser(ctx context.Context, userID, applicationID uuid.UUID) bool {
	user, err := h.managementUseCase.GetUser(ctx, userID, applicationID)
	if err != nil {
		return false
	}
	for _, role := range user.Roles {
		if role.Code == "master" {
			return true
		}
	}
	return false
}

// ListUsers lista usuários da aplicação com filtros
func (h *UserHandler) ListUsers(c *gin.Context) {
	applicationID, err := getApplicationIDFromContext(c)
	if err != nil {
		respondWithError(c, http.StatusUnauthorized, "Aplicação não identificada")
		return
	}

	limit, _ := strconv.Atoi(c.Query("limit"))
	offset, _ := strconv.Atoi(c.Query("offset"))

	var active *bool
	if c.Query("active") != "" {
		activeVal := c.Query("active") == "true"
		active = &activeVal
	}

	req := dto.ListUsersRequest{
		Limit:  limit,
		Offset: offset,
		Email:  c.Query("email"),
		Name:   c.Query("name"),
		Active: active,
		Role:   c.Query("role"),
	}

	response, err := h.listUseCase.Execute(c.Request.Context(), applicationID, req)
	if err != nil {
		respondWithError(c, http.StatusInternalServerError, "Erro ao listar usuários")
		return
	}

	respondWithJSON(c, http.StatusOK, response)
}

// GetUser busca um usuário específico
func (h *UserHandler) GetUser(c *gin.Context) {
	applicationID, err := getApplicationIDFromContext(c)
	if err != nil {
		respondWithError(c, http.StatusUnauthorized, "Aplicação não identificada")
		return
	}

	userID, err := uuid.Parse(c.Param("id"))
	if err != nil {
		respondWithError(c, http.StatusBadRequest, "ID de usuário inválido")
		return
	}

	response, err := h.managementUseCase.GetUser(c.Request.Context(), userID, applicationID)
	if err != nil {
		respondWithError(c, http.StatusNotFound, "Usuário não encontrado")
		return
	}

	respondWithJSON(c, http.StatusOK, response)
}

// CreateUser cria um novo usuário
// O tenant_id do novo usuário é automaticamente definido pelo AUTH usando o tenant_id do JWT do criador
// Qualquer tenant_id enviado no request é ignorado (segurança)
func (h *UserHandler) CreateUser(c *gin.Context) {
	applicationID, err := getApplicationIDFromContext(c)
	if err != nil {
		respondWithError(c, http.StatusUnauthorized, "Aplicação não identificada")
		return
	}

	var req dto.CreateUserRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		respondWithError(c, http.StatusBadRequest, "Dados inválidos: "+err.Error())
		return
	}

	// O contexto já contém o tenant_id do criador (extraído do JWT pelo middleware)
	// O use case usa esse tenant_id para definir o tenant_id do novo usuário
	response, err := h.managementUseCase.CreateUser(c.Request.Context(), applicationID, req)
	if err != nil {
		respondWithError(c, http.StatusBadRequest, err.Error())
		return
	}

	respondWithJSON(c, http.StatusCreated, response)
}

// UpdateUser atualiza um usuário
func (h *UserHandler) UpdateUser(c *gin.Context) {
	applicationID, err := getApplicationIDFromContext(c)
	if err != nil {
		respondWithError(c, http.StatusUnauthorized, "Aplicação não identificada")
		return
	}

	userID, err := uuid.Parse(c.Param("id"))
	if err != nil {
		respondWithError(c, http.StatusBadRequest, "ID de usuário inválido")
		return
	}

	var req dto.UpdateUserRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		respondWithError(c, http.StatusBadRequest, "Dados inválidos")
		return
	}

	response, err := h.managementUseCase.UpdateUser(c.Request.Context(), userID, applicationID, req)
	if err != nil {
		if err.Error() == "versão desatualizada: o usuário foi modificado por outro processo" {
			respondWithError(c, http.StatusConflict, err.Error())
			return
		}
		respondWithError(c, http.StatusBadRequest, err.Error())
		return
	}

	respondWithJSON(c, http.StatusOK, response)
}

// DeleteUser remove um usuário (soft delete)
func (h *UserHandler) DeleteUser(c *gin.Context) {
	applicationID, err := getApplicationIDFromContext(c)
	if err != nil {
		respondWithError(c, http.StatusUnauthorized, "Aplicação não identificada")
		return
	}

	currentUserID, err := getUserIDFromContext(c)
	if err != nil {
		respondWithError(c, http.StatusUnauthorized, "Usuário não identificado")
		return
	}

	userID, err := uuid.Parse(c.Param("id"))
	if err != nil {
		respondWithError(c, http.StatusBadRequest, "ID de usuário inválido")
		return
	}

	// Validações adicionais no handler (mesmas regras do UpdateUserStatus)
	// Verifica se o usuário alvo é master
	targetIsMaster := h.isMasterUser(c.Request.Context(), userID, applicationID)
	if targetIsMaster {
		respondWithError(c, http.StatusForbidden, "Não é possível excluir um usuário master")
		return
	}

	if err := h.managementUseCase.DeleteUser(c.Request.Context(), userID, applicationID, currentUserID); err != nil {
		// Trata erros específicos com códigos HTTP apropriados
		errorMsg := err.Error()
		if errorMsg == "não é possível excluir um usuário master" || 
		   errorMsg == "não é possível excluir um usuário master, incluindo a si mesmo" {
			respondWithError(c, http.StatusForbidden, errorMsg)
			return
		}
		if errorMsg == "não é possível se excluir a si mesmo" {
			respondWithError(c, http.StatusBadRequest, errorMsg)
			return
		}
		respondWithError(c, http.StatusBadRequest, "Erro ao deletar usuário: "+errorMsg)
		return
	}

	c.Status(http.StatusNoContent)
}

// UpdateUserRoles atualiza as roles de um usuário
func (h *UserHandler) UpdateUserRoles(c *gin.Context) {
	applicationID, err := getApplicationIDFromContext(c)
	if err != nil {
		respondWithError(c, http.StatusUnauthorized, "Aplicação não identificada")
		return
	}

	currentUserID, err := getUserIDFromContext(c)
	if err != nil {
		respondWithError(c, http.StatusUnauthorized, "Usuário não identificado")
		return
	}

	userID, err := uuid.Parse(c.Param("id"))
	if err != nil {
		respondWithError(c, http.StatusBadRequest, "ID de usuário inválido")
		return
	}

	var req dto.UpdateUserRolesRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		respondWithError(c, http.StatusBadRequest, "Dados inválidos: "+err.Error())
		return
	}

	isMaster := h.isMasterUser(c.Request.Context(), currentUserID, applicationID)
	response, err := h.managementUseCase.UpdateUserRoles(c.Request.Context(), userID, applicationID, currentUserID, isMaster, req)
	if err != nil {
		if err.Error() == "versão desatualizada: o usuário foi modificado por outro processo" {
			respondWithError(c, http.StatusConflict, err.Error())
			return
		}
		respondWithError(c, http.StatusBadRequest, err.Error())
		return
	}

	respondWithJSON(c, http.StatusOK, response)
}

// UpdateUserStatus atualiza o status do usuário (ativa/desativa)
func (h *UserHandler) UpdateUserStatus(c *gin.Context) {
	applicationID, err := getApplicationIDFromContext(c)
	if err != nil {
		respondWithError(c, http.StatusUnauthorized, "Aplicação não identificada")
		return
	}

	currentUserID, err := getUserIDFromContext(c)
	if err != nil {
		respondWithError(c, http.StatusUnauthorized, "Usuário não identificado")
		return
	}

	userID, err := uuid.Parse(c.Param("id"))
	if err != nil {
		respondWithError(c, http.StatusBadRequest, "ID de usuário inválido")
		return
	}

	var req dto.UpdateUserStatusRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		respondWithError(c, http.StatusBadRequest, "Dados inválidos: "+err.Error())
		return
	}

	// Verifica se está tentando desativar (active = false)
	if !req.Active {
		// Verifica se o usuário alvo é master
		targetIsMaster := h.isMasterUser(c.Request.Context(), userID, applicationID)
		if targetIsMaster {
			respondWithError(c, http.StatusForbidden, "Não é possível desativar um usuário master")
			return
		}
	}

	response, err := h.managementUseCase.UpdateUserStatus(c.Request.Context(), userID, applicationID, currentUserID, req)
	if err != nil {
		if err.Error() == "Versão desatualizada: o usuário foi modificado por outro processo" {
			respondWithError(c, http.StatusConflict, err.Error())
			return
		}
		respondWithError(c, http.StatusBadRequest, err.Error())
		return
	}

	respondWithJSON(c, http.StatusOK, response)
}

// ChangePassword altera a senha do usuário (requer senha atual)
func (h *UserHandler) ChangePassword(c *gin.Context) {
	applicationID, err := getApplicationIDFromContext(c)
	if err != nil {
		respondWithError(c, http.StatusUnauthorized, "Aplicação não identificada")
		return
	}

	currentUserID, err := getUserIDFromContext(c)
	if err != nil {
		respondWithError(c, http.StatusUnauthorized, "Usuário não identificado")
		return
	}

	userID, err := uuid.Parse(c.Param("id"))
	if err != nil {
		respondWithError(c, http.StatusBadRequest, "ID de usuário inválido")
		return
	}

	if userID != currentUserID {
		respondWithError(c, http.StatusForbidden, "Você só pode alterar sua própria senha")
		return
	}

	var req dto.ChangePasswordRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		respondWithError(c, http.StatusBadRequest, "Dados inválidos: "+err.Error())
		return
	}

	if err := h.managementUseCase.ChangePassword(c.Request.Context(), userID, applicationID, req); err != nil {
		if err.Error() == "versão desatualizada: o usuário foi modificado por outro processo" {
			respondWithError(c, http.StatusConflict, err.Error())
			return
		}
		respondWithError(c, http.StatusBadRequest, err.Error())
		return
	}

	respondWithJSON(c, http.StatusOK, map[string]string{"message": "Senha alterada com sucesso"})
}

// ResetPassword reseta a senha do usuário (admin/master)
func (h *UserHandler) ResetPassword(c *gin.Context) {
	applicationID, err := getApplicationIDFromContext(c)
	if err != nil {
		respondWithError(c, http.StatusUnauthorized, "Aplicação não identificada")
		return
	}

	userID, err := uuid.Parse(c.Param("id"))
	if err != nil {
		respondWithError(c, http.StatusBadRequest, "ID de usuário inválido")
		return
	}

	var req dto.ResetPasswordRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		respondWithError(c, http.StatusBadRequest, "Dados inválidos: "+err.Error())
		return
	}

	if err := h.managementUseCase.ResetPassword(c.Request.Context(), userID, applicationID, req); err != nil {
		if err.Error() == "versão desatualizada: o usuário foi modificado por outro processo" {
			respondWithError(c, http.StatusConflict, err.Error())
			return
		}
		respondWithError(c, http.StatusBadRequest, err.Error())
		return
	}

	respondWithJSON(c, http.StatusOK, map[string]string{"message": "Senha resetada com sucesso"})
}
