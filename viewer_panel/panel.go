package viewer_panel

import (
	"github.com/MrSametBurgazoglu/atilgan/network"
	"github.com/MrSametBurgazoglu/atilgan/special_path"
	"github.com/MrSametBurgazoglu/atilgan/viewer"
	"github.com/diamondburned/gotk4/pkg/gtk/v4"
)

type Panel struct {
	*gtk.Box
	Path          string
	FileViewer    *viewer.FileViewer
	NetworkViewer *network.NetworkViewer
	Stack         *gtk.Stack
}

func NewPanel(mainWindow *gtk.Window, path string, pathChanged func(string), specialPathManager *special_path.SpecialPathManager) *Panel {
	panel := &Panel{
		Box:        gtk.NewBox(gtk.OrientationHorizontal, 0),
		Path:       path,
		FileViewer: viewer.NewFileViewer(mainWindow, path, pathChanged, specialPathManager),
	}

	networkManager := specialPathManager.Paths["network"].(*network.Network)
	panel.NetworkViewer = network.NewNetworkViewer(networkManager, pathChanged)

	panel.Stack = gtk.NewStack()
	panel.Stack.SetTransitionType(gtk.StackTransitionTypeCrossfade)
	panel.Stack.AddNamed(panel.FileViewer, "file")
	panel.Stack.AddNamed(panel.NetworkViewer, "network")

	panel.Box.AddCSSClass("viewer-panel")
	panel.SetHExpand(true)
	panel.Append(panel.Stack)
	
	panel.Stack.SetVisibleChildName("file")

	return panel
}

func (p *Panel) SetPath(path string) {
	p.Path = path
	if path == "network://" {
		p.NetworkViewer.Refresh()
		p.Stack.SetVisibleChildName("network")
	} else {
		p.FileViewer.SetPath(path)
		p.Stack.SetVisibleChildName("file")
	}
}
