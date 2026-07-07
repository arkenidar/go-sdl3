// Command launcher lists every other example app as a button and launches
// the chosen one as a sibling process, so the whole cmd/ catalog can be
// browsed without a terminal.
package main

import (
	"github.com/jupiterrider/purego-sdl3/sdl"

	"arkenidar.com/purego-sdl3/internal/sdlapp"
)

func main() {
	// Bootstrap needs a renderer/font before any widget can be built, so
	// the window starts at a placeholder size; it's resized to fit the
	// content below once NewApp has computed how much space it needs.
	sdlApp, cleanup, err := sdlapp.Bootstrap(sdlapp.Config{
		Title:     "Example Launcher",
		Width:     400,
		Height:    300,
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

	sdl.SetWindowSize(sdlApp.Window, int32(app.ContentWidth), int32(app.ContentHeight))

	sdlapp.Run(app.handleEvents, app.render)
}
