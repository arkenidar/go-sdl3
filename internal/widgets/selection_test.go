package widgets

import (
	"reflect"
	"testing"
	"unicode/utf8"
)

func TestDeleteRangeLine(t *testing.T) {
	tests := []struct {
		name     string
		line     string
		a, b     int
		wantLine string
		wantCol  int
	}{
		{"middle", "hello", 1, 3, "hlo", 1},
		{"reversed range", "hello", 3, 1, "hlo", 1},
		{"whole line", "hello", 0, 5, "", 0},
		{"multibyte", "héllo", 1, 3, "hllo", 1}, // é is 2 bytes
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			gotLine, gotCol := deleteRangeLine(tt.line, tt.a, tt.b)
			if gotLine != tt.wantLine || gotCol != tt.wantCol {
				t.Errorf("deleteRangeLine(%q, %d, %d) = (%q, %d), want (%q, %d)",
					tt.line, tt.a, tt.b, gotLine, gotCol, tt.wantLine, tt.wantCol)
			}
		})
	}
}

func TestDeleteRangeLines(t *testing.T) {
	tests := []struct {
		name      string
		lines     []string
		a, b      textPos
		wantLines []string
		wantPos   textPos
	}{
		{"single line", []string{"hello", "world"}, textPos{0, 1}, textPos{0, 3}, []string{"hlo", "world"}, textPos{0, 1}},
		{"two lines", []string{"hello", "world"}, textPos{0, 3}, textPos{1, 2}, []string{"helrld"}, textPos{0, 3}},
		{"reversed", []string{"hello", "world"}, textPos{1, 2}, textPos{0, 3}, []string{"helrld"}, textPos{0, 3}},
		{"spanning middle lines", []string{"aa", "bb", "cc", "dd"}, textPos{0, 1}, textPos{3, 1}, []string{"ad"}, textPos{0, 1}},
		{"everything", []string{"aa", "bb"}, textPos{0, 0}, textPos{1, 2}, []string{""}, textPos{0, 0}},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			gotLines, gotPos := deleteRangeLines(tt.lines, tt.a, tt.b)
			if !reflect.DeepEqual(gotLines, tt.wantLines) || gotPos != tt.wantPos {
				t.Errorf("deleteRangeLines(%v, %v, %v) = (%v, %v), want (%v, %v)",
					tt.lines, tt.a, tt.b, gotLines, gotPos, tt.wantLines, tt.wantPos)
			}
		})
	}
}

func TestLineSelSpan(t *testing.T) {
	lo, hi := textPos{1, 2}, textPos{3, 4}
	tests := []struct {
		name    string
		row     int
		lineLen int
		wantA   int
		wantB   int
		wantOK  bool
	}{
		{"before selection", 0, 10, 0, 0, false},
		{"first row", 1, 10, 2, 10, true},
		{"middle row", 2, 7, 0, 7, true},
		{"last row", 3, 10, 0, 4, true},
		{"after selection", 4, 10, 0, 0, false},
		{"empty middle row", 2, 0, 0, 0, false},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			a, b, ok := lineSelSpan(tt.row, tt.lineLen, lo, hi)
			if a != tt.wantA || b != tt.wantB || ok != tt.wantOK {
				t.Errorf("lineSelSpan(%d, %d) = (%d, %d, %v), want (%d, %d, %v)",
					tt.row, tt.lineLen, a, b, ok, tt.wantA, tt.wantB, tt.wantOK)
			}
		})
	}

	if _, _, ok := lineSelSpan(0, 5, textPos{0, 2}, textPos{0, 2}); ok {
		t.Error("empty selection should select nothing")
	}
}

// fakeWidth pretends every rune renders 10px wide.
func fakeWidth(s string) float32 {
	return float32(utf8.RuneCountInString(s)) * 10
}

func TestByteOffsetForX(t *testing.T) {
	tests := []struct {
		name string
		line string
		x    float32
		want int
	}{
		{"before start", "hello", -3, 0},
		{"first half of first rune", "hello", 4, 0},
		{"second half of first rune", "hello", 6, 1},
		{"middle", "hello", 24, 2},
		{"past end", "hello", 999, 5},
		{"empty line", "", 30, 0},
		{"multibyte snaps to rune boundary", "héllo", 16, 3}, // after é (2 bytes)
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := byteOffsetForX(tt.line, tt.x, fakeWidth); got != tt.want {
				t.Errorf("byteOffsetForX(%q, %v) = %d, want %d", tt.line, tt.x, got, tt.want)
			}
		})
	}
}

func TestInsertAtCursorReplacesSelection(t *testing.T) {
	area := &TextArea{
		Lines:     []string{"hello", "world"},
		cursorRow: 1, cursorCol: 2,
		selAnchor: textPos{0, 3},
		hasSel:    true,
	}
	area.insertAtCursor("XY")
	wantLines := []string{"helXYrld"}
	if !reflect.DeepEqual(area.Lines, wantLines) || area.cursorRow != 0 || area.cursorCol != 5 {
		t.Errorf("got (%v, row %d, col %d), want (%v, row 0, col 5)",
			area.Lines, area.cursorRow, area.cursorCol, wantLines)
	}
	if area.hasSel {
		t.Error("selection should be cleared after replacement")
	}
}

func TestTextInputDeleteSelection(t *testing.T) {
	in := &TextInput{Value: "héllo", cursorCol: 1, selAnchor: 5, hasSel: true}
	if !in.deleteSelection() {
		t.Fatal("expected deleteSelection to report true")
	}
	if in.Value != "ho" || in.cursorCol != 1 || in.hasSel {
		t.Errorf("got (%q, col %d, hasSel %v), want (%q, col 1, false)", in.Value, in.cursorCol, in.hasSel, "ho")
	}
	if in.deleteSelection() {
		t.Error("second deleteSelection should report false")
	}
}
