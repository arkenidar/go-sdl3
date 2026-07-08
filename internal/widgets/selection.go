package widgets

import "unicode/utf8"

// textPos addresses a position in a TextArea as (line index, byte offset
// within that line).
type textPos struct{ row, col int }

func (p textPos) before(q textPos) bool {
	return p.row < q.row || (p.row == q.row && p.col < q.col)
}

// orderInts returns a and b in ascending order, normalizing a single-line
// selection whose anchor may sit after the caret.
func orderInts(a, b int) (int, int) {
	if a > b {
		return b, a
	}
	return a, b
}

// orderPos is the two-dimensional counterpart of orderInts.
func orderPos(a, b textPos) (textPos, textPos) {
	if b.before(a) {
		return b, a
	}
	return a, b
}

// deleteRangeLine removes the bytes between offsets a and b (in either
// order) from line and returns the new line plus the collapsed caret offset.
func deleteRangeLine(line string, a, b int) (string, int) {
	lo, hi := orderInts(a, b)
	return line[:lo] + line[hi:], lo
}

// deleteRangeLines removes the selection between positions a and b (in
// either order) from lines, joining the partial first and last lines, and
// returns the new lines plus the collapsed caret position. It does not
// mutate the input slice.
func deleteRangeLines(lines []string, a, b textPos) ([]string, textPos) {
	lo, hi := orderPos(a, b)
	if lo.row == hi.row {
		newLine, col := deleteRangeLine(lines[lo.row], lo.col, hi.col)
		out := append([]string{}, lines...)
		out[lo.row] = newLine
		return out, textPos{lo.row, col}
	}
	joined := lines[lo.row][:lo.col] + lines[hi.row][hi.col:]
	out := append([]string{}, lines[:lo.row]...)
	out = append(out, joined)
	out = append(out, lines[hi.row+1:]...)
	return out, lo
}

// extractRangeLines returns the text selected between positions a and b (in
// either order), with line breaks rendered as "\n".
func extractRangeLines(lines []string, a, b textPos) string {
	lo, hi := orderPos(a, b)
	if lo.row == hi.row {
		return lines[lo.row][lo.col:hi.col]
	}
	parts := []string{lines[lo.row][lo.col:]}
	parts = append(parts, lines[lo.row+1:hi.row]...)
	parts = append(parts, lines[hi.row][:hi.col])
	out := parts[0]
	for _, p := range parts[1:] {
		out += "\n" + p
	}
	return out
}

// lineSelSpan returns the byte range [a, b) of the given row covered by the
// normalized selection lo..hi, and whether any part of the row is selected.
func lineSelSpan(row, lineLen int, lo, hi textPos) (int, int, bool) {
	if lo == hi || row < lo.row || row > hi.row {
		return 0, 0, false
	}
	a, b := 0, lineLen
	if row == lo.row {
		a = lo.col
	}
	if row == hi.row {
		b = hi.col
	}
	if a >= b {
		return 0, 0, false
	}
	return a, b, true
}

// byteOffsetForX maps a pixel x offset (relative to the line's left edge) to
// the nearest rune boundary in line, measuring prefixes with width. Clicks
// past either end clamp to 0 or len(line).
func byteOffsetForX(line string, x float32, width func(string) float32) int {
	if x <= 0 {
		return 0
	}
	prev := float32(0)
	for i := 0; i < len(line); {
		_, size := utf8.DecodeRuneInString(line[i:])
		w := width(line[:i+size])
		if x < (prev+w)/2 {
			return i
		}
		if x < w {
			return i + size
		}
		prev = w
		i += size
	}
	return len(line)
}
