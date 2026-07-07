// Command blank-window is the bare-minimum starter/template for a new SDL3
// app in this repo. Copy this directory, rename the package/binary, and
// build from here: it opens a window, clears it every frame, and quits on
// Escape or the window's close button. Delete this comment once you start
// building.
package main

import (
	"github.com/jupiterrider/purego-sdl3/sdl"

	"arkenidar.com/purego-sdl3/internal/sdlapp"
)

func main() {
	app, cleanup, err := sdlapp.Bootstrap(sdlapp.Config{
		Title:     "Blank Window",
		Width:     640,
		Height:    480,
		Resizable: true,
		FontPath:  "assets/OpenDyslexic-Regular.ttf",
		FontSize:  24,
	})
	if err != nil {
		panic(err)
	}
	defer cleanup()

	handleEvents := func() bool {
		var event sdl.Event
		for sdl.PollEvent(&event) {
			switch event.Type() {
			case sdl.EventQuit:
				return false
			case sdl.EventKeyDown:
				if event.Key().Scancode == sdl.ScancodeEscape {
					return false
				}
			}
		}
		return true
	}

	render := func() {
		sdl.SetRenderDrawColor(app.Renderer, 30, 30, 40, sdl.AlphaOpaque)
		sdl.RenderClear(app.Renderer)
		sdl.RenderPresent(app.Renderer)
	}

	sdlapp.Run(handleEvents, render)
}
