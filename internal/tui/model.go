// Package tui implements the Bubble Tea TUI for clu.
package tui

import (
	"fmt"
	"os/exec"
	"runtime"
	"strings"
	"time"

	tea "charm.land/bubbletea/v2"
	"charm.land/bubbles/v2/key"
	"charm.land/bubbles/v2/spinner"
	"charm.land/lipgloss/v2"
	"github.com/ali5ter/clu/internal/api"
)

// FetchFn is the signature for the data-loading function passed to Run.
type FetchFn func() ([]api.CatalogItem, []api.Article, *api.About, error)

// dataMsg carries loaded data back into the model.
type dataMsg struct {
	catalog  []api.CatalogItem
	articles []api.Article
	about    *api.About
	err      error
}

// versionMsg carries the latest available version string.
type versionMsg struct{ latest string }

type tab int

const (
	tabCatalog tab = iota
	tabArticles
	tabAbout
)

var tabNames = []string{"Catalog", "Articles", "About"}

// Tab-contextual footer animations — all 4 chars wide so the layout stays stable.
var (
	// animCatalog: dot bouncing along a track — browsing through a list.
	animCatalog = spinner.Spinner{
		Frames: []string{"∘───", "─∘──", "──∘─", "───∘", "──∘─", "─∘──"},
		FPS:    time.Second / 8,
	}
	// animArticles: dots rippling left-to-right — text being read.
	animArticles = spinner.Spinner{
		Frames: []string{"·   ", "··  ", "··· ", "····", " ···", "  ··", "   ·", "  ··", " ···", "····", "··· ", "··  "},
		FPS:    time.Second / 10,
	}
	// animAbout: slow heartbeat pulse — quiet, human.
	animAbout = spinner.Spinner{
		Frames: []string{"    ", "·   ", "●   ", "·   ", "    ", "    "},
		FPS:    time.Second / 3,
	}
)

// Model is the top-level Bubble Tea model.
type Model struct {
	catalog  catalogModel
	articles articlesModel
	about    aboutModel

	activeTab tab
	width     int
	height    int

	catalogItems []api.CatalogItem
	articleItems []api.Article

	loading      bool
	spinner      spinner.Model
	animCatalog  spinner.Model
	animArticles spinner.Model
	animAbout    spinner.Model
	fetch        FetchFn
	err          error
	jsonMode     bool // true only when user pressed ^J

	currentVersion string
	latestVersion  string
	checkVersion   func() string
}

func newLoadingModel(fetch FetchFn, currentVersion string, checkVersion func() string) Model {
	s := spinner.New()
	s.Spinner = spinner.Dot
	s.Style = lipgloss.NewStyle().Foreground(colorGreen)

	ac := spinner.New()
	ac.Spinner = animCatalog
	ac.Style = lipgloss.NewStyle().Foreground(colorSteel)

	ar := spinner.New()
	ar.Spinner = animArticles
	ar.Style = lipgloss.NewStyle().Foreground(colorSteel)

	ab := spinner.New()
	ab.Spinner = animAbout
	ab.Style = lipgloss.NewStyle().Foreground(colorSteel)

	return Model{
		loading:        true,
		spinner:        s,
		animCatalog:    ac,
		animArticles:   ar,
		animAbout:      ab,
		fetch:          fetch,
		currentVersion: currentVersion,
		checkVersion:   checkVersion,
	}
}

func (m Model) Init() tea.Cmd {
	return tea.Batch(
		func() tea.Msg { return m.spinner.Tick() },
		func() tea.Msg { return m.animCatalog.Tick() },
		func() tea.Msg { return m.animArticles.Tick() },
		func() tea.Msg { return m.animAbout.Tick() },
		m.loadData(), m.doVersionCheck(),
	)
}

func (m Model) loadData() tea.Cmd {
	return func() tea.Msg {
		catalog, articles, about, err := m.fetch()
		return dataMsg{catalog: catalog, articles: articles, about: about, err: err}
	}
}

func (m Model) doVersionCheck() tea.Cmd {
	if m.checkVersion == nil {
		return nil
	}
	return func() tea.Msg {
		return versionMsg{latest: m.checkVersion()}
	}
}

