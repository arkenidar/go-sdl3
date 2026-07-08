"""Widget/container classes built on top of pygosdl._native. Every widget
shares Widget's update()/render()/destroy() so a Window.run() call can
treat leaves and containers uniformly.
"""
import ctypes as ct
import sys
import traceback

from . import _native as _n
from ._native import check


def _report_callback_error(name: str) -> None:
    # A callback invoked from C must never let an exception escape back
    # across the C boundary -- print and swallow it instead, matching the
    # convention documented in docs/CFFI.md.
    print(f"pygosdl: exception in {name} callback:", file=sys.stderr)
    traceback.print_exc()


class Widget:
    def __init__(self, handle: int):
        self.handle = handle

    def update(self) -> bool:
        handled = ct.c_int32()
        check(_n.gui.gui_widget_update(self.handle, ct.byref(handled)), "gui_widget_update")
        return handled.value != 0

    def render(self, renderer) -> None:
        check(_n.gui.gui_widget_render(self.handle, renderer.handle), "gui_widget_render")

    def destroy(self) -> None:
        check(_n.gui.gui_widget_destroy(self.handle), "gui_widget_destroy")

    def get_bounds(self):
        rect = _n.GuiRect()
        check(_n.gui.gui_widget_get_bounds(self.handle, ct.byref(rect)), "gui_widget_get_bounds")
        return (rect.x, rect.y, rect.w, rect.h)

    def set_bounds(self, x: float, y: float, w: float, h: float) -> None:
        check(_n.gui.gui_widget_set_bounds(self.handle, _n.GuiRect(x, y, w, h)), "gui_widget_set_bounds")


class Label(Widget):
    def __init__(self, renderer, font, x: float, y: float, text: str):
        handle = ct.c_uint64()
        check(
            _n.gui.gui_label_new(renderer.handle, font.handle, x, y, text.encode(), ct.byref(handle)),
            "gui_label_new",
        )
        super().__init__(handle.value)

    def set_text(self, text: str) -> None:
        check(_n.gui.gui_label_set_text(self.handle, text.encode()), "gui_label_set_text")


class Button(Widget):
    def __init__(self, renderer, font, x: float, y: float, w: float, h: float, text: str):
        handle = ct.c_uint64()
        check(
            _n.gui.gui_button_new(renderer.handle, font.handle, x, y, w, h, text.encode(), ct.byref(handle)),
            "gui_button_new",
        )
        super().__init__(handle.value)
        self._on_click = None
        self._on_click_cb = None

    def was_clicked(self) -> bool:
        flag = ct.c_int32()
        check(_n.gui.gui_button_was_clicked(self.handle, ct.byref(flag)), "gui_button_was_clicked")
        return flag.value != 0

    @property
    def on_click(self):
        return self._on_click

    @on_click.setter
    def on_click(self, callback) -> None:
        self._on_click = callback

        def _trampoline(handle, userdata):
            try:
                callback()
            except Exception:
                _report_callback_error("Button.on_click")

        # Kept referenced on self: ctypes has no way to know Go is holding
        # this function pointer, so letting Python garbage-collect it
        # would be a use-after-free from Go's side. See docs/CFFI.md
        # "Callback lifetime".
        self._on_click_cb = _n.GuiClickCallback(_trampoline)
        check(_n.gui.gui_button_set_onclick(self.handle, self._on_click_cb, None), "gui_button_set_onclick")


class Checkbox(Widget):
    def __init__(self, renderer, font, x: float, y: float, text: str, checked: bool = False):
        handle = ct.c_uint64()
        check(
            _n.gui.gui_checkbox_new(
                renderer.handle, font.handle, x, y, text.encode(), 1 if checked else 0, ct.byref(handle)
            ),
            "gui_checkbox_new",
        )
        super().__init__(handle.value)
        self._on_toggle = None
        self._on_toggle_cb = None

    @property
    def checked(self) -> bool:
        flag = ct.c_int32()
        check(_n.gui.gui_checkbox_get_checked(self.handle, ct.byref(flag)), "gui_checkbox_get_checked")
        return flag.value != 0

    def was_toggled(self) -> bool:
        flag = ct.c_int32()
        check(_n.gui.gui_checkbox_was_toggled(self.handle, ct.byref(flag)), "gui_checkbox_was_toggled")
        return flag.value != 0

    @property
    def on_toggle(self):
        return self._on_toggle

    @on_toggle.setter
    def on_toggle(self, callback) -> None:
        self._on_toggle = callback

        def _trampoline(handle, checked, userdata):
            try:
                callback(checked != 0)
            except Exception:
                _report_callback_error("Checkbox.on_toggle")

        self._on_toggle_cb = _n.GuiToggleCallback(_trampoline)
        check(_n.gui.gui_checkbox_set_ontoggle(self.handle, self._on_toggle_cb, None), "gui_checkbox_set_ontoggle")


