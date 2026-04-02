package view

import (
	"fmt"
	"html/template"
	"strings"

	"familyplan/src/internal/assets"
	"familyplan/src/internal/domain"

	"github.com/labstack/echo/v5"
	"golang.org/x/text/cases"
	"golang.org/x/text/language"
)

// Funcs holds shared template helpers.
var Funcs = template.FuncMap{
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
		af, ok1 := toFloat(a)
		bf, ok2 := toFloat(b)
		if !ok1 || !ok2 || bf == 0 {
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

// RenderPage renders a page inside the shared layout template.
func RenderPage(c echo.Context, page string, data map[string]interface{}) error {
	if data == nil {
		data = map[string]interface{}{}
	}

	if session, ok := c.Get("session").(domain.SessionData); ok {
		setDefault(data, "isAuthenticated", session.IsAuthenticated)
		setDefault(data, "username", session.Username)
		setDefault(data, "name", session.Name)
		setDefault(data, "userId", session.UserID)
	}

	tmpl, err := template.New("layout").Funcs(Funcs).ParseFS(
		assets.TemplatesFS,
		"templates/layout.html",
		"templates/"+page,
	)
	if err != nil {
		return err
	}

	return tmpl.ExecuteTemplate(c.Response().Writer, "layout", data)
}

func setDefault(data map[string]interface{}, key string, value interface{}) {
	if _, exists := data[key]; exists {
		return
	}

	data[key] = value
}

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
