package handler

import (
	"html/template"
	"io/ioutil"
	"net/http"
	"strings"

	"github.com/theretech/retechauth-api/internal/config"
	"github.com/gin-gonic/gin"
)

// DocsHandler serves the static docs preview page `public/docs.html`.
// It keeps DocsConfig validation for enablement and SpecURL presence.
type DocsHandler struct {
	config config.DocsConfig
}

// NewDocsHandler creates a new DocsHandler.
func NewDocsHandler(cfg config.DocsConfig) *DocsHandler {
	return &DocsHandler{config: cfg}
}

// Render serves `public/docs.html` when docs are enabled. Returns 404 if
// docs are disabled, or 500 if SpecURL is missing.
func (h *DocsHandler) Render(c *gin.Context) {
	if !h.config.Enabled {
		c.Status(http.StatusNotFound)
		return
	}

	if h.config.SpecURL == "" {
		c.String(http.StatusInternalServerError, "Documentação não configurada. Defina DOCS_SPEC_URL.")
		return
	}

	path := c.Request.URL.Path
	var docFile string

	if path == "/public/docs.html" {
		docFile = "public/docs.html"
	} else if path == "/public/docs-v1.html" {
		docFile = "public/docs-v1.html"
	} else if path == "/docs" {
		docFile = "public/docs-v1.html"
	} else {
		pathParts := strings.Split(strings.TrimPrefix(path, "/docs/"), "/")
		if len(pathParts) > 0 && pathParts[0] != "" {
			version := pathParts[0]
			docFile = "public/docs-" + version + ".html"
		} else {
			docFile = "public/docs-v1.html"
		}
	}

	htmlContent, err := ioutil.ReadFile(docFile)
	if err != nil {
		c.String(http.StatusInternalServerError, "Erro ao carregar documentação")
		return
	}

	tmpl, err := template.New("docs").Parse(string(htmlContent))
	if err != nil {
		c.String(http.StatusInternalServerError, "Erro ao processar template")
		return
	}

	data := map[string]string{
		"Title":            h.config.Title,
		"Description":      h.config.Description,
		"Version":          h.config.Version,
		"SpecURL":          h.config.SpecURL,
		"VersionLinks":     h.config.VersionLinks,
		"APIBaseURL":       h.config.APIBaseURL,
		"HeroSupportEmail": h.config.HeroSupportEmail,
		"HeroSupportURL":   h.config.HeroSupportURL,
		"HeroLicense":      h.config.HeroLicense,
	}

	c.Header("Content-Type", "text/html; charset=utf-8")
	c.Status(http.StatusOK)
	if err := tmpl.Execute(c.Writer, data); err != nil {
		c.String(http.StatusInternalServerError, "Erro ao renderizar documentação")
		return
	}
}

// RenderOpenAPISpec serve o arquivo OpenAPI YAML processado como template
func (h *DocsHandler) RenderOpenAPISpec(c *gin.Context) {
	yamlFile := strings.TrimPrefix(c.Request.URL.Path, "/public/")
	yamlPath := "public/" + yamlFile

	yamlContent, err := ioutil.ReadFile(yamlPath)
	if err != nil {
		c.String(http.StatusNotFound, "Arquivo OpenAPI não encontrado")
		return
	}

	tmpl, err := template.New("openapi").Parse(string(yamlContent))
	if err != nil {
		c.String(http.StatusInternalServerError, "Erro ao processar OpenAPI spec")
		return
	}

	data := map[string]string{
		"Title":       h.config.Title,
		"Description": h.config.Description,
		"Version":     h.config.Version,
		"APIBaseURL":  h.config.APIBaseURL,
	}

	c.Header("Content-Type", "application/x-yaml; charset=utf-8")
	c.Status(http.StatusOK)
	if err := tmpl.Execute(c.Writer, data); err != nil {
		c.String(http.StatusInternalServerError, "Erro ao renderizar OpenAPI spec")
		return
	}
}

// RegisterVersionRoutes registra dinamicamente as rotas de documentação
func (h *DocsHandler) RegisterVersionRoutes(router *gin.Engine) {
	if h.config.VersionLinks == "" {
		return
	}

	versions := strings.Split(h.config.VersionLinks, ",")
	for _, versionPair := range versions {
		parts := strings.Split(strings.TrimSpace(versionPair), ":")
		if len(parts) != 2 {
			continue
		}
		path := parts[1]
		router.GET(path, h.Render)
	}
}
