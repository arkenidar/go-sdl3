// app.go
package main

// This is a simple SDL3 application written in Go
// It creates a window, handles events, and renders a rectangle

// use purego-sdl3 from jupiterrider
import (
	"fmt"

	"github.com/jupiterrider/purego-sdl3/sdl"
	"github.com/jupiterrider/purego-sdl3/ttf"
)

// Widget interface for UI elements
type Widget interface {
	Update(event sdl.Event, mx, my float32) bool // Returns true if event was handled
	Render(renderer *sdl.Renderer)
	GetBounds() sdl.FRect
}

// Button widget
type Button struct {
	Bounds    sdl.FRect
	Text      string
	Texture   *sdl.Texture
	OnClick   func()
	IsPressed bool
}

func NewButton(x, y, w, h float32, text string, font *ttf.Font, renderer *sdl.Renderer, onClick func()) *Button {
	// Create button text texture
	surface := ttf.RenderTextBlended(font, text, 0, sdl.Color{R: 255, G: 255, B: 255, A: 255})
	if surface == nil {
		panic(sdl.GetError())
	}
	defer sdl.DestroySurface(surface)

	texture := sdl.CreateTextureFromSurface(renderer, surface)
	if texture == nil {
		panic(sdl.GetError())
	}

	// Auto-size button based on text if width/height are 0
	var textW, textH float32
	sdl.GetTextureSize(texture, &textW, &textH)

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

func (b *Button) Destroy() {
	if b.Texture != nil {
		sdl.DestroyTexture(b.Texture)
		b.Texture = nil
	}
}

// Label widget for displaying text
type Label struct {
	Bounds   sdl.FRect
	Text     string
	Texture  *sdl.Texture
	font     *ttf.Font
	renderer *sdl.Renderer
}

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

func (l *Label) Destroy() {
	if l.Texture != nil {
		sdl.DestroyTexture(l.Texture)
		l.Texture = nil
	}
}

// Layout system
type Layout struct {
	X, Y    float32
	Spacing float32
	Widgets []Widget
}

func NewLayout(x, y, spacing float32) *Layout {
	return &Layout{X: x, Y: y, Spacing: spacing, Widgets: make([]Widget, 0)}
}

func (layout *Layout) AddWidget(widget Widget) {
	bounds := widget.GetBounds()

	// Position widget based on layout
	if len(layout.Widgets) == 0 {
		// First widget
		bounds.X = layout.X
		bounds.Y = layout.Y
	} else {
		// Position relative to previous widget
		lastBounds := layout.Widgets[len(layout.Widgets)-1].GetBounds()
		bounds.X = lastBounds.X + lastBounds.W + layout.Spacing
		bounds.Y = layout.Y
	}

	// Update widget bounds (this is a bit hacky, but works for our simple case)
	if btn, ok := widget.(*Button); ok {
		btn.Bounds = bounds
	} else if lbl, ok := widget.(*Label); ok {
		lbl.Bounds = bounds
	}

	layout.Widgets = append(layout.Widgets, widget)
}

func (layout *Layout) Update(event sdl.Event, mx, my float32) bool {
	for _, widget := range layout.Widgets {
		if widget.Update(event, mx, my) {
			return true
		}
	}
	return false
}

func (layout *Layout) Render(renderer *sdl.Renderer) {
	for _, widget := range layout.Widgets {
		widget.Render(renderer)
	}
}

func (layout *Layout) Destroy() {
	for _, widget := range layout.Widgets {
		if btn, ok := widget.(*Button); ok {
			btn.Destroy()
		} else if lbl, ok := widget.(*Label); ok {
			lbl.Destroy()
		}
	}
}

// Helper function to wrap text to fit within a given width
func wrapText(text string, font *ttf.Font, maxWidth float32) []string {
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

// Function to render text at bottom with centering and wrapping
func renderBottomText(renderer *sdl.Renderer, font *ttf.Font, text string, windowWidth, windowHeight, margin float32) {
	maxWidth := windowWidth - (margin * 2) // Available width for text
	lines := wrapText(text, font, maxWidth)

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

func main() {

	// SECTION : Initialize SDL

	defer sdl.Quit()
	if !sdl.Init(sdl.InitVideo) {
		panic(sdl.GetError())
	}

	// Initialize TTF
	defer ttf.Quit()
	if !ttf.Init() {
		panic(sdl.GetError())
	}

	// Load font
	font := ttf.OpenFont("assets/OpenDyslexic-Regular.ttf", 24)
	if font == nil {
		panic(sdl.GetError())
	}
	defer ttf.CloseFont(font)

	// Create a window and renderer
	var window *sdl.Window
	var renderer *sdl.Renderer
	if !sdl.CreateWindowAndRenderer("App built with Go and SDL3", 700, 500, sdl.WindowResizable, &window, &renderer) {
		panic(sdl.GetError())
	}
	defer sdl.DestroyRenderer(renderer)
	defer sdl.DestroyWindow(window)

	// SECTION : Application state
	x, y := float32(150), float32(150)
	counter := 0
	showAlert := false
	alertMessage := "Button clicked! This is a longer message that will demonstrate the text wrapping functionality in alert dialogs."

	// Window dimensions (will be updated on resize)
	windowWidth := float32(700)
	windowHeight := float32(500)

	// Create UI layout with buttons and counter (positioned at top)
	uiLayout := NewLayout(10, 10, 10)
	defer uiLayout.Destroy()

	// Create buttons with callbacks (auto-sized)
	plusButton := NewButton(0, 0, 0, 0, "+", font, renderer, func() {
		counter++
	})
	minusButton := NewButton(0, 0, 0, 0, "-", font, renderer, func() {
		counter--
	})

	// Create counter label
	counterLabel := NewLabel(0, 0, fmt.Sprintf("Counter: %d", counter), font, renderer)

	// Add widgets to main layout
	uiLayout.AddWidget(plusButton)
	uiLayout.AddWidget(minusButton)
	uiLayout.AddWidget(counterLabel)

	// Create a right-aligned button (demonstration of extensibility - auto-sized)
	newButton := NewButton(0, 0, 0, 0, "Click Me", font, renderer, func() {
		showAlert = true
	})
	// Position the button to the right border using dynamic window width
	buttonBounds := newButton.GetBounds()
	newButton.Bounds.X = windowWidth - buttonBounds.W - 10 // 10px margin from right edge
	newButton.Bounds.Y = 10                                // Align with the top button row

	// Drag state variables
	dragging := false
	dragOffsetX, dragOffsetY := float32(0), float32(0)

Outer:
	for {
		var event sdl.Event
		for sdl.PollEvent(&event) {
			mx := float32(0)
			my := float32(0)

			// Get mouse position for widgets
			if event.Type() == sdl.EventMouseButtonDown || event.Type() == sdl.EventMouseButtonUp {
				mx = float32(event.Button().X)
				my = float32(event.Button().Y)
			} else if event.Type() == sdl.EventMouseMotion {
				mx = float32(event.Motion().X)
				my = float32(event.Motion().Y)
			}

			switch event.Type() {
			case sdl.EventQuit:
				break Outer
			case sdl.EventWindowResized:
				windowWidth = float32(event.Window().Data1)
				windowHeight = float32(event.Window().Data2)

				// Reposition right-aligned button when window resizes
				buttonBounds := newButton.GetBounds()
				newButton.Bounds.X = windowWidth - buttonBounds.W - 10 // 10px margin from right edge
				
				// Keep square within new window bounds
				if x < 0 {
					x = 0
				}
				if y < 0 {
					y = 0
				}
				if x + 100 > windowWidth {
					x = windowWidth - 100
				}
				if y + 100 > windowHeight {
					y = windowHeight - 100
				}
			case sdl.EventKeyDown:
				switch event.Key().Scancode {
				case sdl.ScancodeEscape:
					if showAlert {
						showAlert = false // Dismiss alert first
					} else {
						break Outer // Exit application
					}
				case sdl.ScancodeSpace:
					if showAlert {
						showAlert = false // Dismiss alert with spacebar
					}
				case sdl.ScancodeRight:
					x += 15
					if x+100 > windowWidth {
						x = windowWidth - 100
					}
				case sdl.ScancodeLeft:
					x -= 15
					if x < 0 {
						x = 0
					}
				case sdl.ScancodeDown:
					y += 15
					if y+100 > windowHeight {
						y = windowHeight - 100
					}
				case sdl.ScancodeUp:
					y -= 15
					if y < 0 {
						y = 0
					}
				}
			case sdl.EventMouseButtonDown:
				// Check if alert is showing and handle click-to-close
				if showAlert {
					showAlert = false // Dismiss alert on any click
				} else {
					// Check if UI layout handled the event first
					if !uiLayout.Update(event, mx, my) {
						// Check if right-aligned button handled the event
						if !newButton.Update(event, mx, my) {
							// Check if mouse is inside the square for dragging
							if mx >= x && mx <= x+100 && my >= y && my <= y+100 {
								dragging = true
								dragOffsetX = mx - x
								dragOffsetY = my - y
							}
						}
					}
				}
			case sdl.EventMouseButtonUp:
				uiLayout.Update(event, mx, my)
				newButton.Update(event, mx, my) // Handle button release for right-aligned button
				dragging = false

				// Update counter display if counter changed
				newCounterText := fmt.Sprintf("Counter: %d", counter)
				if newCounterText != counterLabel.Text {
					counterLabel.UpdateText(newCounterText)
				}
			case sdl.EventMouseMotion:
				if dragging {
					x = mx - dragOffsetX
					y = my - dragOffsetY

					// Keep square within window bounds
					if x < 0 {
						x = 0
					}
					if y < 0 {
						y = 0
					}
					if x+100 > windowWidth {
						x = windowWidth - 100
					}
					if y+100 > windowHeight {
						y = windowHeight - 100
					}
				}
			}
		}

		// SECTION : Rendering
		sdl.SetRenderDrawColor(renderer, 100, 150, 200, sdl.AlphaOpaque)
		sdl.RenderClear(renderer)

		// Draw rectangle
		rect := sdl.FRect{X: x, Y: y, W: 100, H: 100}
		sdl.SetRenderDrawColor(renderer, 0, 0, 200, sdl.AlphaOpaque)
		sdl.RenderFillRect(renderer, &rect)

		// Render UI elements
		uiLayout.Render(renderer)
		newButton.Render(renderer) // Render the right-aligned button separately

		// Render instruction text at bottom with centering and wrapping
		renderBottomText(renderer, font, "• move the blue square with arrow keys or mouse drag\n • click its buttons to change counter", windowWidth, windowHeight, 10)

		// Render alert if active
		if showAlert {
			// Calculate available width for alert text (with padding)
			maxAlertWidth := windowWidth * 0.8 // Use 80% of window width max
			if maxAlertWidth < 200 {
				maxAlertWidth = 200 // Minimum width
			}

			// Wrap alert text and dismiss text
			alertLines := wrapText(alertMessage, font, maxAlertWidth-40) // Subtract padding
			dismissLines := wrapText("Press ESC/SPACE or click to close", font, maxAlertWidth-40)

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
				surface := ttf.RenderTextBlended(font, line, 0, sdl.Color{R: 0, G: 0, B: 0, A: 255})
				if surface != nil {
					texture := sdl.CreateTextureFromSurface(renderer, surface)
					if texture != nil {
						var textW, textH float32
						sdl.GetTextureSize(texture, &textW, &textH)

						// Center the line horizontally within the alert box
						textX := alertBox.X + (alertBox.W-textW)/2

						alertTextRect := sdl.FRect{X: textX, Y: currentY, W: textW, H: textH}
						sdl.RenderTexture(renderer, texture, nil, &alertTextRect)

						sdl.DestroyTexture(texture)
					}
					sdl.DestroySurface(surface)
				}
				currentY += lineHeight
			}

			// Add spacing between alert text and dismiss text
			currentY += 20

			// Render dismiss instruction lines (centered)
			for _, line := range dismissLines {
				surface := ttf.RenderTextBlended(font, line, 0, sdl.Color{R: 0, G: 0, B: 0, A: 255})
				if surface != nil {
					texture := sdl.CreateTextureFromSurface(renderer, surface)
					if texture != nil {
						var textW, textH float32
						sdl.GetTextureSize(texture, &textW, &textH)

						// Center the line horizontally within the alert box
						textX := alertBox.X + (alertBox.W-textW)/2

						dismissTextRect := sdl.FRect{X: textX, Y: currentY, W: textW, H: textH}
						sdl.RenderTexture(renderer, texture, nil, &dismissTextRect)

						sdl.DestroyTexture(texture)
					}
					sdl.DestroySurface(surface)
				}
				currentY += lineHeight
			}
		}

		sdl.RenderPresent(renderer)
	}
}
