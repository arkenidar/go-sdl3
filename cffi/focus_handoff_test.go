package main

import (
	"encoding/binary"
	"testing"

	"arkenidar.com/purego-sdl3/internal/widgets"
	"github.com/jupiterrider/purego-sdl3/sdl"
)

// setupFocusHandoffScene builds a TextInput (inside a VStack) and a
// TextArea sharing one window, plus mouseDown/typeText/keyDown helpers that
// synthesize real sdl.Events -- mirroring what a live click/keystroke would
// produce, without needing an actual display.
func setupFocusHandoffScene(t *testing.T) (stack, textinput, textarea uint64) {
	t.Helper()
	if s := gui_init(); s != 0 {
		t.Fatalf("gui_init failed: %s", getLastError())
	}
	t.Cleanup(func() { gui_quit() })

	window := testCreateWindow("test", 400, 400)
	var renderer uint64
	gui_window_get_renderer(window, &renderer)
	font := testLoadFont("../assets/OpenDyslexic-Regular.ttf", 16)

	gui_textinput_new(renderer, font, window, 20, 20, 200, 32, &textinput)
	gui_vstack_new(20, 20, 10, &stack)
	gui_vstack_add(stack, textinput)
	gui_textarea_new(renderer, font, window, 20, 100, 200, 100, &textarea)
	return stack, textinput, textarea
}

func textInputValue(t *testing.T, handle uint64) string {
	t.Helper()
	v, ok := handles.get(handle)
	if !ok {
		t.Fatalf("unknown TextInput handle %d", handle)
	}
	ti, ok := v.(*widgets.TextInput)
	if !ok {
		t.Fatalf("handle %d is not a TextInput", handle)
	}
	return ti.Value
}

func synthMouseDown() sdl.Event {
	var raw sdl.Event
	binary.LittleEndian.PutUint32(raw[:4], uint32(sdl.EventMouseButtonDown))
	return raw
}

// TestFocusHandoffUnconditionalDispatch guards two bugs found in sequence
// while building the lua-use/python-use FFI demos, both around a TextInput
// and a TextArea sharing one SDL window's text-input state:
//
//  1. Blur() used to call SDL_StopTextInput. Switching focus TextInput ->
//     TextArea meant TextInput's Focus() called StartTextInput, and the
//     same click, fed to TextArea, called its Blur() -> StopTextInput --
//     racing after the StartTextInput and leaving the newly focused widget
//     with a rendered caret but no real OS-level text composition. Fixed by
//     removing StopTextInput from Blur() entirely: harmless to leave text
//     input enabled since every widget already gates insertion on its own
//     Focused flag.
//
//  2. The first fix attempt (short-circuiting dispatch: skip a container
//     once a prior one handles the event) broke Blur() propagation instead:
//     if `stack` handles a click by focusing the TextInput, `textarea`
//     never sees that click and never learns it should blur -- so both end
//     up Focused at once (two blinking carets). The actual fix is (1); this
//     rules (2) back out by asserting every container is always dispatched
//     to and that only one widget is Focused at a time.
func TestFocusHandoffUnconditionalDispatch(t *testing.T) {
	stack, textinput, textarea := setupFocusHandoffScene(t)

	dispatch := func(kind int32, ev sdl.Event, mx, my float32) {
		currentEvent = queuedEvent{kind: kind, ev: ev, mx: mx, my: my}
		var handled int32
		gui_widget_update(stack, &handled)
		gui_widget_update(textarea, &handled)
	}

	focusedTextInput := func() bool { return textInputFocused(t, textinput) }
	focusedTextArea := func() bool {
		v, _ := handles.get(textarea)
		return v.(*widgets.TextArea).Focused
	}

	dispatch(guiEventMouseDown, synthMouseDown(), 50, 36) // click TextInput
	dispatch(guiEventTextInput, synthTextInputEvent("one"), 50, 36)
	if !focusedTextInput() || focusedTextArea() {
		t.Fatalf("after focusing TextInput: textinput.Focused=%v textarea.Focused=%v, want true/false",
			focusedTextInput(), focusedTextArea())
	}

	dispatch(guiEventMouseDown, synthMouseDown(), 50, 150) // click TextArea
	dispatch(guiEventTextInput, synthTextInputEvent("two"), 50, 150)
	if focusedTextInput() || !focusedTextArea() {
		t.Fatalf("after focusing TextArea: textinput.Focused=%v textarea.Focused=%v, want false/true",
			focusedTextInput(), focusedTextArea())
	}

	dispatch(guiEventMouseDown, synthMouseDown(), 50, 36) // click TextInput again
	dispatch(guiEventTextInput, synthTextInputEvent("three"), 50, 36)
	if !focusedTextInput() || focusedTextArea() {
		t.Fatalf("after re-focusing TextInput: textinput.Focused=%v textarea.Focused=%v, want true/false",
			focusedTextInput(), focusedTextArea())
	}

	if got := textInputValue(t, textinput); got != "onethree" {
		t.Fatalf("expected TextInput value %q, got %q", "onethree", got)
	}
	if got := goStringFromC(gui_textarea_get_text(textarea)); got != "two" {
		t.Fatalf("expected TextArea value %q, got %q", "two", got)
	}
}

func textInputFocused(t *testing.T, handle uint64) bool {
	t.Helper()
	v, ok := handles.get(handle)
	if !ok {
		t.Fatalf("unknown TextInput handle %d", handle)
	}
	return v.(*widgets.TextInput).Focused
}
