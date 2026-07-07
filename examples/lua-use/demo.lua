-- Drives the cffi/libgui.so C ABI from LuaJIT's FFI, with no Go toolchain
-- needed at runtime -- just the built .so.
--
-- Kept in lockstep with examples/go-use/main.go (native Go) and
-- examples/python-use/demo.py (CPython ctypes) -- same widgets, same
-- layout, same behavior. If you change one, change the other two.
--
-- Build the library once into cffi/ (both examples/lua-use and
-- examples/python-use load that single shared copy -- no per-example
-- duplicate), then run from anywhere:
--   go build -buildmode=c-shared -o cffi/libgui.so ./cffi
--   luajit examples/lua-use/demo.lua

local ffi = require("ffi")

-- Resolve paths relative to this script's own location, not the caller's
-- cwd, so `luajit examples/lua-use/demo.lua` works from the repo root too.
local scriptDir = arg[0]:match("(.*/)") or "./"

ffi.cdef[[
typedef struct { int32_t type; float mx, my; float wheel_x, wheel_y; int32_t scancode; char text[32]; } GuiEvent;
typedef struct { float x, y, w, h; } GuiRect;
typedef void (*GuiClickCallback)(uint64_t handle, void* userdata);

int32_t gui_abi_version(void);
int32_t gui_delay_ms(uint32_t ms);
const char* gui_last_error(void);
void gui_free_string(char* s);
int32_t gui_init(void);
int32_t gui_quit(void);
int32_t gui_create_window(const char* title, int32_t w, int32_t h, int32_t resizable, uint64_t* out);
int32_t gui_window_get_renderer(uint64_t window, uint64_t* out);
int32_t gui_load_font(const char* path, float ptsize, uint64_t* out);
int32_t gui_render_clear(uint64_t renderer, uint8_t r, uint8_t g, uint8_t b, uint8_t a);
int32_t gui_render_present(uint64_t renderer);
int32_t gui_pump_events(int32_t* out_count);
int32_t gui_poll_event(GuiEvent* out);

int32_t gui_label_new(uint64_t renderer, uint64_t font, float x, float y, const char* text, uint64_t* out);
int32_t gui_label_set_text(uint64_t handle, const char* text);

int32_t gui_button_new(uint64_t renderer, uint64_t font, float x, float y, float w, float h, const char* text, uint64_t* out);
int32_t gui_button_was_clicked(uint64_t handle, int32_t* out);
int32_t gui_button_set_onclick(uint64_t handle, GuiClickCallback cb, void* userdata);

int32_t gui_checkbox_new(uint64_t renderer, uint64_t font, float x, float y, const char* text, int32_t checked, uint64_t* out);
int32_t gui_checkbox_get_checked(uint64_t handle, int32_t* out);
int32_t gui_checkbox_was_toggled(uint64_t handle, int32_t* out);

int32_t gui_textinput_new(uint64_t renderer, uint64_t font, uint64_t window, float x, float y, float w, float h, uint64_t* out);
int32_t gui_textinput_was_submitted(uint64_t handle, int32_t* out);
char* gui_textinput_get_value(uint64_t handle);
int32_t gui_textinput_set_value(uint64_t handle, const char* text);
int32_t gui_textinput_focus(uint64_t handle);
int32_t gui_textinput_blur(uint64_t handle);

int32_t gui_vstack_new(float x, float y, float spacing, uint64_t* out);
int32_t gui_vstack_add(uint64_t stack, uint64_t widget);
int32_t gui_vstack_content_size(uint64_t handle, float* out_w, float* out_h);

int32_t gui_table_new(float x, float y, float col_spacing, float row_spacing, uint64_t* out);
int32_t gui_table_add_row(uint64_t table, uint64_t* widget_handles, int32_t n);
int32_t gui_table_layout(uint64_t table);
int32_t gui_table_content_size(uint64_t handle, float* out_w, float* out_h);

int32_t gui_textarea_new(uint64_t renderer, uint64_t font, uint64_t window, float x, float y, float w, float h, uint64_t* out);

int32_t gui_widget_update(uint64_t handle, int32_t* out_handled);
int32_t gui_widget_render(uint64_t handle, uint64_t renderer);
int32_t gui_widget_get_bounds(uint64_t handle, GuiRect* out);
int32_t gui_widget_set_bounds(uint64_t handle, GuiRect rect);
int32_t gui_widget_destroy(uint64_t handle);
]]

