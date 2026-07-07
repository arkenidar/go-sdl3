package main

import (
	"encoding/binary"
	"testing"
	"unsafe"

	"github.com/jupiterrider/purego-sdl3/sdl"
)

// synthTextInputEvent builds a raw sdl.Event with Type=EventTextInput and a
// real null-terminated text pointer, replicating the memory layout SDL uses
// (TextInputEvent's `text` field is unexported, so it can't be set via the
// public API -- this pokes the bytes directly at the known offset:
// CommonEvent(16) + WindowID uint32(4) + 4 bytes padding = 24).
func synthTextInputEvent(text string) sdl.Event {
	var raw sdl.Event
	binary.LittleEndian.PutUint32(raw[:4], uint32(sdl.EventTextInput))
	data := append([]byte(text), 0)
	ptr := unsafe.Pointer(&data[0])
	binary.LittleEndian.PutUint64(raw[24:32], uint64(uintptr(ptr)))
	return raw
}

func TestTextAreaInsertRoundTrip(t *testing.T) {
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

	var ta uint64
	if s := gui_textarea_new(renderer, font, window, 0, 0, 200, 80, &ta); s != 0 {
		t.Fatalf("gui_textarea_new failed: %s", getLastError())
	}

	// Focus via a synthetic mouse-down inside bounds.
	var raw sdl.Event
	binary.LittleEndian.PutUint32(raw[:4], uint32(sdl.EventMouseButtonDown))
	currentEvent = queuedEvent{kind: guiEventMouseDown, ev: raw, mx: 50, my: 25}
	var handled int32
	gui_widget_update(ta, &handled)

	v, _ := handles.get(ta)
	t.Logf("focused after click: %v", v)

	// Now dispatch a real text-input event.
	currentEvent = queuedEvent{kind: guiEventTextInput, ev: synthTextInputEvent("abcd"), mx: 50, my: 25}
	if s := gui_widget_update(ta, &handled); s != 0 {
		t.Fatalf("gui_widget_update (text) failed: %s", getLastError())
	}
	t.Logf("handled=%d", handled)

	text := gui_textarea_get_text(ta)
	got := goStringFromC(text)
	if got != "abcd" {
		t.Fatalf("expected TextArea text %q, got %q", "abcd", got)
	}
}
