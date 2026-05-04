package previewer

import (
	"fmt"
	"os"
	"path/filepath"
	"sort"
	"strings"

	"github.com/MrSametBurgazoglu/atilgan/file_list"
	"github.com/MrSametBurgazoglu/atilgan/fileops"
	"github.com/MrSametBurgazoglu/atilgan/preferences"
	"github.com/MrSametBurgazoglu/atilgan/special_path"
	"github.com/MrSametBurgazoglu/atilgan/types"
	"github.com/diamondburned/gotk4/pkg/gtk/v4"
)

type DirPreviewer struct {
	*gtk.Box
	Path               string
	changePath         func(string)
	FileViewerList     *file_list.FileList
	folderIcon         *gtk.Image
	folderName         *gtk.Label
	specialPathManager *special_path.SpecialPathManager
	config           *preferences.Config
}

func NewDirPreviewer(path string, changePath func(string), specialPathManager *special_path.SpecialPathManager, parent *gtk.Window, config *preferences.Config) *DirPreviewer {
	viewer := &DirPreviewer{
		Box:                gtk.NewBox(gtk.OrientationVertical, 2),
		Path:               path,
		changePath:         changePath,
		FileViewerList:     file_list.NewFileList(false, specialPathManager, parent, config),
		specialPathManager: specialPathManager,
		config:            config,
	}
	viewer.folderIcon = gtk.NewImageFromIconName("folder-symbolic")
	viewer.folderIcon.SetPixelSize(20)
	viewer.folderName = gtk.NewLabel("")
	viewer.folderName.AddCSSClass("dir-previewer-title")
	viewer.folderName.SetEllipsize(1)

	viewer.FileViewerList.SetHExpand(true)
	viewer.FileViewerList.SetVExpand(true)
	viewer.FileViewerList.SetVAlign(gtk.AlignFill)
	viewer.Box.SetVExpand(true)
	viewer.Box.SetHExpand(true)
	viewer.Box.SetVAlign(gtk.AlignFill)

	headerBox := gtk.NewBox(gtk.OrientationHorizontal, 6)
	headerBox.AddCSSClass("dir-previewer-header")
	viewer.Box.Append(headerBox)

	leftBox := gtk.NewBox(gtk.OrientationHorizontal, 6)
	leftBox.SetHExpand(true)
	headerBox.Append(leftBox)

	leftBox.Append(viewer.folderIcon)
	leftBox.Append(viewer.folderName)

	viewer.FileViewerList.PathChanged = func(path string) {
		changePath(path)
	}

	viewer.Box.Append(viewer.FileViewerList)

	viewer.Refresh()

	return viewer
}

func (viewer *DirPreviewer) SetPath(path string) {
	viewer.Path = path
	if strings.HasPrefix(path, "tags://") {
		viewer.folderName.SetText("Tag: " + strings.TrimPrefix(path, "tags://"))
	} else {
		viewer.folderName.SetText(filepath.Base(path))
	}
	viewer.Refresh()
}

func (viewer *DirPreviewer) Refresh() {
	if viewer.Path == "" {
		return
	}

	var items []*types.ListItem
	specialPath := viewer.specialPathManager.GetPath(viewer.Path)
	if specialPath != nil {
		items = specialPath.GetItems()
	} else {
		entries, err := os.ReadDir(viewer.Path)
		if err != nil {
			fmt.Println("Error reading directory:", err)
			return
		}

		sort.Slice(entries, func(i, j int) bool {
			return strings.ToLower(entries[i].Name()) < strings.ToLower(entries[j].Name())
		})

		for _, entry := range entries {
			name := entry.Name()

			if !viewer.config.ShowHidden && strings.HasPrefix(name, ".") {
				continue
			}

			fullPath := filepath.Join(viewer.Path, name)
			group := ""
			if len(name) > 0 {
				runes := []rune(strings.ToUpper(name))
				group = string(runes[0])
			}
			
			listItem := &types.ListItem{
				Name:  name,
				Path:  fullPath,
				Group: group,
				IsDir: entry.IsDir(),
			}
			info, err := entry.Info()
			if err == nil {
				listItem.Size = info.Size()
			}
			items = append(items, listItem)
		}
	}

	viewer.FileViewerList.SetItems(items)
	viewer.folderIcon.SetFromIconName(fileops.GetIconForFolderSymbolic(viewer.Path))
}
