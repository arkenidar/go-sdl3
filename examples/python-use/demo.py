#!/usr/bin/env python3
"""Drives the cffi/libgui.so C ABI from CPython's ctypes, with no Go
toolchain needed at runtime -- just the built .so.

Kept in lockstep with examples/go-use/main.go (native Go) and
examples/lua-use/demo.lua (LuaJIT FFI) -- same widgets, same layout, same
behavior. If you change one, change the other two.

Build the library once into cffi/ (both examples/python-use and
examples/lua-use load that single shared copy -- no per-example
duplicate), then run from anywhere:
  go build -buildmode=c-shared -o cffi/libgui.so ./cffi
  python3 examples/python-use/demo.py
"""
import ctypes as ct
import os
import sys

# Resolve paths relative to this script's own location, not the caller's
# cwd, so `python3 examples/python-use/demo.py` works from the repo root too.
SCRIPT_DIR = os.path.dirname(os.path.abspath(__file__))


class GuiEvent(ct.Structure):
    _fields_ = [
        ("type", ct.c_int32),
        ("mx", ct.c_float),
        ("my", ct.c_float),
        ("wheel_x", ct.c_float),
        ("wheel_y", ct.c_float),
        ("scancode", ct.c_int32),
        ("text", ct.c_char * 32),
    ]


class GuiRect(ct.Structure):
    _fields_ = [("x", ct.c_float), ("y", ct.c_float), ("w", ct.c_float), ("h", ct.c_float)]


GUI_EVENT_QUIT = 1

# cffi/build.sh names the library per-OS: .dll on Windows, .dylib on
# macOS, .so elsewhere.
_LIB_EXT = {"win32": "dll", "darwin": "dylib"}.get(sys.platform, "so")
gui = ct.CDLL(os.path.join(SCRIPT_DIR, "..", "..", "cffi", f"libgui.{_LIB_EXT}"))

gui.gui_abi_version.argtypes = []
gui.gui_abi_version.restype = ct.c_int32

gui.gui_delay_ms.argtypes = [ct.c_uint32]
gui.gui_delay_ms.restype = ct.c_int32

gui.gui_last_error.argtypes = []
gui.gui_last_error.restype = ct.c_char_p

gui.gui_free_string.argtypes = [ct.c_void_p]
gui.gui_free_string.restype = None

gui.gui_init.argtypes = []
gui.gui_init.restype = ct.c_int32

gui.gui_quit.argtypes = []
gui.gui_quit.restype = ct.c_int32

gui.gui_create_window.argtypes = [ct.c_char_p, ct.c_int32, ct.c_int32, ct.c_int32, ct.POINTER(ct.c_uint64)]
gui.gui_create_window.restype = ct.c_int32

gui.gui_window_get_renderer.argtypes = [ct.c_uint64, ct.POINTER(ct.c_uint64)]
gui.gui_window_get_renderer.restype = ct.c_int32

gui.gui_load_font.argtypes = [ct.c_char_p, ct.c_float, ct.POINTER(ct.c_uint64)]
gui.gui_load_font.restype = ct.c_int32

gui.gui_render_clear.argtypes = [ct.c_uint64, ct.c_uint8, ct.c_uint8, ct.c_uint8, ct.c_uint8]
gui.gui_render_clear.restype = ct.c_int32

gui.gui_render_present.argtypes = [ct.c_uint64]
gui.gui_render_present.restype = ct.c_int32

gui.gui_pump_events.argtypes = [ct.POINTER(ct.c_int32)]
gui.gui_pump_events.restype = ct.c_int32

gui.gui_poll_event.argtypes = [ct.POINTER(GuiEvent)]
gui.gui_poll_event.restype = ct.c_int32

gui.gui_label_new.argtypes = [ct.c_uint64, ct.c_uint64, ct.c_float, ct.c_float, ct.c_char_p, ct.POINTER(ct.c_uint64)]
gui.gui_label_new.restype = ct.c_int32

gui.gui_label_set_text.argtypes = [ct.c_uint64, ct.c_char_p]
gui.gui_label_set_text.restype = ct.c_int32

