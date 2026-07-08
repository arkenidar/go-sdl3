package main

import (
	"encoding/binary"
	"strings"
	"testing"
	"unicode/utf8"
	"unsafe"

	"github.com/jupiterrider/purego-sdl3/sdl"
)

// TestTextEventSurvivesBufferReuse guards the queued-dispatch lifetime bug:
// SDL owns a text-input event's string only temporarily, and this FFI layer
// dispatches events after pumping, so the raw pointer inside the stored
// sdl.Event can dangle by dispatch time. retainText must copy the string
// into Go memory and repoint the event, so clobbering the original buffer
// (as SDL reusing it would) leaves dispatch unaffected.
func TestTextEventSurvivesBufferReuse(t *testing.T) {
	_, textinput, _ := setupFocusHandoffScene(t)

	// Focus the TextInput with a click, like a real session would.
	currentEvent = queuedEvent{kind: guiEventMouseDown, ev: synthMouseDown(), mx: 50, my: 36}
	var handled int32
	gui_widget_update(textinput, &handled)

	// Build a text event whose pointer targets a buffer we control,
	// mirroring what gui_pump_events receives from SDL.
	const typed = "héllo€"
	buf := append([]byte(typed), 0)
	var raw sdl.Event
	binary.LittleEndian.PutUint32(raw[:4], uint32(sdl.EventTextInput))
	binary.LittleEndian.PutUint64(raw[sdlTextInputEventTextOffset:sdlTextInputEventTextOffset+8],
		uint64(uintptr(unsafe.Pointer(&buf[0]))))

	qe := queuedEvent{kind: guiEventTextInput, ev: raw}
	textEv := qe.ev.Text()
	qe.text = textEv.Text()
	qe.retainText()

	// Simulate SDL reusing the buffer between pump and dispatch.
	for i := range buf {
		buf[i] = 'X'
	}

	currentEvent = qe
	gui_widget_update(textinput, &handled)

	if got := textInputValue(t, textinput); got != typed {
		t.Errorf("TextInput received %q, want %q", got, typed)
	}
}

func TestCStringLenRuneBoundary(t *testing.T) {
	tests := []struct {
		name string
		s    string
		max  int
		want string
	}{
		{"fits", "héllo", 31, "héllo"},
		{"ascii cut", "abcdef", 3, "abc"},
		{"cut lands mid-rune", "aé", 2, "a"}, // é = 2 bytes at offset 1..2
		{"cut on boundary", "aé", 3, "aé"},
		{"long multibyte", strings.Repeat("é", 20), 31, strings.Repeat("é", 15)},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			n := cStringLen([]byte(tt.s), tt.max)
			got := tt.s[:n]
			if got != tt.want {
				t.Errorf("cStringLen(%q, %d) keeps %q, want %q", tt.s, tt.max, got, tt.want)
			}
			if !utf8.ValidString(got) {
				t.Errorf("cStringLen(%q, %d) produced invalid UTF-8 %q", tt.s, tt.max, got)
			}
		})
	}
}
