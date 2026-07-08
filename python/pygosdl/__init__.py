"""pygosdl -- an easy, Pythonic wrapper around this repo's cffi/libgui C
ABI (docs/CFFI.md). Trades strict parity with examples/go-use and
examples/lua-use for ergonomics: plain classes, callback properties, and
a Window.run() that owns the whole frame loop.

Minimal usage:

    import pygosdl as gui

    with gui.App() as app:
        window = app.create_window("Hello", 300, 120)
        font = window.load_font("assets/OpenDyslexic-Regular.ttf", 20)

        label = gui.Label(window.renderer, font, 20, 20, "Clicks: 0")
        button = gui.Button(window.renderer, font, 20, 60, 0, 0, "Click me")

        clicks = 0
        def on_click():
            global clicks
            clicks += 1
            label.set_text(f"Clicks: {clicks}")
        button.on_click = on_click

        stack = gui.VStack(0, 0, 10)
        stack.add(label)
        stack.add(button)

        window.run([stack])
"""
from ._native import GuiError
from .core import App, Event, EventType, Font, Renderer, Window
from .widgets import (
    Button,
    Checkbox,
    HLayout,
    Label,
    Table,
    TextArea,
    TextInput,
    VStack,
    Widget,
)

__all__ = [
    "App",
    "Button",
    "Checkbox",
    "Event",
    "EventType",
    "Font",
    "GuiError",
    "HLayout",
    "Label",
    "Renderer",
    "Table",
    "TextArea",
    "TextInput",
    "VStack",
    "Widget",
    "Window",
]
