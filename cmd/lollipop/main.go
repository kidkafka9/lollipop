// lollipop is a terminal dashboard for the tools a DevOps or
// a platform engineer needs on a daily basis.
// It reads the sources to show from a TOML config file.
//
// Usage:
//
//	lollipop
//
// Press q or ctrl+c to quit.
package main

import (
	"flag"
	"fmt"
	"os"
	"strings"

	tea "charm.land/bubbletea/v2"
	"github.com/kidkafka9/lollipop/internal/config"
)

type model struct {
	width, height int
	cfg           config.Config
}

func (m model) Init() tea.Cmd {
	return nil
}

func (m model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.WindowSizeMsg:
		m.width = msg.Width
		m.height = msg.Height
		return m, nil

	case tea.KeyPressMsg:
		switch msg.String() {
		case "ctrl+c", "q":
			return m, tea.Quit
		}
	}

	return m, nil
}

func (m model) View() tea.View {
	const (
		footer = "\n(ctrl+c or q to quit)\n"
	)

	var sourceNames []string
	for _, s := range m.cfg.Sources {
		sourceNames = append(sourceNames, s.Name)
	}

	f := strings.Join(sourceNames, ", ") + "\n" + footer

	v := tea.NewView(f)
	v.AltScreen = true
	return v
}

func newModel(cfg config.Config) model {
	return model{cfg: cfg}
}

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
		_, err := fmt.Fprintln(os.Stderr, err)
		if err != nil {
			return err
		}
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
