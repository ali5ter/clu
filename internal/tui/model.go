// Package tui implements the Bubble Tea TUI for clu.
package tui

import (
	"fmt"
	"os/exec"
	"runtime"
	"strings"

	"github.com/charmbracelet/bubbles/key"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	"github.com/ali5ter/clu/internal/api"
)

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

	catalogItems  []api.CatalogItem
	articleItems  []api.Article
	aboutData     *api.About
}

// New constructs the top-level model with fetched data.
func New(catalog []api.CatalogItem, articles []api.Article, about *api.About, width, height int) Model {
	return Model{
		catalog:      newCatalogModel(catalog, width, height),
		articles:     newArticlesModel(articles, width, height),
		about:        newAboutModel(about, width, height),
		catalogItems: catalog,
		articleItems: articles,
		aboutData:    about,
		width:        width,
		height:       height,
	}
}

func (m Model) Init() tea.Cmd {
	return nil
}

func (m Model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.WindowSizeMsg:
		m.width = msg.Width
		m.height = msg.Height
		m.catalog.SetSize(msg.Width, msg.Height)
		m.articles.SetSize(msg.Width, msg.Height)
		m.about.SetSize(msg.Width, msg.Height)
		return m, nil

	case tea.KeyMsg:
		switch {
		case key.Matches(msg, keys.Quit):
			return m, tea.Quit

		case key.Matches(msg, keys.TabNext):
			m.activeTab = (m.activeTab + 1) % tab(len(tabNames))
			return m, nil

		case key.Matches(msg, keys.TabPrev):
			m.activeTab = (m.activeTab + tab(len(tabNames)) - 1) % tab(len(tabNames))
			return m, nil

		case key.Matches(msg, keys.Open):
			url := m.activeURL()
			if url != "" {
				openBrowser(url)
			}
			return m, nil

		case key.Matches(msg, keys.JSON):
			if m.activeTab == tabCatalog {
				if it := m.catalog.SelectedJSON(); it != nil {
					return m, tea.Quit
				}
			}
			return m, nil
		}
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
	header := m.headerView()
	content := m.contentView()
	status := m.statusView()
	return lipgloss.JoinVertical(lipgloss.Left, header, content, status)
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

// SelectedCatalogItem returns the currently selected catalog item (for ^J JSON output).
func (m Model) SelectedCatalogItem() *api.CatalogItem {
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

// Run starts the Bubble Tea program and returns the final model.
func Run(catalog []api.CatalogItem, articles []api.Article, about *api.About) (*Model, error) {
	m := New(catalog, articles, about, 0, 0)
	p := tea.NewProgram(m, tea.WithAltScreen())
	final, err := p.Run()
	if err != nil {
		return nil, err
	}
	fm := final.(Model)
	return &fm, nil
}
