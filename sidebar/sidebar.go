package sidebar

import (
	"os"
	"os/user"
	"runtime"

	"github.com/adrg/xdg"
	"github.com/diamondburned/gotk4/pkg/gtk/v4"
)

type Sidebar struct {
	*gtk.Box
	buttons     map[string]*gtk.Button
	currentPath string
}

func NewSidebar(pathChanged func(string)) *Sidebar {
	box := gtk.NewBox(gtk.OrientationHorizontal, 6)
	box.SetHExpand(false)
	box.SetHAlign(gtk.AlignCenter)
	box.AddCSSClass("sidebar")

	sidebar := &Sidebar{
		Box:     box,
		buttons: make(map[string]*gtk.Button),
	}

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

	homeButton := gtk.NewButtonFromIconName("user-home-symbolic")
	homeButton.AddCSSClass("sidebar-button")
	homeButton.SetTooltipText("home")
	sidebar.buttons[homeDir] = homeButton

	recentButton := gtk.NewButtonFromIconName("document-open-recent-symbolic")
	recentButton.AddCSSClass("sidebar-button")
	recentButton.SetTooltipText("recent")
	sidebar.buttons["recent://"] = recentButton

	tagsButton := gtk.NewButtonFromIconName("tag-symbolic")
	tagsButton.AddCSSClass("sidebar-button")
	tagsButton.SetTooltipText("tags")
	sidebar.buttons["tags://"] = tagsButton

	trashButton := gtk.NewButtonFromIconName("user-trash-symbolic")
	trashButton.AddCSSClass("sidebar-button")
	trashButton.SetTooltipText("trash")
	sidebar.buttons["trash://"] = trashButton

	desktopButton := gtk.NewButtonFromIconName("user-desktop-symbolic")
	desktopButton.AddCSSClass("sidebar-button")
	desktopButton.SetTooltipText("desktop")
	sidebar.buttons[desktop] = desktopButton

	downloadsButton := gtk.NewButtonFromIconName("folder-download-symbolic")
	downloadsButton.AddCSSClass("sidebar-button")
	downloadsButton.SetTooltipText("downloads")
	sidebar.buttons[downloads] = downloadsButton

	documentsButton := gtk.NewButtonFromIconName("folder-documents-symbolic")
	documentsButton.AddCSSClass("sidebar-button")
	documentsButton.SetTooltipText("documents")
	sidebar.buttons[documents] = documentsButton

	picturesButton := gtk.NewButtonFromIconName("folder-pictures-symbolic")
	picturesButton.AddCSSClass("sidebar-button")
	picturesButton.SetTooltipText("pictures")
	sidebar.buttons[pictures] = picturesButton

	musicButton := gtk.NewButtonFromIconName("folder-music-symbolic")
	musicButton.AddCSSClass("sidebar-button")
	musicButton.SetTooltipText("music")
	sidebar.buttons[music] = musicButton

	videosButton := gtk.NewButtonFromIconName("folder-videos-symbolic")
	videosButton.AddCSSClass("sidebar-button")
	videosButton.SetTooltipText("videos")
	sidebar.buttons[videos] = videosButton

	if homeDir != "" {
		box.Append(homeButton)
	}
	box.Append(recentButton)
	box.Append(trashButton)
	box.Append(desktopButton)
	box.Append(downloadsButton)
	box.Append(documentsButton)
	box.Append(picturesButton)
	box.Append(musicButton)
	box.Append(videosButton)
	box.Append(tagsButton)

	homeButton.ConnectClicked(func() {
		pathChanged(homeDir)
	})

	recentButton.ConnectClicked(func() {
		pathChanged("recent://")
	})

	tagsButton.ConnectClicked(func() {
		pathChanged("tags://")
	})

	trashButton.ConnectClicked(func() {
		pathChanged("trash://")
	})

	desktopButton.ConnectClicked(func() {
		pathChanged(desktop)
	})
	downloadsButton.ConnectClicked(func() {
		pathChanged(downloads)
	})
	documentsButton.ConnectClicked(func() {
		pathChanged(documents)
	})
	picturesButton.ConnectClicked(func() {
		pathChanged(pictures)
	})
	musicButton.ConnectClicked(func() {
		pathChanged(music)
	})
	videosButton.ConnectClicked(func() {
		pathChanged(videos)
	})

	return sidebar
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
