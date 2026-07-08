# cffi — C-ABI FFI layer for internal/widgets

`cffi` exposes this repo's SDL3 widget toolkit (`internal/widgets`) as a flat
C ABI, so it can be driven from any language with a C FFI — CPython
`ctypes`, LuaJIT `ffi`, or similar — without a Go toolchain or cgo on the
*consumer* side. Only opaque `uint64` handles and a couple of small fixed
C structs cross the boundary; no Go types leak through.

Working, side-by-side examples of the same scenario (a `VStack` of a
`Label`/`Button`/`Checkbox`/`TextInput`, a `Table`, and a `TextArea`) exist
in three languages:

- [`examples/go-use`](../examples/go-use) — native Go, no FFI (the reference behavior)
- [`examples/lua-use`](../examples/lua-use) — LuaJIT `ffi`
- [`examples/python-use`](../examples/python-use) — CPython `ctypes`

Read those first if you just want to see it work. This document is the
reference for the ABI itself.

## Building

```sh
./cffi/build.sh
# equivalent to:
#   go build -buildmode=c-shared -o cffi/libgui.so ./cffi
```

Produces `cffi/libgui.so` (`.dylib` on macOS, `.dll` on Windows) plus a
cgo-generated `cffi/libgui.h`. Both are gitignored — rebuild after changing
any `cffi/*.go` file. `CGO_ENABLED=1` and a working C compiler are required
(the `-buildmode=c-shared` requirement, not anything to do with SDL3 itself
— see below).

