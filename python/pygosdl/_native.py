"""Low-level ctypes bindings for cffi/libgui, the C ABI documented in
docs/CFFI.md. Nothing in this module is meant to be used directly by an
application -- see pygosdl.core / pygosdl.widgets for the ergonomic API
built on top of it.
"""
import ctypes as ct
import pathlib
import sys

REPO_ROOT = pathlib.Path(__file__).resolve().parents[2]


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


EVENT_NONE = 0
EVENT_QUIT = 1
EVENT_MOUSE_DOWN = 2
EVENT_MOUSE_UP = 3
EVENT_MOUSE_MOTION = 4
EVENT_MOUSE_WHEEL = 5
EVENT_KEY_DOWN = 6
EVENT_TEXT_INPUT = 7

GuiClickCallback = ct.CFUNCTYPE(None, ct.c_uint64, ct.c_void_p)
GuiToggleCallback = ct.CFUNCTYPE(None, ct.c_uint64, ct.c_int32, ct.c_void_p)
GuiSubmitCallback = ct.CFUNCTYPE(None, ct.c_uint64, ct.c_char_p, ct.c_void_p)


class GuiError(RuntimeError):
    """Raised whenever a gui_* call returns a non-zero status."""


def _library_extension() -> str:
    if sys.platform.startswith("win"):
        return "dll"
    if sys.platform == "darwin":
        return "dylib"
    return "so"


def _find_library() -> pathlib.Path:
    import os

    ext = _library_extension()
    name = f"libgui.{ext}"

    override = os.environ.get("PYGOSDL_LIBGUI_PATH")
    if override:
        return pathlib.Path(override)

    candidates = [
        REPO_ROOT / "cffi" / name,
        pathlib.Path(__file__).resolve().parent / name,
    ]
    for candidate in candidates:
        if candidate.is_file():
            return candidate

    raise GuiError(
        f"could not find {name} (looked in: {', '.join(str(c) for c in candidates)}). "
        "Build it with `./cffi/build.sh` (or `go build -buildmode=c-shared "
        f"-o cffi/{name} ./cffi`), or point PYGOSDL_LIBGUI_PATH at a built copy."
    )


gui = ct.CDLL(str(_find_library()))

gui.gui_abi_version.argtypes = []
gui.gui_abi_version.restype = ct.c_int32

gui.gui_last_error.argtypes = []
gui.gui_last_error.restype = ct.c_char_p

gui.gui_free_string.argtypes = [ct.c_void_p]
gui.gui_free_string.restype = None

gui.gui_init.argtypes = []
gui.gui_init.restype = ct.c_int32

gui.gui_quit.argtypes = []
gui.gui_quit.restype = ct.c_int32

gui.gui_delay_ms.argtypes = [ct.c_uint32]
gui.gui_delay_ms.restype = ct.c_int32

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

gui.gui_widget_update.argtypes = [ct.c_uint64, ct.POINTER(ct.c_int32)]
gui.gui_widget_update.restype = ct.c_int32

gui.gui_widget_render.argtypes = [ct.c_uint64, ct.c_uint64]
gui.gui_widget_render.restype = ct.c_int32

gui.gui_widget_get_bounds.argtypes = [ct.c_uint64, ct.POINTER(GuiRect)]
gui.gui_widget_get_bounds.restype = ct.c_int32

gui.gui_widget_set_bounds.argtypes = [ct.c_uint64, GuiRect]
gui.gui_widget_set_bounds.restype = ct.c_int32

gui.gui_widget_destroy.argtypes = [ct.c_uint64]
gui.gui_widget_destroy.restype = ct.c_int32

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

gui.gui_checkbox_set_ontoggle.argtypes = [ct.c_uint64, GuiToggleCallback, ct.c_void_p]
gui.gui_checkbox_set_ontoggle.restype = ct.c_int32

gui.gui_textinput_new.argtypes = [
    ct.c_uint64, ct.c_uint64, ct.c_uint64, ct.c_float, ct.c_float, ct.c_float, ct.c_float, ct.POINTER(ct.c_uint64),
]
gui.gui_textinput_new.restype = ct.c_int32

gui.gui_textinput_was_submitted.argtypes = [ct.c_uint64, ct.POINTER(ct.c_int32)]
gui.gui_textinput_was_submitted.restype = ct.c_int32