local gui = ffi.load(scriptDir .. "../../cffi/libgui.so")

local GUI_EVENT_QUIT = 1

local function check(status, what)
    if status ~= 0 then
        error((what or "gui call") .. " failed: " .. ffi.string(gui.gui_last_error()))
    end
end

print("ABI version:", gui.gui_abi_version())

check(gui.gui_init(), "gui_init")

local window = ffi.new("uint64_t[1]")
check(gui.gui_create_window("LuaJIT FFI-parity demo", 420, 460, 0, window), "gui_create_window")
window = window[0]

local renderer = ffi.new("uint64_t[1]")
check(gui.gui_window_get_renderer(window, renderer), "gui_window_get_renderer")
renderer = renderer[0]

local font = ffi.new("uint64_t[1]")
check(gui.gui_load_font(scriptDir .. "../../assets/OpenDyslexic-Regular.ttf", 20, font), "gui_load_font")
font = font[0]

local clicks = 0

local label = ffi.new("uint64_t[1]")
check(gui.gui_label_new(renderer, font, 0, 0, "Clicks: 0", label), "gui_label_new")
label = label[0]

local button = ffi.new("uint64_t[1]")
check(gui.gui_button_new(renderer, font, 0, 0, 0, 0, "Click me", button), "gui_button_new")
button = button[0]

-- Phase 3: native callback (as opposed to the poll-based checkbox/textinput
-- below). ffi.cast turns this Lua closure into a real C function pointer
-- Go can call directly. It MUST be kept alive (assigned to a variable,
-- never re-cast/discarded) for as long as the button exists -- LuaJIT does
-- not know the C side is holding this pointer, so a GC'd callback here
-- would be a use-after-free from Go's perspective. Call :free() once done
-- (after the button is destroyed) to release the callback trampoline slot.
--
-- Deliberately minimal: just bump a plain Lua counter, no FFI calls at
-- all. LuaJIT's FFI callback mechanism can't tolerate calling back into
-- another cdata C function (like gui.gui_label_set_text) from *within* a
-- callback that's itself already a C-into-Lua call -- doing so aborts the
-- whole process with "PANIC: unprotected error in call to Lua API (bad
-- callback)", which pcall cannot catch (the panic happens in the callback
-- trampoline itself, not from an ordinary Lua error). ctypes/CPython does
-- not have this restriction (see examples/python-use/demo.py), so this is
-- a LuaJIT-specific workaround: real work (gui_label_set_text, print) is
-- deferred to the main loop below, safely outside the callback.
local buttonClickCount = 0
local onButtonClick = ffi.cast("GuiClickCallback", function(handle, userdata)
    buttonClickCount = buttonClickCount + 1
end)
check(gui.gui_button_set_onclick(button, onButtonClick, nil), "gui_button_set_onclick")

local checkbox = ffi.new("uint64_t[1]")
check(gui.gui_checkbox_new(renderer, font, 0, 0, "Enable feature", 0, checkbox), "gui_checkbox_new")
checkbox = checkbox[0]

local textinput = ffi.new("uint64_t[1]")
check(gui.gui_textinput_new(renderer, font, window, 0, 0, 200, 32, textinput), "gui_textinput_new")
textinput = textinput[0]

local stack = ffi.new("uint64_t[1]")
check(gui.gui_vstack_new(20, 20, 10, stack), "gui_vstack_new")
stack = stack[0]
check(gui.gui_vstack_add(stack, label))
check(gui.gui_vstack_add(stack, button))
check(gui.gui_vstack_add(stack, checkbox))
check(gui.gui_vstack_add(stack, textinput))

local stackSize = ffi.new("float[2]")
check(gui.gui_vstack_content_size(stack, stackSize + 0, stackSize + 1))
local tableY = 20 + stackSize[1] + 20

