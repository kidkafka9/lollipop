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

func main() {
	if err := run(); err != nil {
		fmt.Fprintln(os.Stderr, "lollipop:", err)
		os.Exit(1)
	}
}

// run reads the flags and the config, then runs the dashboard until the user
// quits. main is the only place that prints errors and exits.
func run() error {
	configPath := flag.String("config", "", "path to the config file")
	flag.Parse()

	path := *configPath
	if path == "" {
		var err error
		path, err = config.DefaultPath()
		if err != nil {
			return err
		}
	}

	cfg, err := config.Load(path)
	if err != nil {
		return err
	}

	_, err = tea.NewProgram(newModel(cfg)).Run()
	return err
}
