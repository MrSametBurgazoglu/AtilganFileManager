package shortcut_popup

import (
	_ "embed"
	"github.com/diamondburned/gotk4/pkg/gtk/v4"
)

//go:embed shortcut_popup.ui
var uiXML string

type ShortcutPopup struct {
	*gtk.ShortcutsWindow
	builder *gtk.Builder
}

func NewShortcutPopup(parent *gtk.Window) *ShortcutPopup {
	builder := gtk.NewBuilderFromString(uiXML)
	shortCutWindowObj := builder.GetObject("shortcuts-window")
	window := shortCutWindowObj.Cast().(*gtk.ShortcutsWindow)
	shortcutPopup := &ShortcutPopup{
		ShortcutsWindow: window,
		builder:         builder,
	}
	shortcutPopup.SetTransientFor(parent)
	shortcutPopup.SetModal(true)

	shortcutPopup.SetVisible(true)
	shortcutPopup.Connect("close-request", func() {
		shortcutPopup.Destroy()
	})

	return shortcutPopup
}
