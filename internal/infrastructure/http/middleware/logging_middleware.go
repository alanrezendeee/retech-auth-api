package middleware

import (
	"log"
	"time"

	"github.com/gin-gonic/gin"
)

// LoggingMiddleware é um middleware para logging de requisições usando Gin
func LoggingMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		start := time.Now()

		// Processa a requisição
		c.Next()

		// Log da requisição após processamento
		duration := time.Since(start)
		log.Printf(
			"[%s] %s %s -> %d (%s)",
			c.Request.Method,
			c.Request.RequestURI,
			c.Request.URL.Path,
			c.Writer.Status(),
			duration,
		)
	}
}