class TextInput(Widget):
    def __init__(self, renderer, font, window, x: float, y: float, w: float, h: float):
        handle = ct.c_uint64()
        check(
            _n.gui.gui_textinput_new(renderer.handle, font.handle, window.handle, x, y, w, h, ct.byref(handle)),
            "gui_textinput_new",
        )
        super().__init__(handle.value)
        self._on_submit = None
        self._on_submit_cb = None

    @property
    def value(self) -> str:
        return _n.take_string(_n.gui.gui_textinput_get_value(self.handle))

    @value.setter
    def value(self, text: str) -> None:
        check(_n.gui.gui_textinput_set_value(self.handle, text.encode()), "gui_textinput_set_value")

    def was_submitted(self) -> bool:
        flag = ct.c_int32()
        check(_n.gui.gui_textinput_was_submitted(self.handle, ct.byref(flag)), "gui_textinput_was_submitted")
        return flag.value != 0

    def focus(self) -> None:
        check(_n.gui.gui_textinput_focus(self.handle), "gui_textinput_focus")

    def blur(self) -> None:
        check(_n.gui.gui_textinput_blur(self.handle), "gui_textinput_blur")

    @property
    def on_submit(self):
        return self._on_submit

    @on_submit.setter
    def on_submit(self, callback) -> None:
        self._on_submit = callback

        def _trampoline(handle, value, userdata):
            try:
                callback(value.decode() if value else "")
            except Exception:
                _report_callback_error("TextInput.on_submit")

        self._on_submit_cb = _n.GuiSubmitCallback(_trampoline)
        check(_n.gui.gui_textinput_set_onsubmit(self.handle, self._on_submit_cb, None), "gui_textinput_set_onsubmit")


class TextArea(Widget):
    def __init__(self, renderer, font, window, x: float, y: float, w: float, h: float):
        handle = ct.c_uint64()
        check(
            _n.gui.gui_textarea_new(renderer.handle, font.handle, window.handle, x, y, w, h, ct.byref(handle)),
            "gui_textarea_new",
        )
        super().__init__(handle.value)

    @property
    def text(self) -> str:
        return _n.take_string(_n.gui.gui_textarea_get_text(self.handle))

    @text.setter
    def text(self, value: str) -> None:
        check(_n.gui.gui_textarea_set_text(self.handle, value.encode()), "gui_textarea_set_text")

    def set_wordwrap(self, wrap: bool) -> None:
        check(_n.gui.gui_textarea_set_wordwrap(self.handle, 1 if wrap else 0), "gui_textarea_set_wordwrap")

    def focus(self) -> None:
        check(_n.gui.gui_textarea_focus(self.handle), "gui_textarea_focus")

    def blur(self) -> None:
        check(_n.gui.gui_textarea_blur(self.handle), "gui_textarea_blur")


class VStack(Widget):
    def __init__(self, x: float, y: float, spacing: float):
        handle = ct.c_uint64()
        check(_n.gui.gui_vstack_new(x, y, spacing, ct.byref(handle)), "gui_vstack_new")
        super().__init__(handle.value)

    def add(self, widget: Widget) -> Widget:
        check(_n.gui.gui_vstack_add(self.handle, widget.handle), "gui_vstack_add")
        return widget

    @property
    def content_size(self):
        w, h = ct.c_float(), ct.c_float()
        check(_n.gui.gui_vstack_content_size(self.handle, ct.byref(w), ct.byref(h)), "gui_vstack_content_size")
        return (w.value, h.value)


class HLayout(Widget):
    def __init__(self, x: float, y: float, spacing: float):
        handle = ct.c_uint64()
        check(_n.gui.gui_hlayout_new(x, y, spacing, ct.byref(handle)), "gui_hlayout_new")
        super().__init__(handle.value)

    def add(self, widget: Widget) -> Widget:
        check(_n.gui.gui_hlayout_add(self.handle, widget.handle), "gui_hlayout_add")
        return widget


class Table(Widget):
    def __init__(self, x: float, y: float, col_spacing: float, row_spacing: float):
        handle = ct.c_uint64()
        check(_n.gui.gui_table_new(x, y, col_spacing, row_spacing, ct.byref(handle)), "gui_table_new")
        super().__init__(handle.value)

    def add_row(self, widgets) -> None:
        handles = (ct.c_uint64 * len(widgets))(*(w.handle for w in widgets))
        check(_n.gui.gui_table_add_row(self.handle, handles, len(widgets)), "gui_table_add_row")

    def layout(self) -> None:
        check(_n.gui.gui_table_layout(self.handle), "gui_table_layout")

    def col_width(self, col: int) -> float:
        w = ct.c_float()
        check(_n.gui.gui_table_col_width(self.handle, col, ct.byref(w)), "gui_table_col_width")
        return w.value

    @property
    def content_size(self):
        w, h = ct.c_float(), ct.c_float()
        check(_n.gui.gui_table_content_size(self.handle, ct.byref(w), ct.byref(h)), "gui_table_content_size")
        return (w.value, h.value)
