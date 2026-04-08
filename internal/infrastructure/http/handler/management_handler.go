package handler

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"github.com/theretech/retech-auth-api/internal/application/usecase"
	"github.com/theretech/retech-auth-api/internal/domain/dto"
)

// ManagementHandler gerencia operações de applications, roles e permissions
type ManagementHandler struct {
	managementUseCase *usecase.ManagementUseCase
}

func NewManagementHandler(managementUseCase *usecase.ManagementUseCase) *ManagementHandler {
	return &ManagementHandler{managementUseCase: managementUseCase}
}

// ==================== APPLICATIONS ====================

func (h *ManagementHandler) ListApplications(c *gin.Context) {
	response, err := h.managementUseCase.ListApplications(c.Request.Context())
	if err != nil {
		respondWithError(c, http.StatusInternalServerError, "Erro ao listar aplicações")
		return
	}
	respondWithJSON(c, http.StatusOK, response)
}

func (h *ManagementHandler) GetApplication(c *gin.Context) {
	id, err := uuid.Parse(c.Param("id"))
	if err != nil {
		respondWithError(c, http.StatusBadRequest, "ID inválido")
		return
	}

	response, err := h.managementUseCase.GetApplication(c.Request.Context(), id)
	if err != nil {
		respondWithError(c, http.StatusNotFound, "Aplicação não encontrada")
		return
	}
	respondWithJSON(c, http.StatusOK, response)
}

func (h *ManagementHandler) CreateApplication(c *gin.Context) {
	var req dto.CreateApplicationRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		respondWithError(c, http.StatusBadRequest, "Dados inválidos")
		return
	}

	response, err := h.managementUseCase.CreateApplication(c.Request.Context(), req)
	if err != nil {
		respondWithError(c, http.StatusBadRequest, "Erro ao criar aplicação")
		return
	}
	respondWithJSON(c, http.StatusCreated, response)
}

func (h *ManagementHandler) UpdateApplication(c *gin.Context) {
	id, err := uuid.Parse(c.Param("id"))
	if err != nil {
		respondWithError(c, http.StatusBadRequest, "ID inválido")
		return
	}

	var req dto.UpdateApplicationRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		respondWithError(c, http.StatusBadRequest, "Dados inválidos")
		return
	}

	response, err := h.managementUseCase.UpdateApplication(c.Request.Context(), id, req)
	if err != nil {
		respondWithError(c, http.StatusBadRequest, "Erro ao atualizar aplicação")
		return
	}
	respondWithJSON(c, http.StatusOK, response)
}

func (h *ManagementHandler) DeleteApplication(c *gin.Context) {
	id, err := uuid.Parse(c.Param("id"))
	if err != nil {
		respondWithError(c, http.StatusBadRequest, "ID inválido")
		return
	}

	if err := h.managementUseCase.DeleteApplication(c.Request.Context(), id); err != nil {
		respondWithError(c, http.StatusBadRequest, "Erro ao deletar aplicação")
		return
	}
	c.Status(http.StatusNoContent)
}

// ==================== ROLES ====================

func (h *ManagementHandler) ListRoles(c *gin.Context) {
	applicationID, err := getApplicationIDFromContext(c)
	if err != nil {
		respondWithError(c, http.StatusUnauthorized, "Aplicação não identificada")
		return
	}

	// Ler query parameters
	var active *bool
	if c.Query("active") != "" {
		activeVal := c.Query("active") == "true"
		active = &activeVal
	}

	includePermissions := c.Query("include_permissions") == "true"

	req := dto.ListRolesRequest{
		Active:             active,
		IncludePermissions: includePermissions,
	}

	response, err := h.managementUseCase.ListRoles(c.Request.Context(), applicationID, req)
	if err != nil {
		respondWithError(c, http.StatusInternalServerError, "Erro ao listar roles")
		return
	}
	respondWithJSON(c, http.StatusOK, response)
}

func (h *ManagementHandler) GetRole(c *gin.Context) {
	applicationID, err := getApplicationIDFromContext(c)
	if err != nil {
		respondWithError(c, http.StatusUnauthorized, "Aplicação não identificada")
		return
	}

	id, err := uuid.Parse(c.Param("id"))
	if err != nil {
		respondWithError(c, http.StatusBadRequest, "ID inválido")
		return
	}

	response, err := h.managementUseCase.GetRole(c.Request.Context(), id, applicationID)
	if err != nil {
		respondWithError(c, http.StatusNotFound, "Role não encontrada")
		return
	}
	respondWithJSON(c, http.StatusOK, response)
}

func (h *ManagementHandler) CreateRole(c *gin.Context) {
	applicationID, err := getApplicationIDFromContext(c)
	if err != nil {
		respondWithError(c, http.StatusUnauthorized, "Aplicação não identificada")
		return
	}

	var req dto.CreateRoleRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		respondWithError(c, http.StatusBadRequest, "Dados inválidos")
		return
	}

	response, err := h.managementUseCase.CreateRole(c.Request.Context(), applicationID, req)
	if err != nil {
		respondWithError(c, http.StatusBadRequest, "Erro ao criar role")
		return
	}
	respondWithJSON(c, http.StatusCreated, response)
}

