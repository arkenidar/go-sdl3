// Command crud-app is a minimal to-do list: add items via a text input,
// toggle a "done" checkbox per row, and delete rows — every mutation is
// persisted to a JSON file (see storage.go) so state survives restarts.
package main

import (
	"arkenidar.com/purego-sdl3/internal/sdlapp"
)

func main() {
	sdlApp, cleanup, err := sdlapp.Bootstrap(sdlapp.Config{
		Title:     "CRUD App",
		Width:     500,
		Height:    500,
		Resizable: true,
		FontPath:  "assets/OpenDyslexic-Regular.ttf",
		FontSize:  20,
	})
	if err != nil {
		panic(err)
	}
	defer cleanup()

	app := NewApp(sdlApp.Window, sdlApp.Renderer, sdlApp.Font, 500, 500)
	defer app.Destroy()

	sdlapp.Run(app.handleEvents, app.render)
}
