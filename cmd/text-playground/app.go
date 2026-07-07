// app.go
package main

import (
	"fmt"

	"github.com/jupiterrider/purego-sdl3/sdl"
	"github.com/jupiterrider/purego-sdl3/ttf"

	"arkenidar.com/purego-sdl3/internal/widgets"
)

// App demonstrates a scrollable, optionally word-wrapped multi-line
// text editor (internal/widgets.TextArea).
type App struct {
	renderer *sdl.Renderer
	font     *ttf.Font

	windowWidth  float32
	windowHeight float32

	area        *widgets.TextArea
	wrapToggle  *widgets.Button
	statusLabel *widgets.Label

	areaTop float32
}

func NewApp(window *sdl.Window, renderer *sdl.Renderer, font *ttf.Font, windowWidth, windowHeight float32) *App {
	app := &App{
		renderer:     renderer,
		font:         font,
		windowWidth:  windowWidth,
		windowHeight: windowHeight,
	}

	app.wrapToggle = widgets.NewButton(10, 10, 0, 0, "Toggle Word-Wrap", font, renderer, func() {
		app.area.ToggleWordWrap()
		app.updateStatus()
	})

	app.statusLabel = widgets.NewLabel(0, 10, "", font, renderer)
	toggleBounds := app.wrapToggle.GetBounds()
	app.statusLabel.Bounds.X = toggleBounds.X + toggleBounds.W + 14

	app.areaTop = toggleBounds.Y + toggleBounds.H + 10

	app.area = widgets.NewTextArea(10, app.areaTop, windowWidth-20, windowHeight-app.areaTop-10, font, renderer, window)
	app.area.Lines = []string{
		"This is a multi-line text area.",
		"Type, use arrow keys, Enter, Backspace.",
		"Scroll with the mouse wheel (Shift+wheel scrolls sideways on most systems).",
		"Toggle word-wrap with the button above to switch between horizontal scrolling and wrapping long lines.",
		"Line 5", "Line 6", "Line 7", "Line 8", "Line 9", "Line 10",
		"Line 11", "Line 12", "Line 13", "Line 14", "Line 15",
		"This very last line is intentionally quite long so that horizontal scrolling has something real to demonstrate when word-wrap is switched off.",
	}
	app.area.Focus()
	app.updateStatus()

	return app
}

func (app *App) updateStatus() {
	app.statusLabel.UpdateText(fmt.Sprintf("word-wrap: %t", app.area.WordWrap))
}

func (app *App) Destroy() {
	app.wrapToggle.Destroy()
	app.statusLabel.Destroy()
}

func (app *App) handleEvents() bool {
	var event sdl.Event
	for sdl.PollEvent(&event) {
		mx, my := float32(0), float32(0)
		switch event.Type() {
		case sdl.EventMouseButtonDown, sdl.EventMouseButtonUp:
			mx = float32(event.Button().X)
			my = float32(event.Button().Y)
		case sdl.EventMouseMotion:
			mx = float32(event.Motion().X)
			my = float32(event.Motion().Y)
		case sdl.EventMouseWheel:
			wheel := event.Wheel()
			mx, my = wheel.MouseX, wheel.MouseY
		}

		switch event.Type() {
		case sdl.EventQuit:
			return false
		case sdl.EventWindowResized:
			app.windowWidth = float32(event.Window().Data1)
			app.windowHeight = float32(event.Window().Data2)
			app.area.Bounds.W = app.windowWidth - 20
			app.area.Bounds.H = app.windowHeight - app.areaTop - 10
		case sdl.EventKeyDown:
			if event.Key().Scancode == sdl.ScancodeEscape {
				return false
			}
			app.area.Update(event, mx, my)
		case sdl.EventTextInput, sdl.EventMouseWheel, sdl.EventMouseMotion:
			app.area.Update(event, mx, my)
		case sdl.EventMouseButtonDown:
			if !app.wrapToggle.Update(event, mx, my) {
				app.area.Update(event, mx, my)
			}
		case sdl.EventMouseButtonUp:
			app.wrapToggle.Update(event, mx, my)
			app.area.Update(event, mx, my)
		}
	}
	return true
}

func (app *App) render() {
	sdl.SetRenderDrawColor(app.renderer, 30, 30, 35, sdl.AlphaOpaque)
	sdl.RenderClear(app.renderer)

	app.wrapToggle.Render(app.renderer)
	app.statusLabel.Render(app.renderer)
	app.area.Render(app.renderer)

	sdl.RenderPresent(app.renderer)
}
