package middleware

import (
	"bytes"
	"context"
	"fmt"
	"io"
	"net/http"
	"strconv"
	"strings"

	"github.com/theretech/retechauth-api/internal/application/service"
	"github.com/gin-gonic/gin"
)

// SyncMiddleware é um middleware flexível para /sync que aceita JWT OU HMAC
type SyncMiddleware struct {
	jwtService    service.JWTService
	bootstrapSecret string
}

// NewSyncMiddleware cria uma nova instância do middleware flexível para /sync
func NewSyncMiddleware(jwtService service.JWTService, bootstrapSecret string) *SyncMiddleware {
	return &SyncMiddleware{
		jwtService:      jwtService,
		bootstrapSecret: bootstrapSecret,
	}
}

// AuthenticateSync valida autenticação para /sync (aceita JWT OU HMAC assinado)
func (m *SyncMiddleware) AuthenticateSync() gin.HandlerFunc {
	return func(c *gin.Context) {
		// 1. Primeiro, tenta HMAC (bootstrap com secret compartilhado)
		signature := c.GetHeader("X-Signature")
		timestampStr := c.GetHeader("X-Timestamp")
		
		if signature != "" && timestampStr != "" {
			// Ler body para validar HMAC
			body, err := io.ReadAll(c.Request.Body)
			if err != nil {
				respondWithError(c, http.StatusBadRequest, "Erro ao ler body da requisição")
				c.Abort()
				return
			}
			
			// Restaurar body para o handler usar depois
			c.Request.Body = io.NopCloser(bytes.NewBuffer(body))

			// Converter timestamp
			timestamp, err := strconv.ParseInt(timestampStr, 10, 64)
			if err != nil {
				respondWithError(c, http.StatusBadRequest, "Timestamp inválido")
				c.Abort()
				return
			}

			// Validar HMAC
			if err := ValidateHMAC(body, timestamp, signature, m.bootstrapSecret); err != nil {
				if m.bootstrapSecret == "" {
					respondWithError(c, http.StatusServiceUnavailable, "Bootstrap não configurado: BOOTSTRAP_SECRET não definido")
				} else {
					respondWithError(c, http.StatusUnauthorized, fmt.Sprintf("Assinatura HMAC inválida: %v", err))
				}
				c.Abort()
				return
			}

			// HMAC válido → permite bootstrap
			// Para bootstrap, não precisamos de application_id no contexto
			// O use case vai criar/atualizar a aplicação baseado no manifest
			c.Next()
			return
		}

		// 2. Se não tem HMAC, tenta JWT (uso normal)
		authHeader := c.GetHeader("Authorization")
		if authHeader == "" {
			respondWithError(c, http.StatusUnauthorized, "Token de autenticação ou assinatura HMAC não fornecidos. Use Authorization: Bearer {token} ou X-Signature + X-Timestamp com secret compartilhado")
			c.Abort()
			return
		}

		parts := strings.Split(authHeader, " ")
		if len(parts) != 2 || parts[0] != "Bearer" {
			respondWithError(c, http.StatusUnauthorized, "Formato de token inválido")
			c.Abort()
			return
		}

		tokenString := parts[1]
		claims, err := m.jwtService.ValidateToken(tokenString)
		if err != nil {
			if err == service.ErrExpiredToken {
				respondWithError(c, http.StatusUnauthorized, "Token expirado")
				c.Abort()
				return
			}
			respondWithError(c, http.StatusUnauthorized, "Token inválido")
			c.Abort()
			return
		}

		// JWT válido → insere application_id e tenant_id no contexto (uso normal)
		ctx := c.Request.Context()
		ctx = context.WithValue(ctx, UserIDKey, claims.UserID)
		ctx = context.WithValue(ctx, UserEmailKey, claims.Email)
		ctx = context.WithValue(ctx, ApplicationIDKey, claims.ApplicationID)
		if claims.TenantID != nil {
			ctx = context.WithValue(ctx, TenantIDKey, *claims.TenantID)
		}

		c.Request = c.Request.WithContext(ctx)
		c.Next()
	}
}

