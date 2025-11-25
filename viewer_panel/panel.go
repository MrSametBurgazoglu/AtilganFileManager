package viewer_panel

import (
	"github.com/MrSametBurgazoglu/atilgan/special_path"
	"github.com/MrSametBurgazoglu/atilgan/viewer"
	"github.com/diamondburned/gotk4/pkg/gtk/v4"
)

type Panel struct {
	*gtk.Box
	Path       string
	FileViewer *viewer.FileViewer
}

func NewPanel(mainWindow *gtk.Window, path string, pathChanged func(string), specialPathManager *special_path.SpecialPathManager) *Panel {
	panel := &Panel{
		Box:        gtk.NewBox(gtk.OrientationHorizontal, 0),
		Path:       path,
		FileViewer: viewer.NewFileViewer(mainWindow, path, pathChanged, specialPathManager),
	}
	panel.Box.AddCSSClass("preview-panel")
	panel.SetHExpand(false)
	panel.Append(panel.FileViewer)
	return panel
}
