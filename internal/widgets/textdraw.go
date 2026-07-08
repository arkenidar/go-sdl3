package widgets

import (
	"github.com/jupiterrider/purego-sdl3/sdl"
	"github.com/jupiterrider/purego-sdl3/ttf"
)

// makeTextTexture renders text into a new texture and returns it with its
// pixel size. Returns nil (and zero size) for empty text or render failure.
// The caller owns the texture and must destroy it. Shared by Label,
// TextInput, TextArea, Button and Checkbox so single-line text renders the
// same way everywhere.
func makeTextTexture(renderer *sdl.Renderer, font *ttf.Font, text string, color sdl.Color) (texture *sdl.Texture, w, h float32) {
	if text == "" {
		return nil, 0, 0
	}
	surface := ttf.RenderTextBlended(font, text, 0, color)
	if surface == nil {
		return nil, 0, 0
	}
	defer sdl.DestroySurface(surface)
	texture = sdl.CreateTextureFromSurface(renderer, surface)
	if texture == nil {
		return nil, 0, 0
	}
	sdl.GetTextureSize(texture, &w, &h)
	return texture, w, h
}

// textPixelWidth measures the rendered width of s in font.
func textPixelWidth(font *ttf.Font, s string) float32 {
	if s == "" {
		return 0
	}
	surface := ttf.RenderTextBlended(font, s, 0, sdl.Color{R: 0, G: 0, B: 0, A: 255})
	if surface == nil {
		return 0
	}
	defer sdl.DestroySurface(surface)
	return float32(surface.W)
}

// drawText immediately renders text with its top-left corner at (x, y),
// creating and destroying a throwaway texture. Returns the drawn size.
func drawText(renderer *sdl.Renderer, font *ttf.Font, text string, color sdl.Color, x, y float32) (w, h float32) {
	texture, w, h := makeTextTexture(renderer, font, text, color)
	if texture == nil {
		return 0, 0
	}
	defer sdl.DestroyTexture(texture)
	rect := sdl.FRect{X: x, Y: y, W: w, H: h}
	sdl.RenderTexture(renderer, texture, nil, &rect)
	return w, h
}
