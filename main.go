package main

import (
	_ "embed"
	"flag"
	"fmt"
	"strings"
)

//go:embed VERSION
var version string

func main() {
	showVersion := flag.Bool("version", false, "print the version and exit")
	flag.Parse()

	if *showVersion {
		fmt.Println(strings.TrimSpace(version))
		return
	}

	fmt.Println("Hello, World!")
}
