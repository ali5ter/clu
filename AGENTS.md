# Repository Guidelines

## Project Overview

`clu` is a Go CLI/TUI for exploring the [commandlineuser.com](https://commandlineuser.com) catalogue of
command-line tools and its feature articles. It is a Go rewrite of a bash v0.3.0 tool (functional spec at
`../command-line-user-site/clu`, a sibling repo, not part of this module).

Two modes selected at startup by TTY detection:

- **TUI mode** (stdout is a terminal) — full Bubble Tea app with Catalog / Articles / About tabs.
- **Pipeline mode** (stdout piped/redirected, or `--no-tty`/`--json`) — dumps the catalogue as JSON, meant
  to be piped into `jq`, `fzf`, etc.

Module: `github.com/ali5ter/clu` · Author: Alister Lewis-Bowen `<alister@lewis-bowen.org>`.

## Architecture & Data Flow

```text
main.go            TTY detection → dispatch (main.go:9 calls cmd.Execute())
cmd/root.go        Cobra flags + run() orchestration
config/config.go   ~/.config/clu/config.toml via Viper
internal/api/       HTTP client + types for catalog/articles/about
internal/cache/     JSON snapshot persistence (os.UserCacheDir()/clu) for --offline
internal/pipeline/  JSON serialisation for pipeline mode
internal/tui/       Bubble Tea app (Elm architecture): model, per-tab views, keys, styles
```

Call chain: `cmd.Execute()` (`cmd/root.go:29,47`) → `run()` loads `config.Load()` → builds
`api.NewClient()` (8s HTTP timeout, `internal/api/client.go:20`) → `loadData()` (`cmd/root.go:87`) reads
from `internal/cache` when `--offline`, otherwise calls `client.FetchCatalog/Articles/About` and
best-effort writes the results back to cache → branches on `--json || --no-tty || !isTTY()`
(`cmd/tty.go:9`) to either `internal/pipeline.EmitCatalog` or `tui.Run(fetchFn, version, checkVersionFn)`
(`internal/tui/model.go:412`). After the TUI exits, `root.go` checks
`finalModel.SelectedCatalogItem()` (set only via Ctrl-J) and emits it via `pipeline.EmitItem`.

**Elm architecture** lives entirely in `internal/tui/model.go`'s `Model` (`Init`/`Update`/`View`
implement `tea.Model`). `Update` type-switches on `WindowSizeMsg` / `dataMsg` / `versionMsg` /
`spinner.TickMsg` / `tea.KeyPressMsg`, then delegates to the active tab's own (non-`tea.Model`) sub-model
`Update`. Sub-models: `catalogModel` (`internal/tui/catalog.go:91`), `articlesModel`
(`internal/tui/articles.go:71`), `aboutModel` (`internal/tui/about.go:12`) — same shape, invoked manually
by the parent, not registered with Bubble Tea directly.

Data flow to the TUI: `api.Client` → `internal/cache` (persisted for `--offline`) → `tui` or `pipeline`.
API endpoints are currently unversioned (`/api/catalog.json`, `/api/articles.json`, `/api/about.json`);
a migration to `/api/v1/` is planned but not yet done.

## Key Directories

| Path              | Purpose                                                                          |
| ------------------ | -------------------------------------------------------------------------------- |
| `cmd/`             | Cobra root command (`root.go`), TTY detection (`tty.go`)                        |
| `config/`          | Viper-backed config loader (`config.go`) — `~/.config/clu/config.toml`          |
| `internal/api/`    | HTTP client (`client.go`) + response types (`types.go`)                         |
| `internal/cache/`  | Read/write JSON snapshots for offline mode (`cache.go`)                         |
| `internal/pipeline/` | JSON serialisation for non-TTY output (`json.go`)                             |
| `internal/tui/`    | Bubble Tea app — model, per-tab views, `keys.go`, `styles.go`                    |
| `examples/`        | VHS demo tooling — `run_vhs.sh` builds `clu` and renders `.tape` scripts into `clu_demo.gif` |

## Development Commands

```bash
# Build
go build ./...
go build -o clu .

# Run
go run . [--json] [--offline] [--no-tty]

# Test (no test suite currently exists in the repo — see Testing & QA)
go test ./...

# Lint (not wired into CI; run locally)
golangci-lint run

# Release dry-run / real release (goreleaser-driven; CI only runs on `v*` tags)
goreleaser release --snapshot --clean
goreleaser release --clean

# Regenerate the README demo GIF (requires https://github.com/charmbracelet/vhs)
cd examples && ./run_vhs.sh          # renders every *.tape (except config.tape)
cd examples && ./run_vhs.sh clu_demo.tape   # render one tape
```

CI (`.github/workflows/release.yml`) only triggers on `v*` tag pushes: checkout → `setup-go` (version
pinned by `go.mod`) → `goreleaser release --clean`. There is **no** build/test/lint gate in CI — verify
locally before tagging.

## Code Conventions & Common Patterns

- **Packages**: short, singular, lowercase, matching their directory (`api`, `cache`, `cmd`, `config`,
  `pipeline`, `tui`). Small deliberate exported surface; everything else unexported.
- **Files**: one responsibility per file — `tty.go`, `keys.go`, `styles.go`, one file per TUI tab
  (`catalog.go`, `articles.go`, `about.go`).
- **Error handling**: `fmt.Errorf("...: %w", err)` wrapping throughout (e.g. `internal/api/client.go:26`,
  `internal/pipeline/json.go:14`, `cmd/root.go:95`). No sentinel errors, no panics. Non-critical failures
  are deliberately swallowed: `config.Load()` errors are logged as a warning only (`cmd/root.go:49-51`);
  `cache.Write` errors are discarded with `_ =` (best-effort persistence, `cmd/root.go:119-121`); Glamour
  render failures fall back to raw text (`internal/tui/articles.go:145-149`,
  `internal/tui/about.go:66-70`); `openBrowser`'s `exec.Start()` error is ignored with `//nolint:errcheck`.
  The only fatal exit path is `cmd.Execute()` → `os.Exit(1)`.
