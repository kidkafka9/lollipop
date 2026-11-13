package config

import (
	"errors"
	"io/fs"
	"os"
	"path/filepath"
	"slices"
	"strings"
	"testing"

	"github.com/BurntSushi/toml"
)

func TestLoad(t *testing.T) {
	tests := []struct {
		name string
		toml string
		want []Source
	}{
		{
			name: "one source",
			toml: `
[[source]]
type  = "gitrepo"
name  = "team"
paths = ["/work/team"]
`,
			want: []Source{{Type: "gitrepo", Name: "team", Paths: []string{"/work/team"}}},
		},
		{
			name: "two sources keep file order",
			toml: `
[[source]]
type  = "gitrepo"
name  = "b"
paths = ["/b"]

[[source]]
type  = "gitrepo"
name  = "a"
paths = ["/a"]
`,
			want: []Source{
				{Type: "gitrepo", Name: "b", Paths: []string{"/b"}},
				{Type: "gitrepo", Name: "a", Paths: []string{"/a"}},
			},
		},
		{
			name: "paths are made absolute",
			toml: `
[[source]]
type  = "gitrepo"
name  = "mixed"
paths = ["~", "~/code", "/abs/../tmp", "notes", "./repos/x"]
`,
			want: []Source{{Type: "gitrepo", Name: "mixed", Paths: []string{
				"{home}", "{home}/code", "/tmp", "{dir}/notes", "{dir}/repos/x",
			}}},
		},
		{
			name: "comments are fine",
			toml: `
# my repos
[[source]]
type  = "gitrepo" # local only
name  = "team"
paths = [
  "/work/a", # first
  "/work/b",
]
`,
			want: []Source{{Type: "gitrepo", Name: "team", Paths: []string{"/work/a", "/work/b"}}},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			home := t.TempDir()
			t.Setenv("HOME", home)
			path := writeConfig(t, tt.toml)
			dir := filepath.Dir(path)

			cfg, err := Load(path)
			if err != nil {
				t.Fatalf("Load returned error: %v", err)
			}
			if len(cfg.Sources) != len(tt.want) {
				t.Fatalf("got %d sources, want %d: %+v", len(cfg.Sources), len(tt.want), cfg.Sources)
			}
			for i, got := range cfg.Sources {
				want := tt.want[i]
				want.Paths = nil
				for _, p := range tt.want[i].Paths {
					want.Paths = append(want.Paths, fill(p, home, dir))
				}
				if got.Type != want.Type || got.Name != want.Name || !slices.Equal(got.Paths, want.Paths) {
					t.Errorf("source %d = %+v, want %+v", i, got, want)
				}
			}
		})
	}
}

func TestLoadErrors(t *testing.T) {
	tests := []struct {
		name string
		toml string
		want []error
	}{
		{
			name: "empty file",
			toml: ``,
			want: []error{ErrNoSources},
		},
		{
			name: "only comments",
			toml: "# nothing yet\n",
			want: []error{ErrNoSources},
		},
		{
			name: "misspelled key",
			toml: `
[[source]]
type  = "gitrepo"
name  = "team"
pahts = ["/work"]
`,
			want: []error{ErrUnknownKeys},
		},
		{
			name: "plural table name",
			toml: `
[[sources]]
type  = "gitrepo"
name  = "team"
paths = ["/work"]
`,
			want: []error{ErrUnknownKeys},
		},
		{
			name: "name missing",
			toml: `
[[source]]
type  = "gitrepo"
paths = ["/work"]
`,
			want: []error{ErrNameEmpty},
		},
		{
			name: "name only spaces",
			toml: `
[[source]]
type  = "gitrepo"
name  = "   "
paths = ["/work"]
`,
			want: []error{ErrNameEmpty},
		},
		{
			name: "same name twice",
			toml: `
[[source]]
type  = "gitrepo"
name  = "team"
paths = ["/a"]

[[source]]
type  = "gitrepo"
name  = "team"
paths = ["/b"]
`,
			want: []error{ErrNameDuplicate},
		},
		{
			name: "type missing",
			toml: `
[[source]]
name  = "team"
paths = ["/work"]
`,
			want: []error{ErrTypeEmpty},
		},
		{
			name: "type unknown",
			toml: `
[[source]]
type  = "jira"
name  = "tickets"
`,
			want: []error{ErrUnsupportedType},
		},
		{
			name: "type in wrong case",
			toml: `
[[source]]
type  = "GitRepo"
name  = "team"
paths = ["/work"]
`,
			want: []error{ErrUnsupportedType},
		},
		{
			name: "paths missing",
			toml: `
[[source]]
type = "gitrepo"
name = "team"
`,
			want: []error{ErrNoPaths},
		},
		{
			name: "paths empty list",
			toml: `
[[source]]
type  = "gitrepo"
name  = "team"
paths = []
`,
			want: []error{ErrNoPaths},
		},
		{
			name: "one path empty",
			toml: `
[[source]]
type  = "gitrepo"
name  = "team"
paths = ["/work", ""]
`,
			want: []error{ErrPathEmpty},
		},
		{
			name: "other user's home",
			toml: `
[[source]]
type  = "gitrepo"
name  = "team"
paths = ["~bob/code"]
`,
			want: []error{ErrPathOtherUser},
		},
		{
			name: "every problem is reported at once",
			toml: `
[[source]]
type = "gitrepo"

[[source]]
type  = "jira"
name  = "tickets"

[[source]]
type  = "gitrepo"
name  = "tickets"
paths = [""]
`,
			want: []error{ErrNameEmpty, ErrNoPaths, ErrUnsupportedType, ErrNameDuplicate, ErrPathEmpty},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			cfg, err := Load(writeConfig(t, tt.toml))
			if err == nil {
				t.Fatalf("got %+v, want an error", cfg)
			}
			for _, want := range tt.want {
				if !errors.Is(err, want) {
					t.Errorf("err = %v\nwant it to contain %q", err, want)
				}
			}
			if cfg.Sources != nil {
				t.Errorf("got sources %+v along with the error, want none", cfg.Sources)
			}
		})
	}
}

