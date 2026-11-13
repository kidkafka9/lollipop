// Package gitrepo finds local git repositories and reads their state with
// git status.
package gitrepo

import (
	"context"
	"errors"
	"fmt"
	"os/exec"
	"strconv"
	"strings"
)

// Repo is the state of one repository.
type Repo struct {
	// Path is the repository's folder, as given to Read.
	Path string
	// Branch is the checked-out branch, or "(detached)" when HEAD is not on
	// a branch.
	Branch string
	// Dirty is the number of changed, staged, conflicted and untracked files.
	Dirty int
	// Ahead and Behind count the commits not yet pushed to, and not yet pulled
	// from, the upstream branch. Both are 0 when there is no upstream.
	Ahead, Behind int
}

// Read runs git status in the repository at path and returns its state.
// It doesn't take git's index lock, so it can run in the background while
// you use git in the same repository. Cancelling ctx stops git.
//
// The error contains git's own message, such as "not a git repository".
func Read(ctx context.Context, path string) (Repo, error) {
	cmd := exec.CommandContext(
		ctx,
		"git",
		"--no-optional-locks",
		"-C", path,
		"status",
		"--porcelain=v2",
		"--branch",
	)

	out, err := cmd.Output()
	if err != nil {
		if errors.Is(err, exec.ErrNotFound) {
			return Repo{}, fmt.Errorf("git executable not found: %w", err)
		}

		if exitErr, ok := errors.AsType[*exec.ExitError](err); ok {
			stderr := strings.TrimSpace(string(exitErr.Stderr))
			if stderr != "" {
				return Repo{}, fmt.Errorf("git status failed: %s: %w", stderr, err)
			}
		}

		return Repo{}, fmt.Errorf("git status failed: %w", err)
	}

	return parseStatus(path, string(out))
}

// parseStatus reads the output of git status --porcelain=v2 --branch.
// It returns an error for a branch.ab line it can't read.
func parseStatus(path, out string) (Repo, error) {
	repo := Repo{Path: path}
	for line := range strings.Lines(out) {
		cleanLine := strings.TrimSuffix(line, "\n")

		if branch, ok := strings.CutPrefix(cleanLine, "# branch.head "); ok {
			repo.Branch = branch
		}

		if aheadClean, ok := strings.CutPrefix(cleanLine, "# branch.ab "); ok {
			fields := strings.Fields(aheadClean)

			if len(fields) != 2 {
				return Repo{}, fmt.Errorf("unexpected field count (%d) in line %q", len(fields), line)
			}

			ahead, err := strconv.Atoi(fields[0])
			if err != nil {
				return Repo{}, fmt.Errorf("parse ahead %q: %w", cleanLine, err)
			}

			behind, err := strconv.Atoi(fields[1])
			if err != nil {
				return Repo{}, fmt.Errorf("parse behind %q: %w", cleanLine, err)
			}

			repo.Ahead = ahead
			repo.Behind = -behind

		}

		if strings.HasPrefix(cleanLine, "# ") {
			continue
		}

		repo.Dirty++
	}
	return repo, nil
}
