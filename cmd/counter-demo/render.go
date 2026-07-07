// render.go
package main

import (
	"github.com/jupiterrider/purego-sdl3/sdl"

	"arkenidar.com/purego-sdl3/internal/textutil"
)

func (app *App) render() {
	renderer := app.renderer

	sdl.SetRenderDrawColor(renderer, 100, 150, 200, sdl.AlphaOpaque)
	sdl.RenderClear(renderer)

	// Draw rectangle
	rect := sdl.FRect{X: app.x, Y: app.y, W: 100, H: 100}
	sdl.SetRenderDrawColor(renderer, 0, 0, 200, sdl.AlphaOpaque)
	sdl.RenderFillRect(renderer, &rect)

	// Render UI elements
	app.uiLayout.Render(renderer)
	app.newButton.Render(renderer) // Render the right-aligned button separately

	// Render instruction text at bottom with centering and wrapping
	textutil.RenderBottomText(renderer, app.font, "• move the blue square with arrow keys or mouse drag\n • click its buttons to change counter", app.windowWidth, app.windowHeight, 10)

	// Render alert if active
	if app.showAlert {
		app.renderAlert()
	}

	sdl.RenderPresent(renderer)
}
