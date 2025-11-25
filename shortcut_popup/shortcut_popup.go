package shortcut_popup

import (
	"os"
	"path/filepath"

	"github.com/diamondburned/gotk4/pkg/gtk/v4"
)

type ShortcutPopup struct {
	*gtk.ShortcutsWindow
	builder *gtk.Builder
}

func NewShortcutPopup(parent *gtk.Window) *ShortcutPopup {
	exePath, err := os.Executable()
	if err != nil {
		return nil
	}
	exeDir := filepath.Dir(exePath)
	uiPath := filepath.Join(exeDir, "shortcut_popup/shortcut_popup.ui")
	builder := gtk.NewBuilderFromFile(uiPath)
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