func (h *ManagementHandler) UpdateRole(c *gin.Context) {
	applicationID, err := getApplicationIDFromContext(c)
	if err != nil {
		respondWithError(c, http.StatusUnauthorized, "Aplicação não identificada")
		return
	}

	id, err := uuid.Parse(c.Param("id"))
	if err != nil {
		respondWithError(c, http.StatusBadRequest, "ID inválido")
		return
	}

	var req dto.UpdateRoleRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		respondWithError(c, http.StatusBadRequest, "Dados inválidos")
		return
	}

	response, err := h.managementUseCase.UpdateRole(c.Request.Context(), id, applicationID, req)
	if err != nil {
		respondWithError(c, http.StatusBadRequest, err.Error())
		return
	}
	respondWithJSON(c, http.StatusOK, response)
}

func (h *ManagementHandler) DeleteRole(c *gin.Context) {
	applicationID, err := getApplicationIDFromContext(c)
	if err != nil {
		respondWithError(c, http.StatusUnauthorized, "Aplicação não identificada")
		return
	}

	id, err := uuid.Parse(c.Param("id"))
	if err != nil {
		respondWithError(c, http.StatusBadRequest, "ID inválido")
		return
	}

	if err := h.managementUseCase.DeleteRole(c.Request.Context(), id, applicationID); err != nil {
		respondWithError(c, http.StatusBadRequest, err.Error())
		return
	}
	c.Status(http.StatusNoContent)
}

func (h *ManagementHandler) UpdateRolePermissions(c *gin.Context) {
	applicationID, err := getApplicationIDFromContext(c)
	if err != nil {
		respondWithError(c, http.StatusUnauthorized, "Aplicação não identificada")
		return
	}

	id, err := uuid.Parse(c.Param("id"))
	if err != nil {
		respondWithError(c, http.StatusBadRequest, "ID inválido")
		return
	}

	var req dto.UpdateRolePermissionsRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		respondWithError(c, http.StatusBadRequest, "Dados inválidos")
		return
	}

	response, err := h.managementUseCase.UpdateRolePermissions(c.Request.Context(), id, applicationID, req)
	if err != nil {
		respondWithError(c, http.StatusBadRequest, err.Error())
		return
	}
	respondWithJSON(c, http.StatusOK, response)
}

// ==================== PERMISSIONS ====================

func (h *ManagementHandler) ListPermissions(c *gin.Context) {
	applicationID, err := getApplicationIDFromContext(c)
	if err != nil {
		respondWithError(c, http.StatusUnauthorized, "Aplicação não identificada")
		return
	}

	response, err := h.managementUseCase.ListPermissions(c.Request.Context(), applicationID)
	if err != nil {
		respondWithError(c, http.StatusInternalServerError, "Erro ao listar permissions")
		return
	}
	respondWithJSON(c, http.StatusOK, response)
}

func (h *ManagementHandler) CreatePermission(c *gin.Context) {
	applicationID, err := getApplicationIDFromContext(c)
	if err != nil {
		respondWithError(c, http.StatusUnauthorized, "Aplicação não identificada")
		return
	}

	var req dto.CreatePermissionRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		respondWithError(c, http.StatusBadRequest, "Dados inválidos")
		return
	}

	response, err := h.managementUseCase.CreatePermission(c.Request.Context(), applicationID, req)
	if err != nil {
		respondWithError(c, http.StatusBadRequest, "Erro ao criar permission")
		return
	}
	respondWithJSON(c, http.StatusCreated, response)
}

func (h *ManagementHandler) UpdatePermission(c *gin.Context) {
	applicationID, err := getApplicationIDFromContext(c)
	if err != nil {
		respondWithError(c, http.StatusUnauthorized, "Aplicação não identificada")
		return
	}

	id, err := uuid.Parse(c.Param("id"))
	if err != nil {
		respondWithError(c, http.StatusBadRequest, "ID inválido")
		return
	}

	var req dto.UpdatePermissionRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		respondWithError(c, http.StatusBadRequest, "Dados inválidos")
		return
	}

	response, err := h.managementUseCase.UpdatePermission(c.Request.Context(), id, applicationID, req)
	if err != nil {
		respondWithError(c, http.StatusBadRequest, err.Error())
		return
	}
	respondWithJSON(c, http.StatusOK, response)
}

func (h *ManagementHandler) DeletePermission(c *gin.Context) {
	applicationID, err := getApplicationIDFromContext(c)
	if err != nil {
		respondWithError(c, http.StatusUnauthorized, "Aplicação não identificada")
		return
	}

	id, err := uuid.Parse(c.Param("id"))
	if err != nil {
		respondWithError(c, http.StatusBadRequest, "ID inválido")
		return
	}

	if err := h.managementUseCase.DeletePermission(c.Request.Context(), id, applicationID); err != nil {
		respondWithError(c, http.StatusBadRequest, err.Error())
		return
	}
	c.Status(http.StatusNoContent)
}

// ==================== SYNC / MANIFEST ====================

func (h *ManagementHandler) SyncManifest(c *gin.Context) {
	var req dto.SyncManifestRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		respondWithError(c, http.StatusBadRequest, "Dados inválidos")
		return
	}

	response, err := h.managementUseCase.SyncManifest(c.Request.Context(), req)
	if err != nil {
		respondWithError(c, http.StatusBadRequest, err.Error())
		return
	}

	respondWithJSON(c, http.StatusOK, response)
}