# c_void_p (not c_char_p): a c_char_p restype makes ctypes auto-copy the
# string into a Python bytes object, losing the original pointer -- passing
# that copy to gui_free_string would then free memory Go never allocated
# and corrupt the heap. Keep the raw pointer, decode via string_at, then
# free that same pointer.
gui.gui_textinput_get_value.argtypes = [ct.c_uint64]
gui.gui_textinput_get_value.restype = ct.c_void_p

gui.gui_textinput_set_value.argtypes = [ct.c_uint64, ct.c_char_p]
gui.gui_textinput_set_value.restype = ct.c_int32

gui.gui_textinput_focus.argtypes = [ct.c_uint64]
gui.gui_textinput_focus.restype = ct.c_int32

gui.gui_textinput_blur.argtypes = [ct.c_uint64]
gui.gui_textinput_blur.restype = ct.c_int32

gui.gui_textinput_set_onsubmit.argtypes = [ct.c_uint64, GuiSubmitCallback, ct.c_void_p]
gui.gui_textinput_set_onsubmit.restype = ct.c_int32

gui.gui_textarea_new.argtypes = [
    ct.c_uint64, ct.c_uint64, ct.c_uint64, ct.c_float, ct.c_float, ct.c_float, ct.c_float, ct.POINTER(ct.c_uint64),
]
gui.gui_textarea_new.restype = ct.c_int32

gui.gui_textarea_get_text.argtypes = [ct.c_uint64]
gui.gui_textarea_get_text.restype = ct.c_void_p  # see gui_textinput_get_value comment above

gui.gui_textarea_set_text.argtypes = [ct.c_uint64, ct.c_char_p]
gui.gui_textarea_set_text.restype = ct.c_int32

gui.gui_textarea_set_wordwrap.argtypes = [ct.c_uint64, ct.c_int32]
gui.gui_textarea_set_wordwrap.restype = ct.c_int32

gui.gui_textarea_focus.argtypes = [ct.c_uint64]
gui.gui_textarea_focus.restype = ct.c_int32

gui.gui_textarea_blur.argtypes = [ct.c_uint64]
gui.gui_textarea_blur.restype = ct.c_int32

gui.gui_vstack_new.argtypes = [ct.c_float, ct.c_float, ct.c_float, ct.POINTER(ct.c_uint64)]
gui.gui_vstack_new.restype = ct.c_int32

gui.gui_vstack_add.argtypes = [ct.c_uint64, ct.c_uint64]
gui.gui_vstack_add.restype = ct.c_int32

gui.gui_vstack_content_size.argtypes = [ct.c_uint64, ct.POINTER(ct.c_float), ct.POINTER(ct.c_float)]
gui.gui_vstack_content_size.restype = ct.c_int32

gui.gui_hlayout_new.argtypes = [ct.c_float, ct.c_float, ct.c_float, ct.POINTER(ct.c_uint64)]
gui.gui_hlayout_new.restype = ct.c_int32

gui.gui_hlayout_add.argtypes = [ct.c_uint64, ct.c_uint64]
gui.gui_hlayout_add.restype = ct.c_int32

gui.gui_table_new.argtypes = [ct.c_float, ct.c_float, ct.c_float, ct.c_float, ct.POINTER(ct.c_uint64)]
gui.gui_table_new.restype = ct.c_int32

gui.gui_table_add_row.argtypes = [ct.c_uint64, ct.POINTER(ct.c_uint64), ct.c_int32]
gui.gui_table_add_row.restype = ct.c_int32

gui.gui_table_layout.argtypes = [ct.c_uint64]
gui.gui_table_layout.restype = ct.c_int32

gui.gui_table_col_width.argtypes = [ct.c_uint64, ct.c_int32, ct.POINTER(ct.c_float)]
gui.gui_table_col_width.restype = ct.c_int32

gui.gui_table_content_size.argtypes = [ct.c_uint64, ct.POINTER(ct.c_float), ct.POINTER(ct.c_float)]
gui.gui_table_content_size.restype = ct.c_int32


def check(status: int, what: str) -> None:
    if status != 0:
        raise GuiError(f"{what} failed: {gui.gui_last_error().decode()}")


def take_string(ptr: int) -> str:
    """Decode + free a caller-owned char* returned by the ABI (e.g.
    gui_textinput_get_value/gui_textarea_get_text)."""
    if not ptr:
        return ""
    text = ct.string_at(ptr).decode()
    gui.gui_free_string(ptr)
    return text
