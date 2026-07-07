// app.go
package main

import (
	"math/rand"

	"github.com/jupiterrider/purego-sdl3/sdl"
	"github.com/jupiterrider/purego-sdl3/ttf"

	"arkenidar.com/purego-sdl3/internal/textutil"
)

const ballCount = 10

// App holds the animated state for the bouncing-balls demo.
type App struct {
	renderer *sdl.Renderer
	font     *ttf.Font

	windowWidth  float32
	windowHeight float32

	balls       []*Ball
	ballTexture *sdl.Texture
	lastTicks   uint64
}

func NewApp(renderer *sdl.Renderer, font *ttf.Font, windowWidth, windowHeight float32) *App {
	app := &App{
		renderer:     renderer,
		font:         font,
		windowWidth:  windowWidth,
		windowHeight: windowHeight,
		lastTicks:    sdl.GetTicks(),
	}

	surface := sdl.LoadBMP("assets/ball.bmp")
	if surface == nil {
		panic(sdl.GetError())
	}
	app.ballTexture = sdl.CreateTextureFromSurface(renderer, surface)
	sdl.DestroySurface(surface)
	if app.ballTexture == nil {
		panic(sdl.GetError())
	}
	sdl.SetTextureBlendMode(app.ballTexture, sdl.BlendModeBlend)

	colors := []sdl.Color{
		{R: 220, G: 60, B: 60, A: 255},
		{R: 60, G: 200, B: 90, A: 255},
		{R: 60, G: 120, B: 220, A: 255},
		{R: 230, G: 200, B: 60, A: 255},
		{R: 200, G: 90, B: 220, A: 255},
	}

	for i := 0; i < ballCount; i++ {
		size := float32(15 + rand.Intn(20))
		app.balls = append(app.balls, &Ball{
			X:     rand.Float32() * (windowWidth - size),
			Y:     rand.Float32() * (windowHeight - size),
			VX:    (rand.Float32()*2 - 1) * 200,
			VY:    (rand.Float32()*2 - 1) * 200,
			Size:  size,
			Color: colors[i%len(colors)],
		})
	}

	return app
}

func (app *App) Destroy() {
	sdl.DestroyTexture(app.ballTexture)
}

func (app *App) handleEvents() bool {
	var event sdl.Event
	for sdl.PollEvent(&event) {
		switch event.Type() {
		case sdl.EventQuit:
			return false
		case sdl.EventWindowResized:
			app.windowWidth = float32(event.Window().Data1)
			app.windowHeight = float32(event.Window().Data2)
		case sdl.EventKeyDown:
			if event.Key().Scancode == sdl.ScancodeEscape {
				return false
			}
		}
	}
	return true
}

func (app *App) render() {
	now := sdl.GetTicks()
	dtSeconds := float32(now-app.lastTicks) / 1000
	app.lastTicks = now

	for _, ball := range app.balls {
		ball.Update(dtSeconds, app.windowWidth, app.windowHeight)
	}

	sdl.SetRenderDrawColor(app.renderer, 20, 20, 25, sdl.AlphaOpaque)
	sdl.RenderClear(app.renderer)

	for _, ball := range app.balls {
		ball.Draw(app.renderer, app.ballTexture)
	}

	textutil.RenderBottomText(app.renderer, app.font, "bouncing balls demo — Escape to quit", app.windowWidth, app.windowHeight, 10)

	sdl.RenderPresent(app.renderer)
}
