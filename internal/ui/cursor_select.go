package ui

import (
	"fmt"
)

// CursorSelect presents a cursor-navigable list and returns the selected index (0-based).
// Returns -1 if the user cancels (q or Esc).
// Controls: j/↓ move down, k/↑ move up, Enter confirm, 1-9 jump and confirm, q/Esc cancel.
func CursorSelect(prompt string, items []string) (int, error) {
	if len(items) == 0 {
		return -1, fmt.Errorf("no items to select")
	}

	rr, err := newRawReader()
	if err != nil {
		return -1, fmt.Errorf("failed to set raw mode: %w", err)
	}
	defer rr.close()

	cursor := 0
	firstRender := true

	render := func() {
		if !firstRender {
			lines := len(items) + 3
			for i := 0; i < lines; i++ {
				fmt.Print("\033[A\033[K")
			}
		}
		firstRender = false

		fmt.Println(prompt)
		fmt.Println()
		for i, item := range items {
			if i == cursor {
				fmt.Printf("> %d. %s\n", i+1, item)
			} else {
				fmt.Printf("  %d. %s\n", i+1, item)
			}
		}
		fmt.Println()
	}

	render()

	for {
		ch, key := rr.readKey()

		switch {
		case key == rawKeyEnter:
			return cursor, nil

		case ch == 'q' || key == rawKeyEscape:
			return -1, nil

		case ch == 'j' || key == rawKeyDown:
			if cursor < len(items)-1 {
				cursor++
			}

		case ch == 'k' || key == rawKeyUp:
			if cursor > 0 {
				cursor--
			}

		case ch >= '1' && ch <= '9':
			idx := int(ch - '1')
			if idx < len(items) {
				return idx, nil
			}
		}

		render()
	}
}
