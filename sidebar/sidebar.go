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
	box := gtk.NewBox(gtk.OrientationVertical, 8)
	box.SetHExpand(false)
	box.SetVExpand(true)
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

	createButton := func(iconName, tooltip string) *gtk.Button {
		btn := gtk.NewButtonFromIconName(iconName)
		btn.AddCSSClass("sidebar-button")
		btn.SetTooltipText(tooltip)
		btn.SetHAlign(gtk.AlignCenter)
		btn.SetSizeRequest(40, 40)
		return btn
	}

	homeButton := createButton("user-home-symbolic", "home")
	sidebar.buttons[homeDir] = homeButton

	recentButton := createButton("document-open-recent-symbolic", "recent")
	sidebar.buttons["recent://"] = recentButton

	tagsButton := createButton("tag-symbolic", "tags")
	sidebar.buttons["tags://"] = tagsButton

	trashButton := createButton("user-trash-symbolic", "trash")
	sidebar.buttons["trash://"] = trashButton

	desktopButton := createButton("user-desktop-symbolic", "desktop")
	sidebar.buttons[desktop] = desktopButton

	downloadsButton := createButton("folder-download-symbolic", "downloads")
	sidebar.buttons[downloads] = downloadsButton

	documentsButton := createButton("folder-documents-symbolic", "documents")
	sidebar.buttons[documents] = documentsButton

	picturesButton := createButton("folder-pictures-symbolic", "pictures")
	sidebar.buttons[pictures] = picturesButton

	musicButton := createButton("folder-music-symbolic", "music")
	sidebar.buttons[music] = musicButton

	videosButton := createButton("folder-videos-symbolic", "videos")
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
