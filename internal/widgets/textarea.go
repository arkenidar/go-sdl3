package widgets

import (
	"github.com/jupiterrider/purego-sdl3/sdl"
	"github.com/jupiterrider/purego-sdl3/ttf"

	"arkenidar.com/purego-sdl3/internal/textutil"
)

// TextArea is a multi-line editable text field. It scrolls both
// horizontally and vertically over its logical lines, and can optionally
// soft-wrap long lines to its width instead of scrolling horizontally.
type TextArea struct {
	Bounds   sdl.FRect
	Lines    []string
	WordWrap bool
	Focused  bool

	cursorRow, cursorCol int
	scrollX, scrollY     float32

	// contentHeight/maxLineWidth are recomputed every Render call and used
	// to clamp scrolling and draw proportional scrollbar thumbs.
	contentHeight float32
	maxLineWidth  float32

	// draggingV/draggingH track an in-progress scrollbar-thumb drag;
	// dragGrabOffset is the pointer's offset from the thumb's top/left
	// edge at the moment the drag started, so the thumb doesn't jump.
	draggingV, draggingH bool
	dragGrabOffset       float32

	window   *sdl.Window
	font     *ttf.Font
	renderer *sdl.Renderer
	lineSkip float32

	cursorVisible  bool
	cursorLastTick uint64
}

const (
	textAreaPadding    = 6
	textAreaWheelPxY   = 24
	textAreaWheelPxX   = 24
	textAreaScrollbarW = 14 // visible thickness of the scrollbar track/thumb
	textAreaHitPad     = 6  // extra invisible margin added to the clickable/draggable area
	textAreaMinThumb   = 32
)

// NewTextArea creates an unfocused, empty TextArea at (x, y) sized (w, h).
// Call Focus to start accepting keystrokes.
func NewTextArea(x, y, w, h float32, font *ttf.Font, renderer *sdl.Renderer, window *sdl.Window) *TextArea {
	return &TextArea{
		Bounds:         sdl.FRect{X: x, Y: y, W: w, H: h},
		Lines:          []string{""},
		window:         window,
		font:           font,
		renderer:       renderer,
		lineSkip:       float32(ttf.GetFontLineSkip(font)),
		cursorLastTick: sdl.GetTicks(),
	}
}

func (t *TextArea) Focus() {
	if t.Focused {
		return
	}
	t.Focused = true
	t.cursorVisible = true
	t.cursorLastTick = sdl.GetTicks()
	sdl.StartTextInput(t.window)
}

func (t *TextArea) Blur() {
	if !t.Focused {
		return
	}
	t.Focused = false
	sdl.StopTextInput(t.window)
}

// ToggleWordWrap flips soft-wrap mode. Horizontal scroll is reset since it
// doesn't apply while wrapping.
func (t *TextArea) ToggleWordWrap() {
	t.WordWrap = !t.WordWrap
	t.scrollX = 0
}

func (t *TextArea) inside(mx, my float32) bool {
	return mx >= t.Bounds.X && mx <= t.Bounds.X+t.Bounds.W &&
		my >= t.Bounds.Y && my <= t.Bounds.Y+t.Bounds.H
}

func (t *TextArea) Update(event sdl.Event, mx, my float32) bool {
	switch event.Type() {
	case sdl.EventMouseButtonDown:
		if track, thumb, visible := t.vScrollbarGeom(); visible && rectContains(expandRect(track, textAreaHitPad), mx, my) {
			t.draggingV = true
			if rectContains(expandRect(thumb, textAreaHitPad), mx, my) {
				t.dragGrabOffset = my - thumb.Y
			} else {
				t.dragGrabOffset = thumb.H / 2
				t.scrollVToPointer(my, track, thumb)
			}
			return true
		}
		if track, thumb, visible := t.hScrollbarGeom(); visible && rectContains(expandRect(track, textAreaHitPad), mx, my) {
			t.draggingH = true
			if rectContains(expandRect(thumb, textAreaHitPad), mx, my) {
				t.dragGrabOffset = mx - thumb.X
			} else {
				t.dragGrabOffset = thumb.W / 2
				t.scrollHToPointer(mx, track, thumb)
			}
			return true
		}

		in := t.inside(mx, my)
		if in {
			t.Focus()
		} else if t.Focused {
			t.Blur()
		}
		return in
	case sdl.EventMouseMotion:
		if t.draggingV {
			track, thumb, _ := t.vScrollbarGeom()
			t.scrollVToPointer(my, track, thumb)
			return true
		}
		if t.draggingH {
			track, thumb, _ := t.hScrollbarGeom()
			t.scrollHToPointer(mx, track, thumb)
			return true
		}
		return false
	case sdl.EventMouseButtonUp:
		wasDragging := t.draggingV || t.draggingH
		t.draggingV = false
		t.draggingH = false
		return wasDragging
	case sdl.EventMouseWheel:
		if !t.inside(mx, my) {
			return false
		}
		wheel := event.Wheel()
		viewportH := t.Bounds.H - 2*textAreaPadding
		t.scrollY = clamp(t.scrollY-wheel.Y*textAreaWheelPxY, 0, maxFloat(0, t.contentHeight-viewportH))
		if !t.WordWrap {
			viewportW := t.Bounds.W - 2*textAreaPadding
			t.scrollX = clamp(t.scrollX-wheel.X*textAreaWheelPxX, 0, maxFloat(0, t.maxLineWidth-viewportW))
		}
		return true
	case sdl.EventTextInput:
		if !t.Focused {
			return false
		}
		textEvent := event.Text()
		t.insertAtCursor(textEvent.Text())
		return true
	case sdl.EventKeyDown:
		if !t.Focused {
			return false
		}
		return t.handleKey(event.Key().Scancode)
	}
	return false
}

