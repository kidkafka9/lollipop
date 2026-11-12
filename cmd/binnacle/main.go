package main

import (
	"fmt"
	"log"

	tea "charm.land/bubbletea/v2"
)

const (
	applicationName  = "binnacle"
	notProvidedValue = "[not-provided]"
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

	case tea.KeyMsg:
		switch msg.String() {
		case "ctrl+c", "q":
			return m, tea.Quit
		}
	}

	return m, nil
}

func (m model) View() tea.View {
	v := tea.NewView(fmt.Sprintf("work in progress..."))
	v.AltScreen = true
	return v
}

func main() {
	if _, err := tea.NewProgram(model{}).Run(); err != nil {
		log.Fatal(err)
	}
}
