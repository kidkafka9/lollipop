package main

import (
	"fmt"
	"strings"

	tea "charm.land/bubbletea/v2"
	"github.com/kidkafka9/lollipop/internal/config"
	"github.com/kidkafka9/lollipop/internal/source/gitrepo"
)

type model struct {
	width, height int
	cfg           config.Config
	sources       []sourceState
}

type sourceState struct {
	loading  bool
	repos    []gitrepo.Repo
	problems []error
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
	const footer = "\n(ctrl+c or q to quit)\n"

	var b strings.Builder

	for i, src := range m.cfg.Sources {
		st := m.sources[i]
		fmt.Fprintf(&b, "%s\n", src.Name)
		if st.loading {
			b.WriteString("  loading…\n")
		} else {
			fmt.Fprintf(&b, "  %d repos\n", len(st.repos))
		}
		b.WriteString("\n")
	}

	f := b.String() + footer

	v := tea.NewView(f)
	v.AltScreen = true
	return v
}

// newModel returns the model before anything has loaded: every source starts
// out loading.
func newModel(cfg config.Config) model {
	sources := make([]sourceState, len(cfg.Sources))
	for i := range sources {
		sources[i].loading = true
	}

	return model{cfg: cfg, sources: sources}
}
