package main

import (
	"fmt"
	"strings"

	"charm.land/lipgloss/v2"
)

// page is one screen of the dashboard.
type page int

// The pages, in the order of the tabs. pageCount is not a page: it equals the
// number of pages above it.
const (
	pageOverview page = iota
	pageGit
	pageGrafana
	pageCount
)

// String returns the page's name, as the tabs show it.
func (p page) String() string {
	switch p {
	case pageOverview:
		return "Overview"
	case pageGit:
		return "Git"
	case pageGrafana:
		return "Grafana"
	default:
		return "Unknown"
	}
}

// open returns m showing page p, scrolled to the top.
func (m model) open(p page) model {
	m.page = p
	m = m.syncViewport()
	m.viewport.GotoTop()
	return m
}

// tabs returns the row of page names, with the current page highlighted.
func (m model) tabs() string {
	names := make([]string, 0, pageCount)

	for p := range pageCount {
		if p == m.page {
			names = append(names, activeTabStyle.Render(p.String()))
		} else {
			names = append(names, mutedStyle.Render(p.String()))
		}
	}

	return strings.Join(names, "   ")
}

// pageBody returns what the current page shows inside the viewport.
func (m model) pageBody() string {
	switch m.page {
	case pageOverview:
		return m.overview()
	case pageGit:
		panels, _ := m.content()
		return panels
	default:
		return mutedStyle.Render("Not connected yet")
	}
}

// overview returns one tile per section. The selected tile has the accent
// border, and enter opens it.
func (m model) overview() string {
	total, attention := m.gitSummary()

	return lipgloss.JoinHorizontal(lipgloss.Top,
		tile("git", fmt.Sprintf("%d repos\n%d need attention", total, attention), m.tile == 0),
		tile("grafana", "not connected\n ", m.tile == 1))
}

// tile returns one overview box.
func tile(title, body string, selected bool) string {
	style := tileStyle
	if selected {
		style = style.BorderForeground(lipgloss.Color("#FF79D0")) // Cheeky
	}
	return style.Render(titleStyle.Render(title) + "\n" + body)
}

// gitSummary returns how many repos there are, and how many have changes or
// commits to push or pull.
func (m model) gitSummary() (total, attention int) {
	for _, st := range m.sources {
		for _, row := range st.repos {
			total++
			if row.repo.Dirty > 0 || row.repo.Ahead > 0 || row.repo.Behind > 0 {
				attention++
			}
		}
	}
	return total, attention
}
