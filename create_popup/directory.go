package create_popup

import (
	"log"
	"os"
	"path/filepath"
	"strings"

	"github.com/diamondburned/gotk4/pkg/gio/v2"
	"github.com/diamondburned/gotk4/pkg/glib/v2"
	"github.com/diamondburned/gotk4/pkg/gtk/v4"
)

type DirectorySelector struct {
	*gtk.Window
	Entry     *gtk.Entry
	ListView  *gtk.ListView
	ListStore *gio.ListStore
	BasePath  string
}

func NewDirectorySelector(path string, pathChanged func(string)) *DirectorySelector {
	ds := &DirectorySelector{
		Window:   gtk.NewWindow(),
		Entry:    gtk.NewEntry(),
		BasePath: path,
	}

	ds.SetResizable(false)
	ds.SetDefaultSize(400, 600)
	ds.SetModal(true)
	headerBar := gtk.NewHeaderBar()
	headerBar.SetTitleWidget(gtk.NewLabel("Directory Creator"))
	ds.SetTitlebar(headerBar)
	box := gtk.NewBox(gtk.OrientationVertical, 5)
	box.Append(ds.Entry)

	ds.ListStore = gio.NewListStore(glib.TypeObject)

	factory := gtk.NewSignalListItemFactory()
	factory.ConnectSetup(func(o *glib.Object) {
		item := o.Cast().(*gtk.ListItem)
		label := gtk.NewLabel("")
		item.SetChild(label)
	})
	factory.ConnectBind(func(o *glib.Object) {
		item := o.Cast().(*gtk.ListItem)
		label := item.Child().(*gtk.Label)
		obj := item.Item()
		str := obj.Cast().(*gtk.StringObject).String()
		label.SetText(str)
	})

	selection := gtk.NewNoSelection(ds.ListStore)
	ds.ListView = gtk.NewListView(selection, &factory.ListItemFactory)
	scrolledWindow := gtk.NewScrolledWindow()
	scrolledWindow.SetChild(ds.ListView)
	scrolledWindow.SetVExpand(true)

	ds.Entry.Connect("notify::text", func() {
		ds.populateList()
		if ds.isEntryValid() {
			ds.Entry.SetIconFromIconName(gtk.EntryIconSecondary, "object-select-symbolic")
		} else {
			ds.Entry.SetIconFromIconName(gtk.EntryIconSecondary, "window-close-symbolic")
		}
	})

	ds.Entry.Connect("activate", func() {
		if ds.isEntryValid() {
			err := os.Mkdir(filepath.Join(ds.BasePath, ds.GetNewName()), 0755)
			if err != nil {
				log.Printf("Error creating directory %s: %v", ds.GetNewName(), err)
			} else {
				ds.Window.Destroy()
			}
			pathChanged("")
		} else {
			ds.Entry.SetIconFromIconName(gtk.EntryIconSecondary, "window-close-symbolic")
		}
	})

	box.Append(scrolledWindow)
	ds.SetChild(box)
	ds.populateList()

	return ds
}

func (ds *DirectorySelector) populateList() {
	ds.ListStore.RemoveAll()
	names := ds.readPathContents()
	for _, name := range names {
		if strings.HasPrefix(name, ds.Entry.Text()) {
			ds.ListStore.Append(gtk.NewStringObject(name).Object)
		}
	}
}

func (ds *DirectorySelector) readPathContents() []string {
	entries, err := os.ReadDir(ds.BasePath)
	if err != nil {
		log.Printf("Error reading path %s: %v", ds.BasePath, err)
		return []string{}
	}

	var names []string
	for _, entry := range entries {
		if entry.IsDir() {
			names = append(names, entry.Name())
		}
	}
	return names
}

func (ds *DirectorySelector) GetNewName() string {
	return ds.Entry.Text()
}

func (ds *DirectorySelector) isEntryValid() bool {
	searchText := ds.Entry.Text()

	alreadyExists := false
	for i := uint(0); i < ds.ListStore.NItems(); i++ {
		item := ds.ListStore.Item(i)
		if item.Cast().(*gtk.StringObject).String() == searchText {
			alreadyExists = true
			break
		}
	}

	hasIllegalChars := false
	if strings.ContainsAny(searchText, "/\\:*?\"<>|") || len(searchText) == 0 {
		hasIllegalChars = true
	}

	return !(alreadyExists || hasIllegalChars)
}

func (ds *DirectorySelector) RefreshList() {
	ds.populateList()
}