func (m Model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.WindowSizeMsg:
		m.width = msg.Width
		m.height = msg.Height
		if !m.loading {
			m.catalog.SetSize(msg.Width, msg.Height)
			m.articles.SetSize(msg.Width, msg.Height)
			m.about.SetSize(msg.Width, msg.Height)
		}
		return m, nil

	case dataMsg:
		if msg.err != nil {
			m.err = msg.err
			m.loading = false
			return m, nil
		}
		m.catalogItems = msg.catalog
		m.articleItems = msg.articles
		m.catalog = newCatalogModel(msg.catalog, m.width, m.height)
		m.articles = newArticlesModel(msg.articles, m.width, m.height)
		m.about = newAboutModel(msg.about, m.width, m.height)
		m.loading = false
		return m, nil

	case versionMsg:
		m.latestVersion = msg.latest
		return m, nil

	case spinner.TickMsg:
		var c1, c2, c3, c4 tea.Cmd
		m.spinner, c1 = m.spinner.Update(msg)
		m.animCatalog, c2 = m.animCatalog.Update(msg)
		m.animArticles, c3 = m.animArticles.Update(msg)
		m.animAbout, c4 = m.animAbout.Update(msg)
		return m, tea.Batch(c1, c2, c3, c4)

	case tea.KeyPressMsg:
		if key.Matches(msg, keys.Quit) {
			return m, tea.Quit
		}
		if m.loading || m.err != nil {
			return m, nil
		}

		switch {
		case key.Matches(msg, keys.TabNext):
			m.activeTab = (m.activeTab + 1) % tab(len(tabNames))
			return m, nil

		case key.Matches(msg, keys.TabPrev):
			m.activeTab = (m.activeTab + tab(len(tabNames)) - 1) % tab(len(tabNames))
			return m, nil

		case key.Matches(msg, keys.Open):
			if url := m.activeURL(); url != "" {
				openBrowser(url)
			}
			return m, nil

		case key.Matches(msg, keys.JSON):
			if m.activeTab == tabCatalog {
				if m.catalog.SelectedJSON() != nil {
					m.jsonMode = true
					return m, tea.Quit
				}
			}
			return m, nil

		case key.Matches(msg, keys.ScrollUp):
			if m.activeTab == tabAbout {
				m.about.viewport.ScrollUp(5)
				return m, nil
			}
		case key.Matches(msg, keys.ScrollDown):
			if m.activeTab == tabAbout {
				m.about.viewport.ScrollDown(5)
				return m, nil
			}
		}
	}

	if m.loading {
		return m, nil
	}

	var cmd tea.Cmd
	switch m.activeTab {
	case tabCatalog:
		m.catalog, cmd = m.catalog.Update(msg)
	case tabArticles:
		m.articles, cmd = m.articles.Update(msg)
	case tabAbout:
		m.about, cmd = m.about.Update(msg)
	}
	return m, cmd
}

func (m Model) View() tea.View {
	var content string
	if m.err != nil {
		content = lipgloss.NewStyle().
			Foreground(colorCopper).
			Padding(1, 2).
			Render(fmt.Sprintf("Error loading data: %v\n\nPress q to quit.", m.err))
	} else if m.loading {
		content = m.loadingView()
	} else {
		sep := styleDim.Render(strings.Repeat("─", m.width))
		content = lipgloss.JoinVertical(lipgloss.Left,
			m.headerView(),
			sep,
			m.contentView(),
			sep,
			m.statusView(),
		)
	}
	v := tea.NewView(content)
	v.AltScreen = true
	v.WindowTitle = "clu"
	return v
}

func (m Model) loadingView() string {
	// Site-palette gradient banner — cyan → blue-violet → magenta
	c1 := lipgloss.NewStyle().Foreground(lipgloss.Color("#50d2ff"))
	c2 := lipgloss.NewStyle().Foreground(lipgloss.Color("#8c96ff"))
	c3 := lipgloss.NewStyle().Foreground(lipgloss.Color("#b464ff"))
	dim := lipgloss.NewStyle().Foreground(lipgloss.Color("#646478"))

	banner := strings.Join([]string{
		c1.Render(" _____  _     _   _ "),
		c1.Render("/  __ \\| |   | | | |"),
		c2.Render("| /  \\/| |   | | | |"),
		c2.Render("| |    | |   | | | |"),
		c3.Render("| \\__/\\| |___| |_| |"),
		c3.Render(" \\____/\\_____/\\___/ "),
		dim.Render(" commandlineuser.com"),
	}, "\n")

	status := lipgloss.NewStyle().
		Foreground(colorMuted).
		Render(m.spinner.View() + " fetching catalog…")

	content := lipgloss.JoinVertical(lipgloss.Center, banner, "", status)

	return lipgloss.Place(
		m.width, m.height,
		lipgloss.Center, lipgloss.Center,
		content,
	)
}