gui.gui_button_new.argtypes = [
    ct.c_uint64, ct.c_uint64, ct.c_float, ct.c_float, ct.c_float, ct.c_float, ct.c_char_p, ct.POINTER(ct.c_uint64),
]
gui.gui_button_new.restype = ct.c_int32

gui.gui_button_was_clicked.argtypes = [ct.c_uint64, ct.POINTER(ct.c_int32)]
gui.gui_button_was_clicked.restype = ct.c_int32

GuiClickCallback = ct.CFUNCTYPE(None, ct.c_uint64, ct.c_void_p)

gui.gui_button_set_onclick.argtypes = [ct.c_uint64, GuiClickCallback, ct.c_void_p]
gui.gui_button_set_onclick.restype = ct.c_int32

gui.gui_checkbox_new.argtypes = [
    ct.c_uint64, ct.c_uint64, ct.c_float, ct.c_float, ct.c_char_p, ct.c_int32, ct.POINTER(ct.c_uint64),
]
gui.gui_checkbox_new.restype = ct.c_int32

gui.gui_checkbox_get_checked.argtypes = [ct.c_uint64, ct.POINTER(ct.c_int32)]
gui.gui_checkbox_get_checked.restype = ct.c_int32

gui.gui_checkbox_was_toggled.argtypes = [ct.c_uint64, ct.POINTER(ct.c_int32)]
gui.gui_checkbox_was_toggled.restype = ct.c_int32

gui.gui_textinput_new.argtypes = [
    ct.c_uint64, ct.c_uint64, ct.c_uint64, ct.c_float, ct.c_float, ct.c_float, ct.c_float, ct.POINTER(ct.c_uint64),
]
gui.gui_textinput_new.restype = ct.c_int32

gui.gui_textinput_was_submitted.argtypes = [ct.c_uint64, ct.POINTER(ct.c_int32)]
gui.gui_textinput_was_submitted.restype = ct.c_int32

gui.gui_textinput_get_value.argtypes = [ct.c_uint64]
# c_void_p (not c_char_p): a c_char_p restype makes ctypes auto-copy the
# string into a Python bytes object, losing the original pointer -- passing
# that copy to gui_free_string would then free memory Go never allocated
# and corrupt the heap. Keep the raw pointer, decode via string_at, then
# free that same pointer.
gui.gui_textinput_get_value.restype = ct.c_void_p

gui.gui_textinput_set_value.argtypes = [ct.c_uint64, ct.c_char_p]
gui.gui_textinput_set_value.restype = ct.c_int32

gui.gui_vstack_new.argtypes = [ct.c_float, ct.c_float, ct.c_float, ct.POINTER(ct.c_uint64)]
gui.gui_vstack_new.restype = ct.c_int32

gui.gui_vstack_add.argtypes = [ct.c_uint64, ct.c_uint64]
gui.gui_vstack_add.restype = ct.c_int32

gui.gui_vstack_content_size.argtypes = [ct.c_uint64, ct.POINTER(ct.c_float), ct.POINTER(ct.c_float)]
gui.gui_vstack_content_size.restype = ct.c_int32

gui.gui_table_new.argtypes = [ct.c_float, ct.c_float, ct.c_float, ct.c_float, ct.POINTER(ct.c_uint64)]
gui.gui_table_new.restype = ct.c_int32

gui.gui_table_add_row.argtypes = [ct.c_uint64, ct.POINTER(ct.c_uint64), ct.c_int32]
gui.gui_table_add_row.restype = ct.c_int32

gui.gui_table_layout.argtypes = [ct.c_uint64]
gui.gui_table_layout.restype = ct.c_int32

gui.gui_table_content_size.argtypes = [ct.c_uint64, ct.POINTER(ct.c_float), ct.POINTER(ct.c_float)]
gui.gui_table_content_size.restype = ct.c_int32

gui.gui_textarea_new.argtypes = [
    ct.c_uint64, ct.c_uint64, ct.c_uint64, ct.c_float, ct.c_float, ct.c_float, ct.c_float, ct.POINTER(ct.c_uint64),
]
gui.gui_textarea_new.restype = ct.c_int32

