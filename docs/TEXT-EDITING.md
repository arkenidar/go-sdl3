# Text editing in the widget toolkit

How `TextInput` (single-line) and `TextArea` (multi-line) implement
editing, selection, and clipboard in `internal/widgets/`. Companion to
[DEVSTATE.md](DEVSTATE.md); this file is the reference for the design and
its invariants.

## File map

| File | Role |
|---|---|
| `textinput.go` | Single-line field: focus, caret, selection, rendering |
| `textarea.go` | Multi-line editor: lines, scrolling, word-wrap, selection |
| `editkeys.go` | Shared key handling (`editLineKey`) + shortcut predicates |
| `selection.go` | Pure selection logic: ranges, deletes, hit-testing |
| `clipboard.go` | App-internal clipboard + Ctrl+C/X/V predicates |
| `textdraw.go` | Shared text texture/measure/draw helpers |
| `*_test.go` | Table-driven pure-logic tests (no SDL/display needed) |

## Core invariants

1. **Byte offsets, rune steps.** Every caret/selection position is a byte
   offset into a `string` (`TextInput.cursorCol`, `TextArea`'s
   `textPos{row, col}`). Movement and hit-testing step rune-by-rune with
   `utf8.DecodeRuneInString` / `DecodeLastRuneInString`; an offset must
   never land inside a multibyte rune.
2. **Pure logic stays pure.** Everything in `selection.go` and
   `editkeys.go` (except the `sdl.KeyboardEvent` predicates) is free of
   SDL calls, so tests run headless. Pixel measurement is injected as a
   `func(string) float32` (`byteOffsetForX`); widgets pass
   `textPixelWidth`-based closures.
3. **Widgets own their focus.** `Blur()` never calls `SDL_StopTextInput`
   (window-level state shared by all widgets — see the comment on `Blur`
   and `cffi/focus_handoff_test.go`). Insertion is gated on each widget's
   own `Focused` flag.
4. **Event dispatch is app-driven.** Apps forward events with
   `Update(event, mx, my)`. Drag selection needs the app loop to forward
   `EventMouseMotion` and `EventMouseButtonUp` (text-playground,
   form-demo, crud-app do). A widget must tolerate never receiving motion
   — selection then degrades to click-to-caret.

## Selection model

- State per widget: an **anchor** (where the selection started:
  `TextInput.selAnchor int`, `TextArea.selAnchor textPos`), the caret as
  the moving end, `hasSel`, and `selecting` (mouse drag in progress).
- The anchor/caret pair is unordered; normalize with `orderInts` /
  `orderPos` at the point of use. `hasSel` is true only when anchor ≠
  caret.
- Mouse: button-down inside the widget sets caret+anchor
  (`byteOffsetForX` for TextInput, `posForPoint` for TextArea — the
  latter walks scroll offset, `lineSkip`, and in WordWrap mode the same
  `textutil.WrapText` sublines Render uses). Motion while `selecting`
  moves the caret; button-up ends the drag. TextArea's scrollbar-drag
  branches run first in `Update`, so scrollbar drags never start a
  selection. Touch works through SDL3's synthesized mouse events.
- Keyboard: Shift + movement key (Left/Right/Home/End, plus Up/Down in
  TextArea) plants the anchor if needed and extends; after the move,
  `hasSel = shift && anchor != caret`. Unshifted Left/Right on an active
  selection collapse to its low/high end without moving further;
  Home/End (and Up/Down) collapse and then move normally. Ctrl+A selects
  all.
- Editing with a selection active: typing (`EventTextInput`), Backspace,
  Delete, Return, and paste all delete the selection first
  (`deleteSelection`, backed by `deleteRangeLine` / `deleteRangeLines`),
  then apply. In TextArea this lives at the top of `handleKey` and
  `insertAtCursor`.
- Rendering: highlight rects (RGB 179,212,255) are drawn after the
  background, before the text. TextArea computes per-visual-line spans
  with `lineSelSpan` clipped to each wrapped subline; drawing is clipped
  to widget bounds.

## Key handling flow

`TextInput.Update` / `TextArea.Update` on `EventKeyDown`:

1. Ctrl+C / Ctrl+X → copy/cut selection to the app clipboard (below).
2. Ctrl+V → delete selection, insert `clipboardRead()` (TextInput
   flattens newlines to spaces; TextArea splits into real lines,
   normalizing CRLF/CR).
3. Ctrl+A → select all.
4. Selection-aware pre-step (collapse/delete as described above).
5. `editLineKey(scancode, line, col)` — the shared intra-line handler
   for Backspace/Delete/Left/Right/Home/End. It reports *unhandled* for
   line-crossing cases (Backspace at col 0, Right at end of line, …) so
   TextArea's `applyKey` switch can join/hop lines; it also handles
   Return, Up, Down there.

`editLineKey`'s signature deliberately stays modifier-free; shift logic
wraps it at the widget level.

## Clipboard

Upstream `purego-sdl3` exposes `GetClipboardText` but **not**
`SetClipboardText` (commented out upstream), so the OS clipboard is
read-only from this app. `clipboard.go` therefore keeps a package-level
`appClipboard` shared by all widgets:

- `clipboardWrite(s)` stores the text and snapshots the OS clipboard's
  current content.
- `clipboardRead()` → `pickClipboard(osNow, osAtCopy, app)`: the OS text
  wins when it changed since the last in-app copy (user copied in another
  application) or when nothing was ever copied in-app; otherwise the
  internal text wins.
- Consequence: in-app copies can't be pasted into other applications
  until `SDL_SetClipboardText` is available (fork or upstream PR).

## Extending safely

- New pure logic goes in `selection.go`/`editkeys.go` with table-driven
  tests beside the existing ones; keep the width-function injection
  pattern for anything measuring text.
- Anything reading widget state that callers may replace wholesale
  (`Value`, `Lines`) must clamp first: `TextInput.clampCursor` also
  clamps the anchor; `TextArea.clampPos` snaps a `textPos` into `Lines`.
- Known gaps (see DEVSTATE.md next steps): Ctrl+arrow word-jump,
  double/triple-click selection, auto-scroll while drag-selecting,
  scroll-to-caret, caret hidden during selection.

## Manual test bed

`go run ./cmd/text-playground` — TextArea with word-wrap toggle and
scrollbars, plus keyboard/mouse forwarding for every relevant event type.
`cmd/form-demo` and `cmd/crud-app` exercise TextInput.
