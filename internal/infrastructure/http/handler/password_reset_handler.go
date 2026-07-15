package handler

import (
	"errors"
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/theretech/retech-auth-api/internal/application/usecase"
)

// PasswordResetHandler expõe o fluxo "esqueci a senha" para serviços internos.
// As rotas são protegidas por HMAC (mesmo esquema do /applications/sync):
// quem envia o e-mail com o link é o serviço do produto (ex.: retech-meufin-api),
// nunca este serviço — aqui só emitimos e consumimos tokens.
type PasswordResetHandler struct {
	useCase *usecase.PasswordResetUseCase
}

func NewPasswordResetHandler(useCase *usecase.PasswordResetUseCase) *PasswordResetHandler {
	return &PasswordResetHandler{useCase: useCase}
}

type passwordResetRequestBody struct {
	Email string `json:"email" binding:"required"`
}

type passwordResetConfirmBody struct {
	Token       string `json:"token" binding:"required"`
	NewPassword string `json:"new_password" binding:"required"`
}

// Request emite um token de redefinição para o e-mail informado.
// 404 quando o usuário não existe/está inativo — o serviço chamador decide
// como mascarar isso na API pública (sempre 200 genérico).
func (h *PasswordResetHandler) Request(c *gin.Context) {
	var body passwordResetRequestBody
	if err := c.ShouldBindJSON(&body); err != nil {
		respondWithError(c, http.StatusBadRequest, "email é obrigatório")
		return
	}

	result, err := h.useCase.Request(c.Request.Context(), body.Email)
	if err != nil {
		if errors.Is(err, usecase.ErrResetUserNotFound) {
			respondWithError(c, http.StatusNotFound, "usuário não encontrado")
			return
		}
		respondWithError(c, http.StatusInternalServerError, "falha ao emitir token de redefinição")
		return
	}

	respondWithJSON(c, http.StatusOK, gin.H{
		"token":      result.Token,
		"user_name":  result.UserName,
		"email":      result.Email,
		"expires_at": result.ExpiresAt,
	})
}

// Confirm consome o token e troca a senha do usuário.
func (h *PasswordResetHandler) Confirm(c *gin.Context) {
	var body passwordResetConfirmBody
	if err := c.ShouldBindJSON(&body); err != nil {
		respondWithError(c, http.StatusBadRequest, "token e new_password são obrigatórios")
		return
	}

	err := h.useCase.Confirm(c.Request.Context(), body.Token, body.NewPassword)
	switch {
	case err == nil:
		respondWithJSON(c, http.StatusOK, gin.H{"message": "senha redefinida com sucesso"})
	case errors.Is(err, usecase.ErrResetWeakPassword):
		respondWithError(c, http.StatusBadRequest, err.Error())
	case errors.Is(err, usecase.ErrResetTokenInvalid):
		respondWithError(c, http.StatusUnprocessableEntity, "token inválido, expirado ou já utilizado")
	default:
		respondWithError(c, http.StatusInternalServerError, "falha ao redefinir senha")
	}
}
