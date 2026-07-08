#!/usr/bin/env python3
"""A small CRUD app built with pygosdl's high-level API.

Manages a list of "Person" records (name, age) with a form (Name/Age
text inputs), Add/Update buttons, and a table listing every record
with per-row Edit/Delete buttons. Records and the window size persist
to a JSON file next to this script, loaded on startup and saved after
every change -- including the window's live size at that point (e.g.
after the user dragged an edge), via Window.get_size().

Build cffi/libgui once, then run from anywhere:
  ./cffi/build.sh
  python3 python/examples/crud_demo.py
"""
import json
import pathlib
import sys

REPO_ROOT = pathlib.Path(__file__).resolve().parents[2]
sys.path.insert(0, str(REPO_ROOT / "python"))

import pygosdl as gui  # noqa: E402

DATA_FILE = pathlib.Path(__file__).with_name("crud_demo_data.json")
DEFAULT_WINDOW_SIZE = (520, 640)


class CrudApp:
    def __init__(self, app: gui.App):
        self.records, (width, height) = self._load_state()
        self.window = app.create_window("pygosdl CRUD demo", width, height, resizable=True)
        self.font = self.window.load_font(REPO_ROOT / "assets" / "OpenDyslexic-Regular.ttf", 18)
        self.renderer = self.window.renderer

        self.selected_index = None
        self.table = None
        self.table_rows = []  # widgets owned by the current table, for cleanup
        # Rebuilding the table destroys/replaces widget handles. Doing that
        # synchronously from inside a button callback would corrupt the
        # widget list a still-running dispatch loop is iterating over (see
        # run()), so handlers just set this flag and the actual rebuild
        # happens once the frame's event dispatch has finished.
        self._table_dirty = False

        self._build_form()
        self._build_table()

    @staticmethod
    def _load_state():
        """Returns (records, (width, height)), falling back to defaults for
        anything missing or malformed in the JSON file."""
        records, size = [], DEFAULT_WINDOW_SIZE
        if not DATA_FILE.exists():
            return records, size
        try:
            data = json.loads(DATA_FILE.read_text())
        except (json.JSONDecodeError, OSError):
            return records, size
        if not isinstance(data, dict):
            return records, size

        raw_records = data.get("records", [])
        if isinstance(raw_records, list):
            records = [
                {"name": str(item.get("name", "")), "age": str(item.get("age", ""))}
                for item in raw_records
                if isinstance(item, dict)
            ]

        window = data.get("window", {})
        if isinstance(window, dict):
            width = window.get("width", size[0])
            height = window.get("height", size[1])
            if isinstance(width, int) and isinstance(height, int) and width > 0 and height > 0:
                size = (width, height)

        return records, size

    def _save_records(self):
        width, height = self.window.get_size()
        state = {
            "window": {"width": width, "height": height},
            "records": self.records,
        }
        DATA_FILE.write_text(json.dumps(state, indent=2))

    def _build_form(self):
        # Containers (VStack/HLayout/Table) only accept leaf widgets, not
        # other containers, so the form fields and the buttons live in two
        # separate flat containers rather than nested inside one another.
        renderer, font = self.renderer, self.font

        self.status_label = gui.Label(renderer, font, 20, 20, "Ready")

        name_label = gui.Label(renderer, font, 0, 0, "Name:")
        self.name_input = gui.TextInput(renderer, font, self.window, 0, 0, 240, 32)

        age_label = gui.Label(renderer, font, 0, 0, "Age:")
        self.age_input = gui.TextInput(renderer, font, self.window, 0, 0, 240, 32)

        self.add_button = gui.Button(renderer, font, 0, 0, 0, 0, "Add")
        self.add_button.on_click = self.on_add

        self.update_button = gui.Button(renderer, font, 0, 0, 0, 0, "Update selected")
        self.update_button.on_click = self.on_update

        self.form_fields = gui.Table(20, 60, 12, 8)
        self.form_fields.add_row([name_label, self.name_input])
        self.form_fields.add_row([age_label, self.age_input])
        self.form_fields.layout()

        _, fields_h = self.form_fields.content_size
        buttons_y = 60 + fields_h + 16

        self.buttons = gui.HLayout(20, buttons_y, 10)
        self.buttons.add(self.add_button)
        self.buttons.add(self.update_button)

    def _table_y(self) -> float:
        _, fields_h = self.form_fields.content_size
        buttons_y = 60 + fields_h + 16
        button_h = 32
        return buttons_y + button_h + 20

    def _destroy_table(self):
        if self.table is not None:
            for widget in self.table_rows:
                widget.destroy()
            self.table.destroy()
            self.table = None
            self.table_rows = []

    def _build_table(self):
        self._destroy_table()

        renderer, font = self.renderer, self.font
        table = gui.Table(20, self._table_y(), 16, 8)

        header = [
            gui.Label(renderer, font, 0, 0, "Name"),
            gui.Label(renderer, font, 0, 0, "Age"),
            gui.Label(renderer, font, 0, 0, ""),
            gui.Label(renderer, font, 0, 0, ""),
        ]
        table.add_row(header)
        self.table_rows.extend(header)

        for index, record in enumerate(self.records):
            name_lbl = gui.Label(renderer, font, 0, 0, record["name"])
            age_lbl = gui.Label(renderer, font, 0, 0, record["age"])
            edit_btn = gui.Button(renderer, font, 0, 0, 0, 0, "Edit")
            delete_btn = gui.Button(renderer, font, 0, 0, 0, 0, "Delete")

            edit_btn.on_click = self._make_edit_handler(index)
            delete_btn.on_click = self._make_delete_handler(index)

            row = [name_lbl, age_lbl, edit_btn, delete_btn]
            table.add_row(row)
            self.table_rows.extend(row)

        table.layout()
        self.table = table

    def _make_edit_handler(self, index: int):
        def handler():
            self.select_row(index)
        return handler

    def _make_delete_handler(self, index: int):
        def handler():
            self.delete_row(index)
        return handler

    def select_row(self, index: int):
        if not (0 <= index < len(self.records)):
            return
        record = self.records[index]
        self.name_input.value = record["name"]
        self.age_input.value = record["age"]
        self.selected_index = index
        self.status_label.set_text(f"Editing row {index + 1}")

    def delete_row(self, index: int):
        if not (0 <= index < len(self.records)):
            return
        removed = self.records.pop(index)
        if self.selected_index == index:
            self.on_clear()
        elif self.selected_index is not None and self.selected_index > index:
            self.selected_index -= 1
        self.status_label.set_text(f"Deleted {removed['name']}")
        self._save_records()
        self._table_dirty = True

    def on_add(self):
        name = self.name_input.value.strip()
        age = self.age_input.value.strip()
        if not name:
            self.status_label.set_text("Name is required")
            return
        self.records.append({"name": name, "age": age})
        self.status_label.set_text(f"Added {name}")
        self.on_clear()
        self._save_records()
        self._table_dirty = True

    def on_update(self):
        if self.selected_index is None:
            self.status_label.set_text("Select a row to update (click Edit)")
            return
        name = self.name_input.value.strip()
        age = self.age_input.value.strip()
        if not name:
            self.status_label.set_text("Name is required")
            return
        self.records[self.selected_index] = {"name": name, "age": age}
        self.status_label.set_text(f"Updated row {self.selected_index + 1}")
        self._save_records()
        self._table_dirty = True

    def on_clear(self):
        self.name_input.value = ""
        self.age_input.value = ""
        self.selected_index = None
        self.status_label.set_text("Ready")

    def top_level_widgets(self):
        return [self.status_label, self.form_fields, self.buttons, self.table]

    def run(self):
        window = self.window
        delay_ms = max(1, round(1000 / 60))
        running = True
        while running:
            window.pump_events()
            while True:
                event = window.poll_event()
                if event.type == gui.EventType.NONE:
                    break
                if event.type == gui.EventType.QUIT:
                    running = False
                for widget in self.top_level_widgets():
                    widget.update()

            if self._table_dirty:
                self._build_table()
                self._table_dirty = False

            self.renderer.clear(30, 30, 30, 255)
            for widget in self.top_level_widgets():
                widget.render(self.renderer)
            self.renderer.present()
            gui._native.gui.gui_delay_ms(delay_ms)

        # Save once more on the way out so a plain resize-then-quit (with no
        # record edit in between) still persists the window's final size.
        self._save_records()


def main():
    with gui.App() as app:
        crud = CrudApp(app)
        crud.run()
    print(f"done, {len(crud.records)} record(s) remaining")


if __name__ == "__main__":
    main()
