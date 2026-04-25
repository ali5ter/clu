package tui

import (
	"fmt"
	"strings"

	"github.com/charmbracelet/bubbles/key"
	"github.com/charmbracelet/bubbles/list"
	"github.com/charmbracelet/bubbles/textinput"
	"github.com/charmbracelet/bubbles/viewport"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/glamour"
	"github.com/charmbracelet/lipgloss"
	"github.com/ali5ter/clu/internal/api"
)

type articleItem struct {
	api.Article
}

func (i articleItem) Title() string       { return i.Article.Title }
func (i articleItem) Description() string { return i.Published }
func (i articleItem) FilterValue() string {
	return i.Article.Title + " " + strings.Join(i.Tags, " ")
}

type articlesModel struct {
	list      list.Model
	filter    textinput.Model
	detail    viewport.Model
	articles  []api.Article
	renderer  *glamour.TermRenderer
	width     int
	height    int
	filtering bool
}

func newArticlesModel(articles []api.Article, width, height int) articlesModel {
	listItems := make([]list.Item, len(articles))
	for i, a := range articles {
		listItems[i] = articleItem{a}
	}

	delegate := list.NewDefaultDelegate()
	delegate.Styles.SelectedTitle = styleSelected
	delegate.Styles.SelectedDesc = styleDetailMuted
	delegate.Styles.NormalTitle = styleNormal
	delegate.Styles.NormalDesc = styleDetailMuted

	listWidth := width / 2
	contentHeight := height - 4

	l := list.New(listItems, delegate, listWidth, contentHeight)
	l.SetShowTitle(false)
	l.SetShowHelp(false)
	l.SetFilteringEnabled(false)

	fi := textinput.New()
	fi.Placeholder = "filter…"
	fi.PromptStyle = styleFilterPrompt
	fi.TextStyle = styleFilterText
	fi.Prompt = "> "

	vp := viewport.New(width-listWidth, contentHeight)

	renderer, _ := glamour.NewTermRenderer(
		glamour.WithAutoStyle(),
		glamour.WithWordWrap(width-listWidth-2),
	)

	m := articlesModel{
		list:     l,
		filter:   fi,
		detail:   vp,
		articles: articles,
		renderer: renderer,
		width:    width,
		height:   height,
	}
	m.updateDetail()
	return m
}

func (m *articlesModel) selectedArticle() *api.Article {
	if i, ok := m.list.SelectedItem().(articleItem); ok {
		return &i.Article
	}
	return nil
}

func (m *articlesModel) updateDetail() {
	a := m.selectedArticle()
	if a == nil {
		m.detail.SetContent("")
		return
	}

	meta := fmt.Sprintf("**%s**\n\n_%s_ · %s\n\n---\n\n",
		a.Title, a.Published, strings.Join(a.Tags, ", "))

	rendered, err := m.renderer.Render(meta + a.Body)
	if err != nil {
		rendered = meta + a.Body
	}
	m.detail.SetContent(rendered)
	m.detail.GotoTop()
}

func (m articlesModel) Update(msg tea.Msg) (articlesModel, tea.Cmd) {
	var cmds []tea.Cmd

	switch msg := msg.(type) {
	case tea.KeyMsg:
		if m.filtering {
			switch msg.String() {
			case "enter", "esc":
				m.filtering = false
				m.filter.Blur()
				m.applyFilter()
				return m, nil
			default:
				var cmd tea.Cmd
				m.filter, cmd = m.filter.Update(msg)
				cmds = append(cmds, cmd)
				m.applyFilter()
				return m, tea.Batch(cmds...)
			}
		}

		switch {
		case key.Matches(msg, keys.Filter):
			m.filtering = true
			m.filter.Focus()
			return m, textinput.Blink
		case key.Matches(msg, keys.Up):
			var cmd tea.Cmd
			m.list, cmd = m.list.Update(msg)
			cmds = append(cmds, cmd)
			m.updateDetail()
		case key.Matches(msg, keys.Down):
			var cmd tea.Cmd
			m.list, cmd = m.list.Update(msg)
			cmds = append(cmds, cmd)
			m.updateDetail()
		case key.Matches(msg, keys.ScrollUp):
			m.detail.LineUp(5)
		case key.Matches(msg, keys.ScrollDown):
			m.detail.LineDown(5)
		}
	}

	var cmd tea.Cmd
	m.detail, cmd = m.detail.Update(msg)
	cmds = append(cmds, cmd)

	return m, tea.Batch(cmds...)
}

func (m *articlesModel) applyFilter() {
	q := strings.ToLower(m.filter.Value())
	var filtered []list.Item
	for _, a := range m.articles {
		ai := articleItem{a}
		if q == "" || strings.Contains(strings.ToLower(ai.FilterValue()+" "+a.Summary), q) {
			filtered = append(filtered, ai)
		}
	}
	m.list.SetItems(filtered)
	m.updateDetail()
}

func (m articlesModel) View() string {
	listWidth := m.width / 2
	detailWidth := m.width - listWidth

	filterBar := styleFilterPrompt.Render("> ") + m.filter.View()
	if !m.filtering {
		filterBar = styleDetailMuted.Render("> " + m.filter.Value())
		if m.filter.Value() == "" {
			filterBar = styleDetailMuted.Render("> filter…")
		}
	}

	listPane := stylePanelBorder.Width(listWidth).Render(
		lipgloss.JoinVertical(lipgloss.Left,
			filterBar,
			m.list.View(),
		),
	)

	detailPane := lipgloss.NewStyle().Width(detailWidth).Padding(0, 1).Render(m.detail.View())

	return lipgloss.JoinHorizontal(lipgloss.Top, listPane, detailPane)
}

func (m articlesModel) SelectedURL() string {
	if a := m.selectedArticle(); a != nil {
		return "https://commandlineuser.com" + a.URL
	}
	return ""
}

func (m *articlesModel) SetSize(width, height int) {
	m.width = width
	m.height = height
	listWidth := width / 2
	contentHeight := height - 4
	m.list.SetSize(listWidth, contentHeight)
	m.detail.Width = width - listWidth
	m.detail.Height = contentHeight
	if m.renderer != nil {
		m.renderer, _ = glamour.NewTermRenderer(
			glamour.WithAutoStyle(),
			glamour.WithWordWrap(width-listWidth-2),
		)
	}
	m.updateDetail()
}
