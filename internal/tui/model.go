package tui

import (
	"fmt"
	"os/exec"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/santhosh-005/goroviz/internal/editor"
	"github.com/santhosh-005/goroviz/internal/group"
)

// viewState represents which view is currently active.
type viewState int

const (
	listView   viewState = iota // main list of goroutine groups
	detailView                  // detail view of a single group
)

// Model is the Bubble Tea model for the Goroviz TUI.
type Model struct {
	groups []group.Group // grouped goroutines

	// View state
	view viewState

	// List view state
	cursor int // currently highlighted group index
	offset int // scroll offset for the list

	// Detail view state
	frameCursor  int // currently highlighted frame index
	detailOffset int // scroll offset for the frame list

	// Terminal dimensions
	width  int
	height int

	// Status message (e.g. "Opened in vim")
	statusMsg string

	// Quitting flag
	quitting bool
}

// NewModel creates a new TUI model with the given goroutine groups.
func NewModel(groups []group.Group) Model {
	return Model{
		groups: groups,
		view:   listView,
		width:  80,
		height: 24,
	}
}

// Init implements tea.Model.
func (m Model) Init() tea.Cmd {
	return nil
}

// editorFinishedMsg is sent when an editor process finishes.
type editorFinishedMsg struct{ err error }

// Update implements tea.Model. Handles keyboard input and window resize.
func (m Model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.WindowSizeMsg:
		m.width = msg.Width
		m.height = msg.Height
		return m, nil

	case editorFinishedMsg:
		if msg.err != nil {
			m.statusMsg = fmt.Sprintf("Editor error: %v", msg.err)
		} else {
			m.statusMsg = "Editor closed"
		}
		return m, nil

	case tea.KeyMsg:
		// Clear status message on any key press
		m.statusMsg = ""

		// Global keys
		switch msg.String() {
		case "ctrl+c", "q":
			m.quitting = true
			return m, tea.Quit
		}

		// View-specific keys
		switch m.view {
		case listView:
			return m.updateListView(msg)
		case detailView:
			return m.updateDetailView(msg)
		}
	}

	return m, nil
}

// updateListView handles key events in the list view.
func (m Model) updateListView(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
	switch msg.String() {
	case "up", "k":
		if m.cursor > 0 {
			m.cursor--
			// Scroll up if cursor goes above visible area
			if m.cursor < m.offset {
				m.offset = m.cursor
			}
		}

	case "down", "j":
		if m.cursor < len(m.groups)-1 {
			m.cursor++
			// Scroll down if cursor goes below visible area
			visibleHeight := m.height - 10
			if visibleHeight < 3 {
				visibleHeight = 3
			}
			if m.cursor >= m.offset+visibleHeight {
				m.offset = m.cursor - visibleHeight + 1
			}
		}

	case "enter":
		if len(m.groups) > 0 {
			m.view = detailView
			m.frameCursor = 0
			m.detailOffset = 0
		}

	case "home", "g":
		m.cursor = 0
		m.offset = 0

	case "end", "b":
		m.cursor = len(m.groups) - 1
		visibleHeight := m.height - 10
		if visibleHeight < 3 {
			visibleHeight = 3
		}
		if m.cursor >= visibleHeight {
			m.offset = m.cursor - visibleHeight + 1
		}
	}

	return m, nil
}

// updateDetailView handles key events in the detail view.
func (m Model) updateDetailView(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
	g := m.groups[m.cursor]
	frameCount := 0
	if len(g.Goroutines) > 0 {
		frameCount = len(g.Goroutines[0].Frames)
	}

	switch msg.String() {
	case "up", "k":
		if m.frameCursor > 0 {
			m.frameCursor--
			if m.frameCursor < m.detailOffset {
				m.detailOffset = m.frameCursor
			}
		}

	case "down", "j":
		if m.frameCursor < frameCount-1 {
			m.frameCursor++
			visibleHeight := m.height - 16
			if visibleHeight < 4 {
				visibleHeight = 4
			}
			maxVisible := visibleHeight / 2
			if m.frameCursor >= m.detailOffset+maxVisible {
				m.detailOffset = m.frameCursor - maxVisible + 1
			}
		}

	case "esc", "backspace":
		m.view = listView
		m.frameCursor = 0
		m.detailOffset = 0

	case "e", "o":
		// Open the selected frame's file in the editor
		if len(g.Goroutines) > 0 && m.frameCursor < len(g.Goroutines[0].Frames) {
			frame := g.Goroutines[0].Frames[m.frameCursor]
			editorCmd, args := editor.DetectEditor(frame.File, frame.Line)
			if editorCmd == "" {
				m.statusMsg = "No editor found. Set $EDITOR env var."
				return m, nil
			}
			c := exec.Command(editorCmd, args...)
			return m, tea.ExecProcess(c, func(err error) tea.Msg {
				return editorFinishedMsg{err}
			})
		}
	}

	return m, nil
}

// View implements tea.Model. Renders the current view.
func (m Model) View() string {
	if m.quitting {
		return ""
	}

	switch m.view {
	case detailView:
		return renderDetailView(m)
	default:
		return renderListView(m)
	}
}

// Run starts the TUI application with the given groups.
func Run(groups []group.Group) error {
	model := NewModel(groups)
	p := tea.NewProgram(model, tea.WithAltScreen())
	_, err := p.Run()
	return err
}
