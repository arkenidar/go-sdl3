package widgets

import (
	"github.com/jupiterrider/purego-sdl3/sdl"
	"github.com/jupiterrider/purego-sdl3/ttf"
)

// Button is a clickable, texture-labeled widget that calls OnClick on press.
type Button struct {
	Bounds    sdl.FRect
	Text      string
	Texture   *sdl.Texture
	OnClick   func()
	IsPressed bool
}

// NewButton creates a Button labeled with text. Pass w<=0 or h<=0 to
// auto-size the button to fit its rendered text plus padding.
func NewButton(x, y, w, h float32, text string, font *ttf.Font, renderer *sdl.Renderer, onClick func()) *Button {
	texture, textW, textH := makeTextTexture(renderer, font, text, sdl.Color{R: 255, G: 255, B: 255, A: 255})
	if texture == nil {
		panic(sdl.GetError())
	}

	// Auto-size button based on text if width/height are 0
	if w <= 0 {
		w = textW + 20 // Add padding
	}
	if h <= 0 {
		h = textH + 16 // Add padding
	}

	return &Button{
		Bounds:  sdl.FRect{X: x, Y: y, W: w, H: h},
		Text:    text,
		Texture: texture,
		OnClick: onClick,
	}
}

// Update handles a mouse click inside the button's bounds, firing OnClick.
func (b *Button) Update(event sdl.Event, mx, my float32) bool {
	if event.Type() == sdl.EventMouseButtonDown {
		if mx >= b.Bounds.X && mx <= b.Bounds.X+b.Bounds.W &&
			my >= b.Bounds.Y && my <= b.Bounds.Y+b.Bounds.H {
			b.IsPressed = true
			if b.OnClick != nil {
				b.OnClick()
			}
			return true
		}
	} else if event.Type() == sdl.EventMouseButtonUp {
		b.IsPressed = false
	}
	return false
}

func (b *Button) Render(renderer *sdl.Renderer) {
	// Draw button background
	if b.IsPressed {
		sdl.SetRenderDrawColor(renderer, 60, 60, 60, sdl.AlphaOpaque)
	} else {
		sdl.SetRenderDrawColor(renderer, 80, 80, 80, sdl.AlphaOpaque)
	}
	sdl.RenderFillRect(renderer, &b.Bounds)

	// Draw button text (centered)
	var textW, textH float32
	sdl.GetTextureSize(b.Texture, &textW, &textH)
	textRect := sdl.FRect{
		X: b.Bounds.X + (b.Bounds.W-textW)/2,
		Y: b.Bounds.Y + (b.Bounds.H-textH)/2,
		W: textW,
		H: textH,
	}
	sdl.RenderTexture(renderer, b.Texture, nil, &textRect)
}

func (b *Button) GetBounds() sdl.FRect {
	return b.Bounds
}

func (b *Button) SetBounds(bounds sdl.FRect) {
	b.Bounds = bounds
}

func (b *Button) Destroy() {
	if b.Texture != nil {
		sdl.DestroyTexture(b.Texture)
		b.Texture = nil
	}
}
