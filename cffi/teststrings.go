package main

/*
#include <stdlib.h>
*/
import "C"
import "unsafe"

// goString/cString bridge helpers used by tests, which can't use cgo
// directly (Go disallows `import "C"` in _test.go files).

func testCreateWindow(title string, w, h int32) uint64 {
	c := C.CString(title)
	defer C.free(unsafe.Pointer(c))
	var out uint64
	gui_create_window(c, w, h, 0, &out)
	return out
}

func testLoadFont(path string, ptsize float32) uint64 {
	c := C.CString(path)
	defer C.free(unsafe.Pointer(c))
	var out uint64
	gui_load_font(c, ptsize, &out)
	return out
}

func testButtonNew(renderer, font uint64, x, y, w, h float32, text string) uint64 {
	c := C.CString(text)
	defer C.free(unsafe.Pointer(c))
	var out uint64
	gui_button_new(renderer, font, x, y, w, h, c, &out)
	return out
}

func goStringFromC(s *C.char) string {
	defer C.free(unsafe.Pointer(s))
	return C.GoString(s)
}
