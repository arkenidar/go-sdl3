package widgets

import "github.com/jupiterrider/purego-sdl3/sdl"

// The upstream SDL binding exposes GetClipboardText but not
// SetClipboardText, so copy/cut cannot reach the OS clipboard. Instead the
// widgets share an app-internal clipboard: Ctrl+C/X write here, and Ctrl+V
// prefers whichever source is fresher — the OS clipboard if it changed since
// the last in-app copy (the user copied something outside the app), the
// internal one otherwise.
var (
	appClipboard    string
	osTextAtAppCopy string
)

// isCopyShortcut reports whether the key event is the copy chord (Ctrl+C).
func isCopyShortcut(key sdl.KeyboardEvent) bool {
	return key.Scancode == sdl.ScancodeC && key.Mod&sdl.KeymodCtrl != 0
}

// isCutShortcut reports whether the key event is the cut chord (Ctrl+X).
func isCutShortcut(key sdl.KeyboardEvent) bool {
	return key.Scancode == sdl.ScancodeX && key.Mod&sdl.KeymodCtrl != 0
}

// clipboardWrite stores s in the app clipboard, remembering the OS
// clipboard's current text so clipboardRead can tell later whether the OS
// side changed after this copy.
func clipboardWrite(s string) {
	appClipboard = s
	osTextAtAppCopy = sdl.GetClipboardText()
}

// clipboardRead returns the text a paste should insert.
func clipboardRead() string {
	return pickClipboard(sdl.GetClipboardText(), osTextAtAppCopy, appClipboard)
}

// pickClipboard chooses between the OS clipboard text and the app-internal
// one: the OS text wins when it changed since the last in-app copy or when
// nothing was ever copied in-app.
func pickClipboard(osText, osAtCopy, app string) string {
	if app == "" {
		return osText
	}
	if osText != "" && osText != osAtCopy {
		return osText
	}
	return app
}
