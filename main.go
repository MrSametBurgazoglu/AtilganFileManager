package main

import (
	"embed"
	"os"
	"path/filepath"

	"github.com/MrSametBurgazoglu/atilgan/clipboard"
	"github.com/MrSametBurgazoglu/atilgan/create_popup"
	"github.com/MrSametBurgazoglu/atilgan/fileops"
	"github.com/MrSametBurgazoglu/atilgan/header"
	"github.com/MrSametBurgazoglu/atilgan/pathbar"
	"github.com/MrSametBurgazoglu/atilgan/preferences"
	"github.com/MrSametBurgazoglu/atilgan/previewer"
	"github.com/MrSametBurgazoglu/atilgan/previewer_panel"
	"github.com/MrSametBurgazoglu/atilgan/rename_popup"
	"github.com/MrSametBurgazoglu/atilgan/search"
	"github.com/MrSametBurgazoglu/atilgan/shortcut_popup"
	"github.com/MrSametBurgazoglu/atilgan/sidebar"
	"github.com/MrSametBurgazoglu/atilgan/special_path"
	"github.com/MrSametBurgazoglu/atilgan/types"
	"github.com/MrSametBurgazoglu/atilgan/viewer"
	"github.com/MrSametBurgazoglu/atilgan/viewer_panel"
	"github.com/MrSametBurgazoglu/atilgan/workspace"
	"github.com/diamondburned/gotk4-adwaita/pkg/adw"
	"github.com/diamondburned/gotk4/pkg/gdk/v4"
	"github.com/diamondburned/gotk4/pkg/gio/v2"
	"github.com/diamondburned/gotk4/pkg/glib/v2"
	"github.com/diamondburned/gotk4/pkg/gtk/v4"
)

//go:embed style.css
var styleCSS embed.FS

//go:embed atilgan_icon.svg
var atilganIcon []byte

type MainBox struct {
	*adw.NavigationSplitView
	MainWindow     *adw.ApplicationWindow
	Path           string
	Pathbar        *pathbar.PathBar
	PreviewerPanel *previewer_panel.PreviewPanel
	Workspace      *workspace.Workspace
	SpecialPaths   *special_path.SpecialPathManager
	Search         *search.Search
	SideBar        *sidebar.Sidebar
	HeaderBar      *header.HeaderBar
	Config         *preferences.Config
	ContentPaned   *gtk.Paned
	RightBox       *gtk.Box
	BottomBar      *gtk.Box
}

func (m *MainBox) applyPreferences() {
	styleManager := adw.StyleManagerGetDefault()
	switch m.Config.Theme {
	case "light":
		styleManager.SetColorScheme(adw.ColorSchemeForceLight)
	case "dark":
		styleManager.SetColorScheme(adw.ColorSchemeForceDark)
	default:
		styleManager.SetColorScheme(adw.ColorSchemeDefault)
	}

	if m.RightBox != nil {
		m.RightBox.SetVisible(m.Config.EnablePreviewPane)
	}

	m.NavigationSplitView.SetMaxSidebarWidth(float64(m.Config.SidebarWidth))

	// Refresh all tabs
	if m.Workspace != nil {
		for _, panel := range m.Workspace.PagePanels {
			panel.FileViewer.Refresh(true)
		}
	}
}

