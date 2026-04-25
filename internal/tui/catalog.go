package tui

import (
	"fmt"
	"io"
	"strings"

	"github.com/charmbracelet/bubbles/key"
	"github.com/charmbracelet/bubbles/list"
	"github.com/charmbracelet/bubbles/textinput"
	"github.com/charmbracelet/bubbles/viewport"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	"github.com/ali5ter/clu/internal/api"
)

// catalogItem wraps api.CatalogItem to implement list.Item.
type catalogItem struct{ api.CatalogItem }

func (i catalogItem) Title() string       { return i.Name }
func (i catalogItem) Description() string { return "" }
func (i catalogItem) FilterValue() string {
	return i.Name + " " + i.Category + " " + strings.Join(i.Tags, " ")
}

// catalogDelegate is a custom list delegate for rich item rendering.
type catalogDelegate struct{ width int }

func (d catalogDelegate) Height() int                             { return 2 }
func (d catalogDelegate) Spacing() int                            { return 0 }
func (d catalogDelegate) Update(_ tea.Msg, _ *list.Model) tea.Cmd { return nil }

func (d catalogDelegate) Render(w io.Writer, m list.Model, index int, item list.Item) {
	ci, ok := item.(catalogItem)
	if !ok {
		return
	}
	selected := index == m.Index()

	score := fmt.Sprintf("%.2f", ci.Score)
	scoreStr := scoreStyle(ci.Score).Render(score)
	// nameWidth = total - indicator(1) - space(1) - score(4) - space(1)
	nameWidth := d.width - 7
	if nameWidth < 1 {
		nameWidth = 1
	}

	var indicator, name string
	if selected {
		indicator = styleCopper.Render("▶")
		name = styleSelected.Width(nameWidth).Render(ci.Name)
	} else {
		indicator = " "
		name = styleNormal.Width(nameWidth).Render(ci.Name)
	}

	line1 := fmt.Sprintf("%s %s %s", indicator, name, scoreStr)

	meta := ci.Category + " · " + ci.SourceType + " · " + ci.Maturity
	if ci.IsNew {
		meta = "✦ " + meta
	}
	metaWidth := d.width - 2
	var line2 string
	if selected {
		line2 = "  " + styleDetailMuted.Width(metaWidth).Render(meta)
	} else {
		line2 = "  " + styleDim.Width(metaWidth).Render(meta)
	}

	fmt.Fprintln(w, line1)
	fmt.Fprint(w, line2)
}

func scoreStyle(score float64) lipgloss.Style {
	switch {
	case score >= 9.0:
		return lipgloss.NewStyle().Foreground(colorGreen).Bold(true)
	case score >= 8.5:
		return lipgloss.NewStyle().Foreground(colorSteel)
	default:
		return lipgloss.NewStyle().Foreground(colorMuted)
	}
}

// catalogModel is the catalog tab view.
type catalogModel struct {
	list      list.Model
	filter    textinput.Model
	detail    viewport.Model
	items     []api.CatalogItem
	width     int
	height    int
	filtering bool
}

func newCatalogModel(items []api.CatalogItem, width, height int) catalogModel {
	listItems := make([]list.Item, len(items))
	for i, it := range items {
		listItems[i] = catalogItem{it}
	}

	listWidth := width / 2
	contentHeight := height - 5 // header(1) sep(1) filter(1) sep(1) status(1)

	delegate := catalogDelegate{width: listWidth}
	l := list.New(listItems, delegate, listWidth, contentHeight)
	l.SetShowTitle(false)
	l.SetShowHelp(false)
	l.SetShowStatusBar(false)
	l.SetShowPagination(false)
	l.SetFilteringEnabled(false)

	fi := textinput.New()
	fi.Placeholder = "filter…"
	fi.PromptStyle = styleFilterPrompt
	fi.TextStyle = styleFilterText
	fi.Prompt = "> "

	vp := viewport.New(width-listWidth, contentHeight)

	m := catalogModel{
		list:   l,
		filter: fi,
		detail: vp,
		items:  items,
		width:  width,
		height: height,
	}
	m.updateDetail()
	return m
}

