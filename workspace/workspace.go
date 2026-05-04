package workspace

import (
	"path/filepath"

	"github.com/MrSametBurgazoglu/atilgan/special_path"
	"github.com/MrSametBurgazoglu/atilgan/viewer_panel"
	"github.com/diamondburned/gotk4-adwaita/pkg/adw"
	"github.com/diamondburned/gotk4/pkg/gtk/v4"
)

type Workspace struct {
	*adw.TabView
	TabBar             *adw.TabBar
	MainWindow         *gtk.Window
	SpecialPathManager *special_path.SpecialPathManager
	OnPathChanged      func(string)
	SetupPanel         func(*viewer_panel.Panel)
	PagePanels         map[*adw.TabPage]*viewer_panel.Panel
}

func NewWorkspace(mainWindow *gtk.Window, specialPathManager *special_path.SpecialPathManager, onPathChanged func(string)) *Workspace {
	ws := &Workspace{
		TabView:            adw.NewTabView(),
		TabBar:             adw.NewTabBar(),
		MainWindow:         mainWindow,
		SpecialPathManager: specialPathManager,
		OnPathChanged:      onPathChanged,
		PagePanels:         make(map[*adw.TabPage]*viewer_panel.Panel),
	}
	ws.TabBar.SetView(ws.TabView)

	return ws
}

func (ws *Workspace) NewTab(path string) *adw.TabPage {
	panel := viewer_panel.NewPanel(ws.MainWindow, path, ws.OnPathChanged, ws.SpecialPathManager)
	if ws.SetupPanel != nil {
		ws.SetupPanel(panel)
	}

	panel.SetHExpand(true)
	panel.SetVExpand(true)

	adwPage := ws.TabView.Append(panel)
	adwPage.SetTitle(filepath.Base(path))
	if adwPage.Title() == "." || adwPage.Title() == "/" {
		adwPage.SetTitle("Root")
	}

	ws.PagePanels[adwPage] = panel

	ws.TabView.SetSelectedPage(adwPage)

	return adwPage
}

func (ws *Workspace) GetActivePanel() *viewer_panel.Panel {
	selectedPage := ws.TabView.SelectedPage()
	if selectedPage == nil {
		return nil
	}

	return ws.PagePanels[selectedPage]
}
