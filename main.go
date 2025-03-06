package main

import (
	"embed"
	"fmt"
	"html/template"
	"log"
	"os"

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

	// Configure app settings for Cloudflare
	app.Settings().Meta.AppUrl = "http://0.0.0.0:8090" // Local binding
	app.Settings().Meta.HideControls = true
	app.Settings().Logs.MaxDays = 7
	app.Settings().Smtp.Enabled = false

	// Trust X-Forwarded-* headers from Cloudflare
	app.OnBeforeServe().Add(func(e *core.ServeEvent) error {
		e.Router.Use(func(next echo.HandlerFunc) echo.HandlerFunc {
			return func(c echo.Context) error {
				// Trust Cloudflare headers
				if cfProto := c.Request().Header.Get("CF-Visitor"); cfProto != "" {
					c.Request().Header.Set("X-Forwarded-Proto", "https")
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

	// Set default command to serve with explicit host binding
	os.Args = append([]string{os.Args[0], "serve", "--http=0.0.0.0:8090"}, os.Args[1:]...)

	// Add DEBUG info to help with troubleshooting
	fmt.Println("Server starting, will be accessible at http://0.0.0.0:8090")
	fmt.Println("Command arguments:", os.Args)

	// Start the server
	if err := app.Start(); err != nil {
		log.Fatal(err)
	}
}
