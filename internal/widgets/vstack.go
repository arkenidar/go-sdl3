package widgets

import (
	"github.com/jupiterrider/purego-sdl3/sdl"
)

// VStack arranges widgets top-to-bottom in a single vertical column,
// starting at (X, Y) with Spacing pixels between each widget. It is the
// vertical counterpart to Layout, and additionally reports its total
// content size via ContentWidth/ContentHeight — use those to size a
// window to fit its content instead of guessing constants up front (a
// window sized smaller than its content silently clips/hides widgets).
type VStack struct {
	X, Y    float32
	Spacing float32
	Widgets []Widget
}

// NewVStack creates an empty vertical VStack anchored at (x, y).
func NewVStack(x, y, spacing float32) *VStack {
	return &VStack{X: x, Y: y, Spacing: spacing, Widgets: make([]Widget, 0)}
}

// AddWidget appends widget to the column, positioning it below the
// previous widget (or at the stack's origin if it's the first).
func (stack *VStack) AddWidget(widget Widget) {
	bounds := widget.GetBounds()

	if len(stack.Widgets) == 0 {
		bounds.X = stack.X
		bounds.Y = stack.Y
	} else {
		lastBounds := stack.Widgets[len(stack.Widgets)-1].GetBounds()
		bounds.X = stack.X
		bounds.Y = lastBounds.Y + lastBounds.H + stack.Spacing
	}

	widget.SetBounds(bounds)
	stack.Widgets = append(stack.Widgets, widget)
}

// ContentWidth returns the width of the widest widget in the stack.
func (stack *VStack) ContentWidth() float32 {
	var maxW float32
	for _, w := range stack.Widgets {
		if b := w.GetBounds(); b.W > maxW {
			maxW = b.W
		}
	}
	return maxW
}

// ContentHeight returns the total vertical span from the stack's origin
// to the bottom edge of its last widget.
func (stack *VStack) ContentHeight() float32 {
	if len(stack.Widgets) == 0 {
		return 0
	}
	last := stack.Widgets[len(stack.Widgets)-1].GetBounds()
	return (last.Y - stack.Y) + last.H
}

func (stack *VStack) Update(event sdl.Event, mx, my float32) bool {
	for _, widget := range stack.Widgets {
		if widget.Update(event, mx, my) {
			return true
		}
	}
	return false
}

func (stack *VStack) Render(renderer *sdl.Renderer) {
	for _, widget := range stack.Widgets {
		widget.Render(renderer)
	}
}

func (stack *VStack) Destroy() {
	for _, widget := range stack.Widgets {
		if btn, ok := widget.(*Button); ok {
			btn.Destroy()
		} else if lbl, ok := widget.(*Label); ok {
			lbl.Destroy()
		} else if cb, ok := widget.(*Checkbox); ok {
			cb.Destroy()
		}
	}
}
