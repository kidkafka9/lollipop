package gitrepo

import (
	"errors"
	"io/fs"
	"os"
	"path/filepath"
	"slices"
	"testing"
)

func TestFind(t *testing.T) {
	tests := []struct {
		name  string
		dirs  []string
		files []string
		want  []string
	}{
		{
			name: "no repos",
			dirs: []string{"notes", "photos/2026"},
			want: nil,
		},
		{
			name: "repos at different depths",
			dirs: []string{"api/.git", "team/web/.git", "team/docs"},
			want: []string{"api", "team/web"},
		},
		{
			name: "root is a repo",
			dirs: []string{".git"},
			want: []string{"."},
		},
		{
			name: "does not look inside .git",
			dirs: []string{"api/.git/modules/lib/.git"},
			want: []string{"api"},
		},
		{
			name: "finds a repo nested in another repo",
			dirs: []string{"app/.git", "app/vendor/lib/.git"},
			want: []string{"app", "app/vendor/lib"},
		},
		{
			name:  "ignores a .git file",
			dirs:  []string{"main/.git", "worktree"},
			files: []string{"worktree/.git"},
			want:  []string{"main"},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			root := t.TempDir()
			for _, dir := range tt.dirs {
				mkdir(t, filepath.Join(root, dir))
			}
			for _, file := range tt.files {
				if err := os.WriteFile(filepath.Join(root, file), []byte("gitdir: elsewhere\n"), 0o644); err != nil {
					t.Fatal(err)
				}
			}

			repos, skipped, err := Find(root)
			if err != nil {
				t.Fatalf("Find returned error: %v", err)
			}
			if len(skipped) != 0 {
				t.Errorf("skipped = %v, want none", skipped)
			}
			if got := relative(t, root, repos); !slices.Equal(got, tt.want) {
				t.Errorf("repos = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestFindMissingRoot(t *testing.T) {
	root := filepath.Join(t.TempDir(), "missing")

	repos, skipped, err := Find(root)
	if !errors.Is(err, fs.ErrNotExist) {
		t.Errorf("err = %v, want fs.ErrNotExist", err)
	}
	if repos != nil || skipped != nil {
		t.Errorf("repos = %v, skipped = %v, want both nil", repos, skipped)
	}
}

func TestFindUnreadableRoot(t *testing.T) {
	if os.Geteuid() == 0 {
		t.Skip("root can read every folder")
	}
	root := t.TempDir()
	lock(t, root)

	_, _, err := Find(root)
	if !errors.Is(err, fs.ErrPermission) {
		t.Errorf("err = %v, want fs.ErrPermission", err)
	}
}

func TestFindSkipsUnreadableFolder(t *testing.T) {
	if os.Geteuid() == 0 {
		t.Skip("root can read every folder")
	}
	root := t.TempDir()
	mkdir(t, filepath.Join(root, "open", ".git"))
	locked := filepath.Join(root, "locked")
	mkdir(t, filepath.Join(locked, "hidden", ".git"))
	lock(t, locked)

	repos, skipped, err := Find(root)
	if err != nil {
		t.Fatalf("Find returned error: %v", err)
	}
	if got := relative(t, root, repos); !slices.Equal(got, []string{"open"}) {
		t.Errorf("repos = %v, want [open]", got)
	}
	if len(skipped) != 1 || !errors.Is(skipped[0], fs.ErrPermission) {
		t.Errorf("skipped = %v, want one permission error", skipped)
	}
}

func mkdir(t *testing.T, path string) {
	t.Helper()
	if err := os.MkdirAll(path, 0o755); err != nil {
		t.Fatal(err)
	}
}

func lock(t *testing.T, path string) {
	t.Helper()
	if err := os.Chmod(path, 0o000); err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() { os.Chmod(path, 0o755) })
}

func relative(t *testing.T, root string, paths []string) []string {
	t.Helper()
	var rel []string
	for _, p := range paths {
		r, err := filepath.Rel(root, p)
		if err != nil {
			t.Fatal(err)
		}
		rel = append(rel, filepath.ToSlash(r))
	}
	return rel
}
