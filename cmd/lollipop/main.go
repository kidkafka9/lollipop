// Lollipop is a terminal dashboard for the tools a DevOps or platform
// engineer checks every day. It reads the sources to show from a TOML
// config file.
//
// Usage:
//
//	lollipop [--config path]
//
// Without --config, it reads ~/.config/lollipop/config.toml.
// Press q or ctrl+c to quit.
package main

import (
	"flag"
	"fmt"
	"os"

	tea "charm.land/bubbletea/v2"
	"github.com/kidkafka9/lollipop/internal/config"
)

func run() error {
	configPath := flag.String("config", "", "path to the config file")
	flag.Parse()

	if *configPath == "" {
		path, err := config.DefaultPath()
		if err != nil {
			return err
		}

		*configPath = path
	}

	cfg, err := config.Load(*configPath)
	if err != nil {
		return err
	}

	if _, err := tea.NewProgram(newModel(cfg)).Run(); err != nil {
		return err
	}

	return nil
}

func main() {
	if err := run(); err != nil {
		fmt.Fprintln(os.Stderr, "lollipop:", err)
		os.Exit(1)
	}
}