local nameLabel = ffi.new("uint64_t[1]")
check(gui.gui_label_new(renderer, font, 0, 0, "Name:", nameLabel))
local nameValue = ffi.new("uint64_t[1]")
check(gui.gui_label_new(renderer, font, 0, 0, "Alice", nameValue))
local ageLabel = ffi.new("uint64_t[1]")
check(gui.gui_label_new(renderer, font, 0, 0, "Age:", ageLabel))
local ageValue = ffi.new("uint64_t[1]")
check(gui.gui_label_new(renderer, font, 0, 0, "30", ageValue))

local table_ = ffi.new("uint64_t[1]")
check(gui.gui_table_new(20, tableY, 20, 8, table_), "gui_table_new")
table_ = table_[0]
local row1 = ffi.new("uint64_t[2]", { nameLabel[0], nameValue[0] })
check(gui.gui_table_add_row(table_, row1, 2), "gui_table_add_row")
local row2 = ffi.new("uint64_t[2]", { ageLabel[0], ageValue[0] })
check(gui.gui_table_add_row(table_, row2, 2), "gui_table_add_row")
check(gui.gui_table_layout(table_), "gui_table_layout")

local tableSize = ffi.new("float[2]")
check(gui.gui_table_content_size(table_, tableSize + 0, tableSize + 1))
local textAreaY = tableY + tableSize[1] + 20

local textarea = ffi.new("uint64_t[1]")
check(gui.gui_textarea_new(renderer, font, window, 20, textAreaY, 360, 100, textarea), "gui_textarea_new")
textarea = textarea[0]

local running = true
local count = ffi.new("int32_t[1]")
local ev = ffi.new("GuiEvent")
local handled = ffi.new("int32_t[1]")
local flag = ffi.new("int32_t[1]")

while running do
    check(gui.gui_pump_events(count))
    while true do
        check(gui.gui_poll_event(ev))
        if ev.type == 0 then break end
        if ev.type == GUI_EVENT_QUIT then running = false end
        -- Dispatch to every container unconditionally (not short-circuited):
        -- each focus-managing widget (TextInput, TextArea) needs to see
        -- every click to know whether to blur itself, even when a
        -- different container's widget is the one actually under the
        -- click. (An earlier short-circuited version of this loop caused
        -- widgets to never learn they'd lost focus, showing multiple
        -- simultaneous blinking carets. The actual StartTextInput/
        -- StopTextInput race that motivated the short-circuit in the first
        -- place is fixed at the source now: Blur() no longer calls
        -- SDL_StopTextInput.)
        check(gui.gui_widget_update(stack, handled))
        check(gui.gui_widget_update(table_, handled))
        check(gui.gui_widget_update(textarea, handled))
    end

    -- Button clicks are detected by the native callback above (it just
    -- bumps buttonClickCount), but the actual FFI work -- updating the
    -- label, printing -- happens here in the main loop, safely outside the
    -- callback. Checkbox and TextInput below still use the poll-based
    -- model, to keep both interaction styles demonstrated side by side.
    if buttonClickCount > 0 then
        clicks = clicks + buttonClickCount
        buttonClickCount = 0
        check(gui.gui_label_set_text(label, "Clicks: " .. clicks))
        print("button clicked (native callback), total:", clicks)
    end

    check(gui.gui_checkbox_was_toggled(checkbox, flag))
    if flag[0] == 1 then
        check(gui.gui_checkbox_get_checked(checkbox, flag))
        print("checkbox toggled:", flag[0] == 1)
    end

    check(gui.gui_textinput_was_submitted(textinput, flag))
    if flag[0] == 1 then
        local valuePtr = gui.gui_textinput_get_value(textinput)
        print("text submitted:", ffi.string(valuePtr))
        gui.gui_free_string(valuePtr)
        check(gui.gui_textinput_set_value(textinput, ""))
    end

    check(gui.gui_render_clear(renderer, 30, 30, 30, 255))
    check(gui.gui_widget_render(stack, renderer))
    check(gui.gui_widget_render(table_, renderer))
    check(gui.gui_widget_render(textarea, renderer))
    check(gui.gui_render_present(renderer))
    gui.gui_delay_ms(16) -- ~60fps; without this the loop finishes in a blink
end

gui.gui_widget_destroy(stack)
gui.gui_widget_destroy(table_)
gui.gui_widget_destroy(textarea)
onButtonClick:free() -- release the callback trampoline slot, after the button (its only user) is gone
check(gui.gui_quit())
print("done, total clicks:", clicks)
