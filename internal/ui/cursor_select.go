package ui

import (
	"fmt"
	"strings"

	tea "github.com/charmbracelet/bubbletea"
)

type cursorModel struct {
	prompt string
	items  []string
	cursor int
	choice int // -1 = canceled, >=0 = selected index
	done   bool
}

func newCursorModel(prompt string, items []string) cursorModel {
	return cursorModel{prompt: prompt, items: items, cursor: 0, choice: -1}
}

func (m cursorModel) Init() tea.Cmd { return nil }

func (m cursorModel) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.KeyMsg:
		switch msg.String() {
		case "enter":
			m.choice = m.cursor
			m.done = true
			return m, tea.Quit
		case "q", "esc":
			m.choice = -1
			m.done = true
			return m, tea.Quit
		case "j", "down":
			if m.cursor < len(m.items)-1 {
				m.cursor++
			}
		case "k", "up":
			if m.cursor > 0 {
				m.cursor--
			}
		default:
			// 1-9 jump and confirm
			if len(msg.String()) == 1 {
				ch := msg.String()[0]
				if ch >= '1' && ch <= '9' {
					idx := int(ch - '1')
					if idx < len(m.items) {
						m.choice = idx
						m.done = true
						return m, tea.Quit
					}
				}
			}
		}
	}
	return m, nil
}

func (m cursorModel) View() string {
	var b strings.Builder
	b.WriteString(m.prompt)
	b.WriteString("\n\n")
	for i, item := range m.items {
		if i == m.cursor {
			fmt.Fprintf(&b, "> %d. %s\n", i+1, item)
		} else {
			fmt.Fprintf(&b, "  %d. %s\n", i+1, item)
		}
	}
	b.WriteString("\n")
	return b.String()
}

// CursorSelect presents a cursor-navigable list and returns the selected index (0-based).
// Returns -1 if the user cancels (q or Esc).
// Controls: j/↓ move down, k/↑ move up, Enter confirm, 1-9 jump and confirm, q/Esc cancel.
func CursorSelect(prompt string, items []string) (int, error) {
	if len(items) == 0 {
		return -1, fmt.Errorf("no items to select")
	}

	p := tea.NewProgram(newCursorModel(prompt, items))
	result, err := p.Run()
	if err != nil {
		return -1, err
	}
	return result.(cursorModel).choice, nil
}
