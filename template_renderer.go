package main

import (
	"embed"
	"fmt"
	"html/template"
	"io"
	"strings"

	"github.com/labstack/echo/v5"
	"golang.org/x/text/cases"
	"golang.org/x/text/language"
)

// TemplateRenderer is a custom renderer for Echo
type TemplateRenderer struct {
	templates *template.Template
}

// Common template functions
var templateFuncs = template.FuncMap{
	"upper": strings.ToUpper,
	"lower": strings.ToLower,
	"title": cases.Title(language.English).String,
	"formatMoney": func(amount float64) string {
		return fmt.Sprintf("$%.2f", amount)
	},
	"slice": func(s string, i, j int) string {
		if i >= len(s) {
			return ""
		}
		if j > len(s) {
			j = len(s)
		}
		return s[i:j]
	},
	"div": func(a, b interface{}) float64 {
		// Convert to float64 for division
		af, ok1 := toFloat(a)
		bf, ok2 := toFloat(b)

		if !ok1 || !ok2 || bf == 0 {
			// Return 0 for invalid input or division by zero
			return 0
		}

		return af / bf
	},
	"mul": func(a, b float64) float64 {
		return a * b
	},
	"sub": func(a, b float64) float64 {
		return a - b
	},
	"float64": func(i int) float64 {
		return float64(i)
	},
}

// Convert various types to float64
func toFloat(value interface{}) (float64, bool) {
	switch v := value.(type) {
	case float64:
		return v, true
	case float32:
		return float64(v), true
	case int:
		return float64(v), true
	case int64:
		return float64(v), true
	case int32:
		return float64(v), true
	default:
		return 0, false
	}
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

// NewTemplateRenderer creates a new template renderer using the provided template files
func NewTemplateRenderer(templatesFS embed.FS) (*TemplateRenderer, error) {
	tmpl, err := template.New("base").Funcs(templateFuncs).ParseFS(templatesFS, "templates/*.html")
	if err != nil {
		return nil, err
	}

	return &TemplateRenderer{
		templates: tmpl,
	}, nil
}
