package router

import (
	"fmt"
	"net/http"
	"strings"

	"github.com/gin-gonic/gin"
	"github.com/theretech/retech-auth-api/internal/infrastructure/http/handler"
	"github.com/theretech/retech-auth-api/internal/infrastructure/http/middleware"
)

type APIVersion struct {
	Prefix      string
	Register    func(r *gin.RouterGroup)
	Middlewares []gin.HandlerFunc
}

func normalizeURLPath(path string) string {
	for strings.Contains(path, "//") {
		path = strings.ReplaceAll(path, "//", "/")
	}
	if !strings.HasPrefix(path, "/") {
		path = "/" + path
	}
	if len(path) > 1 && strings.HasSuffix(path, "/") {
		path = strings.TrimSuffix(path, "/")
	}
	return path
}

func urlNormalizationHandler(handler http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		originalPath := r.URL.Path
		normalizedPath := normalizeURLPath(originalPath)
		if normalizedPath != originalPath {
			r.URL.Path = normalizedPath
			if r.URL.RawQuery != "" {
				r.RequestURI = normalizedPath + "?" + r.URL.RawQuery
			} else {
				r.RequestURI = normalizedPath
			}
		}
		handler.ServeHTTP(w, r)
	})
}

func SetupRoutes(
	versions []APIVersion,
	docsHandler *handler.DocsHandler,
	authHandler *handler.AuthHandler,
	defaultVersion string,
) *gin.Engine {
	gin.SetMode(gin.ReleaseMode)
	router := gin.New()

	router.HandleMethodNotAllowed = true
	router.RedirectTrailingSlash = false
	router.RedirectFixedPath = false

	router.Use(gin.Recovery())
	router.Use(middleware.LoggingMiddleware())

	for _, version := range versions {
		if version.Prefix == "" || version.Register == nil {
			continue
		}
		v1Group := router.Group(version.Prefix)
		if len(version.Middlewares) > 0 {
			v1Group.Use(version.Middlewares...)
		}
		version.Register(v1Group)
	}

	router.GET("/docs", docsHandler.Render)
	router.GET("/public/docs.html", docsHandler.Render)
	router.GET("/public/docs-v1.html", docsHandler.Render)
	docsHandler.RegisterVersionRoutes(router)
	router.GET("/public/openapi.yaml", docsHandler.RenderOpenAPISpec)
	router.GET("/public/openapi-v1.yaml", docsHandler.RenderOpenAPISpec)

	router.StaticFile("/public/redoc.standalone.js", "./public/redoc.standalone.js")
	router.StaticFile("/public/logo-single.png", "./public/logo-single.png")

	// JWKS endpoint (padrão da indústria)
	// GET /.well-known/jwks.json
	if authHandler != nil {
		router.GET("/.well-known/jwks.json", authHandler.JWKS)
	}

	// Health check na raiz (para compatibilidade com Railway e outros serviços)
	// GET /health
	if authHandler != nil {
		router.GET("/health", authHandler.Health)
		router.GET("/health/live", authHandler.Health)
		router.GET("/health/ready", authHandler.Health)
	}

	if defaultVersion != "" {
		target := fmt.Sprintf("%s/health", strings.TrimRight(defaultVersion, "/"))
		router.GET("/", func(c *gin.Context) {
			c.Redirect(http.StatusTemporaryRedirect, target)
		})
	}

	router.NoRoute(handler.NotFoundHandler)
	router.NoMethod(handler.MethodNotAllowedHandler)

	return router
}

func SetupServer(
	versions []APIVersion,
	docsHandler *handler.DocsHandler,
	authHandler *handler.AuthHandler,
	defaultVersion string,
) http.Handler {
	return urlNormalizationHandler(SetupRoutes(versions, docsHandler, authHandler, defaultVersion))
}
