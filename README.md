# Lollipop

<p>
  <img src="assets/banner.png" alt="Lollipop" width="150">
</p>

<p>
  <a href="https://pkg.go.dev/github.com/kidkafka9/lollipop"><img src="https://pkg.go.dev/badge/github.com/kidkafka9/lollipop.svg" alt="Go Reference"></a>
</p>

A terminal dashboard for everything you keep an eye on. 🍭

Lollipop puts the tools you check all day into one screen.

<!-- Demo GIF goes here once the panels exist. Record it with VHS: https://github.com/charmbracelet/vhs -->

> [!NOTE]
> Lollipop is in early development.

## Features

- **Git repositories at a glance.** Lollipop finds every repository under the
  folders you choose, at any depth, and shows its branch, changed files, and
  commits ahead of or behind its upstream.
- **One small config file.** A short TOML file lists what to watch.
- **Polite in the background.** Lollipop reads git without taking its index lock,
  so it never gets in the way of your own commits.
- **Built for many tools.** Every source type is separate, so new ones (Kubernetes,
  Grafana, CI pipelines) are easy to pull in without changing the others.

## Installation

You need [Go](https://go.dev/dl/) 1.27 or newer, and `git` for the git
repository source.

```bash
go install github.com/kidkafka9/lollipop/cmd/lollipop@latest
```

Or build it from source:

```bash
git clone https://github.com/kidkafka9/lollipop.git
cd lollipop
go build ./cmd/lollipop
```

## Usage

```bash
lollipop                          # reads ~/.config/lollipop/config.toml
lollipop --config ./my-config.toml
```

Press `q` or `ctrl+c` to quit.

## Configuration

Lollipop reads a TOML file. Start from the example:

```bash
mkdir -p ~/.config/lollipop
cp config.example.toml ~/.config/lollipop/config.toml
```

Each `[[source]]` block is one thing to watch. Add as many as you like:

```toml
[[source]]
type  = "gitrepo"
name  = "work"
paths = ["~/work/my-team", "~/work/platform"]

[[source]]
type  = "gitrepo"
name  = "personal"
paths = ["~/code"]
```

Every source has a `type` and a `name`. The name is the panel title, so it
must be unique. The other keys depend on the type.

Paths may start with `~/` for your home folder. Relative paths are taken from
the folder the config file is in.

### Source types

| Type      | Shows                                                         | Keys                                  |
| --------- | ------------------------------------------------------------- | ------------------------------------- |
| `gitrepo` | every git repository under `paths`: branch, changes, ahead/behind | `paths`: folders to search, at any depth |

More types are planned. Secrets such as API tokens will never go in this file;
they will be read from environment variables.

## License

[MIT](LICENSE)

---

Lollipop is built with [Bubble Tea](https://github.com/charmbracelet/bubbletea),
[Lip Gloss](https://github.com/charmbracelet/lipgloss) and
[Bubbles](https://github.com/charmbracelet/bubbles), and uses the CharmTone
palette by [Charm](https://charm.sh).
