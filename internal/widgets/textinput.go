package widgets

import (
	"github.com/jupiterrider/purego-sdl3/sdl"
	"github.com/jupiterrider/purego-sdl3/ttf"
)

// TextInput is a single-line editable text field with a blinking cursor.
// Call Focus/Blur to start/stop receiving OS text-input events for it.
type TextInput struct {
	Bounds   sdl.FRect
	Value    string
	Focused  bool
	OnSubmit func(value string)

	window   *sdl.Window
	font     *ttf.Font
	renderer *sdl.Renderer

	cursorVisible  bool
	cursorLastTick uint64
}

const cursorBlinkMillis = 500

// NewTextInput creates an unfocused TextInput at (x, y) sized (w, h). Call
// Focus to start accepting keystrokes.
func NewTextInput(x, y, w, h float32, font *ttf.Font, renderer *sdl.Renderer, window *sdl.Window) *TextInput {
	return &TextInput{
		Bounds:         sdl.FRect{X: x, Y: y, W: w, H: h},
		window:         window,
		font:           font,
		renderer:       renderer,
		cursorLastTick: sdl.GetTicks(),
	}
}

func (t *TextInput) Focus() {
	if t.Focused {
		return
	}
	t.Focused = true
	t.cursorVisible = true
	t.cursorLastTick = sdl.GetTicks()
	sdl.StartTextInput(t.window)
}

func (t *TextInput) Blur() {
	if !t.Focused {
		return
	}
	t.Focused = false
	sdl.StopTextInput(t.window)
}

func (t *TextInput) Update(event sdl.Event, mx, my float32) bool {
	switch event.Type() {
	case sdl.EventMouseButtonDown:
		inside := mx >= t.Bounds.X && mx <= t.Bounds.X+t.Bounds.W &&
			my >= t.Bounds.Y && my <= t.Bounds.Y+t.Bounds.H
		if inside {
			t.Focus()
		} else if t.Focused {
			t.Blur()
		}
		return inside
	case sdl.EventTextInput:
		if !t.Focused {
			return false
		}
		textEvent := event.Text()
		t.Value += textEvent.Text()
		return true
	case sdl.EventKeyDown:
		if !t.Focused {
			return false
		}
		switch event.Key().Scancode {
		case sdl.ScancodeBackspace:
			if len(t.Value) > 0 {
				t.Value = t.Value[:len(t.Value)-1]
			}
			return true
		case sdl.ScancodeReturn:
			if t.OnSubmit != nil {
				t.OnSubmit(t.Value)
			}
			return true
		}
	}
	return false
}

func (t *TextInput) Render(renderer *sdl.Renderer) {
	sdl.SetRenderDrawColor(renderer, 250, 250, 250, sdl.AlphaOpaque)
	sdl.RenderFillRect(renderer, &t.Bounds)

	if t.Focused {
		sdl.SetRenderDrawColor(renderer, 60, 130, 220, sdl.AlphaOpaque)
	} else {
		sdl.SetRenderDrawColor(renderer, 120, 120, 120, sdl.AlphaOpaque)
	}
	sdl.RenderRect(renderer, &t.Bounds)

	text := t.Value
	var textW, textH float32
	if text != "" {
		surface := ttf.RenderTextBlended(t.font, text, 0, sdl.Color{R: 0, G: 0, B: 0, A: 255})
		if surface != nil {
			texture := sdl.CreateTextureFromSurface(renderer, surface)
			sdl.DestroySurface(surface)
			if texture != nil {
				sdl.GetTextureSize(texture, &textW, &textH)
				textRect := sdl.FRect{
					X: t.Bounds.X + 6,
					Y: t.Bounds.Y + (t.Bounds.H-textH)/2,
					W: textW,
					H: textH,
				}
				sdl.RenderTexture(renderer, texture, nil, &textRect)
				sdl.DestroyTexture(texture)
			}
		}
	}

	if t.Focused {
		now := sdl.GetTicks()
		if now-t.cursorLastTick > cursorBlinkMillis {
			t.cursorVisible = !t.cursorVisible
			t.cursorLastTick = now
		}
		if t.cursorVisible {
			cursorX := t.Bounds.X + 6 + textW + 2
			cursor := sdl.FRect{
				X: cursorX,
				Y: t.Bounds.Y + 6,
				W: 2,
				H: t.Bounds.H - 12,
			}
			sdl.SetRenderDrawColor(renderer, 0, 0, 0, sdl.AlphaOpaque)
			sdl.RenderFillRect(renderer, &cursor)
		}
	}
}

func (t *TextInput) GetBounds() sdl.FRect {
	return t.Bounds
}

func (t *TextInput) SetBounds(bounds sdl.FRect) {
	t.Bounds = bounds
}
