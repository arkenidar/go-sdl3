package main

/*
#include <stdint.h>

typedef struct {
    int32_t type;
    float mx, my;
    float wheel_x, wheel_y;
    int32_t scancode;
    char text[32];
} GuiEvent;
*/
import "C"

import (
	"encoding/binary"
	"unsafe"

	"github.com/jupiterrider/purego-sdl3/sdl"
)

const (
	guiEventNone = iota
	guiEventQuit
	guiEventMouseDown
	guiEventMouseUp
	guiEventMouseMotion
	guiEventMouseWheel
	guiEventKeyDown
	guiEventTextInput
)

// queuedEvent bundles the SDL event alongside the pointer position at the
// time it was pumped, and the classified GuiEventType for the C struct.
type queuedEvent struct {
	kind  int32
	ev    sdl.Event
	mx    float32
	my    float32
	wheel sdl.MouseWheelEvent
	text  string
	// textBuf owns the text-input string for the queued event's lifetime;
	// see retainText.
	textBuf []byte
}

// sdlTextInputEventTextOffset is where TextInputEvent's `text` pointer sits
// inside the raw SDL event union: CommonEvent (16 bytes) + WindowID uint32
// (4) + 4 bytes padding.
const sdlTextInputEventTextOffset = 24

// retainText copies a text-input event's string into Go-owned memory and
// repoints the stored event's text pointer at that copy. SDL owns the
// original buffer only temporarily (it is freed/reused by later PollEvent /
// PumpEvents calls), and this queue dispatches events well after pumping —
// without the copy, gui_widget_update would hand widgets an event whose
// text pointer already dangles, garbling input (most visibly multibyte).
func (qe *queuedEvent) retainText() {
	qe.textBuf = append([]byte(qe.text), 0)
	ptr := uint64(uintptr(unsafe.Pointer(&qe.textBuf[0])))
	binary.LittleEndian.PutUint64(qe.ev[sdlTextInputEventTextOffset:sdlTextInputEventTextOffset+8], ptr)
}

var eventQueue []queuedEvent

// currentEvent is the event most recently popped by gui_poll_event. Widget
// dispatch (gui_widget_update) always acts on this event rather than having
// the host hand a reconstructed sdl.Event back in — the underlying SDL
// event is a raw union that can't be faithfully rebuilt from the simplified
// GuiEvent C struct, so the real captured sdl.Event is kept Go-side and
// reused for dispatch immediately after each poll.
var currentEvent queuedEvent

func classify(ev sdl.Event) int32 {
	switch ev.Type() {
	case sdl.EventQuit:
		return guiEventQuit
	case sdl.EventMouseButtonDown:
		return guiEventMouseDown
	case sdl.EventMouseButtonUp:
		return guiEventMouseUp
	case sdl.EventMouseMotion:
		return guiEventMouseMotion
	case sdl.EventMouseWheel:
		return guiEventMouseWheel
	case sdl.EventKeyDown:
		return guiEventKeyDown
	case sdl.EventTextInput:
		return guiEventTextInput
	default:
		return guiEventNone
	}
}

//export gui_pump_events
func gui_pump_events(outCount *int32) int32 {
	return guard(func() error {
		var mx, my float32
		sdl.GetMouseState(&mx, &my)

		var ev sdl.Event
		for sdl.PollEvent(&ev) {
			kind := classify(ev)
			if kind == guiEventNone {
				continue
			}
			qe := queuedEvent{kind: kind, ev: ev, mx: mx, my: my}
			if kind == guiEventMouseWheel {
				qe.wheel = ev.Wheel()
			}
			if kind == guiEventTextInput {
				textEv := ev.Text()
				qe.text = textEv.Text()
				qe.retainText()
			}
			if kind == guiEventMouseMotion {
				motion := ev.Motion()
				qe.mx, qe.my = motion.X, motion.Y
			}
			eventQueue = append(eventQueue, qe)
		}
		if outCount != nil {
			*outCount = int32(len(eventQueue))
		}
		return nil
	})
}

//export gui_poll_event
func gui_poll_event(out *C.GuiEvent) int32 {
	return guard(func() error {
		if out == nil {
			return nil
		}
		*out = C.GuiEvent{}
		if len(eventQueue) == 0 {
			return nil
		}
		qe := eventQueue[0]
		eventQueue = eventQueue[1:]
		currentEvent = qe

		out._type = C.int32_t(qe.kind)
		out.mx = C.float(qe.mx)
		out.my = C.float(qe.my)
		switch qe.kind {
		case guiEventMouseWheel:
			out.wheel_x = C.float(qe.wheel.X)
			out.wheel_y = C.float(qe.wheel.Y)
		case guiEventKeyDown:
			out.scancode = C.int32_t(qe.ev.Key().Scancode)
		case guiEventTextInput:
			writeCString(out.text[:], qe.text)
		}
		return nil
	})
}

func writeCString(dst []C.char, s string) {
	b := []byte(s)
	n := cStringLen(b, len(dst)-1)
	for i := 0; i < n; i++ {
		dst[i] = C.char(b[i])
	}
	dst[n] = 0
}

// cStringLen returns how many bytes of b fit in a C buffer holding max
// bytes plus a NUL. When truncating, it backs up to a rune boundary so a
// multibyte UTF-8 sequence is never split mid-rune.
func cStringLen(b []byte, max int) int {
	if len(b) <= max {
		return len(b)
	}
	n := max
	for n > 0 && b[n]&0xC0 == 0x80 {
		n--
	}
	return n
}
