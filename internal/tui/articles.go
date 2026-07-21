package tui

import (
	"fmt"
	"io"
	"strings"

	tea "charm.land/bubbletea/v2"
	"charm.land/bubbles/v2/key"
	"charm.land/bubbles/v2/list"
	"charm.land/bubbles/v2/textinput"
	"charm.land/bubbles/v2/viewport"
	"charm.land/glamour/v2"
	"charm.land/lipgloss/v2"
	"github.com/ali5ter/clu/internal/api"
)

type articleItem struct{ api.Article }

func (i articleItem) Title() string       { return i.Article.Title }
func (i articleItem) Description() string { return "" }
func (i articleItem) FilterValue() string {
	return i.Article.Title + " " + strings.Join(i.Tags, " ")
}

type articlesDelegate struct{ width int }

func (d articlesDelegate) Height() int                             { return 2 }
func (d articlesDelegate) Spacing() int                            { return 1 }
func (d articlesDelegate) Update(_ tea.Msg, _ *list.Model) tea.Cmd { return nil }

func (d articlesDelegate) Render(w io.Writer, m list.Model, index int, item list.Item) {
	ai, ok := item.(articleItem)
	if !ok {
		return
	}
	selected := index == m.Index()

	titleWidth := d.width - 2
	var indicator, title string
	if selected {
		indicator = styleSelectedBar.Render("▎")
		title = styleListSelected.Width(titleWidth).Render(truncate(ai.Article.Title, titleWidth))
	} else {
		indicator = " "
		title = styleNormal.Width(titleWidth).Render(truncate(ai.Article.Title, titleWidth))
	}
	line1 := fmt.Sprintf("%s %s", indicator, title)

	tags := ""
	if len(ai.Tags) > 0 {
		t := ai.Tags
		if len(t) > 3 {
			t = t[:3]
		}
		tags = " · " + strings.Join(t, ", ")
	}
	meta := ai.Published + tags
	metaWidth := d.width - 2
	var line2 string
	if selected {
		line2 = indicator + " " + styleDetailMuted.Width(metaWidth).Render(truncate(meta, metaWidth))
	} else {
		line2 = "  " + styleDim.Width(metaWidth).Render(truncate(meta, metaWidth))
	}

	fmt.Fprintln(w, line1)
	fmt.Fprint(w, line2)
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

	listWidth := width * 35 / 100
	contentHeight := height - 5
	// subtract 3: filter bar(1) + filter separator(1) + pagination dots(1)
	itemsHeight := ((contentHeight - 3) / 3) * 3
	listHeight := itemsHeight + 1
	delegate := articlesDelegate{width: listWidth}
	l := list.New(listItems, delegate, listWidth, listHeight)
	l.SetShowTitle(false)
	l.SetShowHelp(false)
	l.SetShowStatusBar(false)
	l.SetFilteringEnabled(false)

	fi := textinput.New()
	fi.Placeholder = "filter…"
	fi.Prompt = "> "
	styles := textinput.DefaultDarkStyles()
	styles.Focused.Prompt = styleFilterPrompt
	styles.Focused.Text = styleFilterText
	styles.Blurred.Prompt = styleFilterPrompt
	styles.Blurred.Text = styleFilterText
	fi.SetStyles(styles)

	detailWidth := width - listWidth
	vp := viewport.New(viewport.WithWidth(detailWidth), viewport.WithHeight(contentHeight))

	renderer, _ := glamour.NewTermRenderer(
		glamour.WithStylePath("dark"),
		glamour.WithWordWrap(detailWidth-2),
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

	meta := fmt.Sprintf("**%s**\n\n_%s_", a.Title, a.Published)
	if len(a.Tags) > 0 {
		meta += " · " + strings.Join(a.Tags, ", ")
	}
	meta += "\n\n---\n\n"

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
	case tea.KeyPressMsg:
		if m.filtering {
			switch msg.String() {
			case "enter", "esc":
				m.filtering = false
				m.filter.Blur()
				m.applyFilter()
				return m, nil
			case "up", "k", "down", "j":
				m.filtering = false
				m.filter.Blur()
				m.applyFilter()
				// don't return — navigate the list below
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
			return m, m.filter.Focus()
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
			m.detail.ScrollUp(5)
		case key.Matches(msg, keys.ScrollDown):
			m.detail.ScrollDown(5)
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
		if q == "" || strings.Contains(strings.ToLower(ai.FilterValue()), q) {
			filtered = append(filtered, ai)
		}
	}
	m.list.SetItems(filtered)
	m.updateDetail()
}

func (m articlesModel) View() string {
	listWidth := m.width * 35 / 100
	detailWidth := m.width - listWidth

	var filterBar string
	if m.filtering {
		filterBar = m.filter.View()
	} else if m.filter.Value() != "" {
		filterBar = styleFilterPrompt.Render("> ") + styleFilterText.Render(m.filter.Value())
	} else {
		filterBar = styleDim.Render("> filter…")
	}

	filterSep := styleDim.Render(strings.Repeat("─", listWidth))
	listPane := stylePanelBorder.Width(listWidth).Render(
		lipgloss.JoinVertical(lipgloss.Left, filterBar, filterSep, m.list.View()),
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
	listWidth := width * 35 / 100
	detailWidth := width - listWidth
	contentHeight := height - 5
	itemsHeight := ((contentHeight - 3) / 3) * 3
	listHeight := itemsHeight + 1
	d := articlesDelegate{width: listWidth}
	m.list.SetDelegate(d)
	m.list.SetSize(listWidth, listHeight)
	m.detail.SetWidth(detailWidth)
	m.detail.SetHeight(contentHeight)
	if m.renderer != nil {
		m.renderer, _ = glamour.NewTermRenderer(
			glamour.WithStylePath("dark"),
			glamour.WithWordWrap(detailWidth-2),
		)
	}
	m.updateDetail()
}
