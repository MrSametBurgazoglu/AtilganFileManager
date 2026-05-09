package viewer

import (
	"os"
	"os/user"
	"runtime"

	"github.com/MrSametBurgazoglu/atilgan/devices"
	"github.com/MrSametBurgazoglu/atilgan/preferences"
	"github.com/adrg/xdg"
	"github.com/diamondburned/gotk4/pkg/gtk/v4"
	"github.com/diamondburned/gotk4/pkg/pango"
)

type HomeViewer struct {
	*gtk.Box
	pathChanged   func(string)
	config        *preferences.Config
	devicesBox    *gtk.FlowBox
	deviceManager *devices.DeviceManager
}

func NewHomeViewer(mainWindow *gtk.Window, pathChanged func(string), config *preferences.Config) *HomeViewer {
	box := gtk.NewBox(gtk.OrientationVertical, 0)
	box.AddCSSClass("home-page-bg")
	box.SetHExpand(true)
	box.SetVExpand(true)

	viewer := &HomeViewer{
		Box:         box,
		pathChanged: pathChanged,
		config:      config,
	}

	scrolled := gtk.NewScrolledWindow()
	scrolled.SetHExpand(true)
	scrolled.SetVExpand(true)
	scrolled.SetPolicy(gtk.PolicyNever, gtk.PolicyAutomatic)

	contentBox := gtk.NewBox(gtk.OrientationVertical, 0)
	contentBox.SetHAlign(gtk.AlignCenter)
	contentBox.SetMarginTop(32)
	contentBox.SetMarginBottom(32)
	contentBox.SetMarginStart(32)
	contentBox.SetMarginEnd(32)
	contentBox.SetSizeRequest(800, -1)

	// Logo / Welcome
	welcomeLabel := gtk.NewLabel("Atilgan File Manager")
	welcomeLabel.AddCSSClass("title-1")
	welcomeLabel.SetMarginBottom(16)
	contentBox.Append(welcomeLabel)

	// Search Entry
	searchEntry := gtk.NewSearchEntry()
	searchEntry.SetPlaceholderText("Search in Home Directory...")
	searchEntry.AddCSSClass("home-search-bar")
	searchEntry.SetHExpand(true)
	searchEntry.ConnectActivate(func() {
		query := searchEntry.Text()
		if query != "" {
			// Trigger navigation to home and we could set search query, 
			// but for now we navigate to home and log. A future enhancement could pass the query.
			homeDir, _ := getHomeDir()
			viewer.pathChanged(homeDir)
		}
	})
	contentBox.Append(searchEntry)

	// User Directories
	dirsTitle := gtk.NewLabel("Locations")
	dirsTitle.SetHAlign(gtk.AlignStart)
	dirsTitle.AddCSSClass("home-section-title")
	contentBox.Append(dirsTitle)

	dirsFlow := gtk.NewFlowBox()
	dirsFlow.SetSelectionMode(gtk.SelectionNone)
	dirsFlow.SetMaxChildrenPerLine(4)
	dirsFlow.SetMinChildrenPerLine(2)
	dirsFlow.SetRowSpacing(12)
	dirsFlow.SetColumnSpacing(12)
	contentBox.Append(dirsFlow)

	homeDir, _ := getHomeDir()
	viewer.addCard(dirsFlow, "user-home-symbolic", "Home", homeDir)
	viewer.addCard(dirsFlow, "user-desktop-symbolic", "Desktop", xdg.UserDirs.Desktop)
	viewer.addCard(dirsFlow, "folder-download-symbolic", "Downloads", xdg.UserDirs.Download)
	viewer.addCard(dirsFlow, "folder-documents-symbolic", "Documents", xdg.UserDirs.Documents)
	viewer.addCard(dirsFlow, "folder-pictures-symbolic", "Pictures", xdg.UserDirs.Pictures)
	viewer.addCard(dirsFlow, "folder-videos-symbolic", "Videos", xdg.UserDirs.Videos)
	viewer.addCard(dirsFlow, "folder-music-symbolic", "Music", xdg.UserDirs.Music)

	// Quick Access
	quickTitle := gtk.NewLabel("Quick Access")
	quickTitle.SetHAlign(gtk.AlignStart)
	quickTitle.AddCSSClass("home-section-title")
	contentBox.Append(quickTitle)

	quickFlow := gtk.NewFlowBox()
	quickFlow.SetSelectionMode(gtk.SelectionNone)
	quickFlow.SetMaxChildrenPerLine(4)
	quickFlow.SetMinChildrenPerLine(2)
	quickFlow.SetRowSpacing(12)
	quickFlow.SetColumnSpacing(12)
	contentBox.Append(quickFlow)

	viewer.addCard(quickFlow, "document-open-recent-symbolic", "Recent Files", "recent://")
	viewer.addCard(quickFlow, "user-trash-symbolic", "Trash", "trash://")
	viewer.addCard(quickFlow, "user-bookmarks-symbolic", "Tags", "tags://")

	scrolled.SetChild(contentBox)
	viewer.Append(scrolled)

	return viewer
}

func (v *HomeViewer) addCard(flowBox *gtk.FlowBox, iconName, labelText, path string) {
	if path == "" {
		return
	}

	card := gtk.NewBox(gtk.OrientationVertical, 8)
	card.AddCSSClass("home-card")
	card.SetHAlign(gtk.AlignFill)

	icon := gtk.NewImageFromIconName(iconName)
	icon.SetPixelSize(32)
	icon.AddCSSClass("home-card-icon")
	
	label := gtk.NewLabel(labelText)
	label.SetEllipsize(pango.EllipsizeEnd)
	label.AddCSSClass("home-card-label")

	card.Append(icon)
	card.Append(label)

	btn := gtk.NewButton()
	btn.SetChild(card)
	btn.AddCSSClass("flat")
	btn.ConnectClicked(func() {
		v.pathChanged(path)
	})

	flowBox.Append(btn)
}

func (v *HomeViewer) Refresh(force bool) {
	// Future: refresh devices or other dynamic content here.
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