gui.gui_widget_update.argtypes = [ct.c_uint64, ct.POINTER(ct.c_int32)]
gui.gui_widget_update.restype = ct.c_int32

gui.gui_widget_render.argtypes = [ct.c_uint64, ct.c_uint64]
gui.gui_widget_render.restype = ct.c_int32

gui.gui_widget_destroy.argtypes = [ct.c_uint64]
gui.gui_widget_destroy.restype = ct.c_int32


def check(status, what="gui call"):
    if status != 0:
        raise RuntimeError(f"{what} failed: {gui.gui_last_error().decode()}")


def main():
    print("ABI version:", gui.gui_abi_version())

    check(gui.gui_init(), "gui_init")

    window = ct.c_uint64()
    check(gui.gui_create_window(b"CPython ctypes FFI-parity demo", 420, 460, 0, ct.byref(window)), "gui_create_window")

    renderer = ct.c_uint64()
    check(gui.gui_window_get_renderer(window, ct.byref(renderer)), "gui_window_get_renderer")

    font = ct.c_uint64()
    font_path = os.path.join(SCRIPT_DIR, "..", "..", "assets", "OpenDyslexic-Regular.ttf")
    check(gui.gui_load_font(font_path.encode(), 20, ct.byref(font)), "gui_load_font")

    clicks = 0

    label = ct.c_uint64()
    check(gui.gui_label_new(renderer, font, 0, 0, b"Clicks: 0", ct.byref(label)), "gui_label_new")

    button = ct.c_uint64()
    check(gui.gui_button_new(renderer, font, 0, 0, 0, 0, b"Click me", ct.byref(button)), "gui_button_new")

    # Phase 3: native callback (as opposed to the poll-based checkbox/
    # textinput below). CFUNCTYPE turns this Python closure into a real C
    # function pointer Go can call directly. on_button_click MUST stay
    # referenced (assigned to a variable, never a throwaway expression) for
    # as long as the button exists -- ctypes does not know the C side holds
    # this pointer, so letting Python garbage-collect it would be a
    # use-after-free from Go's perspective.
    #
    # The body is wrapped in try/except: a callback invoked from C should
    # never let an exception escape back across the C boundary. ctypes is
    # forgiving here (it reports the exception via sys.unraisablehook and
    # returns a default value rather than crashing, unlike LuaJIT's FFI,
    # which aborts the whole process) but swallowing it explicitly keeps
    # behavior consistent and avoids relying on that leniency.
    def _on_button_click(handle, userdata):
        try:
            nonlocal clicks
            clicks += 1
            gui.gui_label_set_text(label, f"Clicks: {clicks}".encode())
            print("button clicked (native callback), total:", clicks)
        except Exception as e:
            print("onButtonClick error:", e, file=sys.stderr)

    on_button_click = GuiClickCallback(_on_button_click)
    check(gui.gui_button_set_onclick(button, on_button_click, None), "gui_button_set_onclick")

    checkbox = ct.c_uint64()
    check(
        gui.gui_checkbox_new(renderer, font, 0, 0, b"Enable feature", 0, ct.byref(checkbox)),
        "gui_checkbox_new",
    )

    textinput = ct.c_uint64()
    check(
        gui.gui_textinput_new(renderer, font, window, 0, 0, 200, 32, ct.byref(textinput)),
        "gui_textinput_new",
    )

    stack = ct.c_uint64()
    check(gui.gui_vstack_new(20, 20, 10, ct.byref(stack)), "gui_vstack_new")
    check(gui.gui_vstack_add(stack, label))
    check(gui.gui_vstack_add(stack, button))
    check(gui.gui_vstack_add(stack, checkbox))
    check(gui.gui_vstack_add(stack, textinput))

    stack_w, stack_h = ct.c_float(), ct.c_float()
    check(gui.gui_vstack_content_size(stack, ct.byref(stack_w), ct.byref(stack_h)))
    table_y = 20 + stack_h.value + 20

    name_label, name_value = ct.c_uint64(), ct.c_uint64()
    check(gui.gui_label_new(renderer, font, 0, 0, b"Name:", ct.byref(name_label)))
    check(gui.gui_label_new(renderer, font, 0, 0, b"Alice", ct.byref(name_value)))
    age_label, age_value = ct.c_uint64(), ct.c_uint64()
    check(gui.gui_label_new(renderer, font, 0, 0, b"Age:", ct.byref(age_label)))
    check(gui.gui_label_new(renderer, font, 0, 0, b"30", ct.byref(age_value)))

    table = ct.c_uint64()
    check(gui.gui_table_new(20, table_y, 20, 8, ct.byref(table)), "gui_table_new")
    row1 = (ct.c_uint64 * 2)(name_label.value, name_value.value)
    check(gui.gui_table_add_row(table, row1, 2), "gui_table_add_row")
    row2 = (ct.c_uint64 * 2)(age_label.value, age_value.value)
    check(gui.gui_table_add_row(table, row2, 2), "gui_table_add_row")
    check(gui.gui_table_layout(table), "gui_table_layout")

    table_w, table_h = ct.c_float(), ct.c_float()
    check(gui.gui_table_content_size(table, ct.byref(table_w), ct.byref(table_h)))
    textarea_y = table_y + table_h.value + 20

    textarea = ct.c_uint64()
    check(
        gui.gui_textarea_new(renderer, font, window, 20, textarea_y, 360, 100, ct.byref(textarea)),
        "gui_textarea_new",
    )

    running = True
    count = ct.c_int32()
    ev = GuiEvent()
    handled = ct.c_int32()
    flag = ct.c_int32()

    while running:
        check(gui.gui_pump_events(ct.byref(count)))
        while True:
            check(gui.gui_poll_event(ct.byref(ev)))
            if ev.type == 0:
                break
            if ev.type == GUI_EVENT_QUIT:
                running = False
            # Dispatch to every container unconditionally (not
            # short-circuited): each focus-managing widget (TextInput,
            # TextArea) needs to see every click to know whether to blur
            # itself, even when a different container's widget is the one
            # actually under the click. (An earlier short-circuited version
            # of this loop caused widgets to never learn they'd lost focus,
            # showing multiple simultaneous blinking carets. The actual
            # StartTextInput/StopTextInput race that motivated the
            # short-circuit in the first place is fixed at the source now:
            # Blur() no longer calls SDL_StopTextInput.)
            check(gui.gui_widget_update(stack, ct.byref(handled)))
            check(gui.gui_widget_update(table, ct.byref(handled)))
            check(gui.gui_widget_update(textarea, ct.byref(handled)))

        # Button clicks are handled by the native callback registered above
        # (fires during gui_widget_update, inside the dispatch loop).
        # Checkbox and TextInput below still use the poll-based model, to
        # keep both interaction styles demonstrated side by side.

        check(gui.gui_checkbox_was_toggled(checkbox, ct.byref(flag)))
        if flag.value == 1:
            check(gui.gui_checkbox_get_checked(checkbox, ct.byref(flag)))
            print("checkbox toggled:", flag.value == 1)

        check(gui.gui_textinput_was_submitted(textinput, ct.byref(flag)))
        if flag.value == 1:
            raw = gui.gui_textinput_get_value(textinput)
            print("text submitted:", ct.string_at(raw).decode())
            gui.gui_free_string(raw)
            check(gui.gui_textinput_set_value(textinput, b""))

        check(gui.gui_render_clear(renderer, 30, 30, 30, 255))
        check(gui.gui_widget_render(stack, renderer))
        check(gui.gui_widget_render(table, renderer))
        check(gui.gui_widget_render(textarea, renderer))
        check(gui.gui_render_present(renderer))
        gui.gui_delay_ms(16)  # ~60fps; without this the loop finishes in a blink

    gui.gui_widget_destroy(stack)
    gui.gui_widget_destroy(table)
    gui.gui_widget_destroy(textarea)
    check(gui.gui_quit())
    print("done, total clicks:", clicks)


if __name__ == "__main__":
    main()
