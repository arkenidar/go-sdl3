package widgets

import (
	"strings"

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

	cursorCol      int // byte offset of the cursor within Value
	selAnchor      int // byte offset of the selection anchor; meaningful when hasSel
	hasSel         bool
	selecting      bool // a mouse-drag selection is in progress
	cursorVisible  bool
	cursorLastTick uint64
}

// textInputPadX is the gap between the field's left border and its text.
const textInputPadX = 6

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

// Blur clears focus but deliberately does not call SDL_StopTextInput: the
// window's text-input state is shared by every focusable widget on it, and
// each widget already gates character insertion on its own Focused flag, so
// leaving text input enabled is harmless. Calling StopTextInput here would
// race with another widget's Focus (called for the same click event) and
// could disable text input for whichever one just gained it.
func (t *TextInput) Blur() {
	t.Focused = false
}

func (t *TextInput) Update(event sdl.Event, mx, my float32) bool {
	switch event.Type() {
	case sdl.EventMouseButtonDown:
		inside := mx >= t.Bounds.X && mx <= t.Bounds.X+t.Bounds.W &&
			my >= t.Bounds.Y && my <= t.Bounds.Y+t.Bounds.H
		if inside {
			t.Focus()
			t.clampCursor()
			t.cursorCol = byteOffsetForX(t.Value, mx-(t.Bounds.X+textInputPadX), func(s string) float32 {
				return textPixelWidth(t.font, s)
			})
			t.selAnchor = t.cursorCol
			t.hasSel = false
			t.selecting = true
		} else if t.Focused {
			t.Blur()
		}
		return inside
	case sdl.EventMouseMotion:
		if !t.selecting {
			return false
		}
		t.cursorCol = byteOffsetForX(t.Value, mx-(t.Bounds.X+textInputPadX), func(s string) float32 {
			return textPixelWidth(t.font, s)
		})
		t.hasSel = t.cursorCol != t.selAnchor
		return true
	case sdl.EventMouseButtonUp:
		wasSelecting := t.selecting
		t.selecting = false
		return wasSelecting
	case sdl.EventTextInput:
		if !t.Focused {
			return false
		}
		textEvent := event.Text()
		t.clampCursor()
		t.deleteSelection()
		inserted := textEvent.Text()
		t.Value = t.Value[:t.cursorCol] + inserted + t.Value[t.cursorCol:]
		t.cursorCol += len(inserted)
		return true
	case sdl.EventKeyDown:
		if !t.Focused {
			return false
		}
		t.clampCursor()
		key := event.Key()
		if isCopyShortcut(key) || isCutShortcut(key) {
			if t.hasSel {
				lo, hi := orderInts(t.selAnchor, t.cursorCol)
				clipboardWrite(t.Value[lo:hi])
				if isCutShortcut(key) {
					t.deleteSelection()
				}
			}
			return true
		}
		if isPasteShortcut(key) {
			t.deleteSelection()
			// Single-line field: flatten pasted newlines to spaces.
			pasted := strings.NewReplacer("\r\n", " ", "\n", " ", "\r", " ").Replace(clipboardRead())
			t.Value = t.Value[:t.cursorCol] + pasted + t.Value[t.cursorCol:]
			t.cursorCol += len(pasted)
			return true
		}
		if isSelectAllShortcut(key) {
			t.selAnchor = 0
			t.cursorCol = len(t.Value)
			t.hasSel = t.cursorCol != t.selAnchor
			return true
		}
		scancode := key.Scancode
		shift := key.Mod&sdl.KeymodShift != 0
		move := isMovementKey(scancode)
		if t.hasSel {
			switch {
			case scancode == sdl.ScancodeBackspace || scancode == sdl.ScancodeDelete:
				t.deleteSelection()
				return true
			case move && !shift:
				// Unshifted arrows collapse the selection to one end
				// instead of moving the caret a further step.
				lo, hi := orderInts(t.selAnchor, t.cursorCol)
				t.hasSel = false
				if scancode == sdl.ScancodeLeft {
					t.cursorCol = lo
					return true
				}
				if scancode == sdl.ScancodeRight {
					t.cursorCol = hi
					return true
				}
				// Home/End fall through and move normally.
			}
		}
		if move && shift && !t.hasSel {
			t.selAnchor = t.cursorCol
			t.hasSel = true
		}
		// Home/End/Left/Right/Backspace/Delete are shared with TextArea.
		if newValue, newCol, handled := editLineKey(scancode, t.Value, t.cursorCol); handled {
			t.Value = newValue
			t.cursorCol = newCol
			if move {
				t.hasSel = shift && t.hasSel && t.selAnchor != t.cursorCol
			}
			return true
		}
		if move {
			t.hasSel = t.hasSel && t.selAnchor != t.cursorCol
		}
		if scancode == sdl.ScancodeReturn {
			if t.OnSubmit != nil {
				t.OnSubmit(t.Value)
			}
			return true
		}
	}
	return false
}

// deleteSelection removes the selected range from Value and collapses the
// caret to its start. Reports whether there was a selection to delete.
func (t *TextInput) deleteSelection() bool {
	if !t.hasSel {
		return false
	}
	t.hasSel = false
	t.Value, t.cursorCol = deleteRangeLine(t.Value, t.selAnchor, t.cursorCol)
	return true
}

// clampCursor keeps cursorCol inside Value, which callers may have replaced
// wholesale since the last event.
func (t *TextInput) clampCursor() {
	if t.cursorCol > len(t.Value) {
		t.cursorCol = len(t.Value)
	}
	if t.cursorCol < 0 {
		t.cursorCol = 0
	}
	if t.selAnchor > len(t.Value) {
		t.selAnchor = len(t.Value)
	}
	if t.selAnchor < 0 {
		t.selAnchor = 0
	}
	t.hasSel = t.hasSel && t.selAnchor != t.cursorCol
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

	t.clampCursor()
	if t.hasSel {
		lo, hi := orderInts(t.selAnchor, t.cursorCol)
		x0 := t.Bounds.X + textInputPadX + textPixelWidth(t.font, t.Value[:lo])
		x1 := t.Bounds.X + textInputPadX + textPixelWidth(t.font, t.Value[:hi])
		highlight := sdl.FRect{X: x0, Y: t.Bounds.Y + 4, W: x1 - x0, H: t.Bounds.H - 8}
		sdl.SetRenderDrawColor(renderer, 179, 212, 255, sdl.AlphaOpaque)
		sdl.RenderFillRect(renderer, &highlight)
	}

	if texture, textW, textH := makeTextTexture(renderer, t.font, t.Value, sdl.Color{R: 0, G: 0, B: 0, A: 255}); texture != nil {
		textRect := sdl.FRect{
			X: t.Bounds.X + 6,
			Y: t.Bounds.Y + (t.Bounds.H-textH)/2,
			W: textW,
			H: textH,
		}
		sdl.RenderTexture(renderer, texture, nil, &textRect)
		sdl.DestroyTexture(texture)
	}

	if t.Focused {
		now := sdl.GetTicks()
		if now-t.cursorLastTick > cursorBlinkMillis {
			t.cursorVisible = !t.cursorVisible
			t.cursorLastTick = now
		}
		if t.cursorVisible {
			t.clampCursor()
			cursorX := t.Bounds.X + 6 + textPixelWidth(t.font, t.Value[:t.cursorCol]) + 2
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
