package main

import (
	"encoding/binary"
	"testing"

	"github.com/jupiterrider/purego-sdl3/sdl"
)

// TestButtonNativeCallback proves the opt-in native-callback path (Phase 3):
// gui_button_set_onclick registers a real C function pointer
// (test_click_callback_ptr, in callback_test_support.go), and a synthetic
// click dispatches through gui_widget_update exactly as ctypes/LuaJIT would,
// ending up invoking native C code (not just Go-side state) via the cgo
// trampoline in callbacks.go.
func TestButtonNativeCallback(t *testing.T) {
	if s := gui_init(); s != 0 {
		t.Fatalf("gui_init failed: %s", getLastError())
	}
	defer gui_quit()

	window := testCreateWindow("test", 200, 100)
	var renderer uint64
	gui_window_get_renderer(window, &renderer)
	font := testLoadFont("../assets/OpenDyslexic-Regular.ttf", 16)

	button := testButtonNew(renderer, font, 10, 10, 80, 30, "Click")
	if button == 0 {
		t.Fatalf("gui_button_new failed: %s", getLastError())
	}

	testResetClickCallback()
	if s := testRegisterClickCallback(button); s != 0 {
		t.Fatalf("gui_button_set_onclick failed: %s", getLastError())
	}

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

	if got := testClickCallbackCount(); got != 1 {
		t.Fatalf("expected native callback to fire exactly once, got %d", got)
	}
	if got := testClickCallbackLastHandle(); got != button {
		t.Fatalf("expected callback to receive handle %d, got %d", button, got)
	}

	// The poll-based check still works independently of the callback.
	var clicked int32
	if s := gui_button_was_clicked(button, &clicked); s != 0 {
		t.Fatalf("gui_button_was_clicked failed: %s", getLastError())
	}
	if clicked != 1 {
		t.Fatalf("expected poll-based click to still report true, got %d", clicked)
	}

	// gui_widget_destroy must clean up the callback registration too, so a
	// stale/reused handle can never fire a callback for a destroyed widget.
	if s := gui_widget_destroy(button); s != 0 {
		t.Fatalf("gui_widget_destroy failed: %s", getLastError())
	}
	callbackMu.Lock()
	_, stillRegistered := buttonCallbacks[button]
	callbackMu.Unlock()
	if stillRegistered {
		t.Fatalf("expected callback registration to be removed after destroy")
	}
}

// TestButtonSetOnclickRejectsNonButton guards the type check: registering a
// callback on a handle that isn't a Button should fail cleanly rather than
// panicking or silently no-op-ing.
func TestButtonSetOnclickRejectsNonButton(t *testing.T) {
	if s := gui_init(); s != 0 {
		t.Fatalf("gui_init failed: %s", getLastError())
	}
	defer gui_quit()

	window := testCreateWindow("test", 200, 100)
	var renderer uint64
	gui_window_get_renderer(window, &renderer)
	font := testLoadFont("../assets/OpenDyslexic-Regular.ttf", 16)

	label := testLabelNew(renderer, font, 0, 0, "hi")

	if s := testRegisterClickCallback(label); s == 0 {
		t.Fatalf("expected gui_button_set_onclick to fail for a non-Button handle")
	}
}
