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
	} else if path == fileops.GetVideosPath() {
		p.VideoViewer.SetPath(path)
		p.Stack.SetVisibleChildName("video_viewer")
	} else if path == fileops.GetPicturesPath() {
		p.PictureViewer.SetPath(path)
		p.Stack.SetVisibleChildName("picture_viewer")
	} else if path == fileops.GetMusicPath() {
		p.MusicViewer.SetPath(path)
		p.Stack.SetVisibleChildName("music_viewer")
	} else {
		p.FileViewer.SetPath(path)
		p.Stack.SetVisibleChildName("file")
	}
}
