// app.go
package main

import (
	"github.com/jupiterrider/purego-sdl3/sdl"
	"github.com/jupiterrider/purego-sdl3/ttf"

	"arkenidar.com/purego-sdl3/internal/widgets"
)

const dataFile = "crud-app-data.json"

const (
	rowHeight   = 36
	rowSpacing  = 8
	listStartY  = 60
	rowPaddingX = 10
)

type row struct {
	item     Item
	checkbox *widgets.Checkbox
	delete   *widgets.Button
}

// App implements a minimal CRUD list: add items via a text input, toggle
// "done" per row, delete a row — every mutation is persisted to dataFile.
type App struct {
	window   *sdl.Window
	renderer *sdl.Renderer
	font     *ttf.Font

	windowWidth  float32
	windowHeight float32

	nextID int
	items  []Item
	rows   []row

	input *widgets.TextInput
}

func NewApp(window *sdl.Window, renderer *sdl.Renderer, font *ttf.Font, windowWidth, windowHeight float32) *App {
	app := &App{
		window:       window,
		renderer:     renderer,
		font:         font,
		windowWidth:  windowWidth,
		windowHeight: windowHeight,
	}

	items, err := LoadItems(dataFile)
	if err == nil {
		app.items = items
	}
	for _, item := range app.items {
		if item.ID >= app.nextID {
			app.nextID = item.ID + 1
		}
	}

	app.input = widgets.NewTextInput(rowPaddingX, 10, windowWidth-2*rowPaddingX, 36, font, renderer, window)
	app.input.OnSubmit = func(value string) {
		if value == "" {
			return
		}
		app.items = append(app.items, Item{ID: app.nextID, Text: value})
		app.nextID++
		app.input.Value = ""
		app.rebuildRows()
		app.save()
	}
	app.input.Focus()

	app.rebuildRows()

	return app
}

func (app *App) save() {
	_ = SaveItems(dataFile, app.items)
}

// rebuildRows recreates the per-row widgets from app.items. Called whenever
// the item list changes shape (add/delete) since each row widget's callback
// closes over that row's item ID.
func (app *App) rebuildRows() {
	for _, r := range app.rows {
		r.checkbox.Destroy()
		r.delete.Destroy()
	}
	app.rows = app.rows[:0]

	for i, item := range app.items {
		item := item
		y := float32(listStartY + i*(rowHeight+rowSpacing))

		checkbox := widgets.NewCheckbox(rowPaddingX, y, item.Text, item.Done, app.font, app.renderer, func(checked bool) {
			app.setDone(item.ID, checked)
		})
		deleteButton := widgets.NewButton(0, 0, 0, 0, "Delete", app.font, app.renderer, func() {
			app.deleteItem(item.ID)
		})
		deleteBounds := deleteButton.GetBounds()
		deleteButton.Bounds.X = app.windowWidth - deleteBounds.W - rowPaddingX
		deleteButton.Bounds.Y = y

		app.rows = append(app.rows, row{item: item, checkbox: checkbox, delete: deleteButton})
	}
}

func (app *App) setDone(id int, done bool) {
	for i := range app.items {
		if app.items[i].ID == id {
			app.items[i].Done = done
			break
		}
	}
	app.save()
}

func (app *App) deleteItem(id int) {
	for i, item := range app.items {
		if item.ID == id {
			app.items = append(app.items[:i], app.items[i+1:]...)
			break
		}
	}
	app.rebuildRows()
	app.save()
}

func (app *App) Destroy() {
	for _, r := range app.rows {
		r.checkbox.Destroy()
		r.delete.Destroy()
	}
}

func (app *App) handleEvents() bool {
	var event sdl.Event
	for sdl.PollEvent(&event) {
		mx, my := float32(0), float32(0)
		if event.Type() == sdl.EventMouseButtonDown || event.Type() == sdl.EventMouseButtonUp {
			mx = float32(event.Button().X)
			my = float32(event.Button().Y)
		} else if event.Type() == sdl.EventMouseMotion {
			mx = float32(event.Motion().X)
			my = float32(event.Motion().Y)
		}

		switch event.Type() {
		case sdl.EventQuit:
			return false
		case sdl.EventWindowResized:
			app.windowWidth = float32(event.Window().Data1)
			app.windowHeight = float32(event.Window().Data2)
			app.input.Bounds.W = app.windowWidth - 2*rowPaddingX
			app.rebuildRows()
		case sdl.EventKeyDown:
			if event.Key().Scancode == sdl.ScancodeEscape {
				return false
			}
			app.input.Update(event, mx, my)
		case sdl.EventTextInput:
			app.input.Update(event, mx, my)
		case sdl.EventMouseButtonDown:
			if app.input.Update(event, mx, my) {
				break
			}
			for _, r := range app.rows {
				if r.checkbox.Update(event, mx, my) {
					break
				}
				if r.delete.Update(event, mx, my) {
					break
				}
			}
		case sdl.EventMouseMotion:
			app.input.Update(event, mx, my)
		case sdl.EventMouseButtonUp:
			app.input.Update(event, mx, my)
			for _, r := range app.rows {
				r.delete.Update(event, mx, my)
			}
		}
	}
	return true
}

func (app *App) render() {
	sdl.SetRenderDrawColor(app.renderer, 30, 30, 35, sdl.AlphaOpaque)
	sdl.RenderClear(app.renderer)

	app.input.Render(app.renderer)

	for _, r := range app.rows {
		r.checkbox.Render(app.renderer)
		r.delete.Render(app.renderer)
	}

	sdl.RenderPresent(app.renderer)
}
