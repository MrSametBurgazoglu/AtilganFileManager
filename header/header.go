package header

import (
	"github.com/diamondburned/gotk4-adwaita/pkg/adw"
	"github.com/diamondburned/gotk4/pkg/gtk/v4"
)

type HeaderBar struct {
	LeftHeader           *adw.HeaderBar
	RightHeader          *adw.HeaderBar
	ActionsButton        *gtk.MenuButton
	ShortcutsButton      *gtk.Button
	SearchButton         *gtk.Button
	CircularProgressBar  *CircularProgressBar
}

func NewHeaderBar(mainWindow *adw.ApplicationWindow) *HeaderBar {
	leftHeader := adw.NewHeaderBar()
	leftHeader.AddCSSClass("left-header")
	leftHeader.SetShowStartTitleButtons(true)
	leftHeader.SetShowEndTitleButtons(false)

	rightHeader := adw.NewHeaderBar()
	rightHeader.AddCSSClass("right-header")
	rightHeader.SetShowStartTitleButtons(false)
	rightHeader.SetShowEndTitleButtons(true)

	// Sync heights of the two headers
	sizeGroup := gtk.NewSizeGroup(gtk.SizeGroupVertical)
	sizeGroup.AddWidget(leftHeader)
	sizeGroup.AddWidget(rightHeader)

	// Set an empty widget to prevent the default window title from showing in the center
	rightHeader.SetTitleWidget(gtk.NewBox(gtk.OrientationHorizontal, 0))

	searchButton := gtk.NewButtonFromIconName("system-search-symbolic")
	leftHeader.PackStart(searchButton)

	actionsButton := gtk.NewMenuButton()
	actionsButton.SetIconName("open-menu-symbolic")
	leftHeader.PackEnd(actionsButton)

	circularProgressBar := NewCircularProgressBar()
	circularProgressBar.SetVisible(false)
	rightHeader.PackStart(circularProgressBar)

	aboutButton := gtk.NewButtonFromIconName("help-about-symbolic")
	aboutButton.ConnectClicked(func() {
		aboutDialog := adw.NewAboutWindow()
		aboutDialog.SetApplicationName("Atilgan")
		aboutDialog.SetVersion("0.1.0")
		aboutDialog.SetApplicationIcon("atilgan_icon")
		aboutDialog.SetCopyright("Copyright © 2025 MrSametBurgazoglu")
		aboutDialog.SetWebsite("https://github.com/MrSametBurgazoglu/AtilganFileManager")
		aboutDialog.SetVisible(true)
	})
	rightHeader.PackEnd(aboutButton)

	shortcutsButton := gtk.NewButtonFromIconName("preferences-desktop-keyboard-shortcuts-symbolic")
	rightHeader.PackEnd(shortcutsButton)

	return &HeaderBar{
		LeftHeader:           leftHeader,
		RightHeader:          rightHeader,
		ActionsButton:        actionsButton,
		ShortcutsButton:      shortcutsButton,
		SearchButton:         searchButton,
		CircularProgressBar:  circularProgressBar,
	}
}

func (h *HeaderBar) ShowProgress() {
	h.CircularProgressBar.SetVisible(true)
}

func (h *HeaderBar) HideProgress() {
	h.CircularProgressBar.SetVisible(false)
}

func (h *HeaderBar) SetProgress(fraction float64) {
	h.CircularProgressBar.SetFraction(fraction)
}

func (h *HeaderBar) SetTitleWidget(widget gtk.Widgetter) {
	h.RightHeader.SetTitleWidget(widget)
}

func (h *HeaderBar) Remove(widget gtk.Widgetter) {
	h.LeftHeader.Remove(widget)
}

func (h *HeaderBar) PackStart(widget gtk.Widgetter) {
	h.RightHeader.PackStart(widget)
}

func (h *HeaderBar) PackStartLeft(widget gtk.Widgetter) {
	h.LeftHeader.PackStart(widget)
}

func (h *HeaderBar) PackStartRight(widget gtk.Widgetter) {
	h.RightHeader.PackStart(widget)
}

func (h *HeaderBar) PackEndLeft(widget gtk.Widgetter) {
	h.LeftHeader.PackEnd(widget)
}
