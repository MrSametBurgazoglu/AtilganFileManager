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

	ds.SetTitle("Create New Folder")
	ds.SetResizable(false)
	ds.SetDefaultSize(350, 450)
	ds.SetModal(true)

	headerBar := gtk.NewHeaderBar()
	headerBar.SetShowTitleButtons(false)

	titleLabel := gtk.NewLabel("Create New Folder")
	titleLabel.AddCSSClass("title-4")
	headerBar.SetTitleWidget(titleLabel)

	cancelBtn := gtk.NewButtonWithLabel("Cancel")
	cancelBtn.ConnectClicked(func() {
		ds.Window.Destroy()
	})
	headerBar.PackStart(cancelBtn)

	createBtn := gtk.NewButtonWithLabel("Create")
	createBtn.AddCSSClass("suggested-action")
	createBtn.SetSensitive(false)

	headerBar.PackEnd(createBtn)
	ds.SetTitlebar(headerBar)

	mainBox := gtk.NewBox(gtk.OrientationVertical, 12)
	mainBox.SetMarginTop(16)
	mainBox.SetMarginBottom(16)
	mainBox.SetMarginStart(16)
	mainBox.SetMarginEnd(16)

	entryBox := gtk.NewBox(gtk.OrientationVertical, 4)
	entryLabel := gtk.NewLabel("Folder Name")
	entryLabel.SetHAlign(gtk.AlignStart)
	entryLabel.AddCSSClass("caption")

	ds.Entry.SetPlaceholderText("Enter folder name...")
	ds.Entry.SetHExpand(true)

	entryBox.Append(entryLabel)
	entryBox.Append(ds.Entry)
	mainBox.Append(entryBox)

	listLabel := gtk.NewLabel("Existing Folders")
	listLabel.SetHAlign(gtk.AlignStart)
	listLabel.AddCSSClass("caption")
	mainBox.Append(listLabel)

	ds.ListStore = gio.NewListStore(glib.TypeObject)

	factory := gtk.NewSignalListItemFactory()
	factory.ConnectSetup(func(o *glib.Object) {
		item := o.Cast().(*gtk.ListItem)
		label := gtk.NewLabel("")
		label.SetHAlign(gtk.AlignStart)
		label.SetMarginStart(8)
		label.SetMarginEnd(8)
		label.SetMarginTop(4)
		label.SetMarginBottom(4)
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
	scrolledWindow.AddCSSClass("view")
	scrolledWindow.SetHasFrame(true)

	mainBox.Append(scrolledWindow)

	validate := func() {
		ds.populateList()
		isValid := ds.isEntryValid()
		createBtn.SetSensitive(isValid)
		if ds.Entry.Text() == "" {
			ds.Entry.SetIconFromIconName(gtk.EntryIconSecondary, "")
		} else if isValid {
			ds.Entry.SetIconFromIconName(gtk.EntryIconSecondary, "object-select-symbolic")
			ds.Entry.RemoveCSSClass("error")
		} else {
			ds.Entry.SetIconFromIconName(gtk.EntryIconSecondary, "window-close-symbolic")
			ds.Entry.AddCSSClass("error")
		}
	}

	ds.Entry.Connect("notify::text", validate)

	createAction := func() {
		if ds.isEntryValid() {
			err := os.Mkdir(filepath.Join(ds.BasePath, ds.GetNewName()), 0755)
			if err != nil {
				log.Printf("Error creating directory %s: %v", ds.GetNewName(), err)
			} else {
				ds.Window.Destroy()
			}
			pathChanged("")
		}
	}

	ds.Entry.Connect("activate", createAction)
	createBtn.ConnectClicked(createAction)

	ds.SetChild(mainBox)
	ds.populateList()
	ds.Entry.GrabFocus()

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
