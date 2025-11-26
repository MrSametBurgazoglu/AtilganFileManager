package main

import (
	"embed"
	"os"
	"path/filepath"

	"github.com/MrSametBurgazoglu/atilgan/clipboard"
	"github.com/MrSametBurgazoglu/atilgan/header"
	"github.com/MrSametBurgazoglu/atilgan/pathbar"
	"github.com/MrSametBurgazoglu/atilgan/previewer"
	"github.com/MrSametBurgazoglu/atilgan/previewer_panel"
	"github.com/MrSametBurgazoglu/atilgan/rename_popup"
	"github.com/MrSametBurgazoglu/atilgan/search"
	"github.com/MrSametBurgazoglu/atilgan/shortcut_popup"
	"github.com/MrSametBurgazoglu/atilgan/sidebar"
	"github.com/MrSametBurgazoglu/atilgan/special_path"
	"github.com/MrSametBurgazoglu/atilgan/viewer"
	"github.com/MrSametBurgazoglu/atilgan/viewer_panel"
	"github.com/diamondburned/gotk4/pkg/gdk/v4"
	"github.com/diamondburned/gotk4/pkg/gio/v2"
	"github.com/diamondburned/gotk4/pkg/glib/v2"
	"github.com/diamondburned/gotk4/pkg/gtk/v4"
)

//go:embed style.css
var styleCSS embed.FS

type MainBox struct {
	*gtk.Box
	Path           string
	Pathbar        *pathbar.PathBar
	PreviewerPanel *previewer_panel.PreviewPanel
	ViewerPanel    *viewer_panel.Panel
	SpecialPaths   *special_path.SpecialPathManager
	Search         *search.Search
	SideBar        *sidebar.Sidebar
}

