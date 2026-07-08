package widgets

import "testing"

func TestExtractRangeLines(t *testing.T) {
	lines := []string{"hello", "world", "héllo"}
	tests := []struct {
		name string
		a, b textPos
		want string
	}{
		{"single line", textPos{0, 1}, textPos{0, 3}, "el"},
		{"reversed single line", textPos{0, 3}, textPos{0, 1}, "el"},
		{"two lines", textPos{0, 3}, textPos{1, 2}, "lo\nwo"},
		{"reversed two lines", textPos{1, 2}, textPos{0, 3}, "lo\nwo"},
		{"three lines with multibyte", textPos{0, 4}, textPos{2, 3}, "o\nworld\nhé"},
		{"empty", textPos{1, 2}, textPos{1, 2}, ""},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := extractRangeLines(lines, tt.a, tt.b); got != tt.want {
				t.Errorf("extractRangeLines(%v, %v) = %q, want %q", tt.a, tt.b, got, tt.want)
			}
		})
	}
}

func TestPickClipboard(t *testing.T) {
	tests := []struct {
		name                  string
		osText, osAtCopy, app string
		want                  string
	}{
		{"nothing anywhere", "", "", "", ""},
		{"only OS text", "os", "", "", "os"},
		{"app copy, OS unchanged since", "os", "os", "app", "app"},
		{"app copy, OS empty", "", "", "app", "app"},
		{"OS changed after app copy", "newer", "os", "app", "newer"},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := pickClipboard(tt.osText, tt.osAtCopy, tt.app); got != tt.want {
				t.Errorf("pickClipboard(%q, %q, %q) = %q, want %q", tt.osText, tt.osAtCopy, tt.app, got, tt.want)
			}
		})
	}
}
