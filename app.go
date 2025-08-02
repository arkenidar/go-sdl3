package main

// app.go
// This is a simple SDL3 application written in Go
// It creates a window, handles events, and renders a rectangle

// use purego-sdl3 from jupiterrider
import "github.com/jupiterrider/purego-sdl3/sdl"

func main() {

	// SECTION : Initialize SDL

	defer sdl.Quit()
	if !sdl.Init(sdl.InitVideo) {
		panic(sdl.GetError())
	}

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

	// SECTION : Application state

	x, y := float32(150), float32(150)

Outer:
	// SECTION : Main loop

	// Handle events and render
	// Exit on Escape key or window close
	// Arrow keys move the rectangle
	for {

		// SECTION : Event handling

		// Poll for events and handle them
		// This is where we check for user input
		// and update the application state accordingly
		var event sdl.Event
		for sdl.PollEvent(&event) {
			switch event.Type() {
			// Handle quit event (window close)
			case sdl.EventQuit:
				break Outer
			// Handle key down events
			// This is where we check for keyboard input
			// and update the rectangle position
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
			}
		}

		// SECTION : Rendering

		// Clear the renderer
		sdl.SetRenderDrawColor(renderer, 100, 150, 200, sdl.AlphaOpaque)
		sdl.RenderClear(renderer)

		// Draw rectangle
		// Use FRect for floating-point coordinates
		rect := sdl.FRect{X: x, Y: y, W: 100, H: 100}
		// Set color and fill the rectangle
		sdl.SetRenderDrawColor(renderer, 0, 0, 200, sdl.AlphaOpaque)
		sdl.RenderFillRect(renderer, &rect)

		// Present the renderer
		// This updates the window with the rendered content
		sdl.RenderPresent(renderer)
	}
}