- **Dependency injection**: no custom Go interfaces — the seam is function types. `tui.FetchFn`
  (`internal/tui/model.go:19`) is injected by `cmd/root.go` so `internal/tui` never imports
  `internal/api`/`internal/cache` directly; a `checkVersion func() string` closure is injected the same
  way. Bubbles library interfaces are satisfied via small adapter types: `catalogItem`/`articleItem`
  implement `list.Item`; `catalogDelegate`/`articlesDelegate` implement `list.ItemDelegate`.
- **Concurrency**: no raw goroutines/channels/mutexes. All async work is `tea.Cmd` (`func() tea.Msg`)
  closures run by the Bubble Tea runtime — `m.loadData()` and `m.doVersionCheck()`
  (`internal/tui/model.go:129,136`) wrap blocking synchronous HTTP calls and return `dataMsg`/`versionMsg`.
  Four `spinner.Model` instances (one per tab, distinct animation each) tick concurrently via
  `tea.Batch(...)`; only the active tab's animation renders.
- **TUI state**: `type tab int` enum (`tabCatalog`/`tabArticles`/`tabAbout`, `internal/tui/model.go:32-37`)
  with `Model.activeTab` cycled by `TabNext`/`TabPrev` (modulo wraparound). Each tab embeds its own
  Bubbles components (`list.Model`, `textinput.Model` filter, `viewport.Model` detail) plus a
  `filtering bool` mode flag that toggles between list-navigation and filter-input key handling.
  `jsonMode bool` records a Ctrl-J press for post-exit JSON emission; width/height propagate top-down via
  `WindowSizeMsg` into each sub-model.
- **Filter scope**: matches name, category, and tags only — the summary field is intentionally excluded
  to avoid false-positive matches.

## Important Files

| File                          | Role                                                                 |
| ------------------------------ | --------------------------------------------------------------------- |
| `main.go`                      | Entry point; TTY-based dispatch                                      |
| `cmd/root.go`                  | Cobra command, flags (`--json`, `--offline`, `--no-tty`), `run()` orchestration |
| `cmd/tty.go`                   | `isTTY()` helper (`golang.org/x/term`)                               |
| `config/config.go`             | Viper config loader; defaults `theme=auto`, `cache_ttl=3600`, `browser=""` |
| `internal/api/client.go`       | `api.Client`, `NewClient()`, `FetchCatalog/Articles/About`, `FetchLatestVersion` |
| `internal/api/types.go`        | `CatalogItem`, `Article`, `About` response structs                   |
| `internal/cache/cache.go`      | `Read`/`Write` JSON snapshots under `os.UserCacheDir()/clu`           |
| `internal/pipeline/json.go`    | `EmitCatalog`, `EmitItem`                                             |
| `internal/tui/model.go`        | Top-level `Model` (Elm architecture), `Run()` entry point             |
| `internal/tui/keys.go`         | Centralised key bindings                                              |
| `internal/tui/styles.go`       | Lip Gloss colour palette (`#e9eff3` text, `#66b08a` accent, `#c9895e` copper) |
| `.goreleaser.yml`               | Cross-build (darwin/linux × amd64/arm64), Homebrew tap publish        |
| `.github/workflows/release.yml` | Sole CI workflow — release-on-tag only                                |

## Runtime/Tooling Preferences

- Go **1.26.2** (pinned in `go.mod`; CI uses `go-version-file: go.mod`).
- Standard `go` toolchain — no Bun/Node involved; this is a pure Go module.
- No Makefile/justfile/task runner — commands are run directly (`go build`, `go run`, `goreleaser`).
- Key dependencies (`go.mod`): `charm.land/bubbletea/v2` (TUI runtime), `charm.land/bubbles/v2` (list,
  viewport, textinput, spinner, key components), `charm.land/lipgloss/v2` (styling), `charm.land/glamour/v2`
  (Markdown rendering for Articles/About), `github.com/muesli/reflow` (word wrap), `github.com/spf13/cobra`
  (CLI), `github.com/spf13/viper` (TOML config), `golang.org/x/term` (TTY detection).
- Config file: `~/.config/clu/config.toml`. Cache dir: `os.UserCacheDir()/clu` (e.g.
  `~/Library/Caches/clu` on macOS, `~/.cache/clu` on Linux — despite the doc comment in `cache.go:1`
  saying `~/.cache/clu/` specifically).
- Demo GIF regeneration requires the external [`vhs`](https://github.com/charmbracelet/vhs) binary
  (`examples/run_vhs.sh`).

## Testing & QA

- **No `*_test.go` files exist in this repo currently**, and no test framework is wired into `go.mod`
  (a stale `stretchr/testify` checksum sits unused in `go.sum`). CI does not run `go test`, `go vet`, or
  any linter — `.github/workflows/release.yml` goes straight from checkout to `goreleaser release --clean`
  on tag push.
- If adding tests, use the standard library `testing` package with colocated `_test.go` files
  (`go test ./...`, `go test ./internal/api/...` for one package, `go test -run TestName ./...` for one
  test) — this is the convention documented in the project's local `CLAUDE.md` even though no tests exist
  yet.
- Lint locally with `golangci-lint run` before tagging a release; it is not enforced by CI.
- `goreleaser release --snapshot --clean` is the closest thing to an integration check — it exercises the
  full darwin/linux × amd64/arm64 build matrix without publishing.
</content>