func (m *catalogModel) selectedItem() *api.CatalogItem {
	if i, ok := m.list.SelectedItem().(catalogItem); ok {
		return &i.CatalogItem
	}
	return nil
}

func (m *catalogModel) updateDetail() {
	it := m.selectedItem()
	if it == nil {
		m.detail.SetContent("")
		return
	}

	var sb strings.Builder
	label := func(k, v string) {
		sb.WriteString(styleDetailLabel.Render(k+":") + "  " + styleDetailValue.Render(v) + "\n")
	}

	sb.WriteString(styleSelected.Render(it.Name) + "\n")
	sb.WriteString(styleDim.Render(strings.Repeat("─", 40)) + "\n\n")
	sb.WriteString(styleDetailValue.Render(it.Summary) + "\n\n")
	label("Score", scoreStyle(it.Score).Render(fmt.Sprintf("%.2f", it.Score)))
	label("Confidence", fmt.Sprintf("%.2f", it.Confidence))
	label("Maturity", it.Maturity)
	label("Maintenance", it.Maintenance)
	label("Category", it.Category)
	label("Source", it.SourceType)
	if it.AuthorOrOrg != "" {
		label("Author/Org", it.AuthorOrOrg)
	}
	if it.License != "" {
		label("License", it.License)
	}
	if len(it.Tags) > 0 {
		label("Tags", strings.Join(it.Tags, ", "))
	}
	sb.WriteString("\n" + styleDetailMuted.Render(it.SourceURL) + "\n")

	m.detail.SetContent(sb.String())
}

func (m catalogModel) Update(msg tea.Msg) (catalogModel, tea.Cmd) {
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
		}
	}

	var cmd tea.Cmd
	m.detail, cmd = m.detail.Update(msg)
	cmds = append(cmds, cmd)

	return m, tea.Batch(cmds...)
}

func (m *catalogModel) applyFilter() {
	q := strings.ToLower(m.filter.Value())
	var filtered []list.Item
	for _, it := range m.items {
		ci := catalogItem{it}
		if q == "" || strings.Contains(strings.ToLower(ci.FilterValue()+" "+it.Summary), q) {
			filtered = append(filtered, ci)
		}
	}
	m.list.SetItems(filtered)
	m.updateDetail()
}

func (m catalogModel) View() string {
	listWidth := m.width / 2
	detailWidth := m.width - listWidth

	var filterBar string
	if m.filtering {
		filterBar = styleFilterPrompt.Render("> ") + m.filter.View()
	} else if m.filter.Value() != "" {
		filterBar = styleFilterPrompt.Render("> ") + styleFilterText.Render(m.filter.Value())
	} else {
		filterBar = styleDim.Render("> filter…")
	}

	listPane := stylePanelBorder.Width(listWidth).Render(
		lipgloss.JoinVertical(lipgloss.Left, filterBar, m.list.View()),
	)
	detailPane := lipgloss.NewStyle().Width(detailWidth).Padding(0, 1).Render(m.detail.View())

	return lipgloss.JoinHorizontal(lipgloss.Top, listPane, detailPane)
}

func (m catalogModel) SelectedURL() string {
	if it := m.selectedItem(); it != nil {
		return it.SourceURL
	}
	return ""
}

func (m catalogModel) SelectedJSON() *api.CatalogItem { return m.selectedItem() }

func (m *catalogModel) SetSize(width, height int) {
	m.width = width
	m.height = height
	listWidth := width / 2
	contentHeight := height - 5
	d := catalogDelegate{width: listWidth}
	m.list.SetDelegate(d)
	m.list.SetSize(listWidth, contentHeight)
	m.detail.Width = width - listWidth
	m.detail.Height = contentHeight
	m.updateDetail()
}
