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

type FileSelector struct {
	*gtk.Window
	Entry     *gtk.Entry
	ListView  *gtk.ListView
	ListStore *gio.ListStore
	BasePath  string
}

func NewFileSelector(path string, pathChanged func(string)) *FileSelector {
	fs := &FileSelector{
		Window:   gtk.NewWindow(),
		Entry:    gtk.NewEntry(),
		BasePath: path,
	}
	fs.SetResizable(false)
	fs.SetDefaultSize(400, 600)
	fs.SetModal(true)
	headerBar := gtk.NewHeaderBar()
	headerBar.SetTitleWidget(gtk.NewLabel("File Creator"))
	fs.SetTitlebar(headerBar)

	box := gtk.NewBox(gtk.OrientationVertical, 5)
	box.Append(fs.Entry)

	fs.ListStore = gio.NewListStore(glib.TypeObject)

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

	selection := gtk.NewNoSelection(fs.ListStore)
	fs.ListView = gtk.NewListView(selection, &factory.ListItemFactory)
	scrolledWindow := gtk.NewScrolledWindow()
	scrolledWindow.SetChild(fs.ListView)
	scrolledWindow.SetVExpand(true)

	fs.Entry.Connect("notify::text", func() {
		fs.populateList()
		if fs.isEntryValid() {
			fs.Entry.SetIconFromIconName(gtk.EntryIconSecondary, "object-select-symbolic")
		} else {
			fs.Entry.SetIconFromIconName(gtk.EntryIconSecondary, "window-close-symbolic")
		}
	})

	fs.Entry.Connect("activate", func() {
		if fs.isEntryValid() {
			file, err := os.Create(filepath.Join(fs.BasePath, fs.GetNewName()))
			if err != nil {
				log.Printf("Error creating file %s: %v", fs.GetNewName(), err)
			} else {
				fs.Window.Destroy()
				file.Close()
			}
			pathChanged("")
		} else {
			fs.Entry.SetIconFromIconName(gtk.EntryIconSecondary, "window-close-symbolic")
		}
	})

	box.Append(scrolledWindow)
	fs.SetChild(box)
	fs.populateList()
	fs.Entry.GrabFocus()

	return fs
}

func (fs *FileSelector) populateList() {
	fs.ListStore.RemoveAll()
	names := fs.readPathContents()
	for _, name := range names {
		if strings.HasPrefix(name, fs.Entry.Text()) {
			fs.ListStore.Append(gtk.NewStringObject(name).Object)
		}
	}
}

func (fs *FileSelector) readPathContents() []string {
	entries, err := os.ReadDir(fs.BasePath)
	if err != nil {
		log.Printf("Error reading path %s: %v", fs.BasePath, err)
		return []string{}
	}

	var names []string
	for _, entry := range entries {
		if !entry.IsDir() {
			names = append(names, entry.Name())
		}
	}
	return names
}

func (fs *FileSelector) GetNewName() string {
	return fs.Entry.Text()
}

func (fs *FileSelector) isEntryValid() bool {
	searchText := fs.Entry.Text()

	alreadyExists := false
	for i := uint(0); i < fs.ListStore.NItems(); i++ {
		item := fs.ListStore.Item(i)
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

func (fs *FileSelector) RefreshList() {
	fs.populateList()
}
