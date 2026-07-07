// Package widgets provides a small SDL3-backed UI toolkit (buttons, labels, layout).
package widgets

import (
	"github.com/jupiterrider/purego-sdl3/sdl"
)

// Widget interface for UI elements
type Widget interface {
	Update(event sdl.Event, mx, my float32) bool // Returns true if event was handled
	Render(renderer *sdl.Renderer)
	GetBounds() sdl.FRect
	SetBounds(bounds sdl.FRect) // Repositions/resizes the widget, used by container layouts
}
