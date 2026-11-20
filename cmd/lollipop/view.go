package main

import (
	"fmt"
	"unicode/utf8"

	"charm.land/lipgloss/v2"
)

// Styles for the whole screen. The colours are from Charm's CharmTone palette.
var (
	panelStyle = lipgloss.NewStyle().
			Border(lipgloss.RoundedBorder()).
			BorderForeground(lipgloss.Color("#FF79D0")). // Cheeky
			Padding(0, 1)
	titleStyle = lipgloss.NewStyle().Bold(true)
	warnStyle  = lipgloss.NewStyle().Foreground(lipgloss.Color("#FF577D")) // Coral
	mutedStyle = lipgloss.NewStyle().Foreground(lipgloss.Color("#A2A0AD")) // Steam
)

// content returns every source's panel stacked, and the problems of every
// source as "name: error". A source with no repos but with problems gets no
// panel; its problems show in the status line instead.
func (m model) content() (string, []string) {
	var problems []string
	panels := make([]string, 0, len(m.cfg.Sources))

	for i, src := range m.cfg.Sources {
		st := m.sources[i]

		for _, err := range st.problems {
			problems = append(problems, fmt.Sprintf("%s: %v", src.Name, err))
		}

		if !st.loading && len(st.repos) == 0 && len(st.problems) > 0 {
			continue
		}

		style := panelStyle
		if m.width > 0 {
			style = style.Width(m.width)
		}

		panels = append(panels, style.Render(titleStyle.Render(src.Name)+"\n"+sourceBody(st)))
	}

	return lipgloss.JoinVertical(lipgloss.Left, panels...), problems
}

// sourceBody returns the lines inside one source's panel: one row per repo
// with its name, branch, changed files and commits ahead/behind, in aligned
// columns.
func sourceBody(st sourceState) string {
	var lines []string

	switch {
	case st.loading:
		lines = append(lines, "Loading...")
	case len(st.repos) == 0:
		lines = append(lines, "No repos found.")
	default:
		nameWidth := 0
		branchWidth := 0
		for _, row := range st.repos {
			nameWidth = max(nameWidth, utf8.RuneCountInString(row.name))
			branchWidth = max(branchWidth, utf8.RuneCountInString(row.repo.Branch))
		}

		for _, row := range st.repos {
			lines = append(lines, fmt.Sprintf("%-*s  %-*s  %d  ↑%d↓%d",
				nameWidth, row.name,
				branchWidth, row.repo.Branch,
				row.repo.Dirty,
				row.repo.Ahead,
				row.repo.Behind))
		}
	}

	return lipgloss.JoinVertical(lipgloss.Left, lines...)
}

// statusLine returns the bottom of the screen: a summary of the problems, if
// there are any, and the keys. A long problem is cut at width; 0 means no
// limit.
func statusLine(problems []string, width int) string {
	help := mutedStyle.Render("↑/↓ scroll · q quit")

	var text string
	switch {
	case len(problems) == 0:
		return help
	case len(problems) == 1:
		text = fmt.Sprintf("⚠ 1 problem: %s", problems[0])
	default:
		text = fmt.Sprintf("⚠ %d problems: %s (and %d more)",
			len(problems), problems[0], len(problems)-1)
	}

	return lipgloss.JoinVertical(lipgloss.Left,
		warnStyle.MaxWidth(width).Render(text),
		help,
	)
}
