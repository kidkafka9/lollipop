// sundae is a terminal dashboard for the tools a DevOps or a platform engineer needs on a daily basis. It reads the sources to show from a TOML config file.
//
// Usage:
//
//	sundae
//
// Press q or ctrl+c to quit.
package main

import (
	"fmt"
	"os"

	tea "charm.land/bubbletea/v2"
)

type model struct {
	width, height int
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
	v := tea.NewView(fmt.Sprintf("Window size: %dx%d", m.width, m.height))
	v.AltScreen = true
	return v
}

func main() {
	if _, err := tea.NewProgram(model{}).Run(); err != nil {
		_, err := fmt.Fprintln(os.Stderr, err)
		if err != nil {
			return
		}
		os.Exit(1)
	}
}
