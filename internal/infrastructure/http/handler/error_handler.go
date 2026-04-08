package handler

import (
	"net/http"
	"strings"

	"github.com/gin-gonic/gin"
	"github.com/theretech/retech-auth-api/internal/domain/dto"
)

// MethodNotAllowedHandler trata requisições com método HTTP não permitido (405)
// O Gin chama este handler quando encontra uma rota mas o método não corresponde
func MethodNotAllowedHandler(c *gin.Context) {
	// O Gin já define o header Allow automaticamente quando HandleMethodNotAllowed = true
	// Mas vamos garantir que está correto
	allowHeader := c.Writer.Header().Get("Allow")

	// Se o Gin não definiu, inferimos baseado no path
	if allowHeader == "" {
		path := normalizePathForInference(c.Request.URL.Path)
		allowedMethods := inferAllowedMethods(path)

		if len(allowedMethods) == 0 {
			// Fallback: infere baseado no padrão do path
			if strings.Contains(path, "/") && strings.Count(path, "/") >= 3 {
				allowedMethods = []string{"GET", "PUT", "DELETE", "OPTIONS"}
			} else if strings.HasSuffix(path, "/password") {
				allowedMethods = []string{"PUT", "OPTIONS"}
			} else {
				allowedMethods = []string{"GET", "POST", "OPTIONS"}
			}
		}

		c.Header("Allow", strings.Join(allowedMethods, ", "))
		allowHeader = strings.Join(allowedMethods, ", ")
	}

	c.JSON(http.StatusMethodNotAllowed, dto.ErrorResponse{
		Error:   "Method Not Allowed",
		Message: "O método HTTP '" + c.Request.Method + "' não é permitido para este endpoint. Métodos permitidos: " + allowHeader,
		Code:    http.StatusMethodNotAllowed,
	})
}

// normalizePathForInference normaliza o path para inferência de métodos
func normalizePathForInference(path string) string {
	// Remove query string
	if idx := strings.Index(path, "?"); idx != -1 {
		path = path[:idx]
	}

	// Remove barras duplicadas
	for strings.Contains(path, "//") {
		path = strings.ReplaceAll(path, "//", "/")
	}

	// Remove trailing slash (exceto raiz)
	if len(path) > 1 && strings.HasSuffix(path, "/") {
		path = strings.TrimSuffix(path, "/")
	}

	return path
}

// NotFoundHandler trata requisições para rotas não encontradas (404)
func NotFoundHandler(c *gin.Context) {
	if strings.HasPrefix(c.Request.URL.Path, "/v1/") {
		allowedMethods := inferAllowedMethods(c.Request.URL.Path)
		if len(allowedMethods) > 0 {
			c.Header("Allow", strings.Join(allowedMethods, ", "))
			c.JSON(http.StatusMethodNotAllowed, dto.ErrorResponse{
				Error:   "Method Not Allowed",
				Message: "O método HTTP '" + c.Request.Method + "' não é permitido para este endpoint. Métodos permitidos: " + strings.Join(allowedMethods, ", "),
				Code:    http.StatusMethodNotAllowed,
			})
			return
		}

		c.JSON(http.StatusNotFound, dto.ErrorResponse{
			Error:   "Not Found",
			Message: "O endpoint solicitado não foi encontrado",
			Code:    http.StatusNotFound,
		})
		return
	}

	c.Status(http.StatusNotFound)
}

// inferAllowedMethods infere métodos permitidos baseado no path
func inferAllowedMethods(path string) []string {
	// Path já deve estar normalizado antes de chamar esta função

	patterns := []struct {
		pattern string
		methods []string
	}{
		{"/v1/users/{id}/password", []string{"PUT", "OPTIONS"}},
		{"/v1/users/{id}", []string{"GET", "PUT", "DELETE", "OPTIONS"}},
		{"/v1/users", []string{"GET", "POST", "OPTIONS"}},
		{"/v1/applications/{id}", []string{"GET", "PUT", "DELETE", "OPTIONS"}},
		{"/v1/applications", []string{"GET", "POST", "OPTIONS"}},
		{"/v1/roles/{id}", []string{"GET", "PUT", "DELETE", "OPTIONS"}},
		{"/v1/roles", []string{"GET", "POST", "OPTIONS"}},
		{"/v1/permissions/{id}", []string{"PUT", "DELETE", "OPTIONS"}},
		{"/v1/permissions", []string{"GET", "POST", "OPTIONS"}},
		{"/v1/health", []string{"GET", "OPTIONS"}},
		{"/v1/authenticate", []string{"POST", "OPTIONS"}},
		{"/v1/refresh", []string{"POST", "OPTIONS"}},
		{"/v1/me", []string{"GET", "OPTIONS"}},
	}

	for _, p := range patterns {
		if matchesPattern(path, p.pattern) {
			return p.methods
		}
	}

	return nil
}

// matchesPattern verifica se um path corresponde a um padrão
func matchesPattern(path, pattern string) bool {
	pathParts := strings.Split(strings.Trim(path, "/"), "/")
	patternParts := strings.Split(strings.Trim(pattern, "/"), "/")

	if len(pathParts) != len(patternParts) {
		return false
	}

	for i, part := range patternParts {
		if strings.HasPrefix(part, "{") && strings.HasSuffix(part, "}") {
			if pathParts[i] == "" {
				return false
			}
			continue
		}
		if pathParts[i] != part {
			return false
		}
	}

	return true
}
