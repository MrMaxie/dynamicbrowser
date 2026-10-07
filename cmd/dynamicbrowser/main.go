package main

import (
	"flag"
	"fmt"
	"strings"

	"github.com/MrMaxie/dynamicbrowser/internal/app"
)

// version is set from VERSION by the build recipe.
var version = "dev"

func main() {
	showVersion := flag.Bool("version", false, "print the version and exit")
	flag.Parse()

	if *showVersion {
		fmt.Println(strings.TrimSpace(version))
		return
	}

	if err := app.Run(); err != nil {
		app.ShowStartupError(err)
	}
}
