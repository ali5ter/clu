package tui

import (
	"strings"

	"charm.land/lipgloss/v2"
)

// dialogContentWidth returns a responsive dialog width — roughly half the
// terminal width, clamped so the dialog stays readable on both huge and
// small terminals.
func dialogContentWidth(termWidth int) int {
	w := termWidth / 2
	if w < 44 {
		w = 44
	}
	if w > 64 {
		w = 64
	}
	return w
}

// renderDialog builds a rounded-border modal box from a title, body, and
// hint line. Every row is padded to the widest of the three so the panel
// background (not the terminal's default background) fills every blank
// cell — lipgloss.JoinVertical pads shorter rows with unstyled spaces,
// which would otherwise let the terminal background show through.
func renderDialog(title, body, hint string) string {
	width := lipgloss.Width(body)
	if w := lipgloss.Width(title); w > width {
		width = w
	}
	if w := lipgloss.Width(hint); w > width {
		width = w
	}

	pad := func(s string) string {
		lines := strings.Split(s, "\n")
		for i, line := range lines {
			gap := width - lipgloss.Width(line)
			if gap > 0 {
				line += strings.Repeat(" ", gap)
			}
			lines[i] = styleDialogFill.Render(line)
		}
		return strings.Join(lines, "\n")
	}

	content := lipgloss.JoinVertical(lipgloss.Left,
		pad(styleDialogTitle.Render(title)),
		pad(""),
		pad(body),
		pad(""),
		pad(styleDialogHint.Render(hint)),
	)
	return styleDialogBox.Render(content)
}

// overlayDialog draws dialog centered on top of the already-rendered bg
// using Lip Gloss v2's compositor, so the underlying view stays visible
// around the modal instead of being replaced.
func overlayDialog(bg, dialog string, width, height int) string {
	dw, dh := lipgloss.Size(dialog)
	x := (width - dw) / 2
	y := (height - dh) / 2
	if x < 0 {
		x = 0
	}
	if y < 0 {
		y = 0
	}
	return lipgloss.NewCompositor(
		lipgloss.NewLayer(bg),
		lipgloss.NewLayer(dialog).X(x).Y(y).Z(1),
	).Render()
}
