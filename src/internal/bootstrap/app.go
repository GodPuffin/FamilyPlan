package bootstrap

import (
	"fmt"
	"os"

	"familyplan/src/internal/assets"
	"familyplan/src/internal/http/router"

	"github.com/pocketbase/pocketbase"
	"github.com/pocketbase/pocketbase/apis"
	"github.com/pocketbase/pocketbase/core"
	"github.com/pocketbase/pocketbase/plugins/migratecmd"

	_ "familyplan/migrations"
)

// Run boots the PocketBase app and starts serving HTTP traffic.
func Run() error {
	fmt.Println("Starting FamilyPlan application...")

	app := pocketbase.New()

	migratecmd.MustRegister(app, app.RootCmd, migratecmd.Config{
		Automigrate: true,
	})

	if err := app.Bootstrap(); err != nil {
		return err
	}

	app.Settings().Meta.HideControls = true
	app.Settings().Logs.MaxDays = 7
	app.Settings().Smtp.Enabled = false

	app.OnBeforeServe().Add(func(e *core.ServeEvent) error {
		e.Router.GET("/static/*", apis.StaticDirectoryHandler(assets.StaticFS, false))
		router.Setup(app, e.Router)
		return nil
	})

	os.Args = append([]string{os.Args[0], "serve", "--http=0.0.0.0:8090"}, os.Args[1:]...)

	fmt.Println("Server starting, will be accessible at http://0.0.0.0:8090")
	fmt.Println("Command arguments:", os.Args)

	return app.Start()
}
