package tui

import (
	"fmt"
	"strings"

	"github.com/charmbracelet/lipgloss"
	"github.com/santhosh-005/goroviz/internal/group"
)

// renderListView renders the main list view showing all goroutine groups.
func renderListView(m Model) string {
	var b strings.Builder

	// Title
	title := titleStyle.Render(" 🔍 Goroviz — Goroutine Visualizer")
	b.WriteString(title)
	b.WriteString("\n")

	// Handle empty state
	if len(m.groups) == 0 {
		b.WriteString(emptyStyle.Render("No goroutine groups found.\nCheck that the input file contains valid goroutine dump data."))
		b.WriteString("\n\n")
		b.WriteString(helpStyle.Render("  q quit"))
		return b.String()
	}

	// Summary info
	totalGoroutines := 0
	for _, g := range m.groups {
		totalGoroutines += g.Count()
	}
	summary := fmt.Sprintf("  📊 %d goroutines  •  %d groups", totalGoroutines, len(m.groups))
	b.WriteString(infoStyle.Render(summary))
	b.WriteString("\n\n")

	// Column header
	header := fmt.Sprintf("  %-40s   %-6s  %s", "FUNCTION", "COUNT", "STATE")
	b.WriteString(columnHeaderStyle.Render(header))
	b.WriteString("\n")
	sep := separatorStyle.Render("  " + strings.Repeat("─", clamp(m.width-4, 40, 120)))
	b.WriteString(sep)
	b.WriteString("\n")

	// Calculate visible range for scrolling
	visibleHeight := m.height - 10
	if visibleHeight < 3 {
		visibleHeight = 3
	}

	start := m.offset
	end := start + visibleHeight
	if end > len(m.groups) {
		end = len(m.groups)
	}

	// Render group items
	for i := start; i < end; i++ {
		g := m.groups[i]
		line := formatGroupLine(g)

		if i == m.cursor {
			b.WriteString(selectedStyle.Render("▸ " + line))
		} else {
			b.WriteString(normalStyle.Render("  " + line))
		}
		b.WriteString("\n")
	}

	// Scroll indicator
	if len(m.groups) > visibleHeight {
		pct := 0
		if len(m.groups)-1 > 0 {
			pct = (m.cursor * 100) / (len(m.groups) - 1)
		}
		scrollInfo := fmt.Sprintf("  ↕ %d/%d (%d%%)", m.cursor+1, len(m.groups), pct)
		b.WriteString("\n")
		b.WriteString(infoStyle.Render(scrollInfo))
	}

	// Help bar
	b.WriteString("\n\n")
	keys := []string{"↑/↓ navigate", "enter select", "g/b top/bottom", "q quit"}
	help := helpStyle.Render("  " + strings.Join(keys, "  •  "))
	b.WriteString(help)

	return b.String()
}

