package widgets

import (
	"unicode/utf8"

	"github.com/jupiterrider/purego-sdl3/sdl"
)

// isPasteShortcut reports whether the key event is the paste chord (Ctrl+V),
// shared by TextInput and TextArea.
func isPasteShortcut(key sdl.KeyboardEvent) bool {
	return key.Scancode == sdl.ScancodeV && key.Mod&sdl.KeymodCtrl != 0
}

// isSelectAllShortcut reports whether the key event is the select-all chord
// (Ctrl+A), shared by TextInput and TextArea.
func isSelectAllShortcut(key sdl.KeyboardEvent) bool {
	return key.Scancode == sdl.ScancodeA && key.Mod&sdl.KeymodCtrl != 0
}

// isMovementKey reports whether scancode moves the caret, i.e. whether
// Shift+key should extend the selection instead of editing.
func isMovementKey(scancode sdl.Scancode) bool {
	switch scancode {
	case sdl.ScancodeLeft, sdl.ScancodeRight, sdl.ScancodeHome, sdl.ScancodeEnd,
		sdl.ScancodeUp, sdl.ScancodeDown:
		return true
	}
	return false
}

// editLineKey applies a single-line editing key to line at byte offset col
// and returns the new line, the new cursor offset, and whether the key was
// consumed. It is shared by TextInput and TextArea so keys like Home/End
// behave identically in both. Keys that would cross a line boundary
// (Backspace at column 0, Left at column 0, Right/Delete at end of line)
// are reported as unhandled so multi-line widgets can join/hop lines.
func editLineKey(scancode sdl.Scancode, line string, col int) (string, int, bool) {
	switch scancode {
	case sdl.ScancodeBackspace:
		if col > 0 {
			_, size := utf8.DecodeLastRuneInString(line[:col])
			return line[:col-size] + line[col:], col - size, true
		}
	case sdl.ScancodeDelete:
		if col < len(line) {
			_, size := utf8.DecodeRuneInString(line[col:])
			return line[:col] + line[col+size:], col, true
		}
	case sdl.ScancodeLeft:
		if col > 0 {
			_, size := utf8.DecodeLastRuneInString(line[:col])
			return line, col - size, true
		}
	case sdl.ScancodeRight:
		if col < len(line) {
			_, size := utf8.DecodeRuneInString(line[col:])
			return line, col + size, true
		}
	case sdl.ScancodeHome:
		return line, 0, true
	case sdl.ScancodeEnd:
		return line, len(line), true
	}
	return line, col, false
}
