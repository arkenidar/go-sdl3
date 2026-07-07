package widgets

import (
	"github.com/jupiterrider/purego-sdl3/sdl"
	"github.com/jupiterrider/purego-sdl3/ttf"
)

// Label widget for displaying text
type Label struct {
	Bounds   sdl.FRect
	Text     string
	Texture  *sdl.Texture
	font     *ttf.Font
	renderer *sdl.Renderer
}

// NewLabel creates a single-line, non-interactive text Label at (x, y).
func NewLabel(x, y float32, text string, font *ttf.Font, renderer *sdl.Renderer) *Label {
	label := &Label{
		Text:     text,
		font:     font,
		renderer: renderer,
	}
	label.UpdateText(text)
	label.Bounds.X = x
	label.Bounds.Y = y
	return label
}

// UpdateText replaces the label's displayed text, re-rendering its texture.
func (l *Label) UpdateText(text string) {
	if l.Texture != nil {
		sdl.DestroyTexture(l.Texture)
	}

	// For now, render as single line - multiline support would require more complex text layout
	surface := ttf.RenderTextBlended(l.font, text, 0, sdl.Color{R: 255, G: 255, B: 255, A: 255})
	if surface != nil {
		l.Texture = sdl.CreateTextureFromSurface(l.renderer, surface)
		sdl.GetTextureSize(l.Texture, &l.Bounds.W, &l.Bounds.H)
		sdl.DestroySurface(surface)
	}
	l.Text = text
}

func (l *Label) Update(event sdl.Event, mx, my float32) bool {
	return false // Labels don't handle events
}

func (l *Label) Render(renderer *sdl.Renderer) {
	if l.Texture != nil {
		sdl.RenderTexture(renderer, l.Texture, nil, &l.Bounds)
	}
}

func (l *Label) GetBounds() sdl.FRect {
	return l.Bounds
}

func (l *Label) SetBounds(bounds sdl.FRect) {
	l.Bounds = bounds
}

func (l *Label) Destroy() {
	if l.Texture != nil {
		sdl.DestroyTexture(l.Texture)
		l.Texture = nil
	}
}
