"""App/Window/Font/Renderer/Event -- the pieces needed to open a window and
drive its frame loop. See pygosdl.widgets for the widgets themselves.
"""
import ctypes as ct
from dataclasses import dataclass

from . import _native as _n
from ._native import check


class EventType:
    NONE = _n.EVENT_NONE
    QUIT = _n.EVENT_QUIT
    MOUSE_DOWN = _n.EVENT_MOUSE_DOWN
    MOUSE_UP = _n.EVENT_MOUSE_UP
    MOUSE_MOTION = _n.EVENT_MOUSE_MOTION
    MOUSE_WHEEL = _n.EVENT_MOUSE_WHEEL
    KEY_DOWN = _n.EVENT_KEY_DOWN
    TEXT_INPUT = _n.EVENT_TEXT_INPUT


@dataclass
class Event:
    type: int
    mx: float
    my: float
    wheel_x: float
    wheel_y: float
    scancode: int
    text: str


class Font:
    def __init__(self, handle: int):
        self.handle = handle


class Renderer:
    def __init__(self, handle: int):
        self.handle = handle

    def clear(self, r: int = 30, g: int = 30, b: int = 30, a: int = 255) -> None:
        check(_n.gui.gui_render_clear(self.handle, r, g, b, a), "gui_render_clear")

    def present(self) -> None:
        check(_n.gui.gui_render_present(self.handle), "gui_render_present")


class Window:
    def __init__(self, handle: int, title: str, width: int, height: int):
        self.handle = handle
        self.title = title
        self.width = width
        self.height = height

        renderer_handle = ct.c_uint64()
        check(_n.gui.gui_window_get_renderer(handle, ct.byref(renderer_handle)), "gui_window_get_renderer")
        self.renderer = Renderer(renderer_handle.value)

    def get_size(self):
        """Current window size in screen coordinates -- reflects live
        resizes (e.g. the user dragging an edge), unlike self.width/height
        which are just what was passed to create_window."""
        w, h = ct.c_int32(), ct.c_int32()
        check(_n.gui.gui_window_get_size(self.handle, ct.byref(w), ct.byref(h)), "gui_window_get_size")
        return (w.value, h.value)

    def load_font(self, path, ptsize: float = 20.0) -> Font:
        handle = ct.c_uint64()
        check(_n.gui.gui_load_font(str(path).encode(), ptsize, ct.byref(handle)), "gui_load_font")
        return Font(handle.value)

    def pump_events(self) -> int:
        count = ct.c_int32()
        check(_n.gui.gui_pump_events(ct.byref(count)), "gui_pump_events")
        return count.value

    def poll_event(self) -> Event:
        ev = _n.GuiEvent()
        check(_n.gui.gui_poll_event(ct.byref(ev)), "gui_poll_event")
        return Event(ev.type, ev.mx, ev.my, ev.wheel_x, ev.wheel_y, ev.scancode, ev.text.decode(errors="replace"))

    def run(self, widgets, background=(30, 30, 30, 255), fps: int = 60) -> None:
        """Own the whole frame loop for `widgets` (a list of top-level
        containers/widgets): pump events, dispatch every event to every
        widget unconditionally, clear, render, present, delay -- until a
        quit event arrives.

        Dispatch is deliberately unconditional (not short-circuited):
        focus-managing widgets (TextInput, TextArea) need to see every
        click to know whether *they* should blur, even when the click
        actually landed on a different widget. See docs/CFFI.md's frame
        loop section for the bug this avoids.
        """
        delay_ms = max(1, round(1000 / fps))
        running = True
        while running:
            self.pump_events()
            while True:
                event = self.poll_event()
                if event.type == EventType.NONE:
                    break
                if event.type == EventType.QUIT:
                    running = False
                for widget in widgets:
                    widget.update()

            self.renderer.clear(*background)
            for widget in widgets:
                widget.render(self.renderer)
            self.renderer.present()
            _n.gui.gui_delay_ms(delay_ms)


class App:
    def __init__(self):
        check(_n.gui.gui_init(), "gui_init")

    def __enter__(self) -> "App":
        return self

    def __exit__(self, exc_type, exc, tb) -> bool:
        self.quit()
        return False

    @property
    def abi_version(self) -> int:
        return _n.gui.gui_abi_version()

    def create_window(self, title: str, width: int, height: int, resizable: bool = False) -> Window:
        handle = ct.c_uint64()
        check(
            _n.gui.gui_create_window(title.encode(), width, height, 1 if resizable else 0, ct.byref(handle)),
            "gui_create_window",
        )
        return Window(handle.value, title, width, height)

    def delay_ms(self, ms: int) -> None:
        check(_n.gui.gui_delay_ms(ms), "gui_delay_ms")

    def quit(self) -> None:
        check(_n.gui.gui_quit(), "gui_quit")