**Note:** SDL3/SDL3_ttf are *not* linked at build time. `internal/widgets`
calls them via [`purego-sdl3`](https://github.com/jupiterrider/purego-sdl3),
which `dlopen`s the real shared libraries at runtime. So `libgui.so` has no
link-time dependency on SDL3 — but the *process loading it* still needs the
real `libSDL3`/`libSDL3_ttf` installed/discoverable at runtime.

## Core model

### Opaque handles, not pointers

Every object you create — window, renderer, font, widget, container — is
returned as a `uint64_t` handle, not a raw pointer. Go's GC can move/free
objects; a bare Go pointer must never cross into C. Handles are
monotonically increasing and never reused, so a stale handle from a
destroyed widget just returns an error, never silently aliases a different
live object.

### Error convention

Every function that can fail returns an `int32_t` status: `0` on success,
non-zero on error. On error, call `gui_last_error()` for a message. Every
exported function recovers from internal Go panics — one can never unwind
across the cgo boundary — and reports them as an ordinary error instead.

```c
int32_t status = gui_button_new(renderer, font, 10, 10, 100, 32, "Click me", &handle);
if (status != 0) {
    fprintf(stderr, "gui_button_new failed: %s\n", gui_last_error());
}
```

### Strings

- **Borrowed** (most `const char*`/`char*` input parameters): read once
  during the call, ownership stays with the caller.
- **Caller-owned** (`gui_last_error()`, `gui_textinput_get_value()`,
  `gui_textarea_get_text()`): these allocate and hand ownership to you.
  Free with `gui_free_string()`. In ctypes, declare the `restype` as
  `c_void_p` (not `c_char_p`) so you get the raw pointer instead of an
  auto-copied Python `bytes` — `c_char_p` throws away the pointer you'd
  need to free, and freeing a stand-in address corrupts the heap. Decode
  via `ctypes.string_at(ptr)`, then `gui_free_string(ptr)`.

### The frame loop is host-driven

There's no background thread or Go-side event loop. The host pumps SDL
events, dispatches them to widgets, and renders — once per frame, on
whichever thread is running the host's own loop:

```c
gui_init();
gui_create_window("My App", 400, 300, /*resizable=*/0, &window);
gui_window_get_renderer(window, &renderer);
gui_load_font("font.ttf", 20, &font);

// ...create widgets...

int32_t running = 1;
while (running) {
    int32_t count;
    gui_pump_events(&count);

    GuiEvent ev;
    while (1) {
        gui_poll_event(&ev);
        if (ev.type == 0) break;              // GUI_EVENT_NONE: queue empty
        if (ev.type == 1) running = 0;         // GUI_EVENT_QUIT
        int32_t handled;
        gui_widget_update(some_container, &handled);
    }

    gui_render_clear(renderer, 30, 30, 30, 255);
    gui_widget_render(some_container, renderer);
    gui_render_present(renderer);
    gui_delay_ms(16); // ~60fps -- without this the loop finishes in a blink
}
gui_quit();
```

**Dispatch every event to every top-level container, unconditionally —
don't short-circuit once one "handles" it.** Focus-managing widgets
(`TextInput`, `TextArea`) need to see every click to know whether *they*
should blur, even when the click actually landed on a different container's
widget. Short-circuiting broke this in an earlier version of the FFI
demos: one container would consume a click that focused a widget, so a
sibling container's already-focused widget never learned it should blur —
resulting in two simultaneously blinking carets. See the git history on
`examples/lua-use/demo.lua` for a worked example of this exact bug.

### `GuiEvent`

```c
typedef struct {
    int32_t type;      // GuiEventType
    float mx, my;       // pointer position at the time of the event
    float wheel_x, wheel_y;
    int32_t scancode;   // for GUI_EVENT_KEY_DOWN
    char text[32];      // UTF-8 chunk, for GUI_EVENT_TEXT_INPUT
} GuiEvent;
```

`GuiEventType`: `0=NONE 1=QUIT 2=MOUSE_DOWN 3=MOUSE_UP 4=MOUSE_MOTION
5=MOUSE_WHEEL 6=KEY_DOWN 7=TEXT_INPUT`.

### `GuiRect`

```c
typedef struct { float x, y, w, h; } GuiRect;
```

## Two interaction models — pick either, per widget

**Poll-based** (default, no function pointers cross the boundary): after
each `gui_widget_update`, check `gui_button_was_clicked`,
`gui_checkbox_was_toggled`, or `gui_textinput_was_submitted`. Each clears
its own flag on read.

**Native callback** (opt-in): register a real C function pointer —
`ctypes.CFUNCTYPE`, LuaJIT `ffi.cast` — via `gui_button_set_onclick`,
`gui_checkbox_set_ontoggle`, or `gui_textinput_set_onsubmit`. It fires
synchronously, on whatever thread called `gui_widget_update` (always the
host's own frame-loop thread, never a Go goroutine). Both models fire on
the same underlying event, so you can even use both at once, though there's
no reason to.

**Callback lifetime**: the function-pointer object your FFI layer creates
(`ffi.cast` result, `CFUNCTYPE` instance) must be kept referenced for as
long as the widget exists — the FFI runtime doesn't know C is holding that
pointer, and letting it get garbage-collected is a use-after-free from
Go's side. `gui_widget_destroy` cleans up the *registration*, but you're
still responsible for the callback object's own lifetime up to that point.

**LuaJIT-specific pitfall**: a callback invoked from C must never let a Lua
error escape, and — more subtly — must not itself call back into another
FFI function. Doing so aborts the whole process with `PANIC: unprotected
error in call to Lua API (bad callback)`, a failure `pcall` cannot catch
(it happens below the level `pcall` operates at). Keep native callbacks in
LuaJIT minimal — just record that something happened — and do any real
`gui_*` work back in the main loop, outside the callback. See
`examples/lua-use/demo.lua`'s button-click callback for a worked example.
CPython's `ctypes` does not have this restriction.

```c
typedef void (*GuiClickCallback)(uint64_t handle, void* userdata);
typedef void (*GuiToggleCallback)(uint64_t handle, int32_t checked, void* userdata);
typedef void (*GuiSubmitCallback)(uint64_t handle, const char* value, void* userdata);
```

## API reference

Lifecycle:

```c
int32_t gui_abi_version(void);
char*   gui_last_error(void);
void    gui_free_string(char* s);
int32_t gui_init(void);
int32_t gui_quit(void);
int32_t gui_delay_ms(uint32_t ms);
int32_t gui_create_window(char* title, int32_t w, int32_t h, int32_t resizable, uint64_t* out);
int32_t gui_window_get_renderer(uint64_t window, uint64_t* out);
int32_t gui_load_font(char* path, float ptsize, uint64_t* out);
int32_t gui_render_clear(uint64_t renderer, uint8_t r, uint8_t g, uint8_t b, uint8_t a);
int32_t gui_render_present(uint64_t renderer);
```

Events:

```c
int32_t gui_pump_events(int32_t* outCount);
int32_t gui_poll_event(GuiEvent* out);
```

Generic widget/container ops — work on any leaf widget or container handle
(`gui_widget_update`/`gui_widget_render` on all of them; bounds only on
leaf widgets, not containers, since `VStack`/`Layout`/`Table` don't
implement `GetBounds`/`SetBounds`):

```c
int32_t gui_widget_update(uint64_t handle, int32_t* outHandled);
int32_t gui_widget_render(uint64_t handle, uint64_t renderer);
int32_t gui_widget_get_bounds(uint64_t handle, GuiRect* out);
int32_t gui_widget_set_bounds(uint64_t handle, GuiRect rect);
int32_t gui_widget_destroy(uint64_t handle); // idempotent; recurses into containers
```

Label:

```c
int32_t gui_label_new(uint64_t renderer, uint64_t font, float x, float y, char* text, uint64_t* out);
int32_t gui_label_set_text(uint64_t handle, char* text);
```

Button:

```c
int32_t gui_button_new(uint64_t renderer, uint64_t font, float x, float y, float w, float h, char* text, uint64_t* out); // w<=0 or h<=0 autosizes to fit the text
int32_t gui_button_was_clicked(uint64_t handle, int32_t* out);          // poll-based
int32_t gui_button_set_onclick(uint64_t handle, GuiClickCallback cb, void* userdata); // native callback; cb=NULL unregisters
```

Checkbox:

```c
int32_t gui_checkbox_new(uint64_t renderer, uint64_t font, float x, float y, char* text, int32_t checked, uint64_t* out);
int32_t gui_checkbox_get_checked(uint64_t handle, int32_t* out);
int32_t gui_checkbox_was_toggled(uint64_t handle, int32_t* out);         // poll-based
int32_t gui_checkbox_set_ontoggle(uint64_t handle, GuiToggleCallback cb, void* userdata);
```

TextInput (single-line):

```c
int32_t gui_textinput_new(uint64_t renderer, uint64_t font, uint64_t window, float x, float y, float w, float h, uint64_t* out);
int32_t gui_textinput_was_submitted(uint64_t handle, int32_t* out);      // poll-based, fires on Enter
char*   gui_textinput_get_value(uint64_t handle);                        // caller-owned -- gui_free_string when done
int32_t gui_textinput_set_value(uint64_t handle, char* text);
int32_t gui_textinput_focus(uint64_t handle);
int32_t gui_textinput_blur(uint64_t handle);
int32_t gui_textinput_set_onsubmit(uint64_t handle, GuiSubmitCallback cb, void* userdata);
```

TextArea (multi-line):

```c
int32_t gui_textarea_new(uint64_t renderer, uint64_t font, uint64_t window, float x, float y, float w, float h, uint64_t* out);
char*   gui_textarea_get_text(uint64_t handle);                          // caller-owned, lines joined with \n
int32_t gui_textarea_set_text(uint64_t handle, char* text);              // split on \n
int32_t gui_textarea_set_wordwrap(uint64_t handle, int32_t wordWrap);
int32_t gui_textarea_focus(uint64_t handle);
int32_t gui_textarea_blur(uint64_t handle);
```

VStack (vertical layout, reports content size so you can size a window to
fit rather than guessing constants):

```c
int32_t gui_vstack_new(float x, float y, float spacing, uint64_t* out);
int32_t gui_vstack_add(uint64_t stackHandle, uint64_t widgetHandle);
int32_t gui_vstack_content_size(uint64_t handle, float* outW, float* outH);
```

Layout (horizontal row):

```c
int32_t gui_hlayout_new(float x, float y, float spacing, uint64_t* out);
int32_t gui_hlayout_add(uint64_t layoutHandle, uint64_t widgetHandle);
```

Table (grid; column widths/row heights computed from content — call
`gui_table_layout` once after adding all rows, not incrementally):

```c
int32_t gui_table_new(float x, float y, float colSpacing, float rowSpacing, uint64_t* out);
int32_t gui_table_add_row(uint64_t tableHandle, uint64_t* widgetHandles, int32_t n); // one row, n widgets, one per column
int32_t gui_table_layout(uint64_t handle);
int32_t gui_table_col_width(uint64_t handle, int32_t col, float* out);
int32_t gui_table_content_size(uint64_t handle, float* outW, float* outH);
```

## ABI stability

- Never change an existing exported function's signature — add a new
  `_v2` export instead.
- Never expose a Go struct or interface by value — only handles and the
  fixed C structs above.
- `gui_abi_version()` lets a host detect a mismatched header/library pair.

## Testing

```sh
SDL_VIDEODRIVER=dummy go test ./cffi/...
```

The test suite exercises the widget dispatch logic (clicks, focus
handoff, native callbacks) using synthesized `sdl.Event` values, so it runs
headlessly without a real display. It cannot observe genuine
platform-level behavior that depends on a real windowing system (e.g. the
exact `SDL_StartTextInput`/`SDL_StopTextInput` sequence reaching the OS) —
see the comments on `TestFocusHandoffUnconditionalDispatch` and
`TestButtonNativeCallback` for what each test does and doesn't prove.

## What's not here yet

- **`Layout` (horizontal row) has no demo usage** — `go-use`/`lua-use`/
  `python-use` exercise `VStack`, `Table`, and `TextArea` together, but
  nothing calls the `gui_hlayout_*` exports from a demo yet.
- **Checkbox/TextInput native callbacks** have Go-level test coverage but
  aren't wired into the FFI demos (only the poll-based path is, there,
  deliberately, to show both styles) — only the button's native callback
  is demonstrated end-to-end from Lua/Python.
- **macOS is untested** (Homebrew SDL3 package naming unverified).
  **Windows CI** (MSYS2/MinGW64) exists in `.github/workflows/ci.yml` but
  is unverified against a real run.
