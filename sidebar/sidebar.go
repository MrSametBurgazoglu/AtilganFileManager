package sidebar

import (
	"os"
	"os/user"
	"runtime"

	"github.com/adrg/xdg"
	"github.com/MrSametBurgazoglu/atilgan/devices"
	"github.com/diamondburned/gotk4/pkg/gtk/v4"
)

type Sidebar struct {
	*gtk.Box
	locationsBox  *gtk.Box
	devicesBox    *devices.DeviceManager
	buttons       map[string]*gtk.Button
	currentPath   string
	pathChanged   func(string)
}

func NewSidebar(pathChanged func(string)) *Sidebar {
	box := gtk.NewBox(gtk.OrientationVertical, 8)
	box.SetHExpand(false)
	box.SetVExpand(true)
	box.AddCSSClass("sidebar")

	sidebar := &Sidebar{
		Box:          box,
		locationsBox: gtk.NewBox(gtk.OrientationVertical, 8),
		devicesBox:   devices.NewDeviceManager(pathChanged),
		buttons:      make(map[string]*gtk.Button),
		pathChanged:  pathChanged,
	}

	box.Append(sidebar.locationsBox)
	box.Append(sidebar.devicesBox)

	homeDir, err := getHomeDir()
	if err != nil {
		homeDir = ""
	}

	desktop := xdg.UserDirs.Desktop
	downloads := xdg.UserDirs.Download
	documents := xdg.UserDirs.Documents
	pictures := xdg.UserDirs.Pictures
	music := xdg.UserDirs.Music
	videos := xdg.UserDirs.Videos

	sidebar.addLocationButton("user-home-symbolic", "Home", homeDir)
	sidebar.addLocationButton("document-open-recent-symbolic", "Recent", "recent://")
	sidebar.addLocationButton("user-trash-symbolic", "Trash", "trash://")
	sidebar.addLocationButton("user-desktop-symbolic", "Desktop", desktop)
	sidebar.addLocationButton("folder-download-symbolic", "Downloads", downloads)
	sidebar.addLocationButton("folder-documents-symbolic", "Documents", documents)
	sidebar.addLocationButton("folder-pictures-symbolic", "Pictures", pictures)
	sidebar.addLocationButton("folder-music-symbolic", "Music", music)
	sidebar.addLocationButton("folder-videos-symbolic", "Videos", videos)
	sidebar.addLocationButton("tag-symbolic", "Tags", "tags://")

	return sidebar
}

func (s *Sidebar) addLocationButton(iconName, labelText, path string) {
	if path == "" {
		return
	}
	btn := s.createButton(iconName, labelText, path)
	s.locationsBox.Append(btn)
	s.buttons[path] = btn
}

func (s *Sidebar) createButton(iconName, labelText, path string) *gtk.Button {
	btn := gtk.NewButton()

	box := gtk.NewBox(gtk.OrientationHorizontal, 8)
	box.SetHAlign(gtk.AlignStart)

	icon := gtk.NewImageFromIconName(iconName)
	label := gtk.NewLabel(labelText)

	box.Append(icon)
	box.Append(label)

	btn.SetChild(box)
	btn.AddCSSClass("sidebar-button")
	btn.SetHAlign(gtk.AlignFill)

	btn.ConnectClicked(func() {
		s.pathChanged(path)
	})

	return btn
}

func (s *Sidebar) SetPath(path string) {
	s.currentPath = path
	for btnPath, button := range s.buttons {
		if btnPath == path {
			button.AddCSSClass("selected")
		} else {
			button.RemoveCSSClass("selected")
		}
	}
	s.devicesBox.SetPath(path)
}

func getHomeDir() (string, error) {
	currentUser, err := user.Current()
	if err == nil {
		return currentUser.HomeDir, nil
	}

	if runtime.GOOS == "windows" {
		return os.Getenv("USERPROFILE"), nil
	}
	return os.Getenv("HOME"), nil
}
