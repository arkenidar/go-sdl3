// Command go-use is the native-Go reference implementation of the same
// scenario built by examples/lua-use (LuaJIT FFI) and examples/python-use
// (CPython ctypes): a VStack (Label, Button, Checkbox, TextInput), a Table
// of key/value Labels below it, and a TextArea at the bottom, wired up with
// poll-based click/toggle/submit checks. Keep all three in lockstep — same
// widgets, same layout, same behavior — so they serve as a working
// comparison of the three call styles onto the same underlying toolkit.
package main

import (
	"fmt"
	"path/filepath"
	"runtime"

	"arkenidar.com/purego-sdl3/internal/sdlapp"
	"arkenidar.com/purego-sdl3/internal/widgets"
	"github.com/jupiterrider/purego-sdl3/sdl"
)

func main() {
	// Resolve font path relative to this source file, so the example
	// works regardless of which directory the user runs it from.
	_, filename, _, _ := runtime.Caller(0)
	fontPath := filepath.Join(filepath.Dir(filename), "..", "..", "assets", "OpenDyslexic-Regular.ttf")

	app, cleanup, err := sdlapp.Bootstrap(sdlapp.Config{
		Title:    "Go native FFI-parity demo",
		Width:    420,
		Height:   460,
		FontPath: fontPath,
		FontSize: 20,
	})
	if err != nil {
		panic(err)
	}
	defer cleanup()

	clicks := 0
	label := widgets.NewLabel(0, 0, "Clicks: 0", app.Font, app.Renderer)
	button := widgets.NewButton(0, 0, 0, 0, "Click me", app.Font, app.Renderer, func() {
		clicks++
		label.UpdateText(fmt.Sprintf("Clicks: %d", clicks))
		fmt.Println("button clicked, total:", clicks)
	})
	checkbox := widgets.NewCheckbox(0, 0, "Enable feature", false, app.Font, app.Renderer, func(checked bool) {
		fmt.Println("checkbox toggled:", checked)
	})
	textInput := widgets.NewTextInput(0, 0, 200, 32, app.Font, app.Renderer, app.Window)
	textInput.OnSubmit = func(value string) {
		fmt.Println("text submitted:", value)
	}

	stack := widgets.NewVStack(20, 20, 10)
	stack.AddWidget(label)
	stack.AddWidget(button)
	stack.AddWidget(checkbox)
	stack.AddWidget(textInput)

	tableY := stack.Y + stack.ContentHeight() + 20
	table := widgets.NewTable(20, tableY, 20, 8)
	table.AddRow(
		widgets.NewLabel(0, 0, "Name:", app.Font, app.Renderer),
		widgets.NewLabel(0, 0, "Alice", app.Font, app.Renderer),
	)
	table.AddRow(
		widgets.NewLabel(0, 0, "Age:", app.Font, app.Renderer),
		widgets.NewLabel(0, 0, "30", app.Font, app.Renderer),
	)
	table.Layout()

	textAreaY := tableY + table.ContentHeight() + 20
	textArea := widgets.NewTextArea(20, textAreaY, 360, 100, app.Font, app.Renderer, app.Window)

	running := true
	handleEvents := func() bool {
		var ev sdl.Event
		var mx, my float32
		sdl.GetMouseState(&mx, &my)
		for sdl.PollEvent(&ev) {
			if ev.Type() == sdl.EventQuit {
				running = false
			}
			// Dispatch to every container unconditionally (not
			// short-circuited): each focus-managing widget (TextInput,
			// TextArea) needs to see every click to know whether to blur
			// itself, even when a different container's widget is the one
			// that's actually under the click.
			stack.Update(ev, mx, my)
			table.Update(ev, mx, my)
			textArea.Update(ev, mx, my)
		}
		return running
	}
	render := func() {
		sdl.SetRenderDrawColor(app.Renderer, 30, 30, 30, sdl.AlphaOpaque)
		sdl.RenderClear(app.Renderer)
		stack.Render(app.Renderer)
		table.Render(app.Renderer)
		textArea.Render(app.Renderer)
		sdl.RenderPresent(app.Renderer)
		sdl.DelayNS(16 * 1_000_000)
	}

	sdlapp.Run(handleEvents, render)
	table.Destroy()
	fmt.Println("done, total clicks:", clicks)
}
