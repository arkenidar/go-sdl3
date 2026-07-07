package main

/*
#include <stdint.h>

typedef struct { float x, y, w, h; } GuiRect;
*/
import "C"

import (
	"errors"

	"arkenidar.com/purego-sdl3/internal/widgets"
	"github.com/jupiterrider/purego-sdl3/sdl"
	"github.com/jupiterrider/purego-sdl3/ttf"
)

// updater/renderable are satisfied by both leaf widgets (widgets.Widget) and
// containers (VStack, Layout, Table), which share Update/Render but don't
// implement GetBounds/SetBounds. Generic dispatch (gui_widget_update/render)
// uses these narrower interfaces so it works on containers too; only
// gui_widget_get_bounds/set_bounds require the full widgets.Widget.
type updater interface {
	Update(event sdl.Event, mx, my float32) bool
}

type renderable interface {
	Render(renderer *sdl.Renderer)
}

func lookupWidget(h uint64) (widgets.Widget, error) {
	v, ok := handles.get(h)
	if !ok {
		return nil, errors.New("unknown widget handle")
	}
	w, ok := v.(widgets.Widget)
	if !ok {
		return nil, errors.New("handle is not a widget")
	}
	return w, nil
}

func lookupUpdater(h uint64) (updater, error) {
	v, ok := handles.get(h)
	if !ok {
		return nil, errors.New("unknown widget handle")
	}
	w, ok := v.(updater)
	if !ok {
		return nil, errors.New("handle does not support Update")
	}
	return w, nil
}

func lookupRenderable(h uint64) (renderable, error) {
	v, ok := handles.get(h)
	if !ok {
		return nil, errors.New("unknown widget handle")
	}
	w, ok := v.(renderable)
	if !ok {
		return nil, errors.New("handle does not support Render")
	}
	return w, nil
}

//export gui_label_new
func gui_label_new(renderer, font uint64, x, y float32, text *C.char, out *uint64) int32 {
	return guard(func() error {
		rend, err := lookupRenderer(renderer)
		if err != nil {
			return err
		}
		f, err := lookupFont(font)
		if err != nil {
			return err
		}
		label := widgets.NewLabel(x, y, C.GoString(text), f, rend)
		if out != nil {
			*out = handles.put(label)
		}
		return nil
	})
}

//export gui_label_set_text
func gui_label_set_text(handle uint64, text *C.char) int32 {
	return guard(func() error {
		v, ok := handles.get(handle)
		if !ok {
			return errors.New("unknown widget handle")
		}
		label, ok := v.(*widgets.Label)
		if !ok {
			return errors.New("handle is not a Label")
		}
		label.UpdateText(C.GoString(text))
		return nil
	})
}

//export gui_button_new
func gui_button_new(renderer, font uint64, x, y, w, h float32, text *C.char, out *uint64) int32 {
	return guard(func() error {
		rend, err := lookupRenderer(renderer)
		if err != nil {
			return err
		}
		f, err := lookupFont(font)
		if err != nil {
			return err
		}
		var handle uint64
		btn := widgets.NewButton(x, y, w, h, C.GoString(text), f, rend, func() {
			buttonClicks[handle] = true
			invokeButtonCallback(handle)
		})
		handle = handles.put(btn)
		if out != nil {
			*out = handle
		}
		return nil
	})
}

// buttonClicks tracks buttons clicked since the last gui_button_was_clicked
// poll — the poll-based interaction model from the FFI design (no native
// callbacks crossing the boundary in phase 1).
var buttonClicks = map[uint64]bool{}

//export gui_button_was_clicked
func gui_button_was_clicked(handle uint64, out *int32) int32 {
	return guard(func() error {
		clicked := buttonClicks[handle]
		delete(buttonClicks, handle)
		if out != nil {
			if clicked {
				*out = 1
			} else {
				*out = 0
			}
		}
		return nil
	})
}

var _ = ttf.Font{} // keep ttf imported for lookupFont's type