func NewMainBox(mainWindow *gtk.Window, headerBar *header.HeaderBar) *MainBox {
	mainVBox := gtk.NewBox(gtk.OrientationVertical, 6)
	mainBox := &MainBox{Box: mainVBox}

	curdir, err := os.Getwd()
	if err != nil {
		panic(err)
	}

	headerBar.SearchButton.ConnectClicked(func() {
		mainBox.Search.SetVisible(!mainBox.Search.Visible())
	})
	headerBar.ShortcutsButton.ConnectClicked(func() {
		shortcut_popup.NewShortcutPopup(mainWindow)
	})

	headerBar.PreviewerPanelButton.ConnectClicked(func() {
		mainBox.PreviewerPanel.SetVisible(!mainBox.PreviewerPanel.Visible())
		if mainBox.PreviewerPanel.Visible() {
			mainBox.ViewerPanel.SetHExpand(false)
		} else {
			mainBox.ViewerPanel.SetHExpand(true)
		}
	})

	mainBox.SpecialPaths, err = special_path.NewSpecialPathManager()
	if err != nil {
		println(err.Error())
	}

	mainBox.Path = curdir
	mainBox.Pathbar = pathbar.NewPathBar(mainBox.pathChanged)
	mainBox.Pathbar.UpdatePathBar(curdir)

	mainBox.SideBar = sidebar.NewSidebar(mainBox.pathChanged)
	mainBox.SideBar.SetOrientation(gtk.OrientationVertical)

	mainBox.Search = search.NewSearch(curdir)
	mainBox.Search.SetVisible(false)
	mainBox.Search.PathChanged = mainBox.pathChanged
	mainVBox.Append(mainBox.Search)

	mainHBox := gtk.NewBox(gtk.OrientationHorizontal, 0)
	mainVBox.Append(mainHBox)

	mainHBox.Append(mainBox.SideBar)

	mainBox.ViewerPanel = viewer_panel.NewPanel(mainWindow, curdir, mainBox.pathChanged, mainBox.SpecialPaths)
	mainBox.ViewerPanel.FileViewer.Box.Append(mainBox.Pathbar)
	mainHBox.Append(mainBox.ViewerPanel)

	seperator := gtk.NewSeparator(gtk.OrientationVertical)
	mainHBox.Append(seperator)

	rightBox := gtk.NewBox(gtk.OrientationVertical, 6)
	mainHBox.Append(rightBox)

	mainBox.PreviewerPanel = previewer_panel.NewPreviewPanel(curdir, mainBox.pathChanged, mainBox.SpecialPaths)
	rightBox.Append(mainBox.PreviewerPanel)

	copyCutPreviewer := previewer.NewCopyCutPreviewer()
	copyCutPreviewer.SetVisible(false)
	rightBox.Append(copyCutPreviewer)

	mainBox.ViewerPanel.FileViewer.FileViewerList.SelectionChanged = func(index int) {
		mainBox.updatePreviewer()
	}

	mainBox.ViewerPanel.FileViewer.FileViewerList.PathChanged = mainBox.pathChanged

	mainBox.ViewerPanel.FileViewer.FileViewerList.KeyLeftPressed = func() {
		specialPath := mainBox.SpecialPaths.GetPath(mainBox.ViewerPanel.FileViewer.Path)
		if specialPath != nil {
			mainBox.pathChanged(specialPath.GetParentPath())
		} else {
			parentDir := filepath.Dir(mainBox.ViewerPanel.FileViewer.Path)
			mainBox.pathChanged(parentDir)
			selectHistory, isExist := mainBox.ViewerPanel.FileViewer.FileViewerHistory[parentDir]
			if isExist {
				mainBox.ViewerPanel.FileViewer.FileViewerList.SetItem(selectHistory.Index)
			}
		}
	}

	controller := gtk.NewShortcutController()

	renameTrigger := gtk.NewKeyvalTrigger(
		gdk.KEY_r,
		gdk.ControlMask,
	)

	renameShortcut := gtk.NewShortcut(renameTrigger, gtk.NewCallbackAction(func(widget gtk.Widgetter, args *glib.Variant) (ok bool) {
		selectedItem := mainBox.ViewerPanel.FileViewer.FileViewerList.Items[mainBox.ViewerPanel.FileViewer.FileViewerList.SelectedIDX]
		renameWindow := rename_popup.NewRenameWindow(mainBox.Path, selectedItem.Path)
		renameWindow.SetTransientFor(mainWindow)
		renameWindow.SetVisible(true)
		return true
	}))
	controller.AddShortcut(renameShortcut)

	searchTrigger := gtk.NewKeyvalTrigger(
		gdk.KEY_f,
		gdk.ControlMask,
	)

	searchShortcut := gtk.NewShortcut(searchTrigger, gtk.NewCallbackAction(func(widget gtk.Widgetter, args *glib.Variant) (ok bool) {
		mainBox.ViewerPanel.FileViewer.SearchRevealer.SetRevealChild(!mainBox.ViewerPanel.FileViewer.SearchRevealer.RevealChild())
		if mainBox.ViewerPanel.FileViewer.SearchRevealer.RevealChild() {
			mainBox.ViewerPanel.FileViewer.SearchEntry.GrabFocus()
			mainBox.ViewerPanel.FileViewer.SearchRevealer.SetVisible(true)
			mainBox.ViewerPanel.FileViewer.FileViewerList.CanFocus = false
		} else {
			mainBox.ViewerPanel.FileViewer.SearchRevealer.SetVisible(false)
			mainBox.ViewerPanel.FileViewer.FileViewerList.CanFocus = true
		}
		return true
	}))
	controller.AddShortcut(searchShortcut)

	copyTrigger := gtk.NewKeyvalTrigger(gdk.KEY_c, gdk.ControlMask)
	copyShortcut := gtk.NewShortcut(copyTrigger, gtk.NewCallbackAction(func(widget gtk.Widgetter, args *glib.Variant) (ok bool) {
		if mainBox.SpecialPaths.Paths[mainBox.Path] != nil {
			return true
		}
		mainBox.ViewerPanel.FileViewer.IsCopy = true
		mainBox.ViewerPanel.FileViewer.AddCopyCutItem(mainBox.ViewerPanel.FileViewer.FileViewerList.SelectedIDX)

		copyCutPreviewer.IsCut = false
		copyCutPreviewer.SetFiles(mainBox.ViewerPanel.FileViewer.CopiedCuttedFiles)
		copyCutPreviewer.SetVisible(true)
		clipboard.CopyFileToClipboard(gio.NewFileForPath(mainBox.ViewerPanel.FileViewer.FileViewerList.Items[mainBox.ViewerPanel.FileViewer.FileViewerList.SelectedIDX].Path))
		return true
	}))
	controller.AddShortcut(copyShortcut)

	cutTrigger := gtk.NewKeyvalTrigger(gdk.KEY_x, gdk.ControlMask)
	cutShortcut := gtk.NewShortcut(cutTrigger, gtk.NewCallbackAction(func(widget gtk.Widgetter, args *glib.Variant) (ok bool) {
		if mainBox.SpecialPaths.Paths[mainBox.Path] != nil {
			return true
		}
		mainBox.ViewerPanel.FileViewer.IsCopy = true
		mainBox.ViewerPanel.FileViewer.IsCut = true
		mainBox.ViewerPanel.FileViewer.AddCopyCutItem(mainBox.ViewerPanel.FileViewer.FileViewerList.SelectedIDX)
		copyCutPreviewer.IsCut = true
		copyCutPreviewer.SetFiles(mainBox.ViewerPanel.FileViewer.CopiedCuttedFiles)
		copyCutPreviewer.SetVisible(true)
		return true
	}))
	controller.AddShortcut(cutShortcut)

	pasteTrigger := gtk.NewKeyvalTrigger(gdk.KEY_v, gdk.ControlMask)
	pasteShortcut := gtk.NewShortcut(pasteTrigger, gtk.NewCallbackAction(func(widget gtk.Widgetter, args *glib.Variant) (ok bool) {
		if mainBox.SpecialPaths.Paths[mainBox.Path] != nil {
			return true
		}
		headerBar.ShowProgress()
		go func() error {
			if err := mainBox.ViewerPanel.FileViewer.ExecuteCopyPaste(func(f float64) {
				glib.IdleAdd(func() {
					headerBar.SetProgress(f)
				})
			}); err == nil {
				glib.IdleAdd(func() {
					mainBox.pathChanged(mainBox.Path)
					mainBox.ViewerPanel.FileViewer.CleanCopyCutFiles()
					mainBox.ViewerPanel.FileViewer.IsCopy = false
					mainBox.ViewerPanel.FileViewer.IsCut = false
					copyCutPreviewer.SetVisible(false)
					headerBar.HideProgress()
				})
			}
			return nil
		}()
		return true
	}))
	controller.AddShortcut(pasteShortcut)

	escapeTrigger := gtk.NewKeyvalTrigger(gdk.KEY_Escape, 0)
	escapeShortcut := gtk.NewShortcut(escapeTrigger, gtk.NewCallbackAction(func(widget gtk.Widgetter, args *glib.Variant) (ok bool) {
		mainBox.ViewerPanel.FileViewer.CleanCopyCutFiles()
		mainBox.ViewerPanel.FileViewer.IsCopy = false
		mainBox.ViewerPanel.FileViewer.IsCut = false
		copyCutPreviewer.SetVisible(false)
		return true
	}))
	controller.AddShortcut(escapeShortcut)

	helpTrigger := gtk.NewKeyvalTrigger(gdk.KEY_h, gdk.ControlMask)
	helpShortcut := gtk.NewShortcut(helpTrigger, gtk.NewCallbackAction(func(widget gtk.Widgetter, args *glib.Variant) (ok bool) {
		shortcut_popup.NewShortcutPopup(mainWindow)
		return true
	}))
	controller.AddShortcut(helpShortcut)

	for r := 'A'; r <= 'Z'; r++ {
		s := string(r)
		keyval := gdk.KeyvalFromName(s)

		trigger := gtk.NewKeyvalTrigger(
			keyval,
			gdk.ShiftMask,
		)

		letterShortcut := gtk.NewShortcut(trigger, gtk.NewCallbackAction(func(widget gtk.Widgetter, args *glib.Variant) (ok bool) {
			mainBox.ViewerPanel.FileViewer.FileViewerList.SetSelectedItemWithLetter(s)
			mainBox.updatePreviewer()
			return true
		}))
		controller.AddShortcut(letterShortcut)
	}

	mainWindow.AddController(controller)

	mainBox.ViewerPanel.FileViewer.FileViewerList.KeyRightPressed = func() {
		selectedIndex := mainBox.ViewerPanel.FileViewer.FileViewerList.SelectedIDX
		selectedItem := mainBox.ViewerPanel.FileViewer.FileViewerList.Items[selectedIndex]
		mainBox.ViewerPanel.FileViewer.FileViewerHistory[mainBox.Path] = &viewer.FileViewHistory{
			Path:  selectedItem.Path,
			Index: selectedIndex,
		}
		if selectedItem.IsDir {
			mainBox.pathChanged(selectedItem.Path)
		}
	}

	keyController := gtk.NewEventControllerKey()
	keyController.ConnectKeyReleased(func(keyval uint, keycode uint, state gdk.ModifierType) {
		if keyval == gdk.KEY_space {
			mainBox.PreviewerPanel.ShowSpecificPreviewer()
		}
	})
	mainWindow.AddController(keyController)

	mainBox.updatePreviewer()

	return mainBox
}
func main() {
	app := gtk.NewApplication("com.github.mrsametburgazoglu.atilgan", 0)
	app.ConnectActivate(func() {
		activate(app)
	})
	if code := app.Run(os.Args); code > 0 {
		os.Exit(code)
	}
}

