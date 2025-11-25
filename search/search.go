package search

import (
	"bufio"
	"os"
	"os/exec"
	"path/filepath"

	"github.com/MrSametBurgazoglu/atilgan/file_list"
	"github.com/MrSametBurgazoglu/atilgan/types"
	"github.com/diamondburned/gotk4/pkg/glib/v2"
	"github.com/diamondburned/gotk4/pkg/gtk/v4"
)

type Search struct {
	*gtk.Box
	filenameEntry *gtk.Entry
	contentEntry  *gtk.Entry
	searchButton  *gtk.Button
	fileList      *file_list.FileList
	path          string
	PathChanged   func(path string)
}

func NewSearch(path string) *Search {
	box := gtk.NewBox(gtk.OrientationVertical, 6)
	search := &Search{
		Box:  box,
		path: path,
	}

	hBox := gtk.NewBox(gtk.OrientationHorizontal, 6)
	hBox.SetVExpand(false)

	filenameBox := gtk.NewBox(gtk.OrientationHorizontal, 6)
	filenameLabel := gtk.NewLabel("Filename:")
	search.filenameEntry = gtk.NewEntry()
	search.filenameEntry.AddCSSClass("search-entry")
	filenameBox.Append(filenameLabel)
	filenameBox.Append(search.filenameEntry)
	hBox.Append(filenameBox)

	contentBox := gtk.NewBox(gtk.OrientationHorizontal, 6)
	contentLabel := gtk.NewLabel("Content:")
	search.contentEntry = gtk.NewEntry()
	search.contentEntry.AddCSSClass("search-entry")
	contentBox.Append(contentLabel)
	contentBox.Append(search.contentEntry)
	hBox.Append(contentBox)

	search.searchButton = gtk.NewButtonWithLabel("Search")
	search.searchButton.ConnectClicked(func() {
		search.fileList.SetItems(make([]*types.ListItem, 0))
		go search.performSearch()
	})
	hBox.Append(search.searchButton)
	box.Append(hBox)

	search.fileList = file_list.NewFileList(true, nil, nil)
	search.fileList.PathChanged = func(path string) {
		if search.PathChanged != nil {
			search.PathChanged(filepath.Dir(path))
		}
	}
	search.fileList.KeyRightPressed = func() {
		if search.PathChanged != nil {
			selectedItem := search.fileList.Items[search.fileList.SelectedIDX]
			if !selectedItem.IsDir {
				cmd := exec.Command("xdg-open", selectedItem.Path)
				cmd.Start()
			} else {
				search.PathChanged(filepath.Dir(selectedItem.Path))
			}
		}
	}
	search.fileList.SetMinContentHeight(200)
	search.fileList.SelectionChanged = func(index int) {
	}

	box.Append(search.fileList)

	return search
}

func (s *Search) performSearch() {
	filename := s.filenameEntry.Text()
	content := s.contentEntry.Text()

	var cmd *exec.Cmd

	if content == "" {
		cmd = exec.Command("find", s.path, "-name", "*"+filename+"*")
	} else {
		if filename == "" {
			cmd = exec.Command("grep", "-rl", content, s.path)
		} else {
			cmd = exec.Command("sh", "-c", "find "+s.path+" -name *"+filename+"* -print0 | xargs -0 grep -l "+content)
		}
	}

	stdout, err := cmd.StdoutPipe()
	if err != nil {
		println("Error creating stdout pipe:", err.Error())
		return
	}

	if err := cmd.Start(); err != nil {
		println("Error starting command:", err.Error())
		return
	}

	scanner := bufio.NewScanner(stdout)
	for scanner.Scan() {
		line := scanner.Text()
		glib.IdleAdd(func() {
			s.addItemToList(line)
		})
	}

	if err := cmd.Wait(); err != nil {
		if exitErr, ok := err.(*exec.ExitError); ok {
			if exitErr.ExitCode() == 1 {
				return
			}
		}
		println("Error waiting for command:", err.Error())
	}
}

func (s *Search) addItemToList(line string) {
	fileInfo, err := os.Stat(line)
	if err != nil {
		println("Error getting file info:", err.Error())
		return
	}
	item := &types.ListItem{
		Name:  line,
		Path:  line,
		IsDir: fileInfo.IsDir(),
	}
	s.fileList.AddItem(item)
}

func (s *Search) SetPath(path string) {
	s.path = path
}