// scrollVToPointer sets scrollY so the thumb's top follows the pointer,
// honoring the grab offset captured when the drag started.
func (t *TextArea) scrollVToPointer(my float32, track, thumb sdl.FRect) {
	viewportH := t.Bounds.H - 2*textAreaPadding
	maxScroll := maxFloat(0, t.contentHeight-viewportH)
	span := track.H - thumb.H
	if span <= 0 {
		return
	}
	thumbTop := clamp(my-t.dragGrabOffset, track.Y, track.Y+span)
	t.scrollY = clamp((thumbTop-track.Y)/span*maxScroll, 0, maxScroll)
}

// scrollHToPointer is the horizontal counterpart of scrollVToPointer.
func (t *TextArea) scrollHToPointer(mx float32, track, thumb sdl.FRect) {
	viewportW := t.Bounds.W - 2*textAreaPadding
	maxScroll := maxFloat(0, t.maxLineWidth-viewportW)
	span := track.W - thumb.W
	if span <= 0 {
		return
	}
	thumbLeft := clamp(mx-t.dragGrabOffset, track.X, track.X+span)
	t.scrollX = clamp((thumbLeft-track.X)/span*maxScroll, 0, maxScroll)
}

func (t *TextArea) handleKey(scancode sdl.Scancode) bool {
	line := t.Lines[t.cursorRow]

	switch scancode {
	case sdl.ScancodeBackspace:
		if t.cursorCol > 0 {
			t.Lines[t.cursorRow] = line[:t.cursorCol-1] + line[t.cursorCol:]
			t.cursorCol--
		} else if t.cursorRow > 0 {
			prevLen := len(t.Lines[t.cursorRow-1])
			t.Lines[t.cursorRow-1] += line
			t.Lines = append(t.Lines[:t.cursorRow], t.Lines[t.cursorRow+1:]...)
			t.cursorRow--
			t.cursorCol = prevLen
		}
	case sdl.ScancodeReturn:
		before, after := line[:t.cursorCol], line[t.cursorCol:]
		t.Lines[t.cursorRow] = before
		t.Lines = append(t.Lines[:t.cursorRow+1], append([]string{after}, t.Lines[t.cursorRow+1:]...)...)
		t.cursorRow++
		t.cursorCol = 0
	case sdl.ScancodeLeft:
		if t.cursorCol > 0 {
			t.cursorCol--
		} else if t.cursorRow > 0 {
			t.cursorRow--
			t.cursorCol = len(t.Lines[t.cursorRow])
		}
	case sdl.ScancodeRight:
		if t.cursorCol < len(line) {
			t.cursorCol++
		} else if t.cursorRow < len(t.Lines)-1 {
			t.cursorRow++
			t.cursorCol = 0
		}
	case sdl.ScancodeUp:
		if t.cursorRow > 0 {
			t.cursorRow--
			t.cursorCol = min(t.cursorCol, len(t.Lines[t.cursorRow]))
		}
	case sdl.ScancodeDown:
		if t.cursorRow < len(t.Lines)-1 {
			t.cursorRow++
			t.cursorCol = min(t.cursorCol, len(t.Lines[t.cursorRow]))
		}
	case sdl.ScancodeHome:
		t.cursorCol = 0
	case sdl.ScancodeEnd:
		t.cursorCol = len(line)
	default:
		return false
	}
	return true
}

func clamp(v, lo, hi float32) float32 {
	if v < lo {
		return lo
	}
	if v > hi {
		return hi
	}
	return v
}

func maxFloat(a, b float32) float32 {
	if a > b {
		return a
	}
	return b
}

func (t *TextArea) insertAtCursor(s string) {
	line := t.Lines[t.cursorRow]
	t.Lines[t.cursorRow] = line[:t.cursorCol] + s + line[t.cursorCol:]
	t.cursorCol += len(s)
}

