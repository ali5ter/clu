# CLAUDE Project Context — clu

## Project

- **Name:** clu
- **Purpose:** A TUI + pipeline CLI for exploring the commandlineuser.com catalog and feature articles
- **Language:** Go
- **Distribution:** Binary releases served via commandlineuser.com; install script at commandlineuser.com/install
- **Target repo:** Private GitHub repo `ali5ter/clu` (not yet created — see Bootstrapping below)

---

## Background

`clu` started as a bash script living inside the `ali5ter/command-line-user-site` repo, distributed as a
static file at `commandlineuser.com/clu`. The bash version (v0.3.0) used `fzf` for interactive browsing
with a preview pane. It worked but felt limited — terminal restore issues with `fzf`, no real layout
control, and dependency on `pfb` (bash text formatting library) for all output styling.

The decision was made to rewrite `clu` as a proper Go TUI using the Charmbracelet ecosystem. Inspiration:
K9s, btop, Claude Code, Gemini-CLI, Codex — tools that feel native in the terminal and have real panel
layouts, animations, and keyboard routing.

---

## Goals for the Go Rewrite

1. **Charmbracelet TUI** — side-by-side panels, live filtering, keyboard navigation, markdown rendering
2. **Automatic mode detection** — if stdout is a TTY, launch TUI; if piped/redirected, emit JSON
3. **Pipeline compatibility** — `clu --json | jq ...` works exactly as before; `--no-tty` forces it
4. **Single binary** — no runtime dependencies (no fzf, pfb, jq, bat)
5. **Cross-platform** — macOS arm64/amd64, Linux arm64/amd64
6. **commandlineuser.com as CDN** — binaries pushed to site on release; install script handles platform detection
7. **Homebrew tap** — `brew tap ali5ter/tap && brew install clu` as a first-class install path
8. **Self-update** — `clu update` checks and replaces the binary in place
9. **Versioned API contract** — fetch from `commandlineuser.com/api/v1/catalog.json` (not `/api/catalog.json`)
10. **Config file** — `~/.config/clu/config.toml` for theme, cache TTL, browser preference

---

## Charmbracelet Stack

