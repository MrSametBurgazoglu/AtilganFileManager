package main

import (
	"github.com/MrSametBurgazoglu/atilgan/clipboard"
	"github.com/MrSametBurgazoglu/atilgan/previewer"
	"github.com/MrSametBurgazoglu/atilgan/rename_popup"
	"github.com/MrSametBurgazoglu/atilgan/shortcut_popup"
	"github.com/diamondburned/gotk4-adwaita/pkg/adw"
	"github.com/diamondburned/gotk4/pkg/gdk/v4"
	"github.com/diamondburned/gotk4/pkg/gio/v2"
	"github.com/diamondburned/gotk4/pkg/glib/v2"
	"github.com/diamondburned/gotk4/pkg/gtk/v4"
)

func (m *MainBox) setupShortcuts(mainWindow *adw.ApplicationWindow, copyCutPreviewer *previewer.CopyCutPreviewer) {
	controller := gtk.NewShortcutController()

	renameTrigger := gtk.NewKeyvalTrigger(gdk.KEY_r, gdk.ControlMask)
	renameShortcut := gtk.NewShortcut(renameTrigger, gtk.NewCallbackAction(func(widget gtk.Widgetter, args *glib.Variant) (ok bool) {
		activePanel := m.getActivePanel()
		if activePanel == nil {
			return true
		}
		selectedItem := activePanel.FileViewer.FileViewerList.GetItems()[activePanel.FileViewer.FileViewerList.GetSelectedIDX()]
		renameWindow := rename_popup.NewRenameWindow(m.Path, selectedItem.Path)
		renameWindow.SetTransientFor(&mainWindow.Window)
		renameWindow.SetVisible(true)
		return true
	}))
	controller.AddShortcut(renameShortcut)

	searchTrigger := gtk.NewKeyvalTrigger(gdk.KEY_f, gdk.ControlMask)
	searchShortcut := gtk.NewShortcut(searchTrigger, gtk.NewCallbackAction(func(widget gtk.Widgetter, args *glib.Variant) (ok bool) {
		activePanel := m.getActivePanel()
		if activePanel == nil {
			return true
		}
		activePanel.FileViewer.SearchRevealer.SetRevealChild(!activePanel.FileViewer.SearchRevealer.RevealChild())
		if activePanel.FileViewer.SearchRevealer.RevealChild() {
			activePanel.FileViewer.SearchEntry.GrabFocus()
			activePanel.FileViewer.SearchRevealer.SetVisible(true)
			activePanel.FileViewer.FileViewerList.SetCanFocus(false)
		} else {
			activePanel.FileViewer.SearchRevealer.SetVisible(false)
			activePanel.FileViewer.FileViewerList.SetCanFocus(true)
		}
		return true
	}))
	controller.AddShortcut(searchShortcut)

	selectAllTrigger := gtk.NewKeyvalTrigger(gdk.KEY_a, gdk.ControlMask)
	selectAllShortcut := gtk.NewShortcut(selectAllTrigger, gtk.NewCallbackAction(func(widget gtk.Widgetter, args *glib.Variant) (ok bool) {
		activePanel := m.getActivePanel()
		if activePanel != nil {
			activePanel.FileViewer.FileViewerList.SelectAll()
		}
		return true
	}))
	controller.AddShortcut(selectAllShortcut)

	copyTrigger := gtk.NewKeyvalTrigger(gdk.KEY_c, gdk.ControlMask)
	copyShortcut := gtk.NewShortcut(copyTrigger, gtk.NewCallbackAction(func(widget gtk.Widgetter, args *glib.Variant) (ok bool) {
		activePanel := m.getActivePanel()
		if activePanel == nil || m.SpecialPaths.Paths[m.Path] != nil {
			return true
		}
		activePanel.FileViewer.IsCopy = true
		activePanel.FileViewer.AddSelectedToCopyCut()

		copyCutPreviewer.IsCut = false
		copyCutPreviewer.SetFiles(activePanel.FileViewer.CopiedCuttedFiles)
		copyCutPreviewer.SetVisible(true)
		
		clipboard.CopyFileToClipboard(gio.NewFileForPath(activePanel.FileViewer.FileViewerList.GetItems()[activePanel.FileViewer.FileViewerList.GetSelectedIDX()].Path))
		return true
	}))
	controller.AddShortcut(copyShortcut)

	cutTrigger := gtk.NewKeyvalTrigger(gdk.KEY_x, gdk.ControlMask)
	cutShortcut := gtk.NewShortcut(cutTrigger, gtk.NewCallbackAction(func(widget gtk.Widgetter, args *glib.Variant) (ok bool) {
		activePanel := m.getActivePanel()
		if activePanel == nil || m.SpecialPaths.Paths[m.Path] != nil {
			return true
		}
		activePanel.FileViewer.IsCopy = true
		activePanel.FileViewer.IsCut = true
		activePanel.FileViewer.AddSelectedToCopyCut()
		copyCutPreviewer.IsCut = true
		copyCutPreviewer.SetFiles(activePanel.FileViewer.CopiedCuttedFiles)
		copyCutPreviewer.SetVisible(true)
		return true
	}))
	controller.AddShortcut(cutShortcut)

	pasteTrigger := gtk.NewKeyvalTrigger(gdk.KEY_v, gdk.ControlMask)
	pasteShortcut := gtk.NewShortcut(pasteTrigger, gtk.NewCallbackAction(func(widget gtk.Widgetter, args *glib.Variant) (ok bool) {
		activePanel := m.getActivePanel()
		if activePanel == nil || m.SpecialPaths.Paths[m.Path] != nil {
			return true
		}
		m.HeaderBar.ShowProgress()
		go func() error {
			if err := activePanel.FileViewer.ExecuteCopyPaste(func(f float64) {
				glib.IdleAdd(func() {
					m.HeaderBar.SetProgress(f)
				})
			}); err == nil {
				glib.IdleAdd(func() {
					m.pathChanged(m.Path)
					activePanel.FileViewer.CleanCopyCutFiles()
					activePanel.FileViewer.IsCopy = false
					activePanel.FileViewer.IsCut = false
					copyCutPreviewer.SetVisible(false)
					m.HeaderBar.HideProgress()
				})
			}
			return nil
		}()
		return true
	}))
	controller.AddShortcut(pasteShortcut)

	escapeTrigger := gtk.NewKeyvalTrigger(gdk.KEY_Escape, 0)
	escapeShortcut := gtk.NewShortcut(escapeTrigger, gtk.NewCallbackAction(func(widget gtk.Widgetter, args *glib.Variant) (ok bool) {
		activePanel := m.getActivePanel()
		if activePanel != nil {
			activePanel.FileViewer.CleanCopyCutFiles()
			activePanel.FileViewer.IsCopy = false
			activePanel.FileViewer.IsCut = false
		}
		copyCutPreviewer.SetVisible(false)
		return true
	}))
	controller.AddShortcut(escapeShortcut)

	helpTrigger := gtk.NewKeyvalTrigger(gdk.KEY_h, gdk.ControlMask)
	helpShortcut := gtk.NewShortcut(helpTrigger, gtk.NewCallbackAction(func(widget gtk.Widgetter, args *glib.Variant) (ok bool) {
		shortcut_popup.NewShortcutPopup(&mainWindow.Window)
		return true
	}))
	controller.AddShortcut(helpShortcut)

	for r := 'A'; r <= 'Z'; r++ {
		s := string(r)
		keyval := gdk.KeyvalFromName(s)
		trigger := gtk.NewKeyvalTrigger(keyval, gdk.ShiftMask)
		letterShortcut := gtk.NewShortcut(trigger, gtk.NewCallbackAction(func(widget gtk.Widgetter, args *glib.Variant) (ok bool) {
			activePanel := m.getActivePanel()
			if activePanel != nil {
				activePanel.FileViewer.FileViewerList.SetSelectedItemWithLetter(s)
				m.updatePreviewer()
			}
			return true
		}))
		controller.AddShortcut(letterShortcut)
	}

	mainWindow.AddController(controller)

	keyController := gtk.NewEventControllerKey()
	keyController.ConnectKeyReleased(func(keyval uint, keycode uint, state gdk.ModifierType) {
		if keyval == gdk.KEY_space {
			m.PreviewerPanel.ShowSpecificPreviewer()
		}
	})
	mainWindow.AddController(keyController)
}
