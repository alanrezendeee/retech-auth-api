package middleware

import (
	"strings"

	"github.com/gin-gonic/gin"
)

// URLNormalizationMiddleware normaliza URLs silenciosamente
// Remove barras duplicadas e trailing slashes
// Esta é a prática padrão do Gin e de muitos frameworks web
func URLNormalizationMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		path := c.Request.URL.Path
		originalPath := path
		
		// Remove barras duplicadas (incluindo no início)
		// Exemplo: /////v1/users -> /v1/users
		for strings.Contains(path, "//") {
			path = strings.ReplaceAll(path, "//", "/")
		}
		
		// Garante que começa com /
		if !strings.HasPrefix(path, "/") {
			path = "/" + path
		}
		
		// Remove trailing slash (exceto para raiz)
		if len(path) > 1 && strings.HasSuffix(path, "/") {
			path = strings.TrimSuffix(path, "/")
		}
		
		// Se o path mudou, atualiza na requisição
		if path != originalPath {
			c.Request.URL.Path = path
			// Atualiza RequestURI também
			if c.Request.RequestURI != "" {
				query := c.Request.URL.RawQuery
				if query != "" {
					c.Request.RequestURI = path + "?" + query
				} else {
					c.Request.RequestURI = path
				}
			}
		}
		
		c.Next()
	}
}