| Library | Purpose |
|---------|---------|
| [Bubble Tea](https://github.com/charmbracelet/bubbletea) | Core TUI framework — Elm architecture (Model, Update, View) |
| [Lip Gloss](https://github.com/charmbracelet/lipgloss) | Styling — colours, borders, layout, padding |
| [Bubbles](https://github.com/charmbracelet/bubbles) | Pre-built components: list, viewport, textinput, spinner |
| [Glamour](https://github.com/charmbracelet/glamour) | Terminal markdown rendering — for feature article preview |
| [Harmonica](https://github.com/charmbracelet/harmonica) | Optional: spring-based animation for transitions |

---

## Site Palette

These CSS variables from commandlineuser.com should be reflected in the TUI colour scheme:

```
--text:   #e9eff3   (foreground)
--muted:  #a8b6c0   (secondary text)
--panel:  #161d24   (background panels)
--line:   #2b3742   (borders)
--green:  #66b08a   (accent / selected / highlight)
--copper: #c9895e   (prompt / pointer)
--steel:  #7f93a6   (headers / info)
```

---

## Data API

`clu` fetches three endpoints. These should be versioned at `/api/v1/` going forward:

| Endpoint | Content |
|----------|---------|
| `commandlineuser.com/api/catalog.json` | Array of catalog items (current, unversioned) |
| `commandlineuser.com/api/articles.json` | Array of feature articles (current, unversioned) |
| `commandlineuser.com/api/about.json` | Site/author/methodology metadata (current, unversioned) |

Catalog item shape (from `schemas/item.schema.json` in command-line-user-site repo):
- `name`, `summary`, `source_url`, `source_type` (github/npm/crates/rss/hn/youtube)
- `score` (0–10, two decimal places), `confidence` (0–1, two decimal places)
- `maturity` (emerging/growing/stable), `maintenance` (active/maintained/inactive)
- `category`, `tags[]`, `author_or_org`, `license`
- `trend` (up/down/null), `is_new` (bool)

Articles item shape:
- `slug`, `title`, `summary`, `published`, `tags[]`, `url`, `body` (markdown)

---

## Existing Bash Version Reference

The bash `clu` (v0.3.0) is the functional spec for the Go rewrite. Key behaviours to preserve:

- **Catalog view:** filterable list (left) + detail preview (right). Enter = open URL. Ctrl-J = emit JSON
- **Articles view:** filterable list (left) + rendered article body (right). Enter = open URL
- **About/methodology:** text output via pfb-style formatting (replace with Lip Gloss in Go)
- **Offline mode:** `--offline` flag; falls back to local snapshot files
- **Banner:** ASCII art logo with live catalog/article counts shown inline

The bash source is at `../command-line-user-site/clu` for reference.

---

## Proposed UI Layout

```
┌─────────────────────────────────────────────────────────────────┐
│  clu  ·  commandlineuser.com              142 tools  6 articles │  ← header bar (Lip Gloss)
├──────────────────────────┬──────────────────────────────────────┤
│ > filter...              │  tool name                           │  ← textinput (Bubbles)
├──────────────────────────┤  ─────────────────────────────────── │
│ ripgrep    9.45  search  │  Summary of the tool goes here,      │  ← list (Bubbles) + viewport (Bubbles)
│ fd         9.12  fs      │  wrapping naturally.                 │
│ bat        8.90  view    │                                      │
│ ▶ fzf      8.77  shell   │  Score: 8.77   Confidence: 0.91      │
│   zoxide   8.61  nav     │  Maturity: stable   Maintenance: active│
│   starship 8.44  shell   │  Category: shell    Source: github   │
│   ...                    │  Tags: fuzzy, search, interactive    │
│                          │  URL: github.com/junegunn/fzf        │
├──────────────────────────┴──────────────────────────────────────┤
│  ↑↓ navigate  /  filter  Enter open  ^J json  ? help  q quit   │  ← status bar (Lip Gloss)
└─────────────────────────────────────────────────────────────────┘
```

Tabs or key bindings to switch between Catalog / Articles / About views.

---

## Proposed Project Structure

```
clu/
├── CLAUDE.md
├── README.md
├── go.mod
├── go.sum
├── main.go                  # entry point, TTY detection, mode dispatch
├── cmd/
│   └── root.go              # cobra/flag setup, --json, --offline, --version, --help
├── internal/
│   ├── api/
│   │   ├── client.go        # fetch catalog, articles, about from API
│   │   └── types.go         # CatalogItem, Article, About structs
│   ├── cache/
│   │   └── cache.go         # local snapshot read/write (~/.cache/clu/)
│   ├── tui/
│   │   ├── model.go         # top-level Bubble Tea Model (state, tabs)
│   │   ├── catalog.go       # catalog list + detail view
│   │   ├── articles.go      # articles list + markdown preview view
│   │   ├── about.go         # about / methodology text view
│   │   ├── keys.go          # keybinding definitions
│   │   └── styles.go        # Lip Gloss styles (site palette)
│   └── pipeline/
│       └── json.go          # JSON output for pipeline mode
├── config/
│   └── config.go            # ~/.config/clu/config.toml loading (Viper)
├── scripts/
│   └── install.sh           # platform-detect + download binary from commandlineuser.com/releases/
├── .goreleaser.yaml         # cross-platform build + release config
└── .github/
    └── workflows/
        └── release.yml      # tag → GoReleaser → push binaries to commandlineuser.com
```

---

## Distribution Model

1. **Tag** a release in this repo (`v0.4.0`, etc.)
2. **GoReleaser** builds binaries for macOS arm64/amd64, Linux arm64/amd64
3. **GitHub Actions** pushes tarballs to `commandlineuser.com/releases/` via DreamHost SFTP (or equivalent)
4. **`releases/version.json`** is updated with the latest version string
5. **`commandlineuser.com/install`** script reads `version.json`, detects platform, downloads + installs binary
6. **Homebrew tap** (`ali5ter/homebrew-tap` repo) formula is updated by GoReleaser automatically

Install flows:
```bash
# one-shot, no install
bash <(curl -sL https://commandlineuser.com/clu)   # still works? — TBD; binary can't pipe this way

# install
curl -sL https://commandlineuser.com/install | bash

# homebrew
brew tap ali5ter/tap && brew install clu

# go install (if repo goes public)
go install github.com/ali5ter/clu@latest
```

Note: the `bash <(curl ...)` one-shot pattern doesn't work for binaries. Discuss whether to drop it or
keep a stub bash script that auto-installs and then runs.

---

## Bootstrapping (First Session Steps)

This repo is brand new (just `git init`, no remote yet). First session tasks:

1. Create private GitHub repo `ali5ter/clu` and push this repo to it
2. Initialise Go module: `go mod init github.com/ali5ter/clu`
3. Add Charmbracelet dependencies (bubbletea, lipgloss, bubbles, glamour)
4. Scaffold `main.go` with TTY detection and mode dispatch
5. Build the `internal/api` package — fetch and parse catalog/articles/about
6. Build the `internal/tui/styles.go` — Lip Gloss styles using site palette
7. Build the catalog list + detail view as the first working TUI screen
8. Wire up `--json` pipeline output
9. Set up `.goreleaser.yaml` and `release.yml` workflow

---

## Author

Alister Lewis-Bowen &lt;alister@lewis-bowen.org&gt;

---

## Key References

- Charmbracelet: <https://github.com/charmbracelet>
- Bubble Tea examples: <https://github.com/charmbracelet/bubbletea/tree/main/examples>
- Bubbles list component: <https://github.com/charmbracelet/bubbles/tree/master/list>
- GoReleaser: <https://goreleaser.com>
- commandlineuser.com site repo: `../command-line-user-site` (bash clu source at `../command-line-user-site/clu`)
