// Package config reads binnacle's TOML config file.
package config

import (
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/BurntSushi/toml"
)

// TypeGitRepo is the source type for local git repositories.
const TypeGitRepo = "gitrepo"

// supportedTypes is shown to the user when a source has an unknown type.
// Every type here needs a case in checkSource.
var supportedTypes = []string{TypeGitRepo}

// Errors returned by Load. They come wrapped with the file and the source
// they belong to, so check them with errors.Is.
var (
	ErrUnknownKeys     = errors.New("unknown keys")
	ErrNoSources       = errors.New("no sources configured")
	ErrNameEmpty       = errors.New("name is empty")
	ErrNameDuplicate   = errors.New("name is used by another source")
	ErrTypeEmpty       = errors.New("type is empty")
	ErrUnsupportedType = errors.New("type is unsupported")
	ErrNoPaths         = errors.New("no paths")
	ErrPathEmpty       = errors.New("path is empty")
	ErrPathOtherUser   = errors.New("only ~ and ~/ are supported, not ~user")
)

// Config is the whole config file.
type Config struct {
	Sources []Source `toml:"source"`
}

// Source is one [[source]] block: one thing binnacle watches.
type Source struct {
	Type  string   `toml:"type"`
	Name  string   `toml:"name"`
	Paths []string `toml:"paths"`
}

// DefaultPath returns ~/.config/binnacle/config.toml.
func DefaultPath() (string, error) {
	home, err := os.UserHomeDir()
	if err != nil {
		return "", fmt.Errorf("find home folder: %w", err)
	}
	return filepath.Join(home, ".config", "binnacle", "config.toml"), nil
}

// Load reads and checks the config file at path. It reports every problem in
// the file at once. In the returned Config, every path is absolute, meaning that
// ~ is replaced by the home folder, and relative paths are taken from the folder
// that holds the config file.
func Load(path string) (Config, error) {
	path, err := filepath.Abs(path)
	if err != nil {
		return Config{}, fmt.Errorf("read config %s: %w", path, err)
	}

	var cfg Config
	meta, err := toml.DecodeFile(path, &cfg)
	if err != nil {
		return Config{}, fmt.Errorf("read config %s: %w", path, err)
	}
	if keys := meta.Undecoded(); len(keys) > 0 {
		return Config{}, fmt.Errorf("read config %s: %w: %v", path, ErrUnknownKeys, keys)
	}
	if err := validate(cfg); err != nil {
		return Config{}, fmt.Errorf("check config %s:\n%w", path, err)
	}

	home, err := os.UserHomeDir()
	if err != nil {
		return Config{}, fmt.Errorf("find home folder: %w", err)
	}
	base := filepath.Dir(path)
	for i := range cfg.Sources {
		paths := cfg.Sources[i].Paths
		for j := range paths {
			paths[j] = resolve(paths[j], home, base)
		}
	}
	return cfg, nil
}

// validate checks every source and returns all problems joined, each
// prefixed with the source's number and name.
func validate(cfg Config) error {
	if len(cfg.Sources) == 0 {
		return ErrNoSources
	}

	var errs []error
	seen := make(map[string]bool)
	for i, s := range cfg.Sources {
		problems := checkSource(s)
		if seen[s.Name] && strings.TrimSpace(s.Name) != "" {
			problems = append(problems, ErrNameDuplicate)
		}
		seen[s.Name] = true

		for _, p := range problems {
			errs = append(errs, fmt.Errorf("source %d (%q): %w", i+1, s.Name, p))
		}
	}
	return errors.Join(errs...)
}

// checkSource returns the problems in one source, without its number.
func checkSource(s Source) []error {
	var errs []error
	if strings.TrimSpace(s.Name) == "" {
		errs = append(errs, ErrNameEmpty)
	}

	switch s.Type {
	case "":
		errs = append(errs, ErrTypeEmpty)
	case TypeGitRepo:
		errs = append(errs, checkPaths(s.Paths)...)
	default:
		errs = append(errs, fmt.Errorf("%w: %q (supported: %s)",
			ErrUnsupportedType, s.Type, strings.Join(supportedTypes, ", ")))
	}
	return errs
}

// checkPaths checks a gitrepo source's paths before they are resolved.
func checkPaths(paths []string) []error {
	if len(paths) == 0 {
		return []error{ErrNoPaths}
	}

	var errs []error
	for i, p := range paths {
		switch {
		case strings.TrimSpace(p) == "":
			errs = append(errs, fmt.Errorf("paths[%d]: %w", i, ErrPathEmpty))
		case strings.HasPrefix(p, "~") && p != "~" && !strings.HasPrefix(p, "~/"):
			errs = append(errs, fmt.Errorf("paths[%d] %q: %w", i, p, ErrPathOtherUser))
		}
	}
	return errs
}

// resolve makes path absolute: ~ and ~/ are replaced by home, and a
// relative path is joined to base, the config file's folder.
func resolve(path, home, base string) string {
	if path == "~" {
		return home
	}
	if rest, ok := strings.CutPrefix(path, "~/"); ok {
		return filepath.Join(home, rest)
	}
	if filepath.IsAbs(path) {
		return filepath.Clean(path)
	}
	return filepath.Join(base, path)
}
