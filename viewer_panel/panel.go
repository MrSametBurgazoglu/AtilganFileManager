package viewer_panel

import (
	"github.com/MrSametBurgazoglu/atilgan/fileops"
	"github.com/MrSametBurgazoglu/atilgan/network"
	"github.com/MrSametBurgazoglu/atilgan/preferences"
	"github.com/MrSametBurgazoglu/atilgan/special_path"
	"github.com/MrSametBurgazoglu/atilgan/viewer"
	"github.com/diamondburned/gotk4/pkg/gtk/v4"
)

type Panel struct {
	*gtk.Box
	Path          string
	FileViewer    *viewer.FileViewer
	VideoViewer   *viewer.VideoViewer
	PictureViewer *viewer.PictureViewer
	MusicViewer   *viewer.MusicViewer
	NetworkViewer *network.NetworkViewer
	HomeViewer    *viewer.HomeViewer
	Stack         *gtk.Stack
	Config        *preferences.Config
}

func NewPanel(mainWindow *gtk.Window, path string, pathChanged func(string), specialPathManager *special_path.SpecialPathManager, config *preferences.Config) *Panel {
	panel := &Panel{
		Box:           gtk.NewBox(gtk.OrientationHorizontal, 0),
		Path:          path,
		FileViewer:    viewer.NewFileViewer(mainWindow, path, pathChanged, specialPathManager, config),
		VideoViewer:   viewer.NewVideoViewer(mainWindow, path, pathChanged, specialPathManager, config),
		PictureViewer: viewer.NewPictureViewer(mainWindow, path, pathChanged, specialPathManager, config),
		MusicViewer:   viewer.NewMusicViewer(mainWindow, path, pathChanged, specialPathManager, config),
		HomeViewer:    viewer.NewHomeViewer(mainWindow, pathChanged, config),
		Config:        config,
	}

	networkManager := specialPathManager.Paths["network"].(*network.Network)
	panel.NetworkViewer = network.NewNetworkViewer(networkManager, pathChanged)

	panel.Stack = gtk.NewStack()
	panel.Stack.SetTransitionType(gtk.StackTransitionTypeCrossfade)
	panel.Stack.AddNamed(panel.FileViewer, "file")
	panel.Stack.AddNamed(panel.VideoViewer, "video_viewer")
	panel.Stack.AddNamed(panel.PictureViewer, "picture_viewer")
	panel.Stack.AddNamed(panel.MusicViewer, "music_viewer")
	panel.Stack.AddNamed(panel.NetworkViewer, "network")
	panel.Stack.AddNamed(panel.HomeViewer, "home_viewer")

	panel.Box.AddCSSClass("viewer-panel")
	panel.SetHExpand(true)
	panel.Append(panel.Stack)

	panel.SetPath(path)

	return panel
}

func (p *Panel) SetPath(path string) {
	p.Path = path
	
	var childName string
	switch {
	case path == "home://":
		p.HomeViewer.Refresh(true)
		childName = "home_viewer"
	case path == "network://":
		p.NetworkViewer.Refresh()
		childName = "network"
	case path == fileops.GetVideosPath():
		p.VideoViewer.SetPath(path)
		childName = "video_viewer"
	case path == fileops.GetPicturesPath():
		p.PictureViewer.SetPath(path)
		childName = "picture_viewer"
	case path == fileops.GetMusicPath():
		p.MusicViewer.SetPath(path)
		childName = "music_viewer"
	default:
		p.FileViewer.SetPath(path)
		childName = "file"
	}
	
	p.Stack.SetVisibleChildName(childName)
}
