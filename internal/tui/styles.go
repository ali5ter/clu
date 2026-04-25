package tui

import "github.com/charmbracelet/lipgloss"

// Site palette — mirrors commandlineuser.com CSS variables.
var (
	colorText   = lipgloss.Color("#e9eff3")
	colorMuted  = lipgloss.Color("#a8b6c0")
	colorPanel  = lipgloss.Color("#161d24")
	colorLine   = lipgloss.Color("#2b3742")
	colorGreen  = lipgloss.Color("#66b08a")
	colorCopper = lipgloss.Color("#c9895e")
	colorSteel  = lipgloss.Color("#7f93a6")
)

var (
	styleHeader = lipgloss.NewStyle().
			Background(colorPanel).
			Foreground(colorSteel).
			Padding(0, 1).
			Bold(true)

	styleHeaderAccent = lipgloss.NewStyle().
				Background(colorPanel).
				Foreground(colorGreen).
				Bold(true)

	styleStatusBar = lipgloss.NewStyle().
			Background(colorPanel).
			Foreground(colorMuted).
			Padding(0, 1)

	styleStatusKey = lipgloss.NewStyle().
			Background(colorPanel).
			Foreground(colorCopper)

	stylePanelBorder = lipgloss.NewStyle().
				Border(lipgloss.NormalBorder(), false, true, false, false).
				BorderForeground(colorLine)

	styleDetailLabel = lipgloss.NewStyle().
				Foreground(colorSteel).
				Bold(true)

	styleDetailValue = lipgloss.NewStyle().
				Foreground(colorText)

	styleDetailMuted = lipgloss.NewStyle().
				Foreground(colorMuted)

	styleFilterPrompt = lipgloss.NewStyle().
				Foreground(colorCopper)

	styleFilterText = lipgloss.NewStyle().
			Foreground(colorText)

	styleSelected = lipgloss.NewStyle().
			Foreground(colorGreen).
			Bold(true)

	styleNormal = lipgloss.NewStyle().
			Foreground(colorMuted)

	styleTabActive = lipgloss.NewStyle().
			Foreground(colorGreen).
			Bold(true).
			Padding(0, 1)

	styleTabInactive = lipgloss.NewStyle().
				Foreground(colorMuted).
				Padding(0, 1)

	styleSteel = lipgloss.NewStyle().
			Foreground(colorSteel).
			Bold(true)

	styleCopper = lipgloss.NewStyle().
			Foreground(colorCopper)

	styleDim = lipgloss.NewStyle().
			Foreground(colorLine)
)
