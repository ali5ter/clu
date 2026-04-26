package tui

import (
	"strings"

	"github.com/charmbracelet/bubbles/viewport"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/glamour"
	"github.com/charmbracelet/lipgloss"
	"github.com/ali5ter/clu/internal/api"
)

type aboutModel struct {
	viewport viewport.Model
	about    *api.About
	renderer *glamour.TermRenderer
	width    int
	height   int
}

func newAboutModel(about *api.About, width, height int) aboutModel {
	vp := viewport.New(width, height-5)
	renderer, _ := glamour.NewTermRenderer(
		glamour.WithAutoStyle(),
		glamour.WithWordWrap(width-4),
	)
	m := aboutModel{
		viewport: vp,
		about:    about,
		renderer: renderer,
		width:    width,
		height:   height,
	}
	m.viewport.SetContent(m.render())
	return m
}

func (m aboutModel) render() string {
	if m.about == nil {
		return styleDetailMuted.Render("No about data available.")
	}
	a := m.about
	var sb strings.Builder

	sb.WriteString("# commandlineuser.com\n\n")
	sb.WriteString(a.Site.Description + "\n\n")

	sb.WriteString("## About\n\n")
	for _, bio := range a.Author.Bio {
		sb.WriteString(bio + "\n\n")
	}

	sb.WriteString("## How the catalogue works\n\n")
	sb.WriteString(a.Catalogue.Description + "\n\n")
	sb.WriteString(a.Catalogue.Curation + "\n\n")

	sb.WriteString("## Links\n\n")
	sb.WriteString("- [About](" + a.Site.AboutURL + ")\n")
	sb.WriteString("- [Methodology](" + a.Site.MethodologyURL + ")\n")
	sb.WriteString("- [Features](" + a.Site.FeaturesURL + ")\n")
	sb.WriteString("- [GitHub](" + a.Author.Links.GitHub + ")\n")

	if m.renderer == nil {
		return sb.String()
	}
	rendered, err := m.renderer.Render(sb.String())
	if err != nil {
		return sb.String()
	}
	return rendered
}

func (m aboutModel) Update(msg tea.Msg) (aboutModel, tea.Cmd) {
	var cmd tea.Cmd
	m.viewport, cmd = m.viewport.Update(msg)
	return m, cmd
}

func (m aboutModel) View() string {
	return lipgloss.NewStyle().Padding(0, 1).Render(m.viewport.View())
}

func (m *aboutModel) SetSize(width, height int) {
	m.width = width
	m.height = height
	m.viewport.Width = width
	m.viewport.Height = height - 5
	if m.renderer != nil {
		m.renderer, _ = glamour.NewTermRenderer(
			glamour.WithAutoStyle(),
			glamour.WithWordWrap(width-4),
		)
	}
	m.viewport.SetContent(m.render())
}
