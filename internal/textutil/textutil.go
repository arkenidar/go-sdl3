// Package textutil provides text wrapping and centered-block rendering helpers for SDL3/TTF.
package textutil

import (
	"github.com/jupiterrider/purego-sdl3/sdl"
	"github.com/jupiterrider/purego-sdl3/ttf"
)

// WrapText wraps text to fit within a given width.
func WrapText(text string, font *ttf.Font, maxWidth float32) []string {
	// First split by explicit newlines
	paragraphs := []string{}
	currentParagraph := ""

	for _, char := range text {
		if char == '\n' {
			if currentParagraph != "" {
				paragraphs = append(paragraphs, currentParagraph)
				currentParagraph = ""
			}
		} else {
			currentParagraph += string(char)
		}
	}
	if currentParagraph != "" {
		paragraphs = append(paragraphs, currentParagraph)
	}

	// If no explicit newlines, treat the whole text as one paragraph
	if len(paragraphs) == 0 && text != "" {
		paragraphs = append(paragraphs, text)
	}

	// Now wrap each paragraph
	allLines := []string{}
	for _, paragraph := range paragraphs {
		// Split paragraph into words
		words := []string{}
		currentWord := ""

		for _, char := range paragraph {
			if char == ' ' {
				if currentWord != "" {
					words = append(words, currentWord)
					currentWord = ""
				}
			} else {
				currentWord += string(char)
			}
		}
		if currentWord != "" {
			words = append(words, currentWord)
		}

		// Wrap words in this paragraph
		currentLine := ""
		for _, word := range words {
			testLine := currentLine
			if testLine != "" {
				testLine += " "
			}
			testLine += word

			// Create a temporary surface to measure text width
			surface := ttf.RenderTextBlended(font, testLine, 0, sdl.Color{R: 255, G: 255, B: 255, A: 255})
			if surface != nil {
				textW := float32(surface.W)
				sdl.DestroySurface(surface)

				if textW <= maxWidth {
					currentLine = testLine
				} else {
					// Word doesn't fit, start new line
					if currentLine != "" {
						allLines = append(allLines, currentLine)
					}
					currentLine = word
				}
			}
		}

		if currentLine != "" {
			allLines = append(allLines, currentLine)
		}
	}

	return allLines
}

// RenderBottomText renders text at the bottom of the window with centering and wrapping.
func RenderBottomText(renderer *sdl.Renderer, font *ttf.Font, text string, windowWidth, windowHeight, margin float32) {
	maxWidth := windowWidth - (margin * 2) // Available width for text
	lines := WrapText(text, font, maxWidth)

	if len(lines) == 0 {
		return
	}

	// Calculate total height needed for all lines
	lineHeight := float32(0)
	if len(lines) > 0 {
		surface := ttf.RenderTextBlended(font, lines[0], 0, sdl.Color{R: 255, G: 255, B: 255, A: 255})
		if surface != nil {
			lineHeight = float32(surface.H)
			sdl.DestroySurface(surface)
		}
	}

	totalHeight := lineHeight * float32(len(lines))
	startY := windowHeight - totalHeight - margin

	// Ensure text doesn't go above the window
	if startY < margin {
		startY = margin
	}

	// Render each line
	for i, line := range lines {
		surface := ttf.RenderTextBlended(font, line, 0, sdl.Color{R: 255, G: 255, B: 255, A: 255})
		if surface != nil {
			texture := sdl.CreateTextureFromSurface(renderer, surface)
			if texture != nil {
				var textW, textH float32
				sdl.GetTextureSize(texture, &textW, &textH)

				// Center the line horizontally
				x := (windowWidth - textW) / 2
				if x < margin {
					x = margin
				}

				y := startY + (float32(i) * lineHeight)

				textRect := sdl.FRect{X: x, Y: y, W: textW, H: textH}
				sdl.RenderTexture(renderer, texture, nil, &textRect)

				sdl.DestroyTexture(texture)
			}
			sdl.DestroySurface(surface)
		}
	}
}
