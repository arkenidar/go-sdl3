package main

// app.go
// This is a simple SDL3 application written in Go
// It creates a window, handles events, and renders a rectangle

// use purego-sdl3 from jupiterrider
import (
	"fmt"
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

	// Create button textures
	plusSurface := ttf.RenderTextBlended(font, "+", 0, sdl.Color{R: 255, G: 255, B: 255, A: 255})
	if plusSurface == nil {
		panic(sdl.GetError())
	}
	defer sdl.DestroySurface(plusSurface)

	plusTexture := sdl.CreateTextureFromSurface(renderer, plusSurface)
	if plusTexture == nil {
		panic(sdl.GetError())
	}
	defer sdl.DestroyTexture(plusTexture)

	minusSurface := ttf.RenderTextBlended(font, "-", 0, sdl.Color{R: 255, G: 255, B: 255, A: 255})
	if minusSurface == nil {
		panic(sdl.GetError())
	}
	defer sdl.DestroySurface(minusSurface)

	minusTexture := sdl.CreateTextureFromSurface(renderer, minusSurface)
	if minusTexture == nil {
		panic(sdl.GetError())
	}
	defer sdl.DestroyTexture(minusTexture)

	// Get button dimensions
	var plusW, plusH, minusW, minusH float32
	sdl.GetTextureSize(plusTexture, &plusW, &plusH)
	sdl.GetTextureSize(minusTexture, &minusW, &minusH)

	// SECTION : Application state

	x, y := float32(150), float32(150)
	counter := 0

	// Button positions and sizes
	buttonWidth, buttonHeight := float32(50), float32(40)
	buttonY := textH + 20 // Position buttons with more space below the instruction text
	plusButton := sdl.FRect{X: 10, Y: buttonY, W: buttonWidth, H: buttonHeight}
	minusButton := sdl.FRect{X: 70, Y: buttonY, W: buttonWidth, H: buttonHeight}

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
				
				// Check button clicks
				if mx >= plusButton.X && mx <= plusButton.X+plusButton.W && 
				   my >= plusButton.Y && my <= plusButton.Y+plusButton.H {
					counter++
				} else if mx >= minusButton.X && mx <= minusButton.X+minusButton.W && 
						  my >= minusButton.Y && my <= minusButton.Y+minusButton.H {
					counter--
				} else if mx >= x && mx <= x+100 && my >= y && my <= y+100 {
					// Check if mouse is inside the square
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

		// Draw rectangle
		rect := sdl.FRect{X: x, Y: y, W: 100, H: 100}
		sdl.SetRenderDrawColor(renderer, 0, 0, 200, sdl.AlphaOpaque)
		sdl.RenderFillRect(renderer, &rect)

		// Draw buttons
		sdl.SetRenderDrawColor(renderer, 80, 80, 80, sdl.AlphaOpaque)
		sdl.RenderFillRect(renderer, &plusButton)
		sdl.RenderFillRect(renderer, &minusButton)

		// Draw button text
		plusTextRect := sdl.FRect{
			X: plusButton.X + (plusButton.W-plusW)/2,
			Y: plusButton.Y + (plusButton.H-plusH)/2,
			W: plusW, H: plusH,
		}
		minusTextRect := sdl.FRect{
			X: minusButton.X + (minusButton.W-minusW)/2,
			Y: minusButton.Y + (minusButton.H-minusH)/2,
			W: minusW, H: minusH,
		}
		sdl.RenderTexture(renderer, plusTexture, nil, &plusTextRect)
		sdl.RenderTexture(renderer, minusTexture, nil, &minusTextRect)

		// Draw counter
		counterText := fmt.Sprintf("Counter: %d", counter)
		counterSurface := ttf.RenderTextBlended(font, counterText, 0, sdl.Color{R: 255, G: 255, B: 255, A: 255})
		if counterSurface != nil {
			counterTexture := sdl.CreateTextureFromSurface(renderer, counterSurface)
			if counterTexture != nil {
				var counterW, counterH float32
				sdl.GetTextureSize(counterTexture, &counterW, &counterH)
				counterRect := sdl.FRect{X: 130, Y: buttonY + (buttonHeight-counterH)/2, W: counterW, H: counterH}
				sdl.RenderTexture(renderer, counterTexture, nil, &counterRect)
				sdl.DestroyTexture(counterTexture)
			}
			sdl.DestroySurface(counterSurface)
		}

		// Draw instruction text
		textRect := sdl.FRect{X: 10, Y: 10, W: textW, H: textH}
		sdl.RenderTexture(renderer, textTexture, nil, &textRect)

		sdl.RenderPresent(renderer)
	}
}
