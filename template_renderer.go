package main

import (
	"html/template"
	"io"
	"strings"

	"github.com/labstack/echo/v5"
)

// TemplateRenderer is a custom renderer for Echo
type TemplateRenderer struct {
	templates *template.Template
}

// Common template functions
var templateFuncs = template.FuncMap{
	"upper": strings.ToUpper,
	"lower": strings.ToLower,
	"title": strings.Title,
}

// Render renders a template with the given name and data
func (t *TemplateRenderer) Render(w io.Writer, name string, data interface{}, c echo.Context) error {
	// Add session data to all templates
	dataMap, ok := data.(echo.Map)
	if !ok {
		dataMap = echo.Map{}
	}

	// Get session data if available
	if session, ok := c.Get("session").(SessionData); ok {
		dataMap["isAuthenticated"] = session.IsAuthenticated
		dataMap["username"] = session.Username
	}

	return t.templates.ExecuteTemplate(w, name, dataMap)
}
