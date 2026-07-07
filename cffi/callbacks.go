package main

/*
#include <stdint.h>
#include <stdlib.h>

typedef void (*GuiClickCallback)(uint64_t handle, void* userdata);
typedef void (*GuiToggleCallback)(uint64_t handle, int32_t checked, void* userdata);
typedef void (*GuiSubmitCallback)(uint64_t handle, const char* value, void* userdata);

static inline void gui_call_click_cb(GuiClickCallback cb, uint64_t handle, void* userdata) {
    cb(handle, userdata);
}
static inline void gui_call_toggle_cb(GuiToggleCallback cb, uint64_t handle, int32_t checked, void* userdata) {
    cb(handle, checked, userdata);
}
static inline void gui_call_submit_cb(GuiSubmitCallback cb, uint64_t handle, const char* value, void* userdata) {
    cb(handle, value, userdata);
}
*/
import "C"

import (
	"errors"
	"sync"
	"unsafe"

	"arkenidar.com/purego-sdl3/internal/widgets"
)

// Phase 3 opt-in: native function-pointer callbacks, as an alternative to
// the poll-based gui_*_was_clicked/toggled/submitted checks. ctypes'
// CFUNCTYPE and LuaJIT's ffi.cast can both produce a C function pointer to
// pass in here. The callback runs synchronously on whatever thread called
// gui_widget_update -- always the host's own frame-loop thread, never a Go
// goroutine, so there's no cross-thread re-entrancy to worry about.
//
// Registration is optional and independent of the poll-based checks: both
// fire on the same event, so a host can use either (or, unusually, both).

type clickCallback struct {
	cb       C.GuiClickCallback
	userdata unsafe.Pointer
}

type toggleCallback struct {
	cb       C.GuiToggleCallback
	userdata unsafe.Pointer
}

type submitCallback struct {
	cb       C.GuiSubmitCallback
	userdata unsafe.Pointer
}

var (
	callbackMu       sync.Mutex
	buttonCallbacks  = map[uint64]clickCallback{}
	checkboxCallback = map[uint64]toggleCallback{}
	submitCallbacks  = map[uint64]submitCallback{}
)

//export gui_button_set_onclick
func gui_button_set_onclick(handle uint64, cb C.GuiClickCallback, userdata unsafe.Pointer) int32 {
	return guard(func() error {
		v, ok := handles.get(handle)
		if !ok {
			return errors.New("unknown widget handle")
		}
		if _, ok := v.(*widgets.Button); !ok {
			return errors.New("handle is not a Button")
		}
		callbackMu.Lock()
		defer callbackMu.Unlock()
		if cb == nil {
			delete(buttonCallbacks, handle)
		} else {
			buttonCallbacks[handle] = clickCallback{cb: cb, userdata: userdata}
		}
		return nil
	})
}

//export gui_checkbox_set_ontoggle
func gui_checkbox_set_ontoggle(handle uint64, cb C.GuiToggleCallback, userdata unsafe.Pointer) int32 {
	return guard(func() error {
		v, ok := handles.get(handle)
		if !ok {
			return errors.New("unknown widget handle")
		}
		if _, ok := v.(*widgets.Checkbox); !ok {
			return errors.New("handle is not a Checkbox")
		}
		callbackMu.Lock()
		defer callbackMu.Unlock()
		if cb == nil {
			delete(checkboxCallback, handle)
		} else {
			checkboxCallback[handle] = toggleCallback{cb: cb, userdata: userdata}
		}
		return nil
	})
}

//export gui_textinput_set_onsubmit
func gui_textinput_set_onsubmit(handle uint64, cb C.GuiSubmitCallback, userdata unsafe.Pointer) int32 {
	return guard(func() error {
		v, ok := handles.get(handle)
		if !ok {
			return errors.New("unknown widget handle")
		}
		if _, ok := v.(*widgets.TextInput); !ok {
			return errors.New("handle is not a TextInput")
		}
		callbackMu.Lock()
		defer callbackMu.Unlock()
		if cb == nil {
			delete(submitCallbacks, handle)
		} else {
			submitCallbacks[handle] = submitCallback{cb: cb, userdata: userdata}
		}
		return nil
	})
}

func invokeButtonCallback(handle uint64) {
	callbackMu.Lock()
	entry, ok := buttonCallbacks[handle]
	callbackMu.Unlock()
	if ok {
		C.gui_call_click_cb(entry.cb, C.uint64_t(handle), entry.userdata)
	}
}

func invokeCheckboxCallback(handle uint64, checked bool) {
	callbackMu.Lock()
	entry, ok := checkboxCallback[handle]
	callbackMu.Unlock()
	if ok {
		checkedInt := C.int32_t(0)
		if checked {
			checkedInt = 1
		}
		C.gui_call_toggle_cb(entry.cb, C.uint64_t(handle), checkedInt, entry.userdata)
	}
}

func invokeSubmitCallback(handle uint64, value string) {
	callbackMu.Lock()
	entry, ok := submitCallbacks[handle]
	callbackMu.Unlock()
	if ok {
		cValue := C.CString(value)
		defer C.free(unsafe.Pointer(cValue))
		C.gui_call_submit_cb(entry.cb, C.uint64_t(handle), cValue, entry.userdata)
	}
}

func deleteCallbacks(handle uint64) {
	callbackMu.Lock()
	defer callbackMu.Unlock()
	delete(buttonCallbacks, handle)
	delete(checkboxCallback, handle)
	delete(submitCallbacks, handle)
}
