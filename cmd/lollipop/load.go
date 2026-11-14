package main

import (
	"context"
	"time"

	tea "charm.land/bubbletea/v2"
	"github.com/kidkafka9/lollipop/internal/config"
	"github.com/kidkafka9/lollipop/internal/source/gitrepo"
)

// sourceLoadedMsg carries what was found for the source at index in
// cfg.Sources: its repos, and the folders or repos that couldn't be read.
type sourceLoadedMsg struct {
	index    int
	repos    []gitrepo.Repo
	problems []error
}

// loadSource returns a command that finds and reads every repo under src's
// paths and reports them as a sourceLoadedMsg for index. A folder or repo that
// can't be read becomes a problem; it doesn't stop the others.
func loadSource(index int, src config.Source) tea.Cmd {
	return func() tea.Msg {
		ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
		defer cancel()

		var (
			repos    []gitrepo.Repo
			problems []error
		)

		for _, root := range src.Paths {
			found, skipped, err := gitrepo.Find(root)
			if err != nil {
				problems = append(problems, err)
				continue
			}

			problems = append(problems, skipped...)

			for _, path := range found {
				repo, err := gitrepo.Read(ctx, path)
				if err != nil {
					problems = append(problems, err)
					continue
				}
				repos = append(repos, repo)
			}

		}

		return sourceLoadedMsg{index: index, repos: repos, problems: problems}
	}
}
