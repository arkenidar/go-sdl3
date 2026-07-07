package main

import (
	"encoding/binary"
	"testing"

	"github.com/jupiterrider/purego-sdl3/sdl"
)

// TestButtonClickRoundTrip proves the full FFI round trip end-to-end without
// a real display/click: create a button through the exported C-ABI
// functions, synthesize a mouse-down event inside its bounds, dispatch it
// through gui_widget_update exactly as the LuaJIT/ctypes examples would,
// and confirm gui_button_was_clicked reports it.
func TestButtonClickRoundTrip(t *testing.T) {
	if s := gui_init(); s != 0 {
		t.Fatalf("gui_init failed: %s", getLastError())
	}
	defer gui_quit()

	window := testCreateWindow("test", 200, 100)
	if window == 0 {
		t.Fatalf("gui_create_window failed: %s", getLastError())
	}

	var renderer uint64
	if s := gui_window_get_renderer(window, &renderer); s != 0 {
		t.Fatalf("gui_window_get_renderer failed: %s", getLastError())
	}

	font := testLoadFont("../assets/OpenDyslexic-Regular.ttf", 16)
	if font == 0 {
		t.Fatalf("gui_load_font failed: %s", getLastError())
	}

	button := testButtonNew(renderer, font, 10, 10, 80, 30, "Click")
	if button == 0 {
		t.Fatalf("gui_button_new failed: %s", getLastError())
	}

	// Synthesize a mouse-down event inside the button's bounds (10,10)-(90,40).
	var raw sdl.Event
	binary.LittleEndian.PutUint32(raw[:4], uint32(sdl.EventMouseButtonDown))
	currentEvent = queuedEvent{kind: guiEventMouseDown, ev: raw, mx: 50, my: 25}

	var handled int32
	if s := gui_widget_update(button, &handled); s != 0 {
		t.Fatalf("gui_widget_update failed: %s", getLastError())
	}
	if handled != 1 {
		t.Fatalf("expected click to be handled, got handled=%d", handled)
	}

	var clicked int32
	if s := gui_button_was_clicked(button, &clicked); s != 0 {
		t.Fatalf("gui_button_was_clicked failed: %s", getLastError())
	}
	if clicked != 1 {
		t.Fatalf("expected gui_button_was_clicked to report true, got %d", clicked)
	}

	// A second poll without a new click should report false.
	if s := gui_button_was_clicked(button, &clicked); s != 0 {
		t.Fatalf("gui_button_was_clicked (2nd) failed: %s", getLastError())
	}
	if clicked != 0 {
		t.Fatalf("expected no click on second poll, got %d", clicked)
	}
}
