package widgets

import (
	"github.com/jupiterrider/purego-sdl3/sdl"
)

// Layout arranges widgets left-to-right in a single horizontal row,
// starting at (X, Y) with Spacing pixels between each widget.
type Layout struct {
	X, Y    float32
	Spacing float32
	Widgets []Widget
}

// NewLayout creates an empty horizontal Layout anchored at (x, y).
func NewLayout(x, y, spacing float32) *Layout {
	return &Layout{X: x, Y: y, Spacing: spacing, Widgets: make([]Widget, 0)}
}

// AddWidget appends widget to the row, positioning it after the previous
// widget (or at the layout's origin if it's the first).
func (layout *Layout) AddWidget(widget Widget) {
	bounds := widget.GetBounds()

	// Position widget based on layout
	if len(layout.Widgets) == 0 {
		// First widget
		bounds.X = layout.X
		bounds.Y = layout.Y
	} else {
		// Position relative to previous widget
		lastBounds := layout.Widgets[len(layout.Widgets)-1].GetBounds()
		bounds.X = lastBounds.X + lastBounds.W + layout.Spacing
		bounds.Y = layout.Y
	}

	widget.SetBounds(bounds)
	layout.Widgets = append(layout.Widgets, widget)
}

func (layout *Layout) Update(event sdl.Event, mx, my float32) bool {
	for _, widget := range layout.Widgets {
		if widget.Update(event, mx, my) {
			return true
		}
	}
	return false
}

func (layout *Layout) Render(renderer *sdl.Renderer) {
	for _, widget := range layout.Widgets {
		widget.Render(renderer)
	}
}

func (layout *Layout) Destroy() {
	for _, widget := range layout.Widgets {
		if btn, ok := widget.(*Button); ok {
			btn.Destroy()
		} else if lbl, ok := widget.(*Label); ok {
			lbl.Destroy()
		} else if cb, ok := widget.(*Checkbox); ok {
			cb.Destroy()
		}
	}
}
