package ui

import "github.com/charmbracelet/lipgloss"

var (
	colorAccent   = lipgloss.Color("#7DD3FC") // sky
	colorGood     = lipgloss.Color("#86EFAC") // green
	colorWarn     = lipgloss.Color("#FDE68A") // amber
	colorBad      = lipgloss.Color("#FCA5A5") // red
	colorMuted    = lipgloss.Color("#6B7280") // gray
	colorFg       = lipgloss.Color("#E5E7EB")
	colorBg       = lipgloss.Color("#111827")
	colorSelected = lipgloss.Color("#1E293B")

	titleStyle = lipgloss.NewStyle().
			Bold(true).
			Foreground(colorAccent).
			Padding(0, 1)

	subtitleStyle = lipgloss.NewStyle().
			Foreground(colorMuted)

	helpStyle = lipgloss.NewStyle().
			Foreground(colorMuted).
			Padding(1, 1, 0, 1)

	borderStyle = lipgloss.NewStyle().
			BorderStyle(lipgloss.RoundedBorder()).
			BorderForeground(colorMuted).
			Padding(1, 2)

	selectedItemStyle = lipgloss.NewStyle().
				Background(colorSelected).
				Foreground(colorAccent).
				Bold(true).
				Padding(0, 1)

	itemStyle = lipgloss.NewStyle().
			Foreground(colorFg).
			Padding(0, 1)

	labelStyle = lipgloss.NewStyle().
			Foreground(colorMuted).
			Width(16)

	valueStyle = lipgloss.NewStyle().
			Foreground(colorFg)

	focusedValueStyle = lipgloss.NewStyle().
				Foreground(colorAccent).
				Bold(true)

	errorStyle = lipgloss.NewStyle().
			Foreground(colorBad).
			Bold(true)

	goodStyle = lipgloss.NewStyle().Foreground(colorGood).Bold(true)
	warnStyle = lipgloss.NewStyle().Foreground(colorWarn).Bold(true)
	badStyle  = lipgloss.NewStyle().Foreground(colorBad).Bold(true)
)
