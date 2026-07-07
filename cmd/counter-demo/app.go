// app.go
package main

import (
	"fmt"

	"github.com/jupiterrider/purego-sdl3/sdl"
	"github.com/jupiterrider/purego-sdl3/ttf"

	"arkenidar.com/purego-sdl3/internal/widgets"
)

// App holds all mutable state for the running application.
type App struct {
	window   *sdl.Window
	renderer *sdl.Renderer
	font     *ttf.Font

	// Draggable square
	x, y float32

	// Counter demo
	counter int

	// Alert dialog
	showAlert    bool
	alertMessage string

	// Window dimensions (updated on resize)
	windowWidth  float32
	windowHeight float32

	// UI widgets
	uiLayout     *widgets.Layout
	plusButton   *widgets.Button
	minusButton  *widgets.Button
	counterLabel *widgets.Label
	newButton    *widgets.Button

	// Drag state
	dragging    bool
	dragOffsetX float32
	dragOffsetY float32
}

func NewApp(window *sdl.Window, renderer *sdl.Renderer, font *ttf.Font) *App {
	app := &App{
		window:       window,
		renderer:     renderer,
		font:         font,
		x:            150,
		y:            150,
		showAlert:    false,
		alertMessage: "Button clicked! This is a longer message that will demonstrate the text wrapping functionality in alert dialogs.",
		windowWidth:  700,
		windowHeight: 500,
	}

	// Create UI layout with buttons and counter (positioned at top)
	app.uiLayout = widgets.NewLayout(10, 10, 10)

	// Create buttons with callbacks (auto-sized)
	app.plusButton = widgets.NewButton(0, 0, 0, 0, "+", font, renderer, func() {
		app.counter++
	})
	app.minusButton = widgets.NewButton(0, 0, 0, 0, "-", font, renderer, func() {
		app.counter--
	})

	// Create counter label
	app.counterLabel = widgets.NewLabel(0, 0, fmt.Sprintf("Counter: %d", app.counter), font, renderer)

	// Add widgets to main layout
	app.uiLayout.AddWidget(app.plusButton)
	app.uiLayout.AddWidget(app.minusButton)
	app.uiLayout.AddWidget(app.counterLabel)

	// Create a right-aligned button (demonstration of extensibility - auto-sized)
	app.newButton = widgets.NewButton(0, 0, 0, 0, "Click Me", font, renderer, func() {
		app.showAlert = true
	})
	// Position the button to the right border using dynamic window width
	buttonBounds := app.newButton.GetBounds()
	app.newButton.Bounds.X = app.windowWidth - buttonBounds.W - 10 // 10px margin from right edge
	app.newButton.Bounds.Y = 10                                    // Align with the top button row

	return app
}

func (app *App) Destroy() {
	app.uiLayout.Destroy()
}
