package handler

import (
	"net/http"

	"github.com/theretech/retechauth-api/internal/infrastructure/http/middleware"
	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
)

func getApplicationIDFromContext(c *gin.Context) (uuid.UUID, error) {
	appID, ok := middleware.GetApplicationID(c.Request.Context())
	if !ok {
		return uuid.Nil, http.ErrNoCookie
	}
	return appID, nil
}

func getUserIDFromContext(c *gin.Context) (uuid.UUID, error) {
	userID, ok := middleware.GetUserID(c.Request.Context())
	if !ok {
		return uuid.Nil, http.ErrNoCookie
	}
	return userID, nil
}

func respondWithError(c *gin.Context, code int, message string) {
	respondWithJSON(c, code, map[string]interface{}{
		"error":   http.StatusText(code),
		"message": message,
		"code":    code,
	})
}

func respondWithJSON(c *gin.Context, code int, payload interface{}) {
	c.JSON(code, payload)
}
