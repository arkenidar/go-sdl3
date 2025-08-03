package main

// app.go
// This is a simple SDL3 application written in Go
// It creates a window, handles events, and renders a rectangle

// use purego-sdl3 from jupiterrider
import (
	"github.com/jupiterrider/purego-sdl3/sdl"
	"github.com/jupiterrider/purego-sdl3/ttf"
)

func main() {

	// SECTION : Initialize SDL

	defer sdl.Quit()
	if !sdl.Init(sdl.InitVideo) {
		panic(sdl.GetError())
	}

	// Initialize TTF
	defer ttf.Quit()
	if !ttf.Init() {
		panic(sdl.GetError())
	}

	// Load font
	font := ttf.OpenFont("assets/OpenDyslexic-Regular.ttf", 24)
	if font == nil {
		panic(sdl.GetError())
	}
	defer ttf.CloseFont(font)

	// Create a window and renderer
	// The window is resizable and has a title
	// The renderer is used to draw graphics on the window
	var window *sdl.Window
	var renderer *sdl.Renderer
	if !sdl.CreateWindowAndRenderer("App built with Go and SDL3", 700, 500, sdl.WindowResizable, &window, &renderer) {
		panic(sdl.GetError())
	}
	defer sdl.DestroyRenderer(renderer)
	defer sdl.DestroyWindow(window)

	// Create text texture
	textSurface := ttf.RenderTextBlended(font, "( move with arrow keys or mouse drag )", 0, sdl.Color{R: 255, G: 255, B: 255, A: 255})
	if textSurface == nil {
		panic(sdl.GetError())
	}
	defer sdl.DestroySurface(textSurface)

	textTexture := sdl.CreateTextureFromSurface(renderer, textSurface)
	if textTexture == nil {
		panic(sdl.GetError())
	}
	defer sdl.DestroyTexture(textTexture)

	// Get text dimensions
	var textW, textH float32
	sdl.GetTextureSize(textTexture, &textW, &textH)

	// SECTION : Application state

	x, y := float32(150), float32(150)

	// Drag state variables
	dragging := false
	dragOffsetX, dragOffsetY := float32(0), float32(0)

Outer:
	for {
		var event sdl.Event
		for sdl.PollEvent(&event) {
			switch event.Type() {
			case sdl.EventQuit:
				break Outer
			case sdl.EventKeyDown:
				switch event.Key().Scancode {
				case sdl.ScancodeEscape:
					break Outer
				case sdl.ScancodeRight:
					x += 15
				case sdl.ScancodeLeft:
					x -= 15
				case sdl.ScancodeDown:
					y += 15
				case sdl.ScancodeUp:
					y -= 15
				}
			case sdl.EventMouseButtonDown:
				mx := float32(event.Button().X)
				my := float32(event.Button().Y)
				// Check if mouse is inside the square
				if mx >= x && mx <= x+100 && my >= y && my <= y+100 {
					dragging = true
					dragOffsetX = mx - x
					dragOffsetY = my - y
				}
			case sdl.EventMouseButtonUp:
				dragging = false
			case sdl.EventMouseMotion:
				if dragging {
					mx := float32(event.Motion().X)
					my := float32(event.Motion().Y)
					x = mx - dragOffsetX
					y = my - dragOffsetY
				}
			}
		}

		// SECTION : Rendering
		sdl.SetRenderDrawColor(renderer, 100, 150, 200, sdl.AlphaOpaque)
		sdl.RenderClear(renderer)

		rect := sdl.FRect{X: x, Y: y, W: 100, H: 100}
		sdl.SetRenderDrawColor(renderer, 0, 0, 200, sdl.AlphaOpaque)
		sdl.RenderFillRect(renderer, &rect)

		textRect := sdl.FRect{X: 10, Y: 10, W: textW, H: textH}
		sdl.RenderTexture(renderer, textTexture, nil, &textRect)

		sdl.RenderPresent(renderer)
	}
}
