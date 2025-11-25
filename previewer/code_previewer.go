package previewer

import (
	"fmt"
	"io/ioutil"
	"os"

	"github.com/diamondburned/gotk4-sourceview/pkg/gtksource/v5"
	"github.com/diamondburned/gotk4/pkg/gdk/v4"
	"github.com/diamondburned/gotk4/pkg/glib/v2"
	"github.com/diamondburned/gotk4/pkg/gtk/v4"
)

type CodePreviewer struct {
	*gtk.Box
	sourceView          *gtksource.View
	searchBar           *gtk.Box
	searchEntry         *gtk.SearchEntry
	nextButton          *gtk.Button
	prevButton          *gtk.Button
	matchCountLabel     *gtk.Label
	searchButton        *gtk.Button
	fileNameLabel       *gtk.Label
	fileSizeLabel       *gtk.Label
	searchResults       []gtk.TextIter
	currentSearchResult int
	searchQuery         string
}

func NewCodePreviewer() *CodePreviewer {
	box := gtk.NewBox(gtk.OrientationVertical, 0)
	scrolledWindow := gtk.NewScrolledWindow()
	sourceView := gtksource.NewView()
	sourceView.SetEditable(false)
	sourceView.SetShowLineNumbers(true)
	scrolledWindow.SetChild(sourceView)
	scrolledWindow.SetVExpand(true)

	searchBar := gtk.NewBox(gtk.OrientationHorizontal, 6)
	searchEntry := gtk.NewSearchEntry()
	prevButton := gtk.NewButtonWithLabel("Previous")
	nextButton := gtk.NewButtonWithLabel("Next")
	matchCountLabel := gtk.NewLabel("")
	searchBar.Append(searchEntry)
	searchBar.Append(prevButton)
	searchBar.Append(nextButton)
	searchBar.Append(matchCountLabel)
	searchBar.SetVisible(false)

	box.Append(searchBar)
	box.Append(scrolledWindow)

	infoBar := gtk.NewBox(gtk.OrientationHorizontal, 6)
	searchButton := gtk.NewButtonFromIconName("system-search-symbolic")
	fileNameLabel := gtk.NewLabel("")
	fileSizeLabel := gtk.NewLabel("")
	infoBar.Append(searchButton)
	infoBar.Append(fileNameLabel)
	infoBar.Append(fileSizeLabel)
	box.Append(infoBar)

	cp := &CodePreviewer{
		Box:             box,
		sourceView:      sourceView,
		searchBar:       searchBar,
		searchEntry:     searchEntry,
		nextButton:      nextButton,
		prevButton:      prevButton,
		matchCountLabel: matchCountLabel,
		searchButton:    searchButton,
		fileNameLabel:   fileNameLabel,
		fileSizeLabel:   fileSizeLabel,
	}

	cp.searchButton.ConnectClicked(func() {
		cp.searchBar.SetVisible(!cp.searchBar.Visible())
		if cp.searchBar.Visible() {
			cp.searchEntry.GrabFocus()
		}
	})

	searchEntry.Connect("search-changed", func() {
		cp.searchQuery = searchEntry.Text()
		buffer := cp.sourceView.Buffer()
		buffer.RemoveTagByName("search", buffer.StartIter(), buffer.EndIter())
		cp.searchResults = nil
		cp.currentSearchResult = 0

		if cp.searchQuery == "" {
			cp.matchCountLabel.SetText("")
			return
		}

		iter := buffer.StartIter()
		for {
			matchStart, matchEnd, found := iter.ForwardSearch(cp.searchQuery, gtk.TextSearchTextOnly, nil)
			if !found {
				break
			}
			buffer.ApplyTagByName("search", matchStart, matchEnd)
			cp.searchResults = append(cp.searchResults, *matchStart)
			iter = matchEnd
		}
		cp.matchCountLabel.SetText(fmt.Sprintf("%d matches", len(cp.searchResults)))
		if len(cp.searchResults) > 0 {
			cp.scrollToCurrentSearchResult()
		}
	})

	cp.nextButton.ConnectClicked(func() {
		if len(cp.searchResults) > 0 {
			cp.currentSearchResult = (cp.currentSearchResult + 1) % len(cp.searchResults)
			cp.scrollToCurrentSearchResult()
		}
	})

	cp.prevButton.ConnectClicked(func() {
		if len(cp.searchResults) > 0 {
			cp.currentSearchResult--
			if cp.currentSearchResult < 0 {
				cp.currentSearchResult = len(cp.searchResults) - 1
			}
			cp.scrollToCurrentSearchResult()
		}
	})

	controller := gtk.NewShortcutController()
	trigger := gtk.NewKeyvalTrigger(gdk.KEY_f, gdk.ControlMask)
	shortcut := gtk.NewShortcut(trigger, gtk.NewCallbackAction(func(widget gtk.Widgetter, args *glib.Variant) bool {
		cp.searchBar.SetVisible(!cp.searchBar.Visible())
		if cp.searchBar.Visible() {
			cp.searchEntry.GrabFocus()
		}
		return true
	}))
	controller.AddShortcut(shortcut)
	cp.sourceView.AddController(controller)

	return cp
}

func (cp *CodePreviewer) SetText(filePath string, fileInfo os.FileInfo) {
	content, err := ioutil.ReadFile(filePath)
	if err != nil {
		return
	}
	cp.fileNameLabel.SetText(fileInfo.Name())
	cp.fileSizeLabel.SetText(fmt.Sprintf("%d bytes", fileInfo.Size()))

	langManager := gtksource.LanguageManagerGetDefault()
	lang := langManager.GuessLanguage(filePath, "")
	if lang != nil {
		buffer := gtksource.NewBufferWithLanguage(lang)
		cp.sourceView.SetBuffer(&buffer.TextBuffer)
	}
	cp.sourceView.Buffer().SetText(string(content))
	tagTable := cp.sourceView.Buffer().TagTable()
	tag := gtk.NewTextTag("search")
	tag.SetObjectProperty("background", glib.NewValue("yellow"))
	tagTable.Add(tag)

}

func (cp *CodePreviewer) scrollToCurrentSearchResult() {
	if len(cp.searchResults) > 0 {
		iter := cp.searchResults[cp.currentSearchResult]
		cp.sourceView.ScrollToIter(&iter, 0, true, 0.5, 0.5)
	}
}
