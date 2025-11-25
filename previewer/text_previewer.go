package previewer

import (
	"fmt"
	"io/ioutil"
	"os"
	"path/filepath"

	"github.com/diamondburned/gotk4/pkg/gdk/v4"
	"github.com/diamondburned/gotk4/pkg/glib/v2"
	"github.com/diamondburned/gotk4/pkg/gtk/v4"
)

type TextPreviewer struct {
	*gtk.Box
	TextView            *gtk.TextView
	SearchEntry         *gtk.SearchEntry
	SearchQuery         string
	searchEntryBox      *gtk.Box
	nextButton          *gtk.Button
	prevButton          *gtk.Button
	matchCountLabel     *gtk.Label
	searchResults       []gtk.TextIter
	currentSearchResult int
	searchButton        *gtk.Button
	fileNameLabel       *gtk.Label
	fileTypeLabel       *gtk.Label
	fileSizeLabel       *gtk.Label
}

func NewTextPreviewer() *TextPreviewer {
	box := gtk.NewBox(gtk.OrientationVertical, 0)
	box.SetVExpand(true)
	scrolledWindow := gtk.NewScrolledWindow()
	textView := gtk.NewTextView()
	textView.SetEditable(false)
	textView.SetCursorVisible(true)
	scrolledWindow.SetChild(textView)
	scrolledWindow.SetVExpand(true)

	searchBar := gtk.NewBox(gtk.OrientationHorizontal, 6)
	searchButton := gtk.NewButtonFromIconName("system-search-symbolic")
	searchBar.Append(searchButton)
	searchEntryBox := gtk.NewBox(gtk.OrientationHorizontal, 6)
	searchEntry := gtk.NewSearchEntry()
	prevButton := gtk.NewButtonWithLabel("Previous")
	nextButton := gtk.NewButtonWithLabel("Next")
	matchCountLabel := gtk.NewLabel("")
	searchEntryBox.Append(searchEntry)
	searchEntryBox.Append(prevButton)
	searchEntryBox.Append(nextButton)
	searchEntryBox.Append(matchCountLabel)
	searchEntryBox.SetVisible(false)
	searchBar.Append(searchEntryBox)
	searchBar.SetVisible(true)
	box.Append(searchBar)
	box.Append(scrolledWindow)

	infoBar := gtk.NewBox(gtk.OrientationHorizontal, 6)
	infoBar.SetVAlign(gtk.AlignEnd)
	fileNameLabel := gtk.NewLabel("")
	fileTypeLabel := gtk.NewLabel("")
	fileSizeLabel := gtk.NewLabel("")
	infoBar.Append(fileNameLabel)
	infoBar.Append(fileTypeLabel)
	infoBar.Append(fileSizeLabel)
	box.Append(infoBar)
	tp := &TextPreviewer{
		Box:             box,
		TextView:        textView,
		SearchEntry:     searchEntry,
		searchEntryBox:  searchEntryBox,
		nextButton:      nextButton,
		prevButton:      prevButton,
		matchCountLabel: matchCountLabel,
		searchButton:    searchButton,
		fileNameLabel:   fileNameLabel,
		fileTypeLabel:   fileTypeLabel,
		fileSizeLabel:   fileSizeLabel,
	}

	tp.searchButton.ConnectClicked(func() {
		tp.searchEntryBox.SetVisible(!tp.searchEntryBox.Visible())
		if tp.searchEntryBox.Visible() {
			tp.SearchEntry.GrabFocus()
		}
	})

	tagTable := tp.TextView.Buffer().TagTable()
	tag := gtk.NewTextTag("search")
	tag.SetObjectProperty("background", glib.NewValue("yellow"))
	tagTable.Add(tag)

	searchEntry.Connect("search-changed", func() {
		tp.SearchQuery = searchEntry.Text()
		buffer := tp.TextView.Buffer()
		buffer.RemoveTagByName("search", buffer.StartIter(), buffer.EndIter())
		tp.searchResults = nil
		tp.currentSearchResult = 0

		if tp.SearchQuery == "" {
			tp.matchCountLabel.SetText("")
			return
		}

		iter := buffer.StartIter()
		for {
			matchStart, matchEnd, found := iter.ForwardSearch(tp.SearchQuery, gtk.TextSearchTextOnly, nil)
			if !found {
				break
			}
			buffer.ApplyTagByName("search", matchStart, matchEnd)
			tp.searchResults = append(tp.searchResults, *matchStart)
			iter = matchEnd
		}
		tp.matchCountLabel.SetText(fmt.Sprintf("%d matches", len(tp.searchResults)))
		if len(tp.searchResults) > 0 {
			tp.scrollToCurrentSearchResult()
		}
	})

	tp.nextButton.ConnectClicked(func() {
		if len(tp.searchResults) > 0 {
			tp.currentSearchResult = (tp.currentSearchResult + 1) % len(tp.searchResults)
			tp.scrollToCurrentSearchResult()
		}
	})

	tp.prevButton.ConnectClicked(func() {
		if len(tp.searchResults) > 0 {
			tp.currentSearchResult--
			if tp.currentSearchResult < 0 {
				tp.currentSearchResult = len(tp.searchResults) - 1
			}
			tp.scrollToCurrentSearchResult()
		}
	})

	controller := gtk.NewShortcutController()
	trigger := gtk.NewKeyvalTrigger(gdk.KEY_f, gdk.ControlMask)
	shortcut := gtk.NewShortcut(trigger, gtk.NewCallbackAction(func(widget gtk.Widgetter, args *glib.Variant) bool {
		tp.searchEntryBox.SetVisible(!tp.searchEntryBox.Visible())
		if tp.searchEntryBox.Visible() {
			tp.SearchEntry.GrabFocus()
		}
		return true
	}))
	controller.AddShortcut(shortcut)
	tp.TextView.AddController(controller)

	return tp
}

func (tp *TextPreviewer) scrollToCurrentSearchResult() {
	if len(tp.searchResults) > 0 {
		iter := tp.searchResults[tp.currentSearchResult]
		tp.TextView.ScrollToIter(&iter, 0, true, 0.5, 0.5)
	}
}

func (tp *TextPreviewer) SetText(filePath string, fileInfo os.FileInfo) {
	content, err := ioutil.ReadFile(filePath)
	if err != nil {
		return
	}
	tp.TextView.Buffer().SetText(string(content))
	tp.fileNameLabel.SetText(fileInfo.Name())
	tp.fileTypeLabel.SetText(filepath.Ext(fileInfo.Name()))
	tp.fileSizeLabel.SetText(fmt.Sprintf("%d bytes", fileInfo.Size()))
}
