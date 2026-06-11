// Package tui provides the interactive terminal UI for Goroviz
// using the Bubble Tea framework.
package tui

import (
	"github.com/charmbracelet/lipgloss"
)

// Color palette — curated for dark terminal backgrounds
var (
	colorPrimary    = lipgloss.Color("#8B5CF6") // violet-500
	colorPrimaryDim = lipgloss.Color("#6D28D9") // violet-700
	colorSecondary  = lipgloss.Color("#06B6D4") // cyan-500
	colorAccent     = lipgloss.Color("#F59E0B") // amber-500
	colorMuted      = lipgloss.Color("#6B7280") // gray-500
	colorSuccess    = lipgloss.Color("#34D399") // emerald-400
	colorDanger     = lipgloss.Color("#F87171") // red-400
	colorWarning    = lipgloss.Color("#FBBF24") // amber-400
	colorText       = lipgloss.Color("#F3F4F6") // gray-100
	colorDimText    = lipgloss.Color("#9CA3AF") // gray-400
	colorSubtle     = lipgloss.Color("#4B5563") // gray-600
)

// Styles used throughout the TUI
var (
	// Title bar — bold violet banner
	titleStyle = lipgloss.NewStyle().
			Bold(true).
			Foreground(lipgloss.Color("#FFFFFF")).
			Background(colorPrimary).
			Padding(0, 2).
			MarginBottom(1)

	// Subtitle / info bar
	infoStyle = lipgloss.NewStyle().
			Foreground(colorDimText).
			Italic(true)

	// Selected item in list — highlighted row
	selectedStyle = lipgloss.NewStyle().
			Bold(true).
			Foreground(lipgloss.Color("#FFFFFF")).
			Background(colorPrimaryDim).
			Padding(0, 1)

	// Normal item in list
	normalStyle = lipgloss.NewStyle().
			Foreground(colorText).
			Padding(0, 1)

	// Count badge — stands out with amber
	countStyle = lipgloss.NewStyle().
			Bold(true).
			Foreground(colorAccent)

	// State badge — cool cyan for status
	stateStyle = lipgloss.NewStyle().
			Foreground(colorSecondary).
			Italic(true)

	// Function name in stack trace — green for code
	funcStyle = lipgloss.NewStyle().
			Foreground(colorSuccess).
			Bold(true)

	// File path in stack trace — dimmed
	fileStyle = lipgloss.NewStyle().
			Foreground(colorDimText)

	// Line number — amber highlight
	lineNumStyle = lipgloss.NewStyle().
			Foreground(colorAccent).
			Bold(true)

	// Group header in detail view — cyan banner
	groupHeaderStyle = lipgloss.NewStyle().
				Bold(true).
				Foreground(lipgloss.Color("#FFFFFF")).
				Background(colorSecondary).
				Padding(0, 2)

	// Goroutine ID list
	idStyle = lipgloss.NewStyle().
		Foreground(colorMuted)

	// Help bar at the bottom
	helpStyle = lipgloss.NewStyle().
			Foreground(colorSubtle)

	// Separator line
	separatorStyle = lipgloss.NewStyle().
			Foreground(colorSubtle)

	// Status message — amber for visibility
	statusMsgStyle = lipgloss.NewStyle().
			Foreground(colorWarning).
			Bold(true)

	// Error message
	errorMsgStyle = lipgloss.NewStyle().
			Foreground(colorDanger).
			Bold(true)

	// Empty state
	emptyStyle = lipgloss.NewStyle().
			Foreground(colorDimText).
			Italic(true).
			Padding(2, 4)

	// Column header
	columnHeaderStyle = lipgloss.NewStyle().
				Foreground(colorDimText).
				Bold(true).
				Underline(true)

	// Frame index number
	frameIndexStyle = lipgloss.NewStyle().
			Foreground(colorSubtle)
)

// stateColor returns a color based on the goroutine state.
func stateColor(state string) lipgloss.Color {
	switch state {
	case "running":
		return lipgloss.Color("#34D399") // green
	case "IO wait":
		return lipgloss.Color("#60A5FA") // blue
	case "chan receive", "chan send":
		return lipgloss.Color("#F59E0B") // amber
	case "select":
		return lipgloss.Color("#A78BFA") // violet
	case "semacquire":
		return lipgloss.Color("#F87171") // red
	case "sleep":
		return lipgloss.Color("#9CA3AF") // gray
	default:
		return lipgloss.Color("#06B6D4") // cyan
	}
}
