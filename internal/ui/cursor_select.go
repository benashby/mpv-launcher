package ui

import (
	"fmt"

	"github.com/eiannone/keyboard"
)

// CursorSelect presents a cursor-navigable list and returns the selected index (0-based).
// Returns -1 if the user cancels (q).
// Controls: j/↓ move down, k/↑ move up, Enter confirm, 1-9 jump and confirm, q cancel.
func CursorSelect(prompt string, items []string) (int, error) {
	if len(items) == 0 {
		return -1, fmt.Errorf("no items to select")
	}

	cursor := 0
	firstRender := true

	render := func() {
		if !firstRender {
			// Clear previous render: header + blank + items + blank = len(items) + 3 lines
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

	if err := keyboard.Open(); err != nil {
		return -1, fmt.Errorf("failed to open keyboard: %w", err)
	}
	defer keyboard.Close()

	render()

	for {
		char, key, err := keyboard.GetKey()
		if err != nil {
			return -1, fmt.Errorf("failed to read key: %w", err)
		}

		switch {
		case key == keyboard.KeyEnter:
			return cursor, nil

		case char == 'q':
			return -1, nil

		case char == 'j' || key == keyboard.KeyArrowDown:
			if cursor < len(items)-1 {
				cursor++
			}

		case char == 'k' || key == keyboard.KeyArrowUp:
			if cursor > 0 {
				cursor--
			}

		case char >= '1' && char <= '9':
			idx := int(char-'1')
			if idx < len(items) {
				return idx, nil
			}
		}

		render()
	}
}
