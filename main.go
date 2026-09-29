package main

import (
	"fmt"
	"os"

	tea "charm.land/bubbletea/v2"
)

// version is the build version, injected via -ldflags "-X main.version=$VERSION".
// Plain `go build` leaves it as "dev".
var version = "dev"

// displayVersion renders the version for humans: "v0.1.0" for releases, "dev"
// for un-stamped builds.
func displayVersion() string {
	if version == "dev" {
		return "dev"
	}
	return "v" + version
}

func main() {
	args := os.Args[1:]
	for _, a := range args {
		switch a {
		case "--demo":
			enableDemoMode()
		case "--version", "-v":
			fmt.Println("hazyvpn " + displayVersion())
			return
		case "--help", "-h":
			printUsage()
			return
		}
	}

	if !demoMode {
		initLog()
	}
	initColors()
	p := tea.NewProgram(initialModel())
	if _, err := p.Run(); err != nil {
		fmt.Fprintf(os.Stderr, "Error: %v\n", err)
		os.Exit(1)
	}
}

func printUsage() {
	fmt.Println(`hazyvpn - a keyboard-driven TUI for WireGuard tunnels

Usage:
  hazyvpn            launch the TUI
  hazyvpn --demo     launch with fake sample tunnels, no root or wireguard-tools needed
  hazyvpn --version  print the version
  hazyvpn --help     show this message

Keys (inside the TUI):
  j/k, ↑/↓   move
  enter      connect / disconnect the selected tunnel
  i          import a .conf/.wg file
  r          rename the selected config
  x          delete the selected config
  ?          help
  q          quit`)
}
