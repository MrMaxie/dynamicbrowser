package main

import (
	"errors"
	"os"
	"strings"

	"github.com/MrMaxie/dynamicbrowser/internal/app"
	"github.com/MrMaxie/dynamicbrowser/internal/registration"
	"github.com/alecthomas/kong"
)

// version is set from VERSION by the build recipe.
var version = "dev"

func main() {
	var cli struct {
		Version  kong.VersionFlag `help:"Print the version and exit."`
		Force    bool             `xor:"operation" help:"Stop the running tray and start a new one."`
		Register bool             `xor:"operation" help:"Register as a browser and open Windows Default Apps settings."`
		Restore  bool             `xor:"operation" help:"Restore previous browser choices through Windows Settings and remove registration."`
		URLs     []string         `arg:"" optional:"" name:"url" help:"Web URLs to open without starting a tray."`
	}
	kong.Parse(&cli,
		kong.Name("dynamicbrowser"),
		kong.Description("Route web links or open the dynamicbrowser tray."),
		kong.Vars{"version": strings.TrimSpace(version)},
	)

	var err error
	if (cli.Register || cli.Restore) && len(cli.URLs) != 0 {
		err = errors.New("--register and --restore cannot be combined with URL arguments")
	} else if cli.Register {
		err = registration.Register()
	} else if cli.Restore {
		err = registration.Restore()
	} else {
		err = app.Run(app.Options{Force: cli.Force, URLs: cli.URLs})
	}
	if errors.Is(err, registration.ErrCancelled) {
		os.Exit(2)
	}
	if err != nil {
		app.ShowStartupError(err)
		os.Exit(1)
	}
}
