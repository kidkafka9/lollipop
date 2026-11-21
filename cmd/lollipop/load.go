package main

import (
	"context"
	"path/filepath"
	"sync"
	"time"

	tea "charm.land/bubbletea/v2"
	"github.com/kidkafka9/lollipop/internal/config"
	"github.com/kidkafka9/lollipop/internal/source/gitrepo"
)

// sourceLoadedMsg carries what was found for the source at index in
// cfg.Sources: its repos, and the folders or repos that couldn't be read.
type sourceLoadedMsg struct {
	index    int
	repos    []repoRow
	problems []error
}

// maxGitReads is how many git commands one source runs at the same time.
const maxGitReads = 8

// repoJob is one repo to read, and the root folder it was found under.
type repoJob struct {
	root string
	path string
}

// repoResult is what reading one repoJob produced: its state, or an error.
type repoResult struct {
	repo gitrepo.Repo
	err  error
}

// loadSource returns a command that finds and reads every repo under src's
// paths and reports them as a sourceLoadedMsg for index. A folder or repo that
// can't be read becomes a problem; it doesn't stop the others.
func loadSource(index int, src config.Source) tea.Cmd {
	return func() tea.Msg {
		ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
		defer cancel()

		var (
			jobs     []repoJob
			repos    []repoRow
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
				jobs = append(jobs, repoJob{root: root, path: path})
			}
		}

		// Each goroutine writes only its own results[i], so they never share
		// memory, and the rows keep the order of jobs.
		results := make([]repoResult, len(jobs))
		limit := make(chan struct{}, maxGitReads)

		var wg sync.WaitGroup
		for i, job := range jobs {
			wg.Go(func() {
				limit <- struct{}{}
				defer func() { <-limit }()

				repo, err := gitrepo.Read(ctx, job.path)
				results[i] = repoResult{repo: repo, err: err}
			})
		}
		wg.Wait()

		for i, result := range results {
			if result.err != nil {
				problems = append(problems, result.err)
				continue
			}

			job := jobs[i]
			name, err := filepath.Rel(job.root, job.path)
			if err != nil {
				problems = append(problems, err)
				continue
			}
			if name == "." {
				// The root itself is a repo: show its folder name.
				name = filepath.Base(job.root)
			}

			repos = append(repos, repoRow{name: name, repo: result.repo})
		}

		return sourceLoadedMsg{index: index, repos: repos, problems: problems}
	}
}