func (m Model) headerView() string {
	logo := styleHeaderAccent.Render("clu")
	site := styleHeader.Render(" · commandlineuser.com")
	counts := styleHeader.Render(fmt.Sprintf("%d tools  %d articles",
		len(m.catalogItems), len(m.articleItems)))

	var tabs strings.Builder
	for i, name := range tabNames {
		if tab(i) == m.activeTab {
			tabs.WriteString(styleTabActive.Render(name))
		} else {
			tabs.WriteString(styleTabInactive.Render(name))
		}
	}

	left := lipgloss.JoinHorizontal(lipgloss.Center, logo, site, "  ", tabs.String())
	padding := m.width - lipgloss.Width(left) - lipgloss.Width(counts)
	if padding < 0 {
		padding = 0
	}

	return lipgloss.NewStyle().
		Background(colorPanel).
		Width(m.width).
		Render(left + strings.Repeat(" ", padding) + counts)
}

func (m Model) contentView() string {
	switch m.activeTab {
	case tabCatalog:
		return m.catalog.View()
	case tabArticles:
		return m.articles.View()
	case tabAbout:
		return m.about.View()
	}
	return ""
}

func (m Model) statusView() string {
	bindings := []string{
		styleStatusKey.Render("↑↓") + " navigate",
		styleStatusKey.Render("/") + " filter",
		styleStatusKey.Render("enter") + " open",
		styleStatusKey.Render("^J") + " json",
		styleStatusKey.Render("⇧↑↓") + " scroll",
		styleStatusKey.Render("tab") + " switch",
		styleStatusKey.Render("q") + " quit",
	}
	left := strings.Join(bindings, "  ")

	var right string
	currentNorm := strings.TrimPrefix(m.currentVersion, "v")
	latestNorm := strings.TrimPrefix(m.latestVersion, "v")
	if latestNorm != "" && latestNorm != currentNorm {
		right = lipgloss.NewStyle().
			Background(colorCopper).
			Foreground(colorPanel).
			Padding(0, 1).
			Render("↑ v" + latestNorm + " available")
	} else {
		var anim string
		switch m.activeTab {
		case tabCatalog:
			anim = m.animCatalog.View()
		case tabArticles:
			anim = m.animArticles.View()
		case tabAbout:
			anim = m.animAbout.View()
		}
		right = lipgloss.NewStyle().Foreground(colorSteel).Render(anim)
	}

	// styleStatusBar has Padding(0,1) so inner content width = m.width - 2
	contentWidth := m.width - 2
	gap := contentWidth - lipgloss.Width(left) - lipgloss.Width(right)
	if gap < 1 {
		gap = 1
	}

	return styleStatusBar.Width(m.width).Render(left + strings.Repeat(" ", gap) + right)
}

func (m Model) activeURL() string {
	switch m.activeTab {
	case tabCatalog:
		return m.catalog.SelectedURL()
	case tabArticles:
		return m.articles.SelectedURL()
	}
	return ""
}

// SelectedCatalogItem returns the item to emit as JSON, or nil if ^J was not pressed.
func (m Model) SelectedCatalogItem() *api.CatalogItem {
	if !m.jsonMode {
		return nil
	}
	return m.catalog.SelectedJSON()
}

func openBrowser(url string) {
	var cmd string
	var args []string
	switch runtime.GOOS {
	case "darwin":
		cmd, args = "open", []string{url}
	default:
		cmd, args = "xdg-open", []string{url}
	}
	exec.Command(cmd, args...).Start() //nolint:errcheck
}

// Run starts the Bubble Tea program with async data loading and returns the final model.
func Run(fetch FetchFn, currentVersion string, checkVersion func() string) (*Model, error) {
	m := newLoadingModel(fetch, currentVersion, checkVersion)
	p := tea.NewProgram(m)
	final, err := p.Run()
	if err != nil {
		return nil, err
	}
	fm := final.(Model)
	return &fm, nil
}
