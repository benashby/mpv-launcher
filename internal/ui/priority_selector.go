package ui

import (
	"fmt"
	"strings"

	tea "github.com/charmbracelet/bubbletea"
)

// PrioritySelector displays and manages an interactive priority list
type PrioritySelector struct {
	items    []string
	order    []int // indices into items representing current order
	canceled bool
}

// NewPrioritySelector creates a new priority selector with the given item labels
func NewPrioritySelector(items []string) *PrioritySelector {
	order := make([]int, len(items))
	for i := range items {
		order[i] = i
	}
	return &PrioritySelector{items: items, order: order}
}

// Run displays the UI and returns the final priority order (indices into original items).
// Returns nil if the user cancels (q or Esc).
func (s *PrioritySelector) Run() ([]int, error) {
	p := tea.NewProgram(newPriorityModel(s.items, s.order))
	result, err := p.Run()
	if err != nil {
		return nil, err
	}
	m := result.(priorityModel)
	if m.canceled {
		s.canceled = true
		return nil, nil
	}
	s.order = m.order
	return s.order, nil
}

// Canceled returns true if the user canceled the selection
func (s *PrioritySelector) Canceled() bool {
	return s.canceled
}

// GetLabelsInOrder returns the item labels in priority order
func (s *PrioritySelector) GetLabelsInOrder() []string {
	result := make([]string, len(s.order))
	for i, idx := range s.order {
		result[i] = s.items[idx]
	}
	return result
}

// --- Bubble Tea model ---

type priorityModel struct {
	items    []string
	order    []int
	cursor   int
	canceled bool
	done     bool
}

func newPriorityModel(items []string, order []int) priorityModel {
	orderCopy := make([]int, len(order))
	copy(orderCopy, order)
	return priorityModel{items: items, order: orderCopy}
}

func (m priorityModel) Init() tea.Cmd { return nil }

func (m priorityModel) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.KeyMsg:
		switch msg.String() {
		case "enter":
			m.done = true
			return m, tea.Quit
		case "q", "esc":
			m.canceled = true
			m.done = true
			return m, tea.Quit
		case "j", "down":
			if m.cursor < len(m.order)-1 {
				m.cursor++
			}
		case "k", "up":
			if m.cursor > 0 {
				m.cursor--
			}
		case "J":
			if m.cursor < len(m.order)-1 {
				m.order[m.cursor], m.order[m.cursor+1] = m.order[m.cursor+1], m.order[m.cursor]
				m.cursor++
			}
		case "K":
			if m.cursor > 0 {
				m.order[m.cursor], m.order[m.cursor-1] = m.order[m.cursor-1], m.order[m.cursor]
				m.cursor--
			}
		}
	}
	return m, nil
}

func (m priorityModel) View() string {
	var b strings.Builder
	b.WriteString("Set playback priority (j/k move, J/K reorder, Enter confirm, q quit):\n\n")
	for i, idx := range m.order {
		if i == m.cursor {
			fmt.Fprintf(&b, "> %d. %s <\n", i+1, m.items[idx])
		} else {
			fmt.Fprintf(&b, "  %d. %s\n", i+1, m.items[idx])
		}
	}
	b.WriteString("\n")
	return b.String()
}

// SimpleSelect provides a simple numbered selection without reordering.
// Returns the selected index (0-based) or -1 if canceled.
// This is now an alias for CursorSelect.
func SimpleSelect(prompt string, items []string) (int, error) {
	return CursorSelect(prompt, items)
}
