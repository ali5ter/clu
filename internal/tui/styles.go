package tui

import (
	"fmt"
	"image/color"
	"strconv"
	"strings"

	"charm.land/lipgloss/v2"
)

// truncate clips s to max runes, appending "…" if trimmed.
func truncate(s string, max int) string {
	r := []rune(s)
	if len(r) <= max {
		return s
	}
	return string(r[:max-1]) + "…"
}

// lerpColor linearly interpolates between two "#rrggbb" hex colours at t ∈ [0,1].
func lerpColor(a, b string, t float64) color.Color {
	if t < 0 {
		t = 0
	} else if t > 1 {
		t = 1
	}
	ar, ag, ab := hexRGB(a)
	br, bg, bb := hexRGB(b)
	lerp := func(x, y int) int { return x + int(float64(y-x)*t) }
	return lipgloss.Color(fmt.Sprintf("#%02x%02x%02x", lerp(ar, br), lerp(ag, bg), lerp(ab, bb)))
}

func hexRGB(hex string) (r, g, b int) {
	hex = strings.TrimPrefix(hex, "#")
	rv, _ := strconv.ParseInt(hex[0:2], 16, 32)
	gv, _ := strconv.ParseInt(hex[2:4], 16, 32)
	bv, _ := strconv.ParseInt(hex[4:6], 16, 32)
	return int(rv), int(gv), int(bv)
}

// subcellBlocks give eighth-cell fill precision for progress bars, from empty to full.
var subcellBlocks = []rune{' ', '▏', '▎', '▍', '▌', '▋', '▊', '▉', '█'}

// makeBar renders value/max as a unicode block bar of the given width, at eighth-cell precision.
func makeBar(value, maxVal float64, width int, style lipgloss.Style) string {
	if maxVal <= 0 || width <= 0 {
		return ""
	}
	frac := value / maxVal
	if frac < 0 {
		frac = 0
	} else if frac > 1 {
		frac = 1
	}
	eighths := int(frac*float64(width)*8 + 0.5)
	full := eighths / 8
	rem := eighths % 8

	var sb strings.Builder
	sb.WriteString(style.Render(strings.Repeat("█", full)))
	if full < width && rem > 0 {
		sb.WriteString(style.Render(string(subcellBlocks[rem])))
		full++
	}
	if full < width {
		sb.WriteString(strings.Repeat(" ", width-full))
	}
	return sb.String()
}

// logoLines is the block-letter "CLU" wordmark shown on the loading screen.
var logoLines = []string{
	" _____  _     _   _ ",
	"/  __ \\| |   | | | |",
	"| /  \\/| |   | | | |",
	"| |    | |   | | | |",
	"| \\__/\\| |___| |_| |",
	" \\____/\\_____/\\___/ ",
}

// renderLogo renders the full block-letter wordmark with a top-to-bottom
// green→copper gradient, one Foreground style per line.
func renderLogo() string {
	last := float64(len(logoLines) - 1)
	lines := make([]string, len(logoLines))
	for i, line := range logoLines {
		t := float64(i) / last
		style := lipgloss.NewStyle().Bold(true).Foreground(lerpColor(hexGreen, hexCopper, t))
		lines[i] = style.Render(line)
	}
	return strings.Join(lines, "\n")
}

// renderHeaderLogo renders the compact "clu" wordmark used in the header,
// applying the same green→copper gradient per character so it reads as the
// same identity as the loading screen's block-letter logo.
func renderHeaderLogo() string {
	runes := []rune("clu")
	last := float64(len(runes) - 1)
	var sb strings.Builder
	for i, r := range runes {
		t := float64(i) / last
		style := lipgloss.NewStyle().
			Background(colorPanel).
			Foreground(lerpColor(hexGreen, hexCopper, t)).
			Bold(true)
		sb.WriteString(style.Render(string(r)))
	}
	return sb.String()
}

// Site palette — mirrors commandlineuser.com CSS variables.
const (
	hexText   = "#e9eff3"
	hexMuted  = "#a8b6c0"
	hexPanel  = "#161d24"
	hexLine   = "#2b3742"
	hexGreen  = "#66b08a"
	hexCopper = "#c9895e"
	hexSteel  = "#7f93a6"
)

var (
	colorText   = lipgloss.Color(hexText)
	colorMuted  = lipgloss.Color(hexMuted)
	colorPanel  = lipgloss.Color(hexPanel)
	colorLine   = lipgloss.Color(hexLine)
	colorGreen  = lipgloss.Color(hexGreen)
	colorCopper = lipgloss.Color(hexCopper)
	colorSteel  = lipgloss.Color(hexSteel)
)

var (
	styleHeader = lipgloss.NewStyle().
			Background(colorPanel).
			Foreground(colorSteel).
			Padding(0, 1).
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

	// styleListSelected marks the row text of a selected list item — the
	// accent bar (styleSelectedBar) carries the primary selection cue, so
	// this stays unbold to avoid double emphasis.
	styleListSelected = lipgloss.NewStyle().
				Foreground(colorGreen)

	// styleSelectedBar renders the left accent bar glyph in front of a
	// selected list row (spans both rendered lines of the row).
	styleSelectedBar = lipgloss.NewStyle().
				Foreground(colorGreen)

	styleNormal = lipgloss.NewStyle().
			Foreground(colorMuted)

	styleTabActive = lipgloss.NewStyle().
			Background(colorGreen).
			Foreground(colorPanel).
			Bold(true).
			Padding(0, 1)

	styleTabInactive = lipgloss.NewStyle().
				Background(colorPanel).
				Foreground(colorMuted).
				Padding(0, 1)

	styleDialogBox = lipgloss.NewStyle().
			Border(lipgloss.RoundedBorder()).
			BorderForeground(colorGreen).
			Background(colorPanel).
			Padding(1, 2)

	styleDialogTitle = lipgloss.NewStyle().
				Background(colorPanel).
				Foreground(colorGreen).
				Bold(true)

	styleDialogHint = lipgloss.NewStyle().
			Background(colorPanel).
			Foreground(colorMuted)

	styleDialogFill = lipgloss.NewStyle().
			Background(colorPanel)

	styleSteel = lipgloss.NewStyle().
			Foreground(colorSteel).
			Bold(true)

	styleDim = lipgloss.NewStyle().
			Foreground(colorLine)
)
