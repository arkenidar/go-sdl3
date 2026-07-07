// Command bouncing-balls animates a handful of colored circular sprites
// (a shared alpha-blended assets/ball.bmp, tinted per-ball via
// SetTextureColorMod) bouncing off the window edges, demonstrating
// per-frame state driven by delta time (sdl.GetTicks) with no widgets
// involved.
package main

import (
	"arkenidar.com/purego-sdl3/internal/sdlapp"
)

func main() {
	sdlApp, cleanup, err := sdlapp.Bootstrap(sdlapp.Config{
		Title:     "Bouncing Balls",
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

	app := NewApp(sdlApp.Renderer, sdlApp.Font, 700, 500)
	defer app.Destroy()

	sdlapp.Run(app.handleEvents, app.render)
}