func (t *TextArea) GetBounds() sdl.FRect {
	return t.Bounds
}

func (t *TextArea) SetBounds(bounds sdl.FRect) {
	t.Bounds = bounds
}

func (t *TextArea) textWidth(s string) float32 {
	if s == "" {
		return 0
	}
	surface := ttf.RenderTextBlended(t.font, s, 0, sdl.Color{R: 0, G: 0, B: 0, A: 255})
	if surface == nil {
		return 0
	}
	defer sdl.DestroySurface(surface)
	return float32(surface.W)
}

// recomputeMetrics recalculates contentHeight and maxLineWidth from the
// current Lines/WordWrap/Bounds. It must run before clampScroll and before
// scrollX/scrollY are used to position anything, since either can change
// out from under a stale scroll offset (word-wrap toggled, a resize, an
// edit that removes lines).
func (t *TextArea) recomputeMetrics() {
	if t.WordWrap {
		var displayLines int
		for _, line := range t.Lines {
			wrapped := textutil.WrapText(line, t.font, t.Bounds.W-2*textAreaPadding)
			displayLines += maxInt(1, len(wrapped))
		}
		t.contentHeight = float32(displayLines) * t.lineSkip
		t.maxLineWidth = 0
	} else {
		t.contentHeight = float32(len(t.Lines)) * t.lineSkip
		t.maxLineWidth = 0
		for _, line := range t.Lines {
			t.maxLineWidth = maxFloat(t.maxLineWidth, t.textWidth(line))
		}
	}
}

// clampScroll keeps scrollX/scrollY within [0, content-viewport], using the
// metrics recomputeMetrics just produced.
func (t *TextArea) clampScroll() {
	viewportH := t.Bounds.H - 2*textAreaPadding
	t.scrollY = clamp(t.scrollY, 0, maxFloat(0, t.contentHeight-viewportH))

	if t.WordWrap {
		t.scrollX = 0
		return
	}
	viewportW := t.Bounds.W - 2*textAreaPadding
	t.scrollX = clamp(t.scrollX, 0, maxFloat(0, t.maxLineWidth-viewportW))
}

func maxInt(a, b int) int {
	if a > b {
		return a
	}
	return b
}

func (t *TextArea) drawLine(renderer *sdl.Renderer, text string, x, y float32) {
	if text == "" {
		return
	}
	surface := ttf.RenderTextBlended(t.font, text, 0, sdl.Color{R: 20, G: 20, B: 20, A: 255})
	if surface == nil {
		return
	}
	texture := sdl.CreateTextureFromSurface(renderer, surface)
	sdl.DestroySurface(surface)
	if texture == nil {
		return
	}
	defer sdl.DestroyTexture(texture)

	var w, h float32
	sdl.GetTextureSize(texture, &w, &h)
	rect := sdl.FRect{X: x, Y: y, W: w, H: h}
	sdl.RenderTexture(renderer, texture, nil, &rect)
}

func (t *TextArea) Render(renderer *sdl.Renderer) {
	sdl.SetRenderDrawColor(renderer, 250, 250, 250, sdl.AlphaOpaque)
	sdl.RenderFillRect(renderer, &t.Bounds)
	if t.Focused {
		sdl.SetRenderDrawColor(renderer, 60, 130, 220, sdl.AlphaOpaque)
	} else {
		sdl.SetRenderDrawColor(renderer, 120, 120, 120, sdl.AlphaOpaque)
	}
	sdl.RenderRect(renderer, &t.Bounds)

	clip := sdl.Rect{
		X: int32(t.Bounds.X), Y: int32(t.Bounds.Y),
		W: int32(t.Bounds.W), H: int32(t.Bounds.H),
	}
	sdl.SetRenderClipRect(renderer, &clip)

	// Recompute metrics and clamp scroll *before* drawing: content shape can
	// change between frames (word-wrap toggled, a resize, an edit that
	// removes lines) and a stale scroll offset would otherwise point past
	// the new content, drawing nothing and putting the scrollbar thumb in
	// an inconsistent position.
	t.recomputeMetrics()
	t.clampScroll()

	originX := t.Bounds.X + textAreaPadding - t.scrollX
	originY := t.Bounds.Y + textAreaPadding - t.scrollY

	var cursorX, cursorY float32
	cursorFound := false
	y := originY

	for row, line := range t.Lines {
		if t.WordWrap {
			wrapped := textutil.WrapText(line, t.font, t.Bounds.W-2*textAreaPadding)
			if len(wrapped) == 0 {
				wrapped = []string{""}
			}
			consumed := 0
			for _, sub := range wrapped {
				t.drawLine(renderer, sub, originX, y)
				if row == t.cursorRow && !cursorFound && t.cursorCol >= consumed && t.cursorCol <= consumed+len(sub) {
					cursorX = originX + t.textWidth(sub[:t.cursorCol-consumed])
					cursorY = y
					cursorFound = true
				}
				consumed += len(sub)
				y += t.lineSkip
			}
		} else {
			t.drawLine(renderer, line, originX, y)
			if row == t.cursorRow {
				cursorX = originX + t.textWidth(line[:t.cursorCol])
				cursorY = y
				cursorFound = true
			}
			y += t.lineSkip
		}
	}

	if t.Focused {
		now := sdl.GetTicks()
		if now-t.cursorLastTick > cursorBlinkMillis {
			t.cursorVisible = !t.cursorVisible
			t.cursorLastTick = now
		}
		if t.cursorVisible && cursorFound {
			cursor := sdl.FRect{X: cursorX, Y: cursorY, W: 2, H: t.lineSkip}
			sdl.SetRenderDrawColor(renderer, 0, 0, 0, sdl.AlphaOpaque)
			sdl.RenderFillRect(renderer, &cursor)
		}
	}

	sdl.SetRenderClipRect(renderer, nil)

	t.renderScrollbars(renderer)
}

