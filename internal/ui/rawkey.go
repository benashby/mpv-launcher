package ui

import (
	"os"

	"golang.org/x/sys/unix"
	"golang.org/x/term"
)

type rawKey int

const (
	rawKeyOther  rawKey = iota
	rawKeyEnter
	rawKeyEscape
	rawKeyUp
	rawKeyDown
)

// rawReader puts stdin into raw mode and reads key events, parsing arrow key
// escape sequences correctly. Uses Select with a short timeout to distinguish
// a lone ESC from the start of an escape sequence.
type rawReader struct {
	fd    int
	state *term.State
}

func newRawReader() (*rawReader, error) {
	fd := int(os.Stdin.Fd())
	state, err := term.MakeRaw(fd)
	if err != nil {
		return nil, err
	}
	return &rawReader{fd: fd, state: state}, nil
}

func (r *rawReader) close() {
	term.Restore(r.fd, r.state) //nolint:errcheck
}

// readKey blocks until a key is pressed. Arrow keys are parsed from their
// three-byte escape sequences (\x1b[A etc). Returns the rune for regular
// chars, or 0 + a rawKey constant for special keys.
func (r *rawReader) readKey() (rune, rawKey) {
	buf := make([]byte, 1)
	if _, err := os.Stdin.Read(buf); err != nil {
		return 0, rawKeyOther
	}

	switch buf[0] {
	case 13, 10: // CR / LF
		return 0, rawKeyEnter

	case 0x1b: // ESC — could be lone ESC or start of \x1b[X sequence
		if !r.inputReady(50) {
			return 0, rawKeyEscape
		}
		b1 := make([]byte, 1)
		os.Stdin.Read(b1) //nolint:errcheck
		if b1[0] != '[' || !r.inputReady(50) {
			return 0, rawKeyEscape
		}
		b2 := make([]byte, 1)
		os.Stdin.Read(b2) //nolint:errcheck
		switch b2[0] {
		case 'A':
			return 0, rawKeyUp
		case 'B':
			return 0, rawKeyDown
		}
		return 0, rawKeyOther

	default:
		return rune(buf[0]), rawKeyOther
	}
}

// inputReady reports whether stdin has bytes available within timeoutMs.
func (r *rawReader) inputReady(timeoutMs int) bool {
	var fds unix.FdSet
	fds.Set(r.fd)
	tv := unix.NsecToTimeval(int64(timeoutMs) * 1_000_000)
	n, _ := unix.Select(r.fd+1, &fds, nil, nil, &tv)
	return n > 0
}
