// Command text-playground is a scrollable, optionally word-wrapped
// multi-line text editor demonstrating internal/widgets.TextArea, including
// draggable scrollbar handles and toggleable word-wrap.
package main

import (
	"arkenidar.com/purego-sdl3/internal/sdlapp"
)

func main() {
	sdlApp, cleanup, err := sdlapp.Bootstrap(sdlapp.Config{
		Title:     "Text Playground",
		Width:     700,
		Height:    400,
		Resizable: true,
		FontPath:  "assets/OpenDyslexic-Regular.ttf",
		FontSize:  24,
	})
	if err != nil {
		panic(err)
	}
	defer cleanup()

	app := NewApp(sdlApp.Window, sdlApp.Renderer, sdlApp.Font, 700, 400)
	defer app.Destroy()

	sdlapp.Run(app.handleEvents, app.render)
}