//export gui_widget_update
func gui_widget_update(handle uint64, outHandled *int32) int32 {
	return guard(func() error {
		w, err := lookupUpdater(handle)
		if err != nil {
			return err
		}
		handled := w.Update(currentEvent.ev, currentEvent.mx, currentEvent.my)
		if outHandled != nil {
			if handled {
				*outHandled = 1
			} else {
				*outHandled = 0
			}
		}
		return nil
	})
}

//export gui_widget_render
func gui_widget_render(handle, renderer uint64) int32 {
	return guard(func() error {
		w, err := lookupRenderable(handle)
		if err != nil {
			return err
		}
		rend, err := lookupRenderer(renderer)
		if err != nil {
			return err
		}
		w.Render(rend)
		return nil
	})
}

//export gui_widget_get_bounds
func gui_widget_get_bounds(handle uint64, out *C.GuiRect) int32 {
	return guard(func() error {
		w, err := lookupWidget(handle)
		if err != nil {
			return err
		}
		b := w.GetBounds()
		if out != nil {
			out.x = C.float(b.X)
			out.y = C.float(b.Y)
			out.w = C.float(b.W)
			out.h = C.float(b.H)
		}
		return nil
	})
}

//export gui_widget_set_bounds
func gui_widget_set_bounds(handle uint64, rect C.GuiRect) int32 {
	return guard(func() error {
		w, err := lookupWidget(handle)
		if err != nil {
			return err
		}
		w.SetBounds(sdl.FRect{X: float32(rect.x), Y: float32(rect.y), W: float32(rect.w), H: float32(rect.h)})
		return nil
	})
}

//export gui_checkbox_new
func gui_checkbox_new(renderer, font uint64, x, y float32, text *C.char, checked int32, out *uint64) int32 {
	return guard(func() error {
		rend, err := lookupRenderer(renderer)
		if err != nil {
			return err
		}
		f, err := lookupFont(font)
		if err != nil {
			return err
		}
		var handle uint64
		cb := widgets.NewCheckbox(x, y, C.GoString(text), checked != 0, f, rend, func(newChecked bool) {
			checkboxToggles[handle] = true
			invokeCheckboxCallback(handle, newChecked)
		})
		handle = handles.put(cb)
		if out != nil {
			*out = handle
		}
		return nil
	})
}

// checkboxToggles mirrors buttonClicks: poll-based toggle detection so no
// callback function pointer needs to cross the FFI boundary in phase 1/2.
var checkboxToggles = map[uint64]bool{}

//export gui_checkbox_get_checked
func gui_checkbox_get_checked(handle uint64, out *int32) int32 {
	return guard(func() error {
		v, ok := handles.get(handle)
		if !ok {
			return errors.New("unknown widget handle")
		}
		cb, ok := v.(*widgets.Checkbox)
		if !ok {
			return errors.New("handle is not a Checkbox")
		}
		if out != nil {
			if cb.Checked {
				*out = 1
			} else {
				*out = 0
			}
		}
		return nil
	})
}

//export gui_checkbox_was_toggled
func gui_checkbox_was_toggled(handle uint64, out *int32) int32 {
	return guard(func() error {
		toggled := checkboxToggles[handle]
		delete(checkboxToggles, handle)
		if out != nil {
			if toggled {
				*out = 1
			} else {
				*out = 0
			}
		}
		return nil
	})
}

//export gui_textinput_new
func gui_textinput_new(renderer, font, window uint64, x, y, w, h float32, out *uint64) int32 {
	return guard(func() error {
		rend, err := lookupRenderer(renderer)
		if err != nil {
			return err
		}
		f, err := lookupFont(font)
		if err != nil {
			return err
		}
		win, err := lookupWindow(window)
		if err != nil {
			return err
		}
		var handle uint64
		ti := widgets.NewTextInput(x, y, w, h, f, rend, win)
		ti.OnSubmit = func(value string) {
			textInputSubmits[handle] = true
			invokeSubmitCallback(handle, value)
		}
		handle = handles.put(ti)
		if out != nil {
			*out = handle
		}
		return nil
	})
}