// renderDetailView renders the detail view for a selected goroutine group.
func renderDetailView(m Model) string {
	if m.cursor < 0 || m.cursor >= len(m.groups) {
		return errorMsgStyle.Render("  No group selected")
	}

	g := m.groups[m.cursor]
	var b strings.Builder

	// Group header with state color
	sc := stateColor(g.State)
	headerStyle := lipgloss.NewStyle().
		Bold(true).
		Foreground(lipgloss.Color("#FFFFFF")).
		Background(sc).
		Padding(0, 2)

	header := headerStyle.Render(fmt.Sprintf(
		" %s ", g.TopFunction(),
	))
	b.WriteString(header)
	b.WriteString("\n\n")

	// Group info line
	countBadge := countStyle.Render(fmt.Sprintf("  %d goroutines", g.Count()))
	stateBadge := lipgloss.NewStyle().
		Foreground(sc).
		Bold(true).
		Render(fmt.Sprintf("[%s]", g.State))
	b.WriteString(countBadge)
	b.WriteString("  ")
	b.WriteString(stateBadge)
	b.WriteString("\n")

	// Goroutine IDs
	ids := make([]string, len(g.Goroutines))
	for i, gr := range g.Goroutines {
		ids[i] = fmt.Sprintf("%d", gr.ID)
	}
	idLine := fmt.Sprintf("  IDs: %s", strings.Join(ids, ", "))
	b.WriteString(idStyle.Render(idLine))
	b.WriteString("\n\n")

	// Separator + stack trace header
	sep := separatorStyle.Render("  " + strings.Repeat("─", clamp(m.width-4, 40, 120)))
	b.WriteString(sep)
	b.WriteString("\n")
	b.WriteString(infoStyle.Render("  Stack Trace"))
	b.WriteString("\n\n")

	// Stack frames
	if len(g.Goroutines) == 0 || len(g.Goroutines[0].Frames) == 0 {
		b.WriteString(emptyStyle.Render("  No stack frames available"))
		b.WriteString("\n")
	} else {
		frames := g.Goroutines[0].Frames

		// Calculate visible range for frame scrolling
		visibleHeight := m.height - 18
		if visibleHeight < 4 {
			visibleHeight = 4
		}
		maxVisible := visibleHeight / 2

		frameStart := m.detailOffset
		frameEnd := frameStart + maxVisible
		if frameEnd > len(frames) {
			frameEnd = len(frames)
		}

		for i := frameStart; i < frameEnd; i++ {
			frame := frames[i]
			idx := frameIndexStyle.Render(fmt.Sprintf("#%-2d", i))
			funcName := funcStyle.Render(frame.Function)
			fileLine := fileStyle.Render(frame.File) +
				lineNumStyle.Render(fmt.Sprintf(":%d", frame.Line))

			if i == m.frameCursor {
				marker := lipgloss.NewStyle().
					Foreground(colorAccent).
					Bold(true).
					Render("▸")
				b.WriteString(fmt.Sprintf("  %s %s %s\n", marker, idx, funcName))
				b.WriteString(fmt.Sprintf("          %s\n", fileLine))
			} else {
				b.WriteString(fmt.Sprintf("    %s %s\n", idx, funcName))
				b.WriteString(fmt.Sprintf("          %s\n", fileLine))
			}
		}

		// Scroll indicator for frames
		if len(frames) > maxVisible {
			pct := 0
			if len(frames)-1 > 0 {
				pct = (m.frameCursor * 100) / (len(frames) - 1)
			}
			scrollInfo := fmt.Sprintf("  ↕ frame %d/%d (%d%%)", m.frameCursor+1, len(frames), pct)
			b.WriteString("\n")
			b.WriteString(infoStyle.Render(scrollInfo))
		}
	}

	// Status message (if any)
	if m.statusMsg != "" {
		b.WriteString("\n")
		b.WriteString(statusMsgStyle.Render("  " + m.statusMsg))
	}

	// Help bar
	b.WriteString("\n\n")
	keys := []string{"↑/↓ frames", "e open editor", "esc back", "q quit"}
	help := helpStyle.Render("  " + strings.Join(keys, "  •  "))
	b.WriteString(help)

	return b.String()
}

// cleanFunctionName formats the function name to be more human-readable.
func cleanFunctionName(fn string) string {
	if idx := strings.LastIndex(fn, "/"); idx != -1 {
		fn = fn[idx+1:]
	}
	if strings.HasPrefix(fn, "main.") {
		fn = strings.TrimPrefix(fn, "main.")
	}
	fn = strings.ReplaceAll(fn, "(*", "")
	fn = strings.ReplaceAll(fn, ")", "")
	return fn
}

// formatGroupLine formats a single group line for the list view.
func formatGroupLine(g group.Group) string {
	funcName := cleanFunctionName(g.TopFunction())
	if len(funcName) > 40 {
		funcName = funcName[:37] + "..."
	}

	countStr := fmt.Sprintf("(%d)", g.Count())
	stateStr := fmt.Sprintf("[%s]", g.State)

	styledCount := countStyle.Render(fmt.Sprintf("%-6s", countStr))
	sc := stateColor(g.State)
	styledState := lipgloss.NewStyle().Foreground(sc).Render(stateStr)

	return fmt.Sprintf("%-40s   %s  %s", funcName, styledCount, styledState)
}

// clamp constrains a value between min and max.
func clamp(val, min, max int) int {
	if val < min {
		return min
	}
	if val > max {
		return max
	}
	return val
}
