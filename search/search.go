package search

import (
	"bufio"
	"os"
	"os/exec"
	"strings"

	"github.com/MrSametBurgazoglu/atilgan/file_list"
	"github.com/MrSametBurgazoglu/atilgan/preferences"
	"github.com/MrSametBurgazoglu/atilgan/special_path"
	"github.com/MrSametBurgazoglu/atilgan/types"
	"github.com/diamondburned/gotk4/pkg/glib/v2"
	"github.com/diamondburned/gotk4/pkg/gtk/v4"
)

type Search struct {
	*gtk.Box
	filenameEntry *gtk.Entry
	contentEntry  *gtk.Entry
	dateDropdown  *gtk.DropDown
	sizeDropdown  *gtk.DropDown
	ContentPanel  *gtk.Box
	SearchBar     *gtk.Box
	fileList      *file_list.FileList
	path          string
	PathChanged   func(path string)
}

func NewSearch(path string, specialPathManager *special_path.SpecialPathManager, parent *gtk.Window, config *preferences.Config) *Search {
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
	search.ContentPanel.SetMarginStart(20)
	search.ContentPanel.SetMarginEnd(20)
	search.ContentPanel.SetMarginTop(6)

	search.contentEntry = gtk.NewEntry()
	search.contentEntry.SetHExpand(true)
	search.contentEntry.SetPlaceholderText("Search inside files...")
	search.contentEntry.AddCSSClass("search-entry")
	search.ContentPanel.Append(search.contentEntry)

	search.dateDropdown = gtk.NewDropDownFromStrings([]string{"Any Time", "Last 24h", "Last Week", "Last Month", "Last Year"})
	search.ContentPanel.Append(search.dateDropdown)

	search.sizeDropdown = gtk.NewDropDownFromStrings([]string{"Any Size", "< 1 MB", "1-100 MB", "100MB-1GB", "> 1 GB"})
	search.ContentPanel.Append(search.sizeDropdown)

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

	search.fileList = file_list.NewFileList(true, specialPathManager, parent, config)
	search.fileList.SetVExpand(true)

	box.Append(search.ContentPanel)
	box.Append(search.fileList)

	return search
}

func (s *Search) performSearch(searchContent bool) {
	filename := s.filenameEntry.Text()
	content := s.contentEntry.Text()

	args := []string{s.path}
	if filename != "" {
		args = append(args, "-name", "*"+filename+"*")
	}

	if searchContent {
		dateIdx := s.dateDropdown.Selected()
		switch dateIdx {
		case 1: // Last 24h
			args = append(args, "-mtime", "-1")
		case 2: // Last Week
			args = append(args, "-mtime", "-7")
		case 3: // Last Month
			args = append(args, "-mtime", "-30")
		case 4: // Last Year
			args = append(args, "-mtime", "-365")
		}

		sizeIdx := s.sizeDropdown.Selected()
		switch sizeIdx {
		case 1: // < 1 MB
			args = append(args, "-size", "-1M")
		case 2: // 1-100 MB
			args = append(args, "-size", "+1M", "-size", "-100M")
		case 3: // 100MB-1GB
			args = append(args, "-size", "+100M", "-size", "-1G")
		case 4: // > 1 GB
			args = append(args, "-size", "+1G")
		}
	}

	var cmd *exec.Cmd

	if !searchContent || content == "" {
		cmd = exec.Command("find", args...)
	} else {
		// Use find to get files and then grep
		findArgs := append(args, "-type", "f", "-print0")
		grepCmd := "xargs -0 grep -l " + content
		// We use sh -c to handle the pipe and potential escaping issues
		cmd = exec.Command("sh", "-c", "find "+"\""+s.path+"\" "+" "+strings.Join(findArgs[1:], " ")+" | "+grepCmd)
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
