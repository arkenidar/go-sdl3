package main

/*
#include <stdint.h>

typedef void (*GuiClickCallback)(uint64_t handle, void* userdata);

static int32_t test_click_count = 0;
static uint64_t test_click_last_handle = 0;

static void test_click_callback(uint64_t handle, void* userdata) {
    test_click_count++;
    test_click_last_handle = handle;
}

GuiClickCallback test_click_callback_ptr = test_click_callback;

static int32_t test_get_click_count() { return test_click_count; }
static uint64_t test_get_click_last_handle() { return test_click_last_handle; }
static void test_reset_click_callback() { test_click_count = 0; test_click_last_handle = 0; }
*/
import "C"

// testRegisterClickCallback exercises gui_button_set_onclick with a real C
// function pointer (test_click_callback_ptr, defined above), proving the
// cgo trampoline actually invokes native code -- not just Go-side
// bookkeeping. Used only from _test.go files, which can't use cgo directly.

func testRegisterClickCallback(handle uint64) int32 {
	return gui_button_set_onclick(handle, C.test_click_callback_ptr, nil)
}

func testClickCallbackCount() int32 {
	return int32(C.test_get_click_count())
}

func testClickCallbackLastHandle() uint64 {
	return uint64(C.test_get_click_last_handle())
}

func testResetClickCallback() {
	C.test_reset_click_callback()
}
