package viewer

import (
	"fmt"
	"os"
	"path"

	"github.com/MrSametBurgazoglu/atilgan/file_list"
	"github.com/MrSametBurgazoglu/atilgan/fileops"
	"github.com/MrSametBurgazoglu/atilgan/special_path"
	"github.com/MrSametBurgazoglu/atilgan/types"
	"github.com/diamondburned/gotk4/pkg/gtk/v4"
)

type MusicViewer struct {
	*gtk.Box
	Path               string
	FileViewerList     *file_list.FileList // Using FileList for music
	specialPathManager *special_path.SpecialPathManager
	stack              *gtk.Stack
}

func NewMusicViewer(mainWindow *gtk.Window, path string, pathChanged func(string), specialPathManager *special_path.SpecialPathManager) *MusicViewer {
	viewer := &MusicViewer{
		Box:                gtk.NewBox(gtk.OrientationVertical, 6),
		Path:               path,
		FileViewerList:     file_list.NewFileList(true, specialPathManager, mainWindow),
		specialPathManager: specialPathManager,
	}
	viewer.SetVExpand(true)
	viewer.SetHExpand(true)

	viewer.stack = gtk.NewStack()
	viewer.stack.AddTitled(viewer.FileViewerList, "list", "List")

	emptyLabel := gtk.NewLabel("No Music Found")
	emptyLabel.AddCSSClass("preview-title")
	viewer.stack.AddTitled(emptyLabel, "empty", "Empty")

	viewer.Box.Append(viewer.stack)

	viewer.Refresh()

	return viewer
}

func (viewer *MusicViewer) SetPath(path string) {
	viewer.Path = path
	viewer.Refresh()
}

func (viewer *MusicViewer) Refresh() {
	if viewer.Path == "" {
		return
	}

	entries, err := os.ReadDir(viewer.Path)
	if err != nil {
		fmt.Println("Error reading directory:", err)
		return
	}

	newFiles := make([]*types.ListItem, 0)
	for _, entry := range entries {
		if !entry.IsDir() && !fileops.IsMusic(entry.Name()) {
			continue
		}
		
		fullPath := path.Join(viewer.Path, entry.Name())
		listItem := &types.ListItem{
			Name:  entry.Name(),
			Path:  fullPath,
			Group: "Music",
			IsDir: entry.IsDir(),
		}
		if listItem.IsDir {
			listItem.Group = "Directories"
			listItem.ItemCount = getDirItemCount(fullPath)
		} else {
			info, err := entry.Info()
			if err == nil {
				listItem.Size = info.Size()
			}
		}
		newFiles = append(newFiles, listItem)
	}
	viewer.FileViewerList.SetItems(newFiles)

	if len(newFiles) == 0 {
		viewer.stack.SetVisibleChildName("empty")
	} else {
		viewer.stack.SetVisibleChildName("list")
	}
}