func activate(app *gtk.Application) {
	window := gtk.NewApplicationWindow(app)
	window.SetTitle("Atilgan")
	window.SetDefaultSize(1200, 700)

	display := gdk.DisplayGetDefault()

	cssProvider := gtk.NewCSSProvider()
	cssBytes, err := styleCSS.ReadFile("style.css")
	if err == nil {
		cssProvider.LoadFromData(string(cssBytes))
	}
	gtk.StyleContextAddProviderForDisplay(display, cssProvider, gtk.STYLE_PROVIDER_PRIORITY_APPLICATION)

	iconTheme := gtk.IconThemeGetForDisplay(display)

	exePath, err := os.Executable()
	if err == nil {
		exeDir := filepath.Dir(exePath)
		iconTheme.AddSearchPath(exeDir)
	}
	curDir, err := os.Getwd()
	if err == nil {
		iconTheme.AddSearchPath(curDir)
	}
	window.SetIconName("atilgan_icon")

	headerBar := header.NewHeaderBar(window)
	mainBox := NewMainBox(&window.Window, headerBar)
	window.SetTitlebar(headerBar)
	window.SetChild(mainBox)

	window.SetVisible(true)
	mainBox.ViewerPanel.FileViewer.FileViewerList.DrawingArea.GrabFocus()
}

func (m *MainBox) pathChanged(path string) {
	if path == "" {
		path = m.Path
	}
	specialPath := m.SpecialPaths.GetPath(path)
	if specialPath != nil {
		items := specialPath.GetItems()
		m.ViewerPanel.FileViewer.FileViewerList.SetItems(items)
		m.Path = specialPath.GetPath()
		m.ViewerPanel.FileViewer.SetFolderName(path)
	} else {
		m.Path = path
		m.ViewerPanel.FileViewer.SetPath(path)
		m.Search.SetPath(path)
		m.SpecialPaths.AddRecentPath(path)
	}
	m.updatePreviewer()
	m.Pathbar.UpdatePathBar(path)
	m.SideBar.SetPath(path)
}

func (m *MainBox) updatePreviewer() {
	if len(m.ViewerPanel.FileViewer.FileViewerList.Items) == 0 {
		m.PreviewerPanel.Update("")
		return
	}
	selected := m.ViewerPanel.FileViewer.FileViewerList.Items[m.ViewerPanel.FileViewer.FileViewerList.SelectedIDX]
	m.PreviewerPanel.Update(selected.Path)
}
