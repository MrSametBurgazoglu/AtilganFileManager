package search

import (
	"bufio"
	"os"
	"os/exec"

	"github.com/MrSametBurgazoglu/atilgan/file_list"
	"github.com/MrSametBurgazoglu/atilgan/types"
	"github.com/diamondburned/gotk4/pkg/glib/v2"
	"github.com/diamondburned/gotk4/pkg/gtk/v4"
)

type Search struct {
	*gtk.Box
	filenameEntry *gtk.Entry
	contentEntry  *gtk.Entry
	ContentPanel  *gtk.Box
	SearchBar     *gtk.Box
	fileList      *file_list.FileList
	path          string
	PathChanged   func(path string)
}

func NewSearch(path string) *Search {
	box := gtk.NewBox(gtk.OrientationVertical, 0)
	search := &Search{
		Box:  box,
		path: path,
	}

	search.SearchBar = gtk.NewBox(gtk.OrientationHorizontal, 6)
	search.SearchBar.SetHExpand(true)
	search.SearchBar.SetMarginStart(20)
	search.SearchBar.SetMarginEnd(20)

	search.filenameEntry = gtk.NewEntry()
	search.filenameEntry.SetHExpand(true)
	search.filenameEntry.SetPlaceholderText("Filter by name...")
	search.filenameEntry.AddCSSClass("search-entry")
	search.SearchBar.Append(search.filenameEntry)

	search.ContentPanel = gtk.NewBox(gtk.OrientationHorizontal, 6)
	search.ContentPanel.SetMarginStart(6)
	search.ContentPanel.SetMarginEnd(6)
	search.ContentPanel.SetMarginTop(6)

	search.contentEntry = gtk.NewEntry()
	search.contentEntry.SetHExpand(true)
	search.contentEntry.SetPlaceholderText("Search inside files...")
	search.contentEntry.AddCSSClass("search-entry")
	search.ContentPanel.Append(search.contentEntry)
	search.ContentPanel.SetVisible(false)

	toggleContentBtn := gtk.NewButtonFromIconName("preferences-system-symbolic")
	toggleContentBtn.ConnectClicked(func() {
		visible := !search.ContentPanel.Visible()
		search.ContentPanel.SetVisible(visible)
		if visible {
			toggleContentBtn.AddCSSClass("suggested-action")
		} else {
			toggleContentBtn.RemoveCSSClass("suggested-action")
		}
	})
	search.SearchBar.Append(toggleContentBtn)

	doSearch := func() {
		filename := search.filenameEntry.Text()
		content := search.contentEntry.Text()
		searchContent := search.ContentPanel.Visible()

		if filename == "" && (!searchContent || content == "") {
			return
		}

		search.fileList.SetItems(make([]*types.ListItem, 0))
		go search.performSearch(searchContent)
	}

	search.filenameEntry.ConnectActivate(func() { doSearch() })
	search.contentEntry.ConnectActivate(func() { doSearch() })

	search.fileList = file_list.NewFileList(true, nil, nil)
	search.fileList.SetVExpand(true)

	box.Append(search.ContentPanel)
	box.Append(search.fileList)

	return search
}

func (s *Search) performSearch(searchContent bool) {
	filename := s.filenameEntry.Text()
	content := s.contentEntry.Text()

	var cmd *exec.Cmd

	if !searchContent || content == "" {
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
