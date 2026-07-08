package widgets

import (
	"testing"

	"github.com/jupiterrider/purego-sdl3/sdl"
)

func TestEditLineKey(t *testing.T) {
	tests := []struct {
		name     string
		scancode sdl.Scancode
		line     string
		col      int
		wantLine string
		wantCol  int
		handled  bool
	}{
		{"home from middle", sdl.ScancodeHome, "hello", 3, "hello", 0, true},
		{"end from middle", sdl.ScancodeEnd, "hello", 3, "hello", 5, true},
		{"left", sdl.ScancodeLeft, "hello", 3, "hello", 2, true},
		{"left at start unhandled", sdl.ScancodeLeft, "hello", 0, "hello", 0, false},
		{"right", sdl.ScancodeRight, "hello", 3, "hello", 4, true},
		{"right at end unhandled", sdl.ScancodeRight, "hello", 5, "hello", 5, false},
		{"backspace mid", sdl.ScancodeBackspace, "hello", 3, "helo", 2, true},
		{"backspace at start unhandled", sdl.ScancodeBackspace, "hello", 0, "hello", 0, false},
		{"delete mid", sdl.ScancodeDelete, "hello", 3, "helo", 3, true},
		{"delete at end unhandled", sdl.ScancodeDelete, "hello", 5, "hello", 5, false},
		{"left over multibyte rune", sdl.ScancodeLeft, "héllo", 3, "héllo", 1, true},
		{"backspace multibyte rune", sdl.ScancodeBackspace, "héllo", 3, "hllo", 1, true},
		{"right over multibyte rune", sdl.ScancodeRight, "héllo", 1, "héllo", 3, true},
		{"unrelated key unhandled", sdl.ScancodeUp, "hello", 3, "hello", 3, false},
	}
	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			line, col, handled := editLineKey(tc.scancode, tc.line, tc.col)
			if line != tc.wantLine || col != tc.wantCol || handled != tc.handled {
				t.Errorf("got (%q, %d, %v), want (%q, %d, %v)",
					line, col, handled, tc.wantLine, tc.wantCol, tc.handled)
			}
		})
	}
}

func TestTextAreaInsertAtCursor(t *testing.T) {
	tests := []struct {
		name      string
		lines     []string
		row, col  int
		insert    string
		wantLines []string
		wantRow   int
		wantCol   int
	}{
		{"plain text mid-line", []string{"hello"}, 0, 2, "XY", []string{"heXYllo"}, 0, 4},
		{"single newline splits", []string{"hello"}, 0, 2, "\n", []string{"he", "llo"}, 1, 0},
		{"multi-line paste", []string{"ab"}, 0, 1, "1\n2\n3", []string{"a1", "2", "3b"}, 2, 1},
		{"crlf normalized", []string{""}, 0, 0, "x\r\ny", []string{"x", "y"}, 1, 1},
		{"cr normalized", []string{""}, 0, 0, "x\ry", []string{"x", "y"}, 1, 1},
		{"paste between existing lines", []string{"aa", "bb"}, 0, 2, "1\n2", []string{"aa1", "2", "bb"}, 1, 1},
	}
	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			area := &TextArea{Lines: append([]string{}, tc.lines...), cursorRow: tc.row, cursorCol: tc.col}
			area.insertAtCursor(tc.insert)
			if len(area.Lines) != len(tc.wantLines) || area.cursorRow != tc.wantRow || area.cursorCol != tc.wantCol {
				t.Fatalf("got lines=%q row=%d col=%d, want lines=%q row=%d col=%d",
					area.Lines, area.cursorRow, area.cursorCol, tc.wantLines, tc.wantRow, tc.wantCol)
			}
			for i := range tc.wantLines {
				if area.Lines[i] != tc.wantLines[i] {
					t.Fatalf("line %d: got %q, want %q", i, area.Lines[i], tc.wantLines[i])
				}
			}
		})
	}
}
