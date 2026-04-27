<!-- markdownlint-disable-file MD041 -->

![clu demo](examples/clu_demo.gif)

# clu

Yes, it's a command-line tool for a website about command-line tools. I know.

[commandlineuser.com](https://commandlineuser.com) catalogues and scores CLI tools. It seemed only right
that you should be able to explore it without leaving the terminal — and that the output should be
pipeable into the very tools it's recommending.

That's the joke. It also happens to be useful.

---

## Install

**Homebrew** (macOS and Linux):

```bash
brew tap ali5ter/clu
brew install clu
```

**Binary** (macOS and Linux, no Go required):

```bash
# macOS Apple Silicon
curl -sL https://github.com/ali5ter/clu/releases/latest/download/clu_darwin_arm64.tar.gz | tar -xz
sudo mv clu /usr/local/bin/

# macOS Intel
curl -sL https://github.com/ali5ter/clu/releases/latest/download/clu_darwin_amd64.tar.gz | tar -xz
sudo mv clu /usr/local/bin/

# Linux arm64
curl -sL https://github.com/ali5ter/clu/releases/latest/download/clu_linux_arm64.tar.gz | tar -xz
sudo mv clu /usr/local/bin/

# Linux amd64
curl -sL https://github.com/ali5ter/clu/releases/latest/download/clu_linux_amd64.tar.gz | tar -xz
sudo mv clu /usr/local/bin/
```

**Go**:

```bash
go install github.com/ali5ter/clu@latest
```

---

## Usage

Run it in a terminal and you get the TUI:

```bash
clu
```

Pipe it somewhere and it emits JSON:

```bash
clu | jq '.[] | select(.score >= 9)'
```

### Flags

| Flag        | What it does                                       |
| ----------- | -------------------------------------------------- |
| `--json`    | Force JSON output even in a terminal               |
| `--offline` | Use the local cache — no network request           |
| `--no-tty`  | Pipeline mode regardless of how you're running it  |

---

## TUI

Three tabs — Catalog, Articles, About — switched with `tab` and `shift+tab`.

### Catalog

A scored list of CLI tools on the left; detail on the right. The score colours tell you
something: green means the tool rated highly on utility, maturity, and maintenance.
Grey means it's worth knowing about but didn't break into the top tier.

Filter with `/` — matches name, category, and tags.

### Articles

Feature articles from the site rendered in the terminal. Navigate the list on the left,
read on the right. Scroll the article with `shift+↑` and `shift+↓`.

### About

Site description, author bio, catalogue methodology, and links — rendered with the same
Glamour markdown engine as the Articles pane.

### Key bindings

| Key                      | Action                                |
| ------------------------ | ------------------------------------- |
| `↑` / `↓` or `j` / `k`  | Navigate list                         |
| `/`                      | Filter                                |
| `enter`                  | Open in browser                       |
| `ctrl+j`                 | Output selected item as JSON and exit |
| `shift+↑` / `shift+↓`   | Scroll detail pane                    |
| `tab` / `shift+tab`      | Switch tabs                           |
| `q`                      | Quit                                  |

If a newer version of clu is available, a badge appears in the right of the footer.

---

## Pipeline mode

`ctrl+j` in the TUI exits and emits the selected catalogue item as JSON — so you can
browse visually and hand off to `jq`, `fzf`, or anything else that reads stdin.

`--json` skips the TUI entirely and dumps the full catalogue:

```bash
clu --json | jq '.[0]'
```

Cached data lives in `~/.cache/clu/`. Use `--offline` if you're on a plane.

---

## Built with

- [Bubble Tea](https://github.com/charmbracelet/bubbletea) — Elm-architecture TUI framework
- [Lip Gloss](https://github.com/charmbracelet/lipgloss) — terminal styling and layout
- [Bubbles](https://github.com/charmbracelet/bubbles) — list, viewport, textinput components
- [Glamour](https://github.com/charmbracelet/glamour) — markdown rendering

---

## commandlineuser.com

The catalogue is updated nightly. Every item has a score, a confidence value, and an
explanation of what drove both. No curation by follower count or GitHub stars alone.

[commandlineuser.com](https://commandlineuser.com) ·
[Methodology](https://commandlineuser.com/about/methodology) ·
[About](https://commandlineuser.com/about)
