// Package tui implements the Bubble Tea TUI for clu.
package tui

import (
	"fmt"
	"os/exec"
	"runtime"
	"strings"

	"github.com/charmbracelet/bubbles/key"
	"github.com/charmbracelet/bubbles/spinner"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
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

type tab int

const (
	tabCatalog tab = iota
	tabArticles
	tabAbout
)

var tabNames = []string{"Catalog", "Articles", "About"}

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

	loading  bool
	spinner  spinner.Model
	fetch    FetchFn
	err      error
	jsonMode bool // true only when user pressed ^J
}

func newLoadingModel(fetch FetchFn) Model {
	s := spinner.New()
	s.Spinner = spinner.Dot
	s.Style = lipgloss.NewStyle().Foreground(colorGreen)
	return Model{
		loading: true,
		spinner: s,
		fetch:   fetch,
	}
}

func (m Model) Init() tea.Cmd {
	return tea.Batch(m.spinner.Tick, m.loadData())
}

func (m Model) loadData() tea.Cmd {
	return func() tea.Msg {
		catalog, articles, about, err := m.fetch()
		return dataMsg{catalog: catalog, articles: articles, about: about, err: err}
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

	case spinner.TickMsg:
		if m.loading {
			var cmd tea.Cmd
			m.spinner, cmd = m.spinner.Update(msg)
			return m, cmd
		}
		return m, nil

	case tea.KeyMsg:
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
				m.about.viewport.LineUp(5)
				return m, nil
			}
		case key.Matches(msg, keys.ScrollDown):
			if m.activeTab == tabAbout {
				m.about.viewport.LineDown(5)
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

func (m Model) View() string {
	if m.err != nil {
		return lipgloss.NewStyle().
			Foreground(colorCopper).
			Padding(1, 2).
			Render(fmt.Sprintf("Error loading data: %v\n\nPress q to quit.", m.err))
	}
	if m.loading {
		return m.loadingView()
	}
	sep := styleDim.Render(strings.Repeat("─", m.width))
	return lipgloss.JoinVertical(lipgloss.Left,
		m.headerView(),
		sep,
		m.contentView(),
		sep,
		m.statusView(),
	)
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
	return styleStatusBar.Width(m.width).Render(strings.Join(bindings, "  "))
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
func Run(fetch FetchFn) (*Model, error) {
	m := newLoadingModel(fetch)
	p := tea.NewProgram(m, tea.WithAltScreen())
	final, err := p.Run()
	if err != nil {
		return nil, err
	}
	fm := final.(Model)
	return &fm, nil
}
