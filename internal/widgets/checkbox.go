package widgets

import (
	"github.com/jupiterrider/purego-sdl3/sdl"
	"github.com/jupiterrider/purego-sdl3/ttf"
)

// Checkbox is a labeled toggle widget.
type Checkbox struct {
	Bounds   sdl.FRect
	Text     string
	Texture  *sdl.Texture
	Checked  bool
	OnToggle func(checked bool)

	boxSize float32
}

const checkboxLabelGap = 8

// NewCheckbox creates a labeled Checkbox at (x, y) with the given initial
// checked state; onToggle (optional) fires with the new state on click.
func NewCheckbox(x, y float32, text string, checked bool, font *ttf.Font, renderer *sdl.Renderer, onToggle func(checked bool)) *Checkbox {
	texture, textW, textH := makeTextTexture(renderer, font, text, sdl.Color{R: 255, G: 255, B: 255, A: 255})
	if texture == nil {
		panic(sdl.GetError())
	}

	boxSize := textH
	w := boxSize + checkboxLabelGap + textW
	h := textH

	return &Checkbox{
		Bounds:   sdl.FRect{X: x, Y: y, W: w, H: h},
		Text:     text,
		Texture:  texture,
		Checked:  checked,
		OnToggle: onToggle,
		boxSize:  boxSize,
	}
}

func (c *Checkbox) Update(event sdl.Event, mx, my float32) bool {
	if event.Type() == sdl.EventMouseButtonDown {
		if mx >= c.Bounds.X && mx <= c.Bounds.X+c.Bounds.W &&
			my >= c.Bounds.Y && my <= c.Bounds.Y+c.Bounds.H {
			c.Checked = !c.Checked
			if c.OnToggle != nil {
				c.OnToggle(c.Checked)
			}
			return true
		}
	}
	return false
}

func (c *Checkbox) Render(renderer *sdl.Renderer) {
	box := sdl.FRect{X: c.Bounds.X, Y: c.Bounds.Y, W: c.boxSize, H: c.boxSize}

	sdl.SetRenderDrawColor(renderer, 230, 230, 230, sdl.AlphaOpaque)
	sdl.RenderFillRect(renderer, &box)
	sdl.SetRenderDrawColor(renderer, 100, 100, 100, sdl.AlphaOpaque)
	sdl.RenderRect(renderer, &box)

	if c.Checked {
		sdl.SetRenderDrawColor(renderer, 40, 160, 70, sdl.AlphaOpaque)
		mark := sdl.FRect{
			X: box.X + box.W*0.2,
			Y: box.Y + box.H*0.2,
			W: box.W * 0.6,
			H: box.H * 0.6,
		}
		sdl.RenderFillRect(renderer, &mark)
	}

	var textW, textH float32
	sdl.GetTextureSize(c.Texture, &textW, &textH)
	textRect := sdl.FRect{
		X: c.Bounds.X + c.boxSize + checkboxLabelGap,
		Y: c.Bounds.Y + (c.Bounds.H-textH)/2,
		W: textW,
		H: textH,
	}
	sdl.RenderTexture(renderer, c.Texture, nil, &textRect)
}

func (c *Checkbox) GetBounds() sdl.FRect {
	return c.Bounds
}

func (c *Checkbox) SetBounds(bounds sdl.FRect) {
	c.Bounds = bounds
}

func (c *Checkbox) Destroy() {
	if c.Texture != nil {
		sdl.DestroyTexture(c.Texture)
		c.Texture = nil
	}
}
