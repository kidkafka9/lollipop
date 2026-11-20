package main

import (
	"charm.land/bubbles/v2/viewport"
	tea "charm.land/bubbletea/v2"
	"charm.land/lipgloss/v2"
	"github.com/kidkafka9/lollipop/internal/config"
	"github.com/kidkafka9/lollipop/internal/source/gitrepo"
)

// model is the whole state of the program: the config, the terminal size,
// what has been loaded for each source, and the scrollable area that shows it.
type model struct {
	width, height int
	cfg           config.Config
	sources       []sourceState
	viewport      viewport.Model
}

// sourceState is what has been loaded for one source. It has the same index
// as the source in cfg.Sources.
type sourceState struct {
	loading  bool
	repos    []repoRow
	problems []error
}

// repoRow is one repo as the screen shows it: its name relative to the folder
// it was found in, and its state.
type repoRow struct {
	name string
	repo gitrepo.Repo
}

// newModel returns the model before anything has loaded: every source starts
// out loading.
func newModel(cfg config.Config) model {
	sources := make([]sourceState, len(cfg.Sources))
	for i := range sources {
		sources[i].loading = true
	}

	return model{cfg: cfg, sources: sources, viewport: viewport.New()}
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
		m = m.syncViewport()
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
		m = m.syncViewport()
		return m, nil
	}

	// Everything else (scroll keys, mouse wheel) is for the viewport.
	var cmd tea.Cmd
	m.viewport, cmd = m.viewport.Update(msg)
	return m, cmd
}

func (m model) View() tea.View {
	_, problems := m.content()

	v := tea.NewView(m.viewport.View() + "\n" + statusLine(problems, m.width))
	v.AltScreen = true
	return v
}

// syncViewport returns m with the viewport given the current panels and the
// space left above the status line. Call it whenever the size or the data
// changes.
func (m model) syncViewport() model {
	panels, problems := m.content()
	status := statusLine(problems, m.width)

	m.viewport.SetWidth(m.width)
	m.viewport.SetHeight(max(0, m.height-lipgloss.Height(status)))
	m.viewport.SetContent(panels)

	return m
}
