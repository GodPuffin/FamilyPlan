package main

import (
	"embed"
	"fmt"
	"html/template"
	"log"
	"net/http"
	"os"
	"strings"

	"github.com/labstack/echo/v5"
	"github.com/pocketbase/pocketbase"
	"github.com/pocketbase/pocketbase/apis"
	"github.com/pocketbase/pocketbase/core"
	"github.com/pocketbase/pocketbase/plugins/migratecmd"
)

//go:embed templates/*
var templatesFS embed.FS

//go:embed static/*
var staticFS embed.FS

func main() {
	fmt.Println("Starting FamilyPlan application...")

	app := pocketbase.New()

	// Register migrations with automigration enabled
	migratecmd.MustRegister(app, app.RootCmd, migratecmd.Config{
		Automigrate: true, // Always run migrations automatically
	})

	// Bootstrap the app (no longer using InitDatabase)
	if err := app.Bootstrap(); err != nil {
		log.Fatal(err)
	}

	// Configure app settings to disable HTTPS
	app.Settings().Meta.AppUrl = "http://familyplanmanager.xyz:8090" // Force HTTP
	app.Settings().Meta.HideControls = true
	app.Settings().Logs.MaxDays = 7
	app.Settings().Smtp.Enabled = false

	// Disable HTTPS requirements
	app.OnBeforeServe().Add(func(e *core.ServeEvent) error {
		// Add a middleware to intercept redirects to HTTPS
		e.Router.Use(func(next echo.HandlerFunc) echo.HandlerFunc {
			return func(c echo.Context) error {
				// Override any internal HTTPS redirects
				c.Request().Header.Set("X-Forwarded-Proto", "http")
				c.Response().Header().Set("X-Forwarded-Proto", "http")

				// Capture the response
				originalResponse := c.Response().Writer
				c.Response().Writer = &customResponseWriter{
					ResponseWriter: originalResponse,
				}

				return next(c)
			}
		})

		// Register templates with functions
		tmpl := template.New("").Funcs(templateFuncs)
		templates := template.Must(tmpl.ParseFS(templatesFS, "templates/*.html"))
		e.Router.Renderer = &TemplateRenderer{
			templates: templates,
		}

		// Serve static files
		e.Router.GET("/static/*", apis.StaticDirectoryHandler(staticFS, false))

		// Setup routes
		setupRoutes(app, e.Router, templatesFS)

		return nil
	})

	// Set default command to serve on HTTP only
	os.Args = append([]string{os.Args[0], "serve", "--http=0.0.0.0:8090"}, os.Args[1:]...)

	// Add DEBUG info to help with troubleshooting
	fmt.Println("Server starting, will be accessible via HTTP ONLY at http://familyplanmanager.xyz:8090")
	fmt.Println("HTTPS redirects have been disabled")
	fmt.Println("Command arguments:", os.Args)

	// Start the server
	if err := app.Start(); err != nil {
		log.Fatal(err)
	}
}

// Custom response writer to intercept redirects
type customResponseWriter struct {
	http.ResponseWriter
}

// Override the WriteHeader method to intercept redirects
func (w *customResponseWriter) WriteHeader(statusCode int) {
	// If it's a redirect to HTTPS, change it to 200 OK
	if statusCode == http.StatusFound || statusCode == http.StatusTemporaryRedirect {
		location := w.Header().Get("Location")
		if strings.HasPrefix(location, "https://") {
			// Remove the redirect header
			w.Header().Del("Location")
			// Set status to OK
			w.ResponseWriter.WriteHeader(http.StatusOK)
			return
		}
	}
	w.ResponseWriter.WriteHeader(statusCode)
}
