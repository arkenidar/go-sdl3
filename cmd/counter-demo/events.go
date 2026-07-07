// events.go
package main

import (
	"fmt"

	"github.com/jupiterrider/purego-sdl3/sdl"
)

// handleEvents drains the SDL event queue for one frame.
// It returns false when the application should quit.
func (app *App) handleEvents() bool {
	var event sdl.Event
	for sdl.PollEvent(&event) {
		mx := float32(0)
		my := float32(0)

		// Get mouse position for widgets
		if event.Type() == sdl.EventMouseButtonDown || event.Type() == sdl.EventMouseButtonUp {
			mx = float32(event.Button().X)
			my = float32(event.Button().Y)
		} else if event.Type() == sdl.EventMouseMotion {
			mx = float32(event.Motion().X)
			my = float32(event.Motion().Y)
		}

		switch event.Type() {
		case sdl.EventQuit:
			return false
		case sdl.EventWindowResized:
			app.handleResize(event)
		case sdl.EventKeyDown:
			if !app.handleKeyDown(event) {
				return false
			}
		case sdl.EventMouseButtonDown:
			app.handleMouseButtonDown(event, mx, my)
		case sdl.EventMouseButtonUp:
			app.handleMouseButtonUp(event, mx, my)
		case sdl.EventMouseMotion:
			app.handleMouseMotion(mx, my)
		}
	}
	return true
}

func (app *App) handleResize(event sdl.Event) {
	app.windowWidth = float32(event.Window().Data1)
	app.windowHeight = float32(event.Window().Data2)

	// Reposition right-aligned button when window resizes
	buttonBounds := app.newButton.GetBounds()
	app.newButton.Bounds.X = app.windowWidth - buttonBounds.W - 10 // 10px margin from right edge

	// Keep square within new window bounds
	app.clampSquare()
}

// handleKeyDown returns false if the application should quit.
func (app *App) handleKeyDown(event sdl.Event) bool {
	switch event.Key().Scancode {
	case sdl.ScancodeEscape:
		if app.showAlert {
			app.showAlert = false // Dismiss alert first
		} else {
			return false // Exit application
		}
	case sdl.ScancodeSpace:
		if app.showAlert {
			app.showAlert = false // Dismiss alert with spacebar
		}
	case sdl.ScancodeRight:
		app.x += 15
	case sdl.ScancodeLeft:
		app.x -= 15
	case sdl.ScancodeDown:
		app.y += 15
	case sdl.ScancodeUp:
		app.y -= 15
	}
	app.clampSquare()
	return true
}

func (app *App) handleMouseButtonDown(event sdl.Event, mx, my float32) {
	// Check if alert is showing and handle click-to-close
	if app.showAlert {
		app.showAlert = false // Dismiss alert on any click
		return
	}

	// Check if UI layout handled the event first
	if app.uiLayout.Update(event, mx, my) {
		return
	}
	// Check if right-aligned button handled the event
	if app.newButton.Update(event, mx, my) {
		return
	}
	// Check if mouse is inside the square for dragging
	if mx >= app.x && mx <= app.x+100 && my >= app.y && my <= app.y+100 {
		app.dragging = true
		app.dragOffsetX = mx - app.x
		app.dragOffsetY = my - app.y
	}
}

func (app *App) handleMouseButtonUp(event sdl.Event, mx, my float32) {
	app.uiLayout.Update(event, mx, my)
	app.newButton.Update(event, mx, my) // Handle button release for right-aligned button
	app.dragging = false

	// Update counter display if counter changed
	newCounterText := fmt.Sprintf("Counter: %d", app.counter)
	if newCounterText != app.counterLabel.Text {
		app.counterLabel.UpdateText(newCounterText)
	}
}

func (app *App) handleMouseMotion(mx, my float32) {
	if !app.dragging {
		return
	}
	app.x = mx - app.dragOffsetX
	app.y = my - app.dragOffsetY
	app.clampSquare()
}

// clampSquare keeps the draggable square within the current window bounds.
func (app *App) clampSquare() {
	if app.x < 0 {
		app.x = 0
	}
	if app.y < 0 {
		app.y = 0
	}
	if app.x+100 > app.windowWidth {
		app.x = app.windowWidth - 100
	}
	if app.y+100 > app.windowHeight {
		app.y = app.windowHeight - 100
	}
}