func TestLoadBadTOML(t *testing.T) {
	tests := []struct {
		name string
		toml string
	}{
		{name: "string without quotes", toml: "[[source]]\ntype = gitrepo\n"},
		{name: "paths is not a list", toml: "[[source]]\ntype = \"gitrepo\"\nname = \"a\"\npaths = \"/work\"\n"},
		{name: "key set twice", toml: "[[source]]\nname = \"a\"\nname = \"b\"\n"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			_, err := Load(writeConfig(t, tt.toml))
			if err == nil {
				t.Fatal("got no error, want one")
			}
			t.Log(err)
		})
	}
}

func TestLoadSyntaxErrorType(t *testing.T) {
	_, err := Load(writeConfig(t, "[[source]]\ntype = gitrepo\n"))
	if _, ok := errors.AsType[toml.ParseError](err); !ok {
		t.Errorf("err = %v, want a toml.ParseError inside", err)
	}
}

func TestLoadMissingFile(t *testing.T) {
	_, err := Load(filepath.Join(t.TempDir(), "missing.toml"))
	if !errors.Is(err, fs.ErrNotExist) {
		t.Errorf("err = %v, want fs.ErrNotExist", err)
	}
}

func TestExampleConfigLoads(t *testing.T) {
	cfg, err := Load(filepath.Join("..", "..", "config.example.toml"))
	if err != nil {
		t.Fatalf("config.example.toml does not load: %v", err)
	}
	if len(cfg.Sources) == 0 {
		t.Error("config.example.toml has no sources")
	}
}

func TestDefaultPath(t *testing.T) {
	home := t.TempDir()
	t.Setenv("HOME", home)

	got, err := DefaultPath()
	if err != nil {
		t.Fatal(err)
	}
	want := filepath.Join(home, ".config", "lollipop", "config.toml")
	if got != want {
		t.Errorf("got %q, want %q", got, want)
	}
}

func writeConfig(t *testing.T, content string) string {
	t.Helper()
	path := filepath.Join(t.TempDir(), "config.toml")
	if err := os.WriteFile(path, []byte(content), 0o644); err != nil {
		t.Fatal(err)
	}
	return path
}

func fill(p, home, dir string) string {
	if p == "{home}" {
		return home
	}
	if rest, ok := strings.CutPrefix(p, "{home}/"); ok {
		return filepath.Join(home, rest)
	}
	if rest, ok := strings.CutPrefix(p, "{dir}/"); ok {
		return filepath.Join(dir, rest)
	}
	return p
}

func TestSupportedTypesHaveACase(t *testing.T) {
	for _, typ := range supportedTypes {
		s := Source{Type: typ, Name: "x", Paths: []string{"/x"}}
		for _, err := range checkSource(s) {
			if errors.Is(err, ErrUnsupportedType) {
				t.Errorf("type %q is in supportedTypes but has no case in checkSource", typ)
			}
		}
	}
}
