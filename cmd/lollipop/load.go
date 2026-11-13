package main

import "github.com/kidkafka9/lollipop/internal/source/gitrepo"

// sourceLoadedMsg carries what was found for the source at index in
// cfg.Sources: its repos, and the folders or repos that couldn't be read.
type sourceLoadedMsg struct {
	index    int
	repos    []gitrepo.Repo
	problems []error
}
