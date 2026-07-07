// app.go
package main

import (
	"fmt"
	"os"
	"os/exec"
	"path/filepath"

	"github.com/jupiterrider/purego-sdl3/sdl"
	"github.com/jupiterrider/purego-sdl3/ttf"

	"arkenidar.com/purego-sdl3/internal/widgets"
)

// entry describes one launchable example: its button label, the sibling
// binary name under bin/, and a one-line description shown next to it.
type entry struct {
	Label       string
	Binary      string
	Description string
}

var entries = []entry{
	{"Counter Demo", "app", "Draggable rectangle, counter, alert dialog"},
	{"Blank Window", "blank-window", "Bare-minimum starter template"},
	{"Bouncing Balls", "bouncing-balls", "Animated squares bouncing off edges"},
	{"Form Demo", "form-demo", "Buttons, checkboxes and layout"},
	{"Text Playground", "text-playground", "Scrollable, word-wrap text area"},
	{"CRUD App", "crud-app", "To-do list, JSON-persisted"},
}

const (
	rowY       = 20
	colSpacing = 20
	rowSpacing = 12
	rowPadX    = 20
	statusGapY = 16
)

// App lists every other example app as a two-column table (button,
// description); clicking a button launches its sibling binary as a
// detached child process.
type App struct {
	renderer *sdl.Renderer
	font     *ttf.Font

	table       *widgets.Table
	statusLabel *widgets.Label

	exeDir   string
	repoRoot string

	// ContentWidth/ContentHeight are the window dimensions this App
	// actually needs, computed once from its widgets after construction.
	// The window is created before content exists (Bootstrap needs a
	// renderer first), so callers should resize to fit these afterward
	// instead of guessing a size up front — a fixed guessed size is
	// exactly what let content run off-screen before.
	ContentWidth  float32
	ContentHeight float32
}

func NewApp(renderer *sdl.Renderer, font *ttf.Font) *App {
	app := &App{renderer: renderer, font: font}

	exePath, err := os.Executable()
	if err != nil {
		app.exeDir = "."
	} else {
		app.exeDir = filepath.Dir(exePath)
	}
	app.repoRoot = filepath.Dir(app.exeDir)

	app.table = widgets.NewTable(rowPadX, rowY, colSpacing, rowSpacing)
	for _, e := range entries {
		e := e
		button := widgets.NewButton(0, 0, 0, 0, e.Label, font, renderer, func() {
			app.launch(e)
		})
		desc := widgets.NewLabel(0, 0, e.Description, font, renderer)
		app.table.AddRow(button, desc)
	}
	app.table.Layout()

	statusY := rowY + app.table.ContentHeight() + statusGapY
	app.statusLabel = widgets.NewLabel(rowPadX, statusY, "", font, renderer)

	// The status label starts empty, so its rendered texture (and thus
	// Bounds.H) is 0 right now — sizing the window from that would leave
	// no room for the "Launched X" message that appears after the first
	// click, pushing it off the bottom edge. Reserve a real line height
	// from font metrics instead of the empty label's current size.
	statusLineHeight := float32(ttf.GetFontLineSkip(font))

	// Likewise, size the width for the common "Launched <label>" message
	// (the longest one, specifically), not just the table — an arbitrary
	// OS error message can still run long and overflow, but this covers
	// the expected case.
	longestStatus := widgets.NewLabel(0, 0, "Launched "+longestLabel(), font, renderer)
	statusWidth := longestStatus.Bounds.W
	longestStatus.Destroy()

	app.ContentWidth = rowPadX + maxFloat(app.table.ContentWidth(), statusWidth) + rowPadX
	app.ContentHeight = statusY + statusLineHeight + rowPadX

	return app
}

func longestLabel() string {
	var longest string
	for _, e := range entries {
		if len(e.Label) > len(longest) {
			longest = e.Label
		}
	}
	return longest
}

func maxFloat(a, b float32) float32 {
	if a > b {
		return a
	}
	return b
}

// launch starts entry's sibling binary as a detached, non-blocking child
// process so the launcher itself stays open and usable.
func (app *App) launch(e entry) {
	path := filepath.Join(app.exeDir, e.Binary)
	if _, err := os.Stat(path); err != nil {
		app.statusLabel.UpdateText(fmt.Sprintf("Can't find %s: %v", e.Binary, err))
		return
	}

	cmd := exec.Command(path)
	cmd.Dir = app.repoRoot
	cmd.Env = append(os.Environ(), "LD_LIBRARY_PATH="+filepath.Join(app.repoRoot, "lib"))
	if err := cmd.Start(); err != nil {
		app.statusLabel.UpdateText(fmt.Sprintf("Failed to launch %s: %v", e.Binary, err))
		return
	}
	app.statusLabel.UpdateText(fmt.Sprintf("Launched %s", e.Label))
}

func (app *App) Destroy() {
	app.table.Destroy()
	app.statusLabel.Destroy()
}

func (app *App) handleEvents() bool {
	var event sdl.Event
	for sdl.PollEvent(&event) {
		mx, my := float32(0), float32(0)
		if event.Type() == sdl.EventMouseButtonDown || event.Type() == sdl.EventMouseButtonUp {
			mx = float32(event.Button().X)
			my = float32(event.Button().Y)
		}

		switch event.Type() {
		case sdl.EventQuit:
			return false
		case sdl.EventKeyDown:
			if event.Key().Scancode == sdl.ScancodeEscape {
				return false
			}
		case sdl.EventMouseButtonDown, sdl.EventMouseButtonUp:
			app.table.Update(event, mx, my)
		}
	}
	return true
}

func (app *App) render() {
	sdl.SetRenderDrawColor(app.renderer, 30, 30, 35, sdl.AlphaOpaque)
	sdl.RenderClear(app.renderer)

	app.table.Render(app.renderer)
	app.statusLabel.Render(app.renderer)

	sdl.RenderPresent(app.renderer)
}
