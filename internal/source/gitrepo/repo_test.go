package gitrepo

import (
	"context"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"testing"
)

func TestParseStatus(t *testing.T) {
	tests := []struct {
		name    string
		out     string
		want    Repo
		wantErr bool
	}{
		{
			name: "clean, in sync",
			out: `# branch.oid cb8e74a0f1d2c3b4a5968778695a4b3c2d1e0f9a
# branch.head main
# branch.upstream origin/main
# branch.ab +0 -0
`,
			want: Repo{Path: "repo", Branch: "main"},
		},
		{
			name: "ahead and behind",
			out: `# branch.oid cb8e74a0f1d2c3b4a5968778695a4b3c2d1e0f9a
# branch.head main
# branch.upstream origin/main
# branch.ab +1 -3
`,
			want: Repo{Path: "repo", Branch: "main", Ahead: 1, Behind: 3},
		},
		{
			name: "no upstream",
			out: `# branch.oid cb8e74a0f1d2c3b4a5968778695a4b3c2d1e0f9a
# branch.head feature
`,
			want: Repo{Path: "repo", Branch: "feature"},
		},
		{
			name: "staged and untracked",
			out: `# branch.oid cb8e74a0f1d2c3b4a5968778695a4b3c2d1e0f9a
# branch.head main
1 A. N... 000000 100644 100644 0000000000000000000000000000000000000000 e69de29bb2d1d6434b8b29ae775ad8c2e48c5391 b.txt
? c.txt
`,
			want: Repo{Path: "repo", Branch: "main", Dirty: 2},
		},
		{
			name: "new repo without commits",
			out: `# branch.oid (initial)
# branch.head main
? a.txt
`,
			want: Repo{Path: "repo", Branch: "main", Dirty: 1},
		},
		{
			name: "detached head",
			out: `# branch.oid cb8e74a0f1d2c3b4a5968778695a4b3c2d1e0f9a
# branch.head (detached)
`,
			want: Repo{Path: "repo", Branch: "(detached)"},
		},
		{
			name: "empty output",
			out:  "",
			want: Repo{Path: "repo"},
		},
		{
			name:    "ahead is not a number",
			out:     "# branch.head main\n# branch.ab +x -0\n",
			wantErr: true,
		},
		{
			name:    "behind is missing",
			out:     "# branch.head main\n# branch.ab +1\n",
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := parseStatus("repo", tt.out)
			if tt.wantErr {
				if err == nil {
					t.Fatalf("got %+v, want an error", got)
				}
				return
			}
			if err != nil {
				t.Fatalf("parseStatus returned error: %v", err)
			}
			if got != tt.want {
				t.Errorf("got %+v, want %+v", got, tt.want)
			}
		})
	}
}

func TestRead(t *testing.T) {
	if _, err := exec.LookPath("git"); err != nil {
		t.Skip("git is not installed")
	}
	dir := t.TempDir()
	if out, err := exec.Command("git", "init", "-q", "-b", "main", dir).CombinedOutput(); err != nil {
		t.Fatalf("git init: %v\n%s", err, out)
	}
	if err := os.WriteFile(filepath.Join(dir, "a.txt"), []byte("hi\n"), 0o644); err != nil {
		t.Fatal(err)
	}

	got, err := Read(context.Background(), dir)
	if err != nil {
		t.Fatalf("Read returned error: %v", err)
	}
	want := Repo{Path: dir, Branch: "main", Dirty: 1}
	if got != want {
		t.Errorf("got %+v, want %+v", got, want)
	}
}

func TestReadNotARepo(t *testing.T) {
	if _, err := exec.LookPath("git"); err != nil {
		t.Skip("git is not installed")
	}
	dir := t.TempDir()

	_, err := Read(context.Background(), dir)
	if err == nil {
		t.Fatal("got no error, want one")
	}
	if !strings.Contains(err.Error(), "not a git repository") {
		t.Errorf("err = %v, want git's own message in it", err)
	}
}
