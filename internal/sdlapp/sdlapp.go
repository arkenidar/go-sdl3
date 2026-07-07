// Package sdlapp provides reusable SDL3+TTF bootstrap and run-loop
// scaffolding shared by every example/app in this repo.
package sdlapp

import (
	"errors"

	"github.com/jupiterrider/purego-sdl3/sdl"
	"github.com/jupiterrider/purego-sdl3/ttf"
)

// Config describes how to set up the window, renderer and font for an app.
type Config struct {
	Title     string
	Width     int32
	Height    int32
	Resizable bool
	FontPath  string
	FontSize  float32
}

// App bundles the SDL/TTF handles every app needs.
type App struct {
	Window   *sdl.Window
	Renderer *sdl.Renderer
	Font     *ttf.Font
}

// Bootstrap initializes SDL and TTF, opens the font and creates the window
// and renderer described by cfg. Call the returned cleanup func (typically
// via defer) to release everything in the right order.
func Bootstrap(cfg Config) (app *App, cleanup func(), err error) {
	if !sdl.Init(sdl.InitVideo) {
		return nil, nil, errors.New(sdl.GetError())
	}

	if !ttf.Init() {
		sdl.Quit()
		return nil, nil, errors.New(sdl.GetError())
	}

	font := ttf.OpenFont(cfg.FontPath, cfg.FontSize)
	if font == nil {
		ttf.Quit()
		sdl.Quit()
		return nil, nil, errors.New(sdl.GetError())
	}

	var window *sdl.Window
	var renderer *sdl.Renderer
	windowFlags := sdl.WindowFlags(0)
	if cfg.Resizable {
		windowFlags |= sdl.WindowResizable
	}
	if !sdl.CreateWindowAndRenderer(cfg.Title, cfg.Width, cfg.Height, windowFlags, &window, &renderer) {
		ttf.CloseFont(font)
		ttf.Quit()
		sdl.Quit()
		return nil, nil, errors.New(sdl.GetError())
	}

	app = &App{Window: window, Renderer: renderer, Font: font}
	cleanup = func() {
		sdl.DestroyRenderer(renderer)
		sdl.DestroyWindow(window)
		ttf.CloseFont(font)
		ttf.Quit()
		sdl.Quit()
	}
	return app, cleanup, nil
}

// Run drives the standard "poll events, then render" loop until
// handleEvents returns false.
func Run(handleEvents func() bool, render func()) {
	for handleEvents() {
		render()
	}
}
