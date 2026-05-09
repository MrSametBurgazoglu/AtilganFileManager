package main

import (
	"path/filepath"

	"github.com/MrSametBurgazoglu/atilgan/fileops"
	"github.com/MrSametBurgazoglu/atilgan/header"
	"github.com/MrSametBurgazoglu/atilgan/pathbar"
	"github.com/MrSametBurgazoglu/atilgan/preferences"
	"github.com/MrSametBurgazoglu/atilgan/previewer"
	"github.com/MrSametBurgazoglu/atilgan/previewer_panel"
	"github.com/MrSametBurgazoglu/atilgan/search"
	"github.com/MrSametBurgazoglu/atilgan/sidebar"
	"github.com/MrSametBurgazoglu/atilgan/special_path"
	"github.com/MrSametBurgazoglu/atilgan/types"
	"github.com/MrSametBurgazoglu/atilgan/utils"
	"github.com/MrSametBurgazoglu/atilgan/viewer"
	"github.com/MrSametBurgazoglu/atilgan/viewer_panel"
	"github.com/MrSametBurgazoglu/atilgan/workspace"
	"github.com/diamondburned/gotk4-adwaita/pkg/adw"
	"github.com/diamondburned/gotk4/pkg/gtk/v4"
)

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

func NewMainBox(mainWindow *adw.ApplicationWindow, headerBar *header.HeaderBar, initialPath string) *MainBox {
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

	curdir := initialPath
	if curdir == "" {
		curdir = "home://"
	} else if curdir != "home://" {
		// Ensure it's an absolute path
		if abs, err := filepath.Abs(curdir); err == nil {
			curdir = abs
		}
	}

	var err error
	mainBox.SpecialPaths, err = special_path.NewSpecialPathManager()
	if err != nil {
		utils.LogError(err)
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

	mainBox.setupShortcuts(mainWindow, copyCutPreviewer)

	mainBox.updatePreviewer()

	mainBox.applyPreferences()

	return mainBox
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

	m.PreviewerPanel.SetCurrentDirectory(path)
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


