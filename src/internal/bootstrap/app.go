package bootstrap

import (
	"familyplan/src/internal/assets"
	"familyplan/src/internal/http/router"
	"os"
	"strings"

	"github.com/pocketbase/pocketbase"
	"github.com/pocketbase/pocketbase/apis"
	"github.com/pocketbase/pocketbase/core"
	"github.com/pocketbase/pocketbase/plugins/migratecmd"

	_ "familyplan/migrations"
)

// Run boots the PocketBase app and starts serving HTTP traffic.
func Run() error {
	defaultToServeCommand()

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

	return app.Start()
}

// defaultToServeCommand preserves explicit PocketBase subcommands but restores
// the historical "run the server by default" behavior for bare binary launches.
func defaultToServeCommand() {
	args := os.Args[1:]
	if len(args) == 0 {
		os.Args = []string{os.Args[0], "serve"}
		return
	}

	flagsWithValue := map[string]struct{}{
		"--dir":           {},
		"--encryptionEnv": {},
		"--http":          {},
		"--https":         {},
		"--origins":       {},
	}

	for i := 0; i < len(args); i++ {
		arg := args[i]

		switch arg {
		case "-h", "--help", "-v", "--version":
			return
		}

		if !strings.HasPrefix(arg, "-") {
			return
		}

		flagName := arg
		if equalsIndex := strings.Index(flagName, "="); equalsIndex > 0 {
			flagName = flagName[:equalsIndex]
		}

		if _, ok := flagsWithValue[flagName]; ok && flagName == arg && i+1 < len(args) {
			i++
		}
	}

	os.Args = append([]string{os.Args[0], "serve"}, args...)
}