func NewMainBox(mainWindow *adw.ApplicationWindow, headerBar *header.HeaderBar) *MainBox {
	splitView := adw.NewNavigationSplitView()
	mainBox := &MainBox{
		NavigationSplitView: splitView,
		MainWindow:          mainWindow,
		HeaderBar:           headerBar,
		Config:              preferences.LoadConfig(),
		BottomBar:           gtk.NewBox(gtk.OrientationHorizontal, 6),
	}

	splitView.SetMaxSidebarWidth(float64(mainBox.Config.SidebarWidth))
	splitView.SetSidebarWidthFraction(0.15)

	curdir, err := os.Getwd()
	if err != nil {
		panic(err)
	}

	mainBox.SpecialPaths, err = special_path.NewSpecialPathManager()
	if err != nil {
		println(err.Error())
	}

	mainBox.Path = curdir
	mainBox.Pathbar = pathbar.NewPathBar(mainBox.pathChanged)
	mainBox.Pathbar.UpdatePathBar(curdir)

	mainBox.Workspace = workspace.NewWorkspace(&mainWindow.Window, mainBox.SpecialPaths, mainBox.pathChanged, mainBox.Config)
	mainBox.Workspace.SetupPanel = mainBox.setupPanel

	headerBar.SearchButton.ConnectClicked(func() {
		isVisible := !mainBox.Search.Visible()
		mainBox.Search.SetVisible(isVisible)
		mainBox.Pathbar.SetVisible(!isVisible)
		mainBox.Workspace.SetVisible(!isVisible)
		mainBox.PreviewerPanel.SetVisible(!isVisible)
		
		if isVisible {
			headerBar.SetTitleWidget(mainBox.Search.SearchBar)
		} else {
			headerBar.SetTitleWidget(gtk.NewBox(gtk.OrientationHorizontal, 0))
		}
	})

	mainBox.Search = search.NewSearch(curdir, mainBox.SpecialPaths, &mainWindow.Window, mainBox.Config)
	mainBox.Search.SetVisible(false)
	mainBox.Search.PathChanged = mainBox.pathChanged

	mainBox.SideBar = sidebar.NewSidebar(mainBox.pathChanged, mainBox.Config)
	mainBox.SideBar.SetOrientation(gtk.OrientationVertical)
	mainBox.SideBar.SetHExpand(true)
	mainBox.SideBar.SetHAlign(gtk.AlignFill)
	mainBox.SideBar.OnToggleHidden = func(active bool) {
		mainBox.Config.ShowHidden = active
		mainBox.Config.Save()
		mainBox.applyPreferences()
	}

	sidebarToolbar := adw.NewToolbarView()
	sidebarToolbar.AddTopBar(headerBar.LeftHeader)
	
	sidebarScrolled := gtk.NewScrolledWindow()
	sidebarScrolled.SetChild(mainBox.SideBar)
	sidebarScrolled.SetPolicy(gtk.PolicyNever, gtk.PolicyAutomatic)
	sidebarToolbar.SetContent(sidebarScrolled)

	sidebarPage := adw.NewNavigationPage(sidebarToolbar, "Atilgan")
	splitView.SetSidebar(sidebarPage)

	contentToolbar := adw.NewToolbarView()
	contentToolbar.AddTopBar(headerBar.RightHeader)

	contentVBox := gtk.NewBox(gtk.OrientationVertical, 0)
	contentVBox.Append(mainBox.Search)
	contentVBox.Append(mainBox.Workspace.TabBar)

	contentPaned := gtk.NewPaned(gtk.OrientationHorizontal)
	contentPaned.SetHExpand(true)
	contentPaned.SetVExpand(true)
	contentPaned.SetPosition(950)
	contentPaned.SetWideHandle(true)
	contentVBox.Append(contentPaned)
	mainBox.ContentPaned = contentPaned
	
	contentToolbar.SetContent(contentVBox)

	contentPage := adw.NewNavigationPage(contentToolbar, "Content")
	splitView.SetContent(contentPage)

headerBar.PackStart(mainBox.Pathbar)

previewToggle := gtk.NewToggleButton()
previewToggle.SetIconName("sidebar-show-right-symbolic")
previewToggle.SetTooltipText("Toggle Preview")
previewToggle.AddCSSClass("flat")
previewToggle.SetActive(mainBox.Config.EnablePreviewPane)
previewToggle.ConnectToggled(func() {
	mainBox.Config.EnablePreviewPane = previewToggle.Active()
	mainBox.Config.Save()
	mainBox.applyPreferences()
})

hiddenToggle := gtk.NewToggleButton()
hiddenToggle.SetIconName("view-conceal-symbolic")
hiddenToggle.SetTooltipText("Toggle Hidden Files")
hiddenToggle.AddCSSClass("flat")
hiddenToggle.SetActive(mainBox.Config.ShowHidden)
hiddenToggle.ConnectToggled(func() {
	mainBox.Config.ShowHidden = hiddenToggle.Active()
	mainBox.Config.Save()
	mainBox.applyPreferences()
	mainBox.pathChanged(mainBox.Path)
})

	spacer := gtk.NewBox(gtk.OrientationHorizontal, 0)
	spacer.SetHExpand(true)
	mainBox.BottomBar.Append(spacer)
	mainBox.BottomBar.Append(hiddenToggle)
	mainBox.BottomBar.Append(previewToggle)

	workspaceWrapper := gtk.NewBox(gtk.OrientationVertical, 0)
	workspaceWrapper.SetHExpand(true)
	workspaceWrapper.SetVExpand(true)
	workspaceWrapper.SetSizeRequest(300, -1)
	workspaceWrapper.Append(mainBox.Workspace)
	workspaceWrapper.Append(mainBox.BottomBar)

	contentPaned.SetStartChild(workspaceWrapper)
	contentPaned.SetResizeStartChild(true)
	contentPaned.SetShrinkStartChild(false)
	contentPaned.SetResizeEndChild(true)
	contentPaned.SetShrinkEndChild(false)

	rightBox := gtk.NewBox(gtk.OrientationVertical, 0)
	rightBox.SetHExpand(false)
	rightBox.SetVExpand(true)
	rightBox.SetSizeRequest(215, -1)
	contentPaned.SetEndChild(rightBox)
	mainBox.RightBox = rightBox

	mainBox.PreviewerPanel = previewer_panel.NewPreviewPanel(curdir, mainBox.pathChanged, mainBox.SpecialPaths, &mainWindow.Window, mainBox.Config)
	rightBox.Append(mainBox.PreviewerPanel)

	mainBox.Workspace.NewTab(curdir)

	mainBox.Workspace.TabView.Connect("notify::selected-page", func() {
		activePanel := mainBox.Workspace.GetActivePanel()
		if activePanel != nil {
			mainBox.Path = activePanel.Path
			mainBox.Pathbar.UpdatePathBar(mainBox.Path)
			mainBox.SideBar.SetPath(mainBox.Path)
			mainBox.updatePreviewer()
		}
	})

	mainBox.Workspace.TabView.Connect("close-page", func(page *adw.TabPage) bool {
		delete(mainBox.Workspace.PagePanels, page)
		mainBox.Workspace.TabView.ClosePageFinish(page, true)
		return true
	})

	mainBox.SideBar.OnNewTab = func() {
		mainBox.Workspace.NewTab(mainBox.Path)
	}

	mainBox.SideBar.OnPreferences = func() {
		dialog := preferences.NewPreferencesDialog(&mainWindow.Window, mainBox.Config)
		dialog.OnChanged = mainBox.applyPreferences
		dialog.Show()
	}
	copyCutPreviewer := previewer.NewCopyCutPreviewer()
	copyCutPreviewer.SetVisible(false)
	copyCutPreviewer.OnClear = func() {
		activePanel := mainBox.Workspace.GetActivePanel()
		if activePanel != nil {
			activePanel.FileViewer.CleanCopyCutFiles()
		}
		copyCutPreviewer.SetVisible(false)
	}
	rightBox.Append(copyCutPreviewer)

	mainBox.setupActionsMenu(mainWindow, headerBar)

	controller := gtk.NewShortcutController()

	renameTrigger := gtk.NewKeyvalTrigger(
		gdk.KEY_r,
		gdk.ControlMask,
	)

	renameShortcut := gtk.NewShortcut(renameTrigger, gtk.NewCallbackAction(func(widget gtk.Widgetter, args *glib.Variant) (ok bool) {
		activePanel := mainBox.getActivePanel()
		if activePanel == nil {
			return true
		}
		selectedItem := activePanel.FileViewer.FileViewerList.GetItems()[activePanel.FileViewer.FileViewerList.GetSelectedIDX()]
		renameWindow := rename_popup.NewRenameWindow(mainBox.Path, selectedItem.Path)
		renameWindow.SetTransientFor(&mainWindow.Window)
		renameWindow.SetVisible(true)
		return true
	}))
	controller.AddShortcut(renameShortcut)

	searchTrigger := gtk.NewKeyvalTrigger(
		gdk.KEY_f,
		gdk.ControlMask,
	)

	searchShortcut := gtk.NewShortcut(searchTrigger, gtk.NewCallbackAction(func(widget gtk.Widgetter, args *glib.Variant) (ok bool) {
		activePanel := mainBox.getActivePanel()
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
		activePanel := mainBox.getActivePanel()
		if activePanel != nil {
			activePanel.FileViewer.FileViewerList.SelectAll()
		}
		return true
	}))
	controller.AddShortcut(selectAllShortcut)

	copyTrigger := gtk.NewKeyvalTrigger(gdk.KEY_c, gdk.ControlMask)
	copyShortcut := gtk.NewShortcut(copyTrigger, gtk.NewCallbackAction(func(widget gtk.Widgetter, args *glib.Variant) (ok bool) {
		activePanel := mainBox.getActivePanel()
		if activePanel == nil || mainBox.SpecialPaths.Paths[mainBox.Path] != nil {
			return true
		}
		activePanel.FileViewer.IsCopy = true
		activePanel.FileViewer.AddSelectedToCopyCut()

		copyCutPreviewer.IsCut = false
		copyCutPreviewer.SetFiles(activePanel.FileViewer.CopiedCuttedFiles)
		copyCutPreviewer.SetVisible(true)
		
		// For now, still just copy the primary focused file to system clipboard
		clipboard.CopyFileToClipboard(gio.NewFileForPath(activePanel.FileViewer.FileViewerList.GetItems()[activePanel.FileViewer.FileViewerList.GetSelectedIDX()].Path))
		return true
	}))
	controller.AddShortcut(copyShortcut)

	cutTrigger := gtk.NewKeyvalTrigger(gdk.KEY_x, gdk.ControlMask)
	cutShortcut := gtk.NewShortcut(cutTrigger, gtk.NewCallbackAction(func(widget gtk.Widgetter, args *glib.Variant) (ok bool) {
		activePanel := mainBox.getActivePanel()
		if activePanel == nil || mainBox.SpecialPaths.Paths[mainBox.Path] != nil {
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
		activePanel := mainBox.getActivePanel()
		if activePanel == nil || mainBox.SpecialPaths.Paths[mainBox.Path] != nil {
			return true
		}
		headerBar.ShowProgress()
		go func() error {
			if err := activePanel.FileViewer.ExecuteCopyPaste(func(f float64) {
				glib.IdleAdd(func() {
					headerBar.SetProgress(f)
				})
			}); err == nil {
				glib.IdleAdd(func() {
					mainBox.pathChanged(mainBox.Path)
					activePanel.FileViewer.CleanCopyCutFiles()
					activePanel.FileViewer.IsCopy = false
					activePanel.FileViewer.IsCut = false
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
		activePanel := mainBox.getActivePanel()
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

		trigger := gtk.NewKeyvalTrigger(
			keyval,
			gdk.ShiftMask,
		)

		letterShortcut := gtk.NewShortcut(trigger, gtk.NewCallbackAction(func(widget gtk.Widgetter, args *glib.Variant) (ok bool) {
			activePanel := mainBox.getActivePanel()
			if activePanel != nil {
				activePanel.FileViewer.FileViewerList.SetSelectedItemWithLetter(s)
				mainBox.updatePreviewer()
			}
			return true
		}))
		controller.AddShortcut(letterShortcut)
	}

	mainWindow.AddController(controller)

	keyController := gtk.NewEventControllerKey()
	keyController.ConnectKeyReleased(func(keyval uint, keycode uint, state gdk.ModifierType) {
		if keyval == gdk.KEY_space {
			mainBox.PreviewerPanel.ShowSpecificPreviewer()
		}
	})
	mainWindow.AddController(keyController)

	mainBox.updatePreviewer()

	mainBox.applyPreferences()

	return mainBox
}
func (m *MainBox) getActivePanel() *viewer_panel.Panel {
	panel := m.Workspace.GetActivePanel()
	if panel == nil && len(m.Workspace.PagePanels) > 0 {
		// Fallback to the first available panel if selection is transiently nil
		for _, p := range m.Workspace.PagePanels {
			return p
		}
	}
	return panel
}

func (m *MainBox) setupPanel(panel *viewer_panel.Panel) {
	panel.FileViewer.FileViewerList.SetSelectionChanged(func(index int) {
		m.updatePreviewer()
	})
	panel.FileViewer.FileViewerList.SetPathChanged(m.pathChanged)

	panel.VideoViewer.FileViewerList.SetSelectionChanged(func(index int) {
		m.updatePreviewer()
	})
	panel.VideoViewer.FileViewerList.SetPathChanged(m.pathChanged)

	panel.PictureViewer.SelectionChanged = func(index int) {
		m.updatePreviewer()
	}
	// Note: PathChanged is passed into NewPictureViewer directly now.

	panel.MusicViewer.FileViewerList.SetSelectionChanged(func(index int) {
		m.updatePreviewer()
	})
	panel.MusicViewer.FileViewerList.SetPathChanged(m.pathChanged)

	panel.FileViewer.FileViewerList.SetKeyLeftPressed(func() {
		specialPath := m.SpecialPaths.GetPath(panel.FileViewer.Path)
		if specialPath != nil {
			m.pathChanged(specialPath.GetParentPath())
		} else {
			parentDir := filepath.Dir(panel.FileViewer.Path)
			m.pathChanged(parentDir)
			selectHistory, isExist := panel.FileViewer.FileViewerHistory[parentDir]
			if isExist {
				panel.FileViewer.FileViewerList.SetItem(selectHistory.Index)
			}
		}
	})

	panel.FileViewer.FileViewerList.SetPinRequested(func(path string) {
		m.SideBar.AddPin(path)
	})

	panel.FileViewer.FileViewerList.SetKeyRightPressed(func() {
		selectedIndex := panel.FileViewer.FileViewerList.GetSelectedIDX()
		items := panel.FileViewer.FileViewerList.GetItems()
		if selectedIndex >= 0 && selectedIndex < len(items) {
			selectedItem := items[selectedIndex]
			panel.FileViewer.FileViewerHistory[panel.Path] = &viewer.FileViewHistory{
				Path:  selectedItem.Path,
				Index: selectedIndex,
			}
			if selectedItem.IsDir {
				m.pathChanged(selectedItem.Path)
			}
		}
	})
}

func main() {
	app := adw.NewApplication("io.github.mrsametburgazoglu.AtilganFileManager", gio.ApplicationFlagsNone)
	app.ConnectActivate(func() {
		activate(app)
	})
	if code := app.Run(os.Args); code > 0 {
		os.Exit(code)
	}
}

func activate(app *adw.Application) {
	window := adw.NewApplicationWindow(&app.Application)
	window.SetTitle("Atilgan")
	window.SetDefaultSize(1200, 800)

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
		// Add /app/share/icons for Flatpak
		iconTheme.AddSearchPath(filepath.Join(filepath.Dir(exeDir), "share", "icons"))
	}
	curDir, err := os.Getwd()
	if err == nil {
		iconTheme.AddSearchPath(curDir)
	}
	window.SetIconName("io.github.mrsametburgazoglu.AtilganFileManager")

	headerBar := header.NewHeaderBar(window, atilganIcon)
	mainBox := NewMainBox(window, headerBar)
	window.SetContent(mainBox)

	window.SetVisible(true)
	activePanel := mainBox.getActivePanel()
	if activePanel != nil {
		activePanel.FileViewer.FileViewerList.FocusDrawingArea()
	}
}

func (m *MainBox) pathChanged(path string) {
	if path == "" {
		path = m.Path
	}

	if m.Search.Visible() {
		m.Search.SetVisible(false)
		m.Pathbar.SetVisible(true)
		m.Workspace.SetVisible(true)
		m.PreviewerPanel.SetVisible(true)
		m.HeaderBar.SetTitleWidget(gtk.NewBox(gtk.OrientationHorizontal, 0))
	}

	m.Path = path
	activePanel := m.getActivePanel()
	if activePanel != nil {
		activePanel.SetPath(path)
		
		selectedPage := m.Workspace.TabView.SelectedPage()
		if selectedPage != nil {
			title := filepath.Base(path)
			if title == "." || title == "/" || title == "" {
				title = "Root"
			}
			selectedPage.SetTitle(title)
		}
	}
	
	specialPath := m.SpecialPaths.GetPath(path)
	if specialPath == nil {
		m.Search.SetPath(path)
		m.SpecialPaths.AddRecentPath(path)
	}

	m.updatePreviewer()
	m.Pathbar.UpdatePathBar(path)
	m.SideBar.SetPath(path)
}

func (m *MainBox) updatePreviewer() {
	if m.PreviewerPanel == nil {
		return
	}
	activePanel := m.getActivePanel()
	if activePanel == nil {
		// If no active panel, we can't show a preview.
		// We don't necessarily want to clear it if it's just a transient state during initialization.
		return
	}

	path := activePanel.Path
	var items []*types.ListItem
	var selectedIdxs map[int]bool
	var selectedIdx int

	if path == fileops.GetVideosPath() && activePanel.VideoViewer != nil && activePanel.VideoViewer.FileViewerList != nil {
		items = activePanel.VideoViewer.FileViewerList.Items
		selectedIdxs = activePanel.VideoViewer.FileViewerList.SelectedIdxs
		selectedIdx = activePanel.VideoViewer.FileViewerList.SelectedIDX
	} else if path == fileops.GetPicturesPath() && activePanel.PictureViewer != nil {
		items = activePanel.PictureViewer.Items
		selectedIdxs = activePanel.PictureViewer.SelectedIdxs
		selectedIdx = activePanel.PictureViewer.SelectedIDX
	} else if path == fileops.GetMusicPath() && activePanel.MusicViewer != nil && activePanel.MusicViewer.FileViewerList != nil {
		items = activePanel.MusicViewer.FileViewerList.Items
		selectedIdxs = nil
		selectedIdx = activePanel.MusicViewer.FileViewerList.SelectedIDX
	} else if activePanel.FileViewer != nil && activePanel.FileViewer.FileViewerList != nil {
		items = activePanel.FileViewer.FileViewerList.GetItems()
		selectedIdxs = activePanel.FileViewer.FileViewerList.GetSelectedIdxs()
		selectedIdx = activePanel.FileViewer.FileViewerList.GetSelectedIDX()
	}

	if items == nil || len(items) == 0 {
		m.PreviewerPanel.Update("")
		return
	}

	selectedPaths := []string{}
	if selectedIdxs != nil {
		for i := 0; i < len(items); i++ {
			if selectedIdxs[i] {
				selectedPaths = append(selectedPaths, items[i].Path)
			}
		}
	}

	if len(selectedPaths) > 1 {
		m.PreviewerPanel.UpdateMultiple(selectedPaths)
	} else if len(selectedPaths) == 1 {
		m.PreviewerPanel.Update(selectedPaths[0])
	} else if selectedIdx >= 0 && selectedIdx < len(items) {
		selected := items[selectedIdx]
		m.PreviewerPanel.Update(selected.Path)
	} else {
		m.PreviewerPanel.Update("")
	}
}

func (m *MainBox) setupActionsMenu(mainWindow *adw.ApplicationWindow, headerBar *header.HeaderBar) {
	actionsPopover := gtk.NewPopover()
	headerBar.ActionsButton.SetPopover(actionsPopover)

	actionsBox := gtk.NewBox(gtk.OrientationVertical, 6)
	actionsBox.SetMarginTop(8)
	actionsBox.SetMarginBottom(8)
	actionsBox.SetMarginStart(8)
	actionsBox.SetMarginEnd(8)
	actionsPopover.SetChild(actionsBox)

	createActionBtn := func(iconName, labelText string, action func()) *gtk.Button {
		btn := gtk.NewButton()
		btn.AddCSSClass("flat")
		
		btnBox := gtk.NewBox(gtk.OrientationHorizontal, 8)
		icon := gtk.NewImageFromIconName(iconName)
		label := gtk.NewLabel(labelText)
		label.SetHAlign(gtk.AlignStart)
		
		btnBox.Append(icon)
		btnBox.Append(label)
		btn.SetChild(btnBox)
		
		btn.ConnectClicked(func() {
			action()
			actionsPopover.Popdown()
		})
		return btn
	}

	actionsLabel := gtk.NewLabel("Actions")
	actionsLabel.AddCSSClass("caption")
	actionsLabel.SetHAlign(gtk.AlignStart)
	actionsBox.Append(actionsLabel)

	createBox := gtk.NewBox(gtk.OrientationHorizontal, 4)
	createBox.SetHomogeneous(true)

	createBox.Append(createActionBtn("document-new-symbolic", "New File", func() {
		fs := create_popup.NewFileSelector(m.Path, m.pathChanged)
		fs.SetVisible(true)
		fs.SetTransientFor(&mainWindow.Window)
		fs.SetModal(true)
	}))
	createBox.Append(createActionBtn("folder-new-symbolic", "New Folder", func() {
		ds := create_popup.NewDirectorySelector(m.Path, m.pathChanged)
		ds.SetVisible(true)
		ds.SetTransientFor(&mainWindow.Window)
		ds.SetModal(true)
	}))
	actionsBox.Append(createBox)
	actionsBox.Append(gtk.NewSeparator(gtk.OrientationHorizontal))

	actionsBox.Append(createActionBtn("utilities-terminal-symbolic", "Open Terminal", func() {
		activePanel := m.getActivePanel()
		if activePanel != nil {
			activePanel.FileViewer.OpenTerminal()
		}
	}))

	actionsBox.Append(createActionBtn("tab-new-symbolic", "New Tab", func() {
		m.Workspace.NewTab(m.Path)
	}))

	actionsBox.Append(gtk.NewSeparator(gtk.OrientationHorizontal))

	actionsBox.Append(createActionBtn("preferences-desktop-keyboard-shortcuts-symbolic", "Shortcuts", func() {
		shortcut_popup.NewShortcutPopup(&mainWindow.Window)
	}))
	actionsBox.Append(createActionBtn("preferences-system-symbolic", "Preferences", func() {
		dialog := preferences.NewPreferencesDialog(&mainWindow.Window, m.Config)
		dialog.OnChanged = m.applyPreferences
		dialog.Show()
	}))
	actionsBox.Append(createActionBtn("help-about-symbolic", "About", func() {
		aboutDialog := adw.NewAboutWindow()
		aboutDialog.SetApplicationName("Atilgan")
		aboutDialog.SetVersion("0.2.0")
		aboutDialog.SetApplicationIcon("io.github.mrsametburgazoglu.AtilganFileManager")
		aboutDialog.SetCopyright("Copyright © 2025 MrSametBurgazoglu")
		aboutDialog.SetWebsite("https://github.com/MrSametBurgazoglu/AtilganFileManager")
		aboutDialog.SetTransientFor(&mainWindow.Window)
		aboutDialog.SetVisible(true)
	}))

	actionsBox.Append(gtk.NewSeparator(gtk.OrientationHorizontal))

	sortLabel := gtk.NewLabel("Sort By")
	sortLabel.AddCSSClass("caption")
	sortLabel.SetHAlign(gtk.AlignStart)
	actionsBox.Append(sortLabel)

	sortBox := gtk.NewBox(gtk.OrientationHorizontal, 0)
	sortBox.AddCSSClass("linked")
	
	sortNameBtn := gtk.NewButtonWithLabel("Name")
	sortTimeBtn := gtk.NewButtonWithLabel("Time")
	sortSizeBtn := gtk.NewButtonWithLabel("Size")
	sortTypeBtn := gtk.NewButtonWithLabel("Type")

	sortNameBtn.ConnectClicked(func() {
		activePanel := m.getActivePanel()
		if activePanel != nil {
			activePanel.FileViewer.SortOrder = types.SortByName
			activePanel.FileViewer.Refresh(false)
		}
		actionsPopover.Popdown()
	})
	sortTimeBtn.ConnectClicked(func() {
		activePanel := m.getActivePanel()
		if activePanel != nil {
			activePanel.FileViewer.SortOrder = types.SortByTime
			activePanel.FileViewer.Refresh(false)
		}
		actionsPopover.Popdown()
	})
	sortSizeBtn.ConnectClicked(func() {
		activePanel := m.getActivePanel()
		if activePanel != nil {
			activePanel.FileViewer.SortOrder = types.SortBySize
			activePanel.FileViewer.Refresh(false)
		}
		actionsPopover.Popdown()
	})
	sortTypeBtn.ConnectClicked(func() {
		activePanel := m.getActivePanel()
		if activePanel != nil {
			activePanel.FileViewer.SortOrder = types.SortByType
			activePanel.FileViewer.Refresh(false)
		}
		actionsPopover.Popdown()
	})

	sortBox.Append(sortNameBtn)
	sortBox.Append(sortTimeBtn)
	sortBox.Append(sortSizeBtn)
	sortBox.Append(sortTypeBtn)
	sortNameBtn.SetHExpand(true)
	sortTimeBtn.SetHExpand(true)
	sortSizeBtn.SetHExpand(true)
	sortTypeBtn.SetHExpand(true)
	actionsBox.Append(sortBox)

	actionsBox.Append(gtk.NewSeparator(gtk.OrientationHorizontal))

	filterLabel := gtk.NewLabel("Filters")
	filterLabel.AddCSSClass("caption")
	filterLabel.SetHAlign(gtk.AlignStart)
	actionsBox.Append(filterLabel)
	
	filterWrapper := gtk.NewBox(gtk.OrientationVertical, 0)
	actionsBox.Append(filterWrapper)

	actionsPopover.Connect("map", func() {
		activePanel := m.getActivePanel()
		if activePanel != nil {
			for child := filterWrapper.FirstChild(); child != nil; {
				next := gtk.BaseWidget(child).NextSibling()
				filterWrapper.Remove(child)
				child = next
			}
			filterWrapper.Append(activePanel.FileViewer.FilterBox)
		}
	})
}
