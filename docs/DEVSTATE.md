# Development state

Snapshot of where development stands, so work can be resumed and referred
to later. Last updated: 2026-07-08.

> Update note: the phase-1/2 text-editing work described below is now
> committed (`809eb2b`, `5801fc1`). A follow-up fix landed for the FFI
> demos: queued text-input events used to keep SDL's temporary `char*`
> alive past its lifetime, garbling typed text (most visibly multibyte) in
> the Lua/Python demos — `queuedEvent.retainText` (cffi/events.go) now
> copies the string into Go memory and repoints the stored event, and the
> C-visible `GuiEvent.text` truncation is rune-safe (`cStringLen`).
> Remember to rebuild `cffi/libgui.so` after pulling.

## Committed baseline (branch `main`, up to `c0992d4`)

- **Apps** (`cmd/`): blank-window, bouncing-balls, counter-demo, crud-app,
  form-demo, text-playground, and a launcher that lists and starts the
  others (with Windows `.exe` handling fixed in `c579db7`).
- **Widget toolkit** (`internal/widgets/`): Label, Button, Checkbox,
  TextInput, TextArea, VStack, Table, Layout. Shared `Widget` interface:
  `Update(event, mx, my) bool` + `Render(renderer)`.
- **C-ABI FFI layer** (`cffi/`): `libgui` shared library over the widget
  toolkit, with Go, LuaJIT (`examples/lua-use`), and Python ctypes
  (`examples/python-use`) demos, including opt-in native callbacks
  (Phase 3). Reference in [CFFI.md](CFFI.md).
- **CI**: GitHub Actions workflow building/testing on Linux + Windows
  (`7d4ab22`).
- **Docs**: app catalog with screenshots ([index.html](index.html) /
  [index.md](index.md)), architecture diagrams.

## In progress — uncommitted working tree

Theme: **full text-editing UX for TextInput and TextArea** (selection,
clipboard). All of it builds, `go vet` is clean, and `go test ./...`
passes. Design/invariants reference: [TEXT-EDITING.md](TEXT-EDITING.md).
Plan file: `~/.claude/plans/selection-is-missing-generic-music.md`.

### Phase 1 (done, uncommitted): text selection

- New `internal/widgets/selection.go` — pure, SDL-free selection logic:
  `textPos` (row + byte offset), range normalization (`orderInts`,
  `orderPos`), `deleteRangeLine`/`deleteRangeLines`, `extractRangeLines`,
  `lineSelSpan` (per-line highlight spans), and `byteOffsetForX`
  (pixel → nearest rune boundary, width function injected for testability).
- New `internal/widgets/editkeys.go` — shared single-line key handling
  (`editLineKey`), plus shortcut predicates (`isPasteShortcut`,
  `isSelectAllShortcut`, `isMovementKey`). Extracted from the widgets so
  Home/End/arrows/Backspace/Delete behave identically in both.
- New `internal/widgets/textdraw.go` — shared text texture/measure/draw
  helpers (`makeTextTexture`, `textPixelWidth`, `drawText`) used by all
  text-bearing widgets.
- **TextInput / TextArea** now support:
  - Click positions the caret (`byteOffsetForX` / `posForPoint`); drag
    extends a selection; touch works via SDL3's synthesized mouse events.
  - Shift+Left/Right/Home/End (and Up/Down in TextArea) extend the
    selection; unshifted Left/Right collapse it to an end; Ctrl+A selects
    all.
  - Typing, Backspace, Delete, Return, and paste replace the selection.
  - Light-blue highlight rendered behind the text; wrap-aware and clipped
    in TextArea; scrollbar drags keep priority over text selection.
- Conventions: **all offsets are byte offsets**, stepped by rune with
  `utf8.DecodeRuneInString` — never split a multibyte rune. Selection
  logic stays pure so tests run without a window/display.
- App event loops in `cmd/form-demo` and `cmd/crud-app` now forward
  mouse-motion/up events (required for drag selection; text-playground
  already did).
- One cffi test updated: `cffi/focus_handoff_test.go` assumed clicks don't
  move the caret; it now clicks past the text end.

### Phase 2 (done, uncommitted): copy/cut via app-internal clipboard

Upstream `purego-sdl3` exposes `GetClipboardText` but has
`SetClipboardText` commented out, so the OS clipboard can be read but not
written. Chosen workaround (user decision): an **app-internal clipboard**.

- New `internal/widgets/clipboard.go`: package-level `appClipboard` shared
  by all widgets; `isCopyShortcut`/`isCutShortcut` (Ctrl+C/X);
  `clipboardWrite` snapshots the OS clipboard text at copy time;
  `clipboardRead`/`pickClipboard` prefer the OS clipboard **only if its
  text changed since the last in-app copy** (i.e. the user copied
  something in another application), otherwise the internal one.
- Known limitation: text copied in-app cannot be pasted into other
  applications. Lifting this requires enabling `SDL_SetClipboardText`
  upstream (or a local fork + `go.mod` replace).

### Tests

Table-driven, pure-logic, no SDL init needed (pattern to keep):

- `editkeys_test.go` — line editing keys, insertAtCursor (multi-line
  paste, CRLF normalization).
- `selection_test.go` — range deletes (incl. multibyte), `lineSelSpan`,
  `byteOffsetForX` with a fake width fn, paste-replaces-selection,
  TextInput deleteSelection.
- `clipboard_test.go` — `extractRangeLines`, `pickClipboard` preference
  logic.

## Next steps / open items

- **Commit the working tree** (selection + clipboard work is complete and
  green; `examples/lua-use/demo.lua` also has uncommitted edits from the
  cffi work — review before committing).
- Possible follow-ups, not started:
  - Real OS clipboard write (fork purego-sdl3 or upstream a PR enabling
    `SDL_SetClipboardText`).
  - Word-jump (Ctrl+arrows) and double-click word / triple-click line
    selection.
  - Auto-scroll TextArea while drag-selecting beyond the viewport, and
    scroll-to-caret on keyboard movement.
  - Hide the blinking caret while a selection is active (cosmetic).

## Build / test quickstart

```sh
go build ./... && go vet ./... && go test ./...
go run ./cmd/text-playground   # main manual test bed for text editing
```

Go toolchain lives at `~/apps/go/bin` (see `~/.profile`).
