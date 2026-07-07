// alert.go
package main

import (
	"github.com/jupiterrider/purego-sdl3/sdl"
	"github.com/jupiterrider/purego-sdl3/ttf"

	"arkenidar.com/purego-sdl3/internal/textutil"
)

// renderAlert draws the modal alert dialog and its dismiss instructions.
func (app *App) renderAlert() {
	renderer := app.renderer
	font := app.font
	windowWidth := app.windowWidth
	windowHeight := app.windowHeight

	// Calculate available width for alert text (with padding)
	maxAlertWidth := windowWidth * 0.8 // Use 80% of window width max
	if maxAlertWidth < 200 {
		maxAlertWidth = 200 // Minimum width
	}

	// Wrap alert text and dismiss text
	alertLines := textutil.WrapText(app.alertMessage, font, maxAlertWidth-40) // Subtract padding
	dismissLines := textutil.WrapText("Press ESC/SPACE or click to close", font, maxAlertWidth-40)

	// Calculate dimensions for wrapped text
	var lineHeight float32
	if len(alertLines) > 0 {
		surface := ttf.RenderTextBlended(font, alertLines[0], 0, sdl.Color{R: 0, G: 0, B: 0, A: 255})
		if surface != nil {
			lineHeight = float32(surface.H)
			sdl.DestroySurface(surface)
		}
	}

	// Find the widest line to determine alert box width
	var maxLineWidth float32
	allLines := append(alertLines, dismissLines...)
	for _, line := range allLines {
		surface := ttf.RenderTextBlended(font, line, 0, sdl.Color{R: 0, G: 0, B: 0, A: 255})
		if surface != nil {
			lineWidth := float32(surface.W)
			if lineWidth > maxLineWidth {
				maxLineWidth = lineWidth
			}
			sdl.DestroySurface(surface)
		}
	}

	// Calculate alert box dimensions
	alertBoxW := maxLineWidth + 40 // 20px padding on each side
	totalTextHeight := lineHeight * float32(len(alertLines)+len(dismissLines))
	alertBoxH := totalTextHeight + 60           // Text heights + spacing + padding
	alertBoxX := (windowWidth - alertBoxW) / 2  // Center horizontally
	alertBoxY := (windowHeight - alertBoxH) / 2 // Center vertically

	// Semi-transparent overlay
	sdl.SetRenderDrawColor(renderer, 0, 0, 0, 128)
	overlay := sdl.FRect{X: 0, Y: 0, W: windowWidth, H: windowHeight}
	sdl.RenderFillRect(renderer, &overlay)

	// Auto-sized alert box
	alertBox := sdl.FRect{X: alertBoxX, Y: alertBoxY, W: alertBoxW, H: alertBoxH}
	sdl.SetRenderDrawColor(renderer, 200, 200, 200, sdl.AlphaOpaque)
	sdl.RenderFillRect(renderer, &alertBox)

	// Alert box border
	sdl.SetRenderDrawColor(renderer, 100, 100, 100, sdl.AlphaOpaque)
	sdl.RenderRect(renderer, &alertBox)

	// Render alert text lines (centered)
	currentY := alertBox.Y + 20
	for _, line := range alertLines {
		currentY = renderCenteredLine(renderer, font, line, alertBox, currentY, lineHeight)
	}

	// Add spacing between alert text and dismiss text
	currentY += 20

	// Render dismiss instruction lines (centered)
	for _, line := range dismissLines {
		currentY = renderCenteredLine(renderer, font, line, alertBox, currentY, lineHeight)
	}
}

// renderCenteredLine renders a single line of text horizontally centered within box,
// at the given y position, and returns the y position for the next line.
func renderCenteredLine(renderer *sdl.Renderer, font *ttf.Font, line string, box sdl.FRect, y, lineHeight float32) float32 {
	surface := ttf.RenderTextBlended(font, line, 0, sdl.Color{R: 0, G: 0, B: 0, A: 255})
	if surface == nil {
		return y + lineHeight
	}
	defer sdl.DestroySurface(surface)

	texture := sdl.CreateTextureFromSurface(renderer, surface)
	if texture == nil {
		return y + lineHeight
	}
	defer sdl.DestroyTexture(texture)

	var textW, textH float32
	sdl.GetTextureSize(texture, &textW, &textH)

	textX := box.X + (box.W-textW)/2
	textRect := sdl.FRect{X: textX, Y: y, W: textW, H: textH}
	sdl.RenderTexture(renderer, texture, nil, &textRect)

	return y + lineHeight
}
