// app.go
package main

import (
	"fmt"

	"github.com/jupiterrider/purego-sdl3/sdl"
	"github.com/jupiterrider/purego-sdl3/ttf"

	"arkenidar.com/purego-sdl3/internal/widgets"
)

// App demonstrates a settings-style form built from buttons, labels and
// checkboxes laid out with widgets.Layout.
type App struct {
	renderer *sdl.Renderer
	font     *ttf.Font

	uiLayout    *widgets.Layout
	statusLabel *widgets.Label

	notifications bool
	darkMode      bool
	autoSave      bool
}

func NewApp(renderer *sdl.Renderer, font *ttf.Font) *App {
	app := &App{renderer: renderer, font: font}

	app.uiLayout = widgets.NewLayout(10, 10, 14)

	notificationsBox := widgets.NewCheckbox(0, 0, "Notifications", app.notifications, font, renderer, func(checked bool) {
		app.notifications = checked
		app.updateStatus()
	})
	darkModeBox := widgets.NewCheckbox(0, 0, "Dark mode", app.darkMode, font, renderer, func(checked bool) {
		app.darkMode = checked
		app.updateStatus()
	})
	autoSaveBox := widgets.NewCheckbox(0, 0, "Auto-save", app.autoSave, font, renderer, func(checked bool) {
		app.autoSave = checked
		app.updateStatus()
	})
	resetButton := widgets.NewButton(0, 0, 0, 0, "Reset", font, renderer, func() {
		notificationsBox.Checked = false
		darkModeBox.Checked = false
		autoSaveBox.Checked = false
		app.notifications = false
		app.darkMode = false
		app.autoSave = false
		app.updateStatus()
	})

	app.statusLabel = widgets.NewLabel(0, 0, "", font, renderer)

	app.uiLayout.AddWidget(notificationsBox)
	app.uiLayout.AddWidget(darkModeBox)
	app.uiLayout.AddWidget(autoSaveBox)
	app.uiLayout.AddWidget(resetButton)

	app.updateStatus()
	app.statusLabel.Bounds.X = 10
	app.statusLabel.Bounds.Y = 60

	return app
}

func (app *App) updateStatus() {
	app.statusLabel.UpdateText(fmt.Sprintf(
		"notifications=%t  dark_mode=%t  auto_save=%t",
		app.notifications, app.darkMode, app.autoSave,
	))
}

func (app *App) Destroy() {
	app.uiLayout.Destroy()
	app.statusLabel.Destroy()
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
		case sdl.EventKeyDown:
			if event.Key().Scancode == sdl.ScancodeEscape {
				return false
			}
		case sdl.EventMouseButtonDown, sdl.EventMouseButtonUp, sdl.EventMouseMotion:
			app.uiLayout.Update(event, mx, my)
		}
	}
	return true
}

func (app *App) render() {
	sdl.SetRenderDrawColor(app.renderer, 40, 40, 45, sdl.AlphaOpaque)
	sdl.RenderClear(app.renderer)

	app.uiLayout.Render(app.renderer)
	app.statusLabel.Render(app.renderer)

	sdl.RenderPresent(app.renderer)
}
