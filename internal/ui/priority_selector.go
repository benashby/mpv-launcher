package ui

import (
	"fmt"
	"strings"

	"github.com/eiannone/keyboard"
)

// PrioritySelector displays and manages an interactive priority list
type PrioritySelector struct {
	items    []string
	order    []int // indices into items representing current order
	cursor   int
	canceled bool
}

// NewPrioritySelector creates a new priority selector with the given item labels
func NewPrioritySelector(items []string) *PrioritySelector {
	order := make([]int, len(items))
	for i := range items {
		order[i] = i
	}
	return &PrioritySelector{
		items:  items,
		order:  order,
		cursor: 0,
	}
}

// Run displays the UI and returns the final priority order (indices into original items)
// Returns nil if the user cancels (q)
func (s *PrioritySelector) Run() ([]int, error) {
	if err := keyboard.Open(); err != nil {
		return nil, fmt.Errorf("failed to open keyboard: %w", err)
	}
	defer keyboard.Close()

	s.render()

	for {
		char, key, err := keyboard.GetKey()
		if err != nil {
			return nil, fmt.Errorf("failed to read key: %w", err)
		}

		switch {
		case key == keyboard.KeyEnter:
			// Clear the display and return
			s.clearDisplay()
			return s.order, nil

		case char == 'q' || key == keyboard.KeyEsc:
			s.clearDisplay()
			s.canceled = true
			return nil, nil

		case char == 'j' || key == keyboard.KeyArrowDown:
			s.moveCursorDown()

		case char == 'k' || key == keyboard.KeyArrowUp:
			s.moveCursorUp()

		case char == 'J':
			s.demoteItem()

		case char == 'K':
			s.promoteItem()
		}

		s.render()
	}
}

// Canceled returns true if the user canceled the selection
func (s *PrioritySelector) Canceled() bool {
	return s.canceled
}

func (s *PrioritySelector) moveCursorDown() {
	if s.cursor < len(s.order)-1 {
		s.cursor++
	}
}

func (s *PrioritySelector) moveCursorUp() {
	if s.cursor > 0 {
		s.cursor--
	}
}

func (s *PrioritySelector) demoteItem() {
	if s.cursor < len(s.order)-1 {
		// Swap current with next
		s.order[s.cursor], s.order[s.cursor+1] = s.order[s.cursor+1], s.order[s.cursor]
		s.cursor++
	}
}

func (s *PrioritySelector) promoteItem() {
	if s.cursor > 0 {
		// Swap current with previous
		s.order[s.cursor], s.order[s.cursor-1] = s.order[s.cursor-1], s.order[s.cursor]
		s.cursor--
	}
}

func (s *PrioritySelector) render() {
	// Move cursor up to overwrite previous render
	// First render: just print. Subsequent: clear and reprint
	s.clearDisplay()

	fmt.Println("Set playback priority (j/k move, J/K reorder, Enter confirm, q quit):")
	fmt.Println()

	for i, idx := range s.order {
		prefix := "  "
		suffix := ""
		if i == s.cursor {
			prefix = "> "
			suffix = " <"
		}
		fmt.Printf("%s%d. %s%s\n", prefix, i+1, s.items[idx], suffix)
	}
	fmt.Println()
}

func (s *PrioritySelector) clearDisplay() {
	// Move cursor up and clear lines
	// Total lines: 1 (header) + 1 (blank) + len(items) + 1 (blank) = len(items) + 3
	lines := len(s.order) + 3
	for i := 0; i < lines; i++ {
		fmt.Print("\033[A\033[K") // Move up and clear line
	}
}

// GetLabelsInOrder returns the item labels in priority order
func (s *PrioritySelector) GetLabelsInOrder() []string {
	result := make([]string, len(s.order))
	for i, idx := range s.order {
		result[i] = s.items[idx]
	}
	return result
}

// SimpleSelect provides a simple numbered selection without reordering
// Returns the selected index (0-based) or -1 if canceled
func SimpleSelect(prompt string, items []string) (int, error) {
	fmt.Println(prompt)
	for i, item := range items {
		fmt.Printf("%d. %s\n", i+1, item)
	}
	fmt.Print("\nSelect option (1-", len(items), ", q to quit): ")

	if err := keyboard.Open(); err != nil {
		return -1, fmt.Errorf("failed to open keyboard: %w", err)
	}
	defer keyboard.Close()

	var input strings.Builder
	for {
		char, key, err := keyboard.GetKey()
		if err != nil {
			return -1, err
		}

		if key == keyboard.KeyEnter {
			fmt.Println()
			break
		}

		if char == 'q' || key == keyboard.KeyEsc {
			fmt.Println()
			return -1, nil
		}

		if char >= '0' && char <= '9' {
			input.WriteRune(char)
			fmt.Print(string(char))
		}

		if key == keyboard.KeyBackspace || key == keyboard.KeyBackspace2 {
			s := input.String()
			if len(s) > 0 {
				input.Reset()
				input.WriteString(s[:len(s)-1])
				fmt.Print("\b \b")
			}
		}
	}

	var choice int
	if _, err := fmt.Sscanf(input.String(), "%d", &choice); err != nil {
		return -1, fmt.Errorf("invalid input")
	}

	if choice < 1 || choice > len(items) {
		return -1, fmt.Errorf("selection out of range")
	}

	return choice - 1, nil
}
