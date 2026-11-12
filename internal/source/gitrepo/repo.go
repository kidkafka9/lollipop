package gitrepo

import (
	"context"
	"errors"
	"fmt"
	"os/exec"
	"strconv"
	"strings"
)

type Repo struct {
	Path, Branch         string
	Dirty, Ahead, Behind int
}

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
