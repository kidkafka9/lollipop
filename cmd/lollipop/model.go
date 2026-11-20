package main

import (
	"fmt"

	tea "charm.land/bubbletea/v2"
	"charm.land/lipgloss/v2"
	"github.com/kidkafka9/lollipop/internal/config"
)

// model is the whole state of the program: the config, the terminal size and
// what has been loaded for each source.
type model struct {
	width, height int
	cfg           config.Config
	sources       []sourceState
}

// sourceState is what the screen shows for one source. It has the same index
// as the source in cfg.Sources.
type sourceState struct {
	loading  bool
	repos    []repoRow
	problems []error
}

func (m model) Init() tea.Cmd {
	cmds := make([]tea.Cmd, len(m.cfg.Sources))
	for i, src := range m.cfg.Sources {
		cmds[i] = loadSource(i, src)
	}

	return tea.Batch(cmds...)
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

	case sourceLoadedMsg:
		m.sources[msg.index].repos = msg.repos
		m.sources[msg.index].problems = msg.problems
		m.sources[msg.index].loading = false
	}

	return m, nil
}

func (m model) View() tea.View {
	var problems []string
	panels := make([]string, 0, len(m.cfg.Sources))

	for i, src := range m.cfg.Sources {
		style := panelStyle
		st := m.sources[i]

		for _, err := range st.problems {
			problems = append(problems, fmt.Sprintf("%s: %v", src.Name, err))
		}

		if !st.loading && len(st.repos) == 0 && len(st.problems) > 0 {
			continue
		}

		if m.width > 0 {
			style = style.Width(m.width)
		}

		content := titleStyle.Render(src.Name) + "\n" + sourceBody(m.sources[i])
		panels = append(panels, style.Render(content))
	}

	screen := lipgloss.JoinVertical(lipgloss.Left, panels...)
	v := tea.NewView(screen + "\n" + statusLine(problems, m.width))
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
