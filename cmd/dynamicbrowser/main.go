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
		Force   bool             `help:"Stop the running tray and start a new one."`
		URLs    []string         `arg:"" optional:"" name:"url" help:"Web URLs to open without starting a tray."`
	}
	kong.Parse(&cli,
		kong.Name("dynamicbrowser"),
		kong.Description("Route web links or open the dynamicbrowser tray."),
		kong.Vars{"version": strings.TrimSpace(version)},
	)

	if err := app.Run(app.Options{Force: cli.Force, URLs: cli.URLs}); err != nil {
		app.ShowStartupError(err)
		os.Exit(1)
	}
}
