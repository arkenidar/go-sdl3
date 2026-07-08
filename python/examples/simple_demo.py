#!/usr/bin/env python3
"""Same scenario as examples/go-use, examples/lua-use, and
examples/python-use/demo.py (a VStack of Label/Button/Checkbox/TextInput,
a Table, and a TextArea) -- but built with pygosdl's high-level API
instead of raw ctypes, to show the difference in ergonomics.

This is a plain usage example, not a fourth "parity demo": it isn't kept
in lockstep with the other three, and it's free to look however makes
pygosdl look easiest to use.

Build cffi/libgui once, then run from anywhere:
  ./cffi/build.sh
  python3 python/examples/simple_demo.py
"""
import pathlib
import sys

REPO_ROOT = pathlib.Path(__file__).resolve().parents[2]
sys.path.insert(0, str(REPO_ROOT / "python"))

import pygosdl as gui  # noqa: E402


def main():
    with gui.App() as app:
        window = app.create_window("pygosdl demo", 420, 460)
        font = window.load_font(REPO_ROOT / "assets" / "OpenDyslexic-Regular.ttf", 20)
        renderer = window.renderer

        clicks = 0
        label = gui.Label(renderer, font, 0, 0, "Clicks: 0")
        button = gui.Button(renderer, font, 0, 0, 0, 0, "Click me")

        def on_click():
            nonlocal clicks
            clicks += 1
            label.set_text(f"Clicks: {clicks}")
            print("button clicked, total:", clicks)

        button.on_click = on_click

        checkbox = gui.Checkbox(renderer, font, 0, 0, "Enable feature")
        checkbox.on_toggle = lambda checked: print("checkbox toggled:", checked)

        textinput = gui.TextInput(renderer, font, window, 0, 0, 200, 32)

        def on_submit(value):
            print("text submitted:", value)
            textinput.value = ""

        textinput.on_submit = on_submit

        stack = gui.VStack(20, 20, 10)
        stack.add(label)
        stack.add(button)
        stack.add(checkbox)
        stack.add(textinput)

        _, stack_h = stack.content_size
        table_y = 20 + stack_h + 20

        table = gui.Table(20, table_y, 20, 8)
        table.add_row([gui.Label(renderer, font, 0, 0, "Name:"), gui.Label(renderer, font, 0, 0, "Dario")])
        table.add_row([gui.Label(renderer, font, 0, 0, "Age:"), gui.Label(renderer, font, 0, 0, "39")])
        table.layout()

        _, table_h = table.content_size
        textarea_y = table_y + table_h + 20

        textarea = gui.TextArea(renderer, font, window, 20, textarea_y, 360, 100)

        window.run([stack, table, textarea])

    print("done, total clicks:", clicks)


if __name__ == "__main__":
    main()
