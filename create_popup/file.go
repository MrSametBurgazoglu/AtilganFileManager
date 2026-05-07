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
	fs.SetTitle("Create New File")
	fs.SetResizable(false)
	fs.SetDefaultSize(350, 450)
	fs.SetModal(true)

	headerBar := gtk.NewHeaderBar()
	headerBar.SetShowTitleButtons(false)
	
	titleLabel := gtk.NewLabel("Create New File")
	titleLabel.AddCSSClass("title-4")
	headerBar.SetTitleWidget(titleLabel)
	
	cancelBtn := gtk.NewButtonWithLabel("Cancel")
	cancelBtn.ConnectClicked(func() {
		fs.Window.Destroy()
	})
	headerBar.PackStart(cancelBtn)
	
	createBtn := gtk.NewButtonWithLabel("Create")
	createBtn.AddCSSClass("suggested-action")
	createBtn.SetSensitive(false)
	
	headerBar.PackEnd(createBtn)
	fs.SetTitlebar(headerBar)

	mainBox := gtk.NewBox(gtk.OrientationVertical, 12)
	mainBox.SetMarginTop(16)
	mainBox.SetMarginBottom(16)
	mainBox.SetMarginStart(16)
	mainBox.SetMarginEnd(16)

	entryBox := gtk.NewBox(gtk.OrientationVertical, 4)
	entryLabel := gtk.NewLabel("File Name")
	entryLabel.SetHAlign(gtk.AlignStart)
	entryLabel.AddCSSClass("caption")
	
	fs.Entry.SetPlaceholderText("Enter file name...")
	fs.Entry.SetHExpand(true)
	
	entryBox.Append(entryLabel)
	entryBox.Append(fs.Entry)
	mainBox.Append(entryBox)

	listLabel := gtk.NewLabel("Existing Files")
	listLabel.SetHAlign(gtk.AlignStart)
	listLabel.AddCSSClass("caption")
	mainBox.Append(listLabel)

	fs.ListStore = gio.NewListStore(glib.TypeObject)

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

	selection := gtk.NewNoSelection(fs.ListStore)
	fs.ListView = gtk.NewListView(selection, &factory.ListItemFactory)
	
	scrolledWindow := gtk.NewScrolledWindow()
	scrolledWindow.SetChild(fs.ListView)
	scrolledWindow.SetVExpand(true)
	scrolledWindow.AddCSSClass("view")
	scrolledWindow.SetHasFrame(true)

	mainBox.Append(scrolledWindow)

	validate := func() {
		fs.populateList()
		isValid := fs.isEntryValid()
		createBtn.SetSensitive(isValid)
		if fs.Entry.Text() == "" {
			fs.Entry.SetIconFromIconName(gtk.EntryIconSecondary, "")
		} else if isValid {
			fs.Entry.SetIconFromIconName(gtk.EntryIconSecondary, "object-select-symbolic")
			fs.Entry.RemoveCSSClass("error")
		} else {
			fs.Entry.SetIconFromIconName(gtk.EntryIconSecondary, "window-close-symbolic")
			fs.Entry.AddCSSClass("error")
		}
	}

	fs.Entry.Connect("notify::text", validate)

	createAction := func() {
		if fs.isEntryValid() {
			file, err := os.Create(filepath.Join(fs.BasePath, fs.GetNewName()))
			if err != nil {
				log.Printf("Error creating file %s: %v", fs.GetNewName(), err)
			} else {
				fs.Window.Destroy()
				file.Close()
			}
			pathChanged("")
		}
	}

	fs.Entry.Connect("activate", createAction)
	createBtn.ConnectClicked(createAction)

	fs.SetChild(mainBox)
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
