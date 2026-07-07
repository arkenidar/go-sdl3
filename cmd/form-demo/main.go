// Command form-demo is a settings-style screen demonstrating buttons,
// a status label, and checkboxes arranged with internal/widgets.Layout.
package main

import (
	"arkenidar.com/purego-sdl3/internal/sdlapp"
)

func main() {
	sdlApp, cleanup, err := sdlapp.Bootstrap(sdlapp.Config{
		Title:     "Form Demo",
		Width:     820,
		Height:    220,
		Resizable: true,
		FontPath:  "assets/OpenDyslexic-Regular.ttf",
		FontSize:  20,
	})
	if err != nil {
		panic(err)
	}
	defer cleanup()

	app := NewApp(sdlApp.Renderer, sdlApp.Font)
	defer app.Destroy()

	sdlapp.Run(app.handleEvents, app.render)
}
