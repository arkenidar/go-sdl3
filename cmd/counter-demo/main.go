// Command counter-demo is a simple SDL3 application: it creates a window,
// handles keyboard/mouse events, and renders a draggable rectangle plus a
// button-driven counter and a modal alert dialog.
package main

import (
	"arkenidar.com/purego-sdl3/internal/sdlapp"
)

func main() {
	sdlApp, cleanup, err := sdlapp.Bootstrap(sdlapp.Config{
		Title:     "App built with Go and SDL3",
		Width:     700,
		Height:    500,
		Resizable: true,
		FontPath:  "assets/OpenDyslexic-Regular.ttf",
		FontSize:  24,
	})
	if err != nil {
		panic(err)
	}
	defer cleanup()

	app := NewApp(sdlApp.Window, sdlApp.Renderer, sdlApp.Font)
	defer app.Destroy()

	sdlapp.Run(app.handleEvents, app.render)
}
