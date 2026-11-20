package main

import (
	"fmt"
	"unicode/utf8"

	"charm.land/lipgloss/v2"
)

var (
	panelStyle = lipgloss.NewStyle().
			Border(lipgloss.RoundedBorder()).
			BorderForeground(lipgloss.Color("#FF79D0")).
			Padding(0, 1)
	warnStyle  = lipgloss.NewStyle().Foreground(lipgloss.Color("#FF577D")) // Coral
	mutedStyle = lipgloss.NewStyle().Foreground(lipgloss.Color("#A2A0AD")) // Steam

	titleStyle = lipgloss.NewStyle().Bold(true)
)

// sourceBody returns the lines inside one source's panel.
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
// there are any, and the keys.
func statusLine(problems []string, width int) string {
	help := mutedStyle.Render("ctrl+c or q to quit")

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