// textInputSubmits mirrors buttonClicks/checkboxToggles: poll-based Enter
// detection.
var textInputSubmits = map[uint64]bool{}

//export gui_textinput_was_submitted
func gui_textinput_was_submitted(handle uint64, out *int32) int32 {
	return guard(func() error {
		submitted := textInputSubmits[handle]
		delete(textInputSubmits, handle)
		if out != nil {
			if submitted {
				*out = 1
			} else {
				*out = 0
			}
		}
		return nil
	})
}

//export gui_textinput_get_value
func gui_textinput_get_value(handle uint64) *C.char {
	v, ok := handles.get(handle)
	if !ok {
		return C.CString("")
	}
	ti, ok := v.(*widgets.TextInput)
	if !ok {
		return C.CString("")
	}
	return C.CString(ti.Value)
}

//export gui_textinput_set_value
func gui_textinput_set_value(handle uint64, text *C.char) int32 {
	return guard(func() error {
		v, ok := handles.get(handle)
		if !ok {
			return errors.New("unknown widget handle")
		}
		ti, ok := v.(*widgets.TextInput)
		if !ok {
			return errors.New("handle is not a TextInput")
		}
		ti.Value = C.GoString(text)
		return nil
	})
}

//export gui_textinput_focus
func gui_textinput_focus(handle uint64) int32 {
	return guard(func() error {
		v, ok := handles.get(handle)
		if !ok {
			return errors.New("unknown widget handle")
		}
		ti, ok := v.(*widgets.TextInput)
		if !ok {
			return errors.New("handle is not a TextInput")
		}
		ti.Focus()
		return nil
	})
}

//export gui_textinput_blur
func gui_textinput_blur(handle uint64) int32 {
	return guard(func() error {
		v, ok := handles.get(handle)
		if !ok {
			return errors.New("unknown widget handle")
		}
		ti, ok := v.(*widgets.TextInput)
		if !ok {
			return errors.New("handle is not a TextInput")
		}
		ti.Blur()
		return nil
	})
}

//export gui_vstack_new
func gui_vstack_new(x, y, spacing float32, out *uint64) int32 {
	return guard(func() error {
		stack := widgets.NewVStack(x, y, spacing)
		if out != nil {
			*out = handles.put(stack)
		}
		return nil
	})
}

//export gui_vstack_add
func gui_vstack_add(stackHandle, widgetHandle uint64) int32 {
	return guard(func() error {
		v, ok := handles.get(stackHandle)
		if !ok {
			return errors.New("unknown stack handle")
		}
		stack, ok := v.(*widgets.VStack)
		if !ok {
			return errors.New("handle is not a VStack")
		}
		w, err := lookupWidget(widgetHandle)
		if err != nil {
			return err
		}
		stack.AddWidget(w)
		return nil
	})
}

//export gui_vstack_content_size
func gui_vstack_content_size(handle uint64, outW, outH *float32) int32 {
	return guard(func() error {
		v, ok := handles.get(handle)
		if !ok {
			return errors.New("unknown stack handle")
		}
		stack, ok := v.(*widgets.VStack)
		if !ok {
			return errors.New("handle is not a VStack")
		}
		if outW != nil {
			*outW = stack.ContentWidth()
		}
		if outH != nil {
			*outH = stack.ContentHeight()
		}
		return nil
	})
}

//export gui_widget_destroy
func gui_widget_destroy(handle uint64) int32 {
	return guard(func() error {
		v, ok := handles.get(handle)
		if !ok {
			return nil // already gone; destroy is idempotent
		}
		if d, ok := v.(interface{ Destroy() }); ok {
			d.Destroy()
		}
		handles.delete(handle)
		delete(buttonClicks, handle)
		delete(checkboxToggles, handle)
		delete(textInputSubmits, handle)
		deleteCallbacks(handle)
		return nil
	})
}
