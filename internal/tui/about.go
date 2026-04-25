package tui

import (
	"fmt"
	"strings"

	"github.com/charmbracelet/bubbles/viewport"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	wrap "github.com/muesli/reflow/wordwrap"
	"github.com/ali5ter/clu/internal/api"
)

type aboutModel struct {
	viewport viewport.Model
	about    *api.About
	width    int
	height   int
}

func newAboutModel(about *api.About, width, height int) aboutModel {
	vp := viewport.New(width, height-5)
	m := aboutModel{viewport: vp, about: about, width: width, height: height}
	m.viewport.SetContent(m.render())
	return m
}

func (m aboutModel) render() string {
	if m.about == nil {
		return styleDetailMuted.Render("No about data available.")
	}
	a := m.about
	wrapWidth := m.viewport.Width - 4
	if wrapWidth < 20 {
		wrapWidth = 20
	}

	var sb strings.Builder

	section := func(title string) {
		sb.WriteString("\n" + styleSteel.Render(title) + "\n\n")
	}
	para := func(text string) {
		sb.WriteString("  " + styleDetailValue.Render(wrap.String(text, wrapWidth)) + "\n")
	}
	muted := func(text string) {
		sb.WriteString("  " + styleDetailMuted.Render(wrap.String(text, wrapWidth)) + "\n")
	}
	link := func(label, value string) {
		sb.WriteString(fmt.Sprintf("  %s  %s\n",
			styleDetailLabel.Width(14).Render(label),
			styleDetailMuted.Render(value),
		))
	}

	section("commandlineuser.com")
	para(a.Site.Description)

	section("About")
	for _, bio := range a.Author.Bio {
		para(bio)
	}

	section("How the catalogue works")
	para(a.Catalogue.Description)
	sb.WriteString("\n")
	muted(a.Catalogue.Curation)

	section("Links")
	link("About:", a.Site.AboutURL)
	link("Methodology:", a.Site.MethodologyURL)
	link("Features:", a.Site.FeaturesURL)
	link("GitHub:", a.Author.Links.GitHub)

	return sb.String()
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
	m.viewport.SetContent(m.render())
}