// vScrollbarGeom returns the vertical scrollbar's track/thumb rects and
// whether it should be shown at all (content taller than the viewport).
func (t *TextArea) vScrollbarGeom() (track, thumb sdl.FRect, visible bool) {
	viewportH := t.Bounds.H - 2*textAreaPadding
	if t.contentHeight <= viewportH {
		return track, thumb, false
	}
	trackH := t.Bounds.H
	thumbH := maxFloat(textAreaMinThumb, trackH*viewportH/t.contentHeight)
	maxScroll := t.contentHeight - viewportH
	thumbY := t.Bounds.Y
	if maxScroll > 0 {
		thumbY += (t.scrollY / maxScroll) * (trackH - thumbH)
	}
	track = sdl.FRect{X: t.Bounds.X + t.Bounds.W - textAreaScrollbarW, Y: t.Bounds.Y, W: textAreaScrollbarW, H: trackH}
	thumb = sdl.FRect{X: track.X, Y: thumbY, W: textAreaScrollbarW, H: thumbH}
	return track, thumb, true
}

// hScrollbarGeom is the horizontal counterpart of vScrollbarGeom; it never
// shows while WordWrap is on since there's nothing to scroll sideways.
func (t *TextArea) hScrollbarGeom() (track, thumb sdl.FRect, visible bool) {
	if t.WordWrap {
		return track, thumb, false
	}
	viewportW := t.Bounds.W - 2*textAreaPadding
	if t.maxLineWidth <= viewportW {
		return track, thumb, false
	}
	trackW := t.Bounds.W
	thumbW := maxFloat(textAreaMinThumb, trackW*viewportW/t.maxLineWidth)
	maxScroll := t.maxLineWidth - viewportW
	thumbX := t.Bounds.X
	if maxScroll > 0 {
		thumbX += (t.scrollX / maxScroll) * (trackW - thumbW)
	}
	track = sdl.FRect{X: t.Bounds.X, Y: t.Bounds.Y + t.Bounds.H - textAreaScrollbarW, W: trackW, H: textAreaScrollbarW}
	thumb = sdl.FRect{X: thumbX, Y: track.Y, W: thumbW, H: textAreaScrollbarW}
	return track, thumb, true
}

func rectContains(r sdl.FRect, x, y float32) bool {
	return x >= r.X && x <= r.X+r.W && y >= r.Y && y <= r.Y+r.H
}

// expandRect grows r by pad on every side, used to give scrollbar thumbs a
// larger invisible grab area than their drawn size.
func expandRect(r sdl.FRect, pad float32) sdl.FRect {
	return sdl.FRect{X: r.X - pad, Y: r.Y - pad, W: r.W + 2*pad, H: r.H + 2*pad}
}

// renderScrollbars draws vertical/horizontal scrollbar tracks and thumbs
// along the widget's right/bottom edges, only when content overflows.
func (t *TextArea) renderScrollbars(renderer *sdl.Renderer) {
	if track, thumb, visible := t.vScrollbarGeom(); visible {
		sdl.SetRenderDrawColor(renderer, 210, 210, 210, sdl.AlphaOpaque)
		sdl.RenderFillRect(renderer, &track)
		sdl.SetRenderDrawColor(renderer, 130, 130, 130, sdl.AlphaOpaque)
		sdl.RenderFillRect(renderer, &thumb)
	}
	if track, thumb, visible := t.hScrollbarGeom(); visible {
		sdl.SetRenderDrawColor(renderer, 210, 210, 210, sdl.AlphaOpaque)
		sdl.RenderFillRect(renderer, &track)
		sdl.SetRenderDrawColor(renderer, 130, 130, 130, sdl.AlphaOpaque)
		sdl.RenderFillRect(renderer, &thumb)
	}
}
