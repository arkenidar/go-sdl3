// ball.go
package main

import (
	"github.com/jupiterrider/purego-sdl3/sdl"
)

// Ball is a small circular sprite that moves in a straight line and
// bounces off the window edges. Its Color tints a shared white circle
// texture (assets/ball.bmp) via SetTextureColorMod rather than each ball
// owning its own texture.
type Ball struct {
	X, Y   float32
	VX, VY float32
	Size   float32
	Color  sdl.Color
}

// Update moves the ball by dtSeconds and reflects its velocity off the
// window bounds (windowWidth, windowHeight).
func (b *Ball) Update(dtSeconds, windowWidth, windowHeight float32) {
	b.X += b.VX * dtSeconds
	b.Y += b.VY * dtSeconds

	if b.X < 0 {
		b.X = 0
		b.VX = -b.VX
	} else if b.X+b.Size > windowWidth {
		b.X = windowWidth - b.Size
		b.VX = -b.VX
	}

	if b.Y < 0 {
		b.Y = 0
		b.VY = -b.VY
	} else if b.Y+b.Size > windowHeight {
		b.Y = windowHeight - b.Size
		b.VY = -b.VY
	}
}

// Draw tints texture with the ball's color and stretches it over the
// ball's current position/size.
func (b *Ball) Draw(renderer *sdl.Renderer, texture *sdl.Texture) {
	sdl.SetTextureColorMod(texture, b.Color.R, b.Color.G, b.Color.B)
	rect := sdl.FRect{X: b.X, Y: b.Y, W: b.Size, H: b.Size}
	sdl.RenderTexture(renderer, texture, nil, &rect)
}
