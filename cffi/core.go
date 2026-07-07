package main

/*
#include <stdint.h>
#include <stdlib.h>
*/
import "C"

import (
	"errors"
	"unsafe"

	"github.com/jupiterrider/purego-sdl3/sdl"
	"github.com/jupiterrider/purego-sdl3/ttf"
)

//export gui_abi_version
func gui_abi_version() int32 { return 1 }

//export gui_delay_ms
func gui_delay_ms(ms uint32) int32 {
	return guard(func() error {
		sdl.DelayNS(uint64(ms) * 1_000_000)
		return nil
	})
}

//export gui_last_error
func gui_last_error() *C.char {
	return C.CString(getLastError())
}

// gui_free_string releases a *char returned by any function documented as
// caller-owned (gui_last_error, gui_textinput_get_value). Safe to call on
// NULL.
//export gui_free_string
func gui_free_string(s *C.char) {
	if s != nil {
		C.free(unsafe.Pointer(s))
	}
}

//export gui_init
func gui_init() int32 {
	return guard(func() error {
		if !sdl.Init(sdl.InitVideo) {
			return errors.New(sdl.GetError())
		}
		if !ttf.Init() {
			sdl.Quit()
			return errors.New(sdl.GetError())
		}
		return nil
	})
}

//export gui_quit
func gui_quit() int32 {
	return guard(func() error {
		ttf.Quit()
		sdl.Quit()
		handles.mu.Lock()
		handles.objs = make(map[uint64]any)
		handles.mu.Unlock()
		eventQueue = nil
		return nil
	})
}

//export gui_create_window
func gui_create_window(title *C.char, w, h int32, resizable int32, out *uint64) int32 {
	return guard(func() error {
		flags := sdl.WindowFlags(0)
		if resizable != 0 {
			flags |= sdl.WindowResizable
		}
		var window *sdl.Window
		var renderer *sdl.Renderer
		if !sdl.CreateWindowAndRenderer(C.GoString(title), w, h, flags, &window, &renderer) {
			return errors.New(sdl.GetError())
		}
		wh := handles.put(window)
		rh := handles.put(renderer)
		// pack: caller retrieves renderer via gui_window_get_renderer
		windowRenderers[wh] = rh
		if out != nil {
			*out = wh
		}
		return nil
	})
}

// windowRenderers maps a window handle to the renderer handle created
// alongside it by gui_create_window (SDL creates them as a pair).
var windowRenderers = map[uint64]uint64{}

//export gui_window_get_renderer
func gui_window_get_renderer(window uint64, out *uint64) int32 {
	return guard(func() error {
		rh, ok := windowRenderers[window]
		if !ok {
			return errors.New("unknown window handle")
		}
		if out != nil {
			*out = rh
		}
		return nil
	})
}

//export gui_load_font
func gui_load_font(path *C.char, ptsize float32, out *uint64) int32 {
	return guard(func() error {
		font := ttf.OpenFont(C.GoString(path), ptsize)
		if font == nil {
			return errors.New(sdl.GetError())
		}
		if out != nil {
			*out = handles.put(font)
		}
		return nil
	})
}

//export gui_render_clear
func gui_render_clear(renderer uint64, r, g, b, a uint8) int32 {
	return guard(func() error {
		rend, err := lookupRenderer(renderer)
		if err != nil {
			return err
		}
		sdl.SetRenderDrawColor(rend, r, g, b, a)
		sdl.RenderClear(rend)
		return nil
	})
}

//export gui_render_present
func gui_render_present(renderer uint64) int32 {
	return guard(func() error {
		rend, err := lookupRenderer(renderer)
		if err != nil {
			return err
		}
		sdl.RenderPresent(rend)
		return nil
	})
}

func lookupRenderer(h uint64) (*sdl.Renderer, error) {
	v, ok := handles.get(h)
	if !ok {
		return nil, errors.New("unknown renderer handle")
	}
	rend, ok := v.(*sdl.Renderer)
	if !ok {
		return nil, errors.New("handle is not a renderer")
	}
	return rend, nil
}

func lookupWindow(h uint64) (*sdl.Window, error) {
	v, ok := handles.get(h)
	if !ok {
		return nil, errors.New("unknown window handle")
	}
	win, ok := v.(*sdl.Window)
	if !ok {
		return nil, errors.New("handle is not a window")
	}
	return win, nil
}

func lookupFont(h uint64) (*ttf.Font, error) {
	v, ok := handles.get(h)
	if !ok {
		return nil, errors.New("unknown font handle")
	}
	font, ok := v.(*ttf.Font)
	if !ok {
		return nil, errors.New("handle is not a font")
	}
	return font, nil
}

func main() {} // required for -buildmode=c-shared
