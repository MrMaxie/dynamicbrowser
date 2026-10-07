package main

import (
	"os"
	"strings"

	"github.com/MrMaxie/dynamicbrowser/internal/app"
	"github.com/alecthomas/kong"
)

// version is set from VERSION by the build recipe.
var version = "dev"

func main() {
	var cli struct {
		Version kong.VersionFlag `help:"Print the version and exit."`
		Force   bool             `help:"Stop the running instance and start a new one."`
	}
	kong.Parse(&cli,
		kong.Name("dynamicbrowser"),
		kong.Description("Run the dynamicbrowser tray application."),
		kong.Vars{"version": strings.TrimSpace(version)},
	)

	if err := app.Run(app.Options{Force: cli.Force}); err != nil {
		app.ShowStartupError(err)
		os.Exit(1)
	}
}
