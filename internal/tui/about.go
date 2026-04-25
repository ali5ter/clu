package tui

import (
	"fmt"
	"strings"

	"github.com/charmbracelet/bubbles/viewport"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/ali5ter/clu/internal/api"
)

type aboutModel struct {
	viewport viewport.Model
	about    *api.About
	width    int
	height   int
}

func newAboutModel(about *api.About, width, height int) aboutModel {
	vp := viewport.New(width, height-4)
	m := aboutModel{
		viewport: vp,
		about:    about,
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

	section := func(title string) {
		sb.WriteString("\n" + styleSteel.Render(title) + "\n\n")
	}
	line := func(label, value string) {
		sb.WriteString(fmt.Sprintf("  %s  %s\n",
			styleDetailLabel.Width(14).Render(label),
			styleDetailValue.Render(value),
		))
	}

	section("commandlineuser.com")
	sb.WriteString("  " + styleDetailValue.Render(a.Site.Description) + "\n")

	section("About")
	for _, bio := range a.Author.Bio {
		sb.WriteString("  " + styleDetailValue.Render(bio) + "\n")
	}

	section("How the catalogue works")
	sb.WriteString("  " + styleDetailValue.Render(a.Catalogue.Description) + "\n\n")
	sb.WriteString("  " + styleDetailMuted.Render(a.Catalogue.Curation) + "\n")

	section("Links")
	line("About:", a.Site.AboutURL)
	line("Methodology:", a.Site.MethodologyURL)
	line("Features:", a.Site.FeaturesURL)
	line("GitHub:", a.Author.Links.GitHub)

	return sb.String()
}

func (m aboutModel) Update(msg tea.Msg) (aboutModel, tea.Cmd) {
	var cmd tea.Cmd
	m.viewport, cmd = m.viewport.Update(msg)
	return m, cmd
}

func (m aboutModel) View() string {
	return m.viewport.View()
}

func (m *aboutModel) SetSize(width, height int) {
	m.width = width
	m.height = height
	m.viewport.Width = width
	m.viewport.Height = height - 4
	m.viewport.SetContent(m.render())
}
