# CLAUDE.md

This file provides guidance to Claude Code (claude.ai/code) when working with code in this repository.

## Project

`clu` is a TUI + pipeline CLI for exploring the [commandlineuser.com](https://commandlineuser.com) catalog
and feature articles. It is a Go rewrite of a bash v0.3.0 tool. The bash source lives at
`../command-line-user-site/clu` and serves as the functional spec.

Module: `github.com/ali5ter/clu` | Author: Alister Lewis-Bowen \<<alister@lewis-bowen.org>\>

---

## Commands

```bash
# Build
go build ./...
go build -o clu .

# Run
go run . [--json] [--offline] [--no-tty]

# Test
go test ./...
go test ./internal/api/...          # single package
go test -run TestFetchCatalog ./... # single test

# Lint
golangci-lint run

# Release (via GoReleaser)
goreleaser release --snapshot --clean  # local dry-run
goreleaser release                     # tagged release
```

---

## Architecture

`clu` operates in two modes detected at startup via TTY check:

- **TUI mode** — launched when stdout is a terminal; Bubble Tea app
- **Pipeline mode** — when piped/redirected or `--no-tty`; emits JSON to stdout

```text
main.go           TTY detection → dispatch to tui.Run() or pipeline.EmitJSON()
cmd/root.go       Cobra flag setup (--json, --offline, --no-tty, --version)
internal/api/     HTTP client + data types for catalog, articles, about endpoints
internal/cache/   Read/write JSON snapshots to ~/.cache/clu/ (used by --offline)
internal/tui/     Bubble Tea app: model, views, keys, Lip Gloss styles
internal/pipeline/ JSON serialisation for pipeline mode
config/           ~/.config/clu/config.toml loading via Viper
```

### TUI structure (Bubble Tea / Elm architecture)

`internal/tui/model.go` holds the top-level `Model` with a `currentTab` field routing
`Update` and `View` calls to one of three sub-models. The `jsonMode bool` field is set
only when Ctrl-J is pressed; `SelectedCatalogItem()` returns nil unless `jsonMode` is true.

- `catalog.go` — list (Bubbles) + detail viewport (Bubbles); left/right split
- `articles.go` — list + Glamour-rendered markdown viewport
- `about.go` — static Lip Gloss-styled text

Key bindings are centralised in `keys.go`. All colours/borders come from `styles.go`.

### Data flow

```text
API endpoints (commandlineuser.com/api/v1/)
  └─ internal/api/client.go  → types.go structs
        └─ internal/cache/   (persist for --offline)
              └─ tui or pipeline
```

API targets (migrate to `/api/v1/` — current endpoints are unversioned):

| Endpoint             | Type            |
| -------------------- | --------------- |
| `/api/catalog.json`  | `[]CatalogItem` |
| `/api/articles.json` | `[]Article`     |
| `/api/about.json`    | `About`         |

---

## Key Design Decisions

### Colour palette (site palette → Lip Gloss)

```text
text:   #e9eff3   foreground
muted:  #a8b6c0   secondary text
panel:  #161d24   background panels
line:   #2b3742   borders
green:  #66b08a   accent / selected / highlight
copper: #c9895e   prompt / pointer
steel:  #7f93a6   headers / info
```

All palette constants live in `internal/tui/styles.go`.

### UI layout

```text
┌──────────────────────────────────────────────────────────────────┐
│  clu  ·  commandlineuser.com             142 tools  6 articles   │  header (Lip Gloss)
├──────────────────────────┬───────────────────────────────────────┤
│ > filter...              │  tool name                            │  textinput (Bubbles)
├──────────────────────────┤  ──────────────────────────────────── │
│ ripgrep    9.45  search  │  Summary text, wrapping naturally.    │  list + viewport (Bubbles)
│ ▶ fzf      8.77  shell   │  Score: 8.77   Confidence: 0.91       │
│   zoxide   8.61  nav     │  Maturity: stable   Maintenance: active│
├──────────────────────────┴───────────────────────────────────────┤
│  ↑↓ navigate  / filter  Enter open  ^J json  ? help  q quit     │  status bar (Lip Gloss)
└──────────────────────────────────────────────────────────────────┘
```

Tabs or key bindings switch between Catalog / Articles / About views.

### Charmbracelet stack

| Library                                                  | Role                                  |
| -------------------------------------------------------- | ------------------------------------- |
| [Bubble Tea](https://github.com/charmbracelet/bubbletea) | Core TUI — Elm architecture           |
| [Lip Gloss](https://github.com/charmbracelet/lipgloss)   | Styling, layout, borders              |
| [Bubbles](https://github.com/charmbracelet/bubbles)      | list, viewport, textinput, spinner    |
| [Glamour](https://github.com/charmbracelet/glamour)      | Markdown rendering (articles preview) |

### Behaviours preserved from bash v0.3.0

- Catalog: filterable list + detail right pane; Enter = open URL; Ctrl-J = emit JSON
- Articles: filterable list + rendered article body right pane; Enter = open URL
- `--offline` flag falls back to local cache snapshots
- ASCII art banner with live item counts

### Distribution model

1. Tag release → GoReleaser builds macOS/Linux arm64+amd64 binaries
2. GitHub Actions pushes tarballs to `commandlineuser.com/releases/` (DreamHost SFTP)
3. `releases/version.json` updated with latest version string
4. `commandlineuser.com/install` detects platform and installs binary
5. GoReleaser updates `ali5ter/homebrew-clu` formula automatically via `HOMEBREW_TAP_TOKEN` secret

---

## References

- Bubble Tea examples: <https://github.com/charmbracelet/bubbletea/tree/main/examples>
- Bubbles list: <https://github.com/charmbracelet/bubbles/tree/master/list>
- GoReleaser: <https://goreleaser.com>
- Bash source (functional spec): `../command-line-user-site/clu`

---

## Standards Audit — 2026-04-25

Issues generated from `/audit-standards` review against `~/.claude/CLAUDE.md`.
Suggested fix order:

### Group 1 — Correctness (fix first)

- [#1](https://github.com/ali5ter/clu/issues/1): markdown: add `.markdownlint.json` configuration
- [#2](https://github.com/ali5ter/clu/issues/2): markdown: fix lint errors in `README.md`
- [#3](https://github.com/ali5ter/clu/issues/3): markdown: fix lint errors in `CLAUDE.md`

### Group 2 — Standards Compliance

- [#4](https://github.com/ali5ter/clu/issues/4): bash: add Google-Style header docs to `run_vhs.sh`

### Group 3 — Quality Improvements

- [#5](https://github.com/ali5ter/clu/issues/5): bash: replace `echo` with `pfb` in `run_vhs.sh`
