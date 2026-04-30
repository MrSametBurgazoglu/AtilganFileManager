package header

import (
	"github.com/diamondburned/gotk4/pkg/gtk/v4"
)

type HeaderBar struct {
	*gtk.Box
	LeftHeader           *gtk.HeaderBar
	RightHeader          *gtk.HeaderBar
	NewButton            *gtk.MenuButton
	TerminalButton       *gtk.Button
	ShortcutsButton      *gtk.Button
	SearchButton         *gtk.Button
	PreviewerPanelButton *gtk.Button
	CircularProgressBar  *CircularProgressBar
}

func NewHeaderBar(mainWindow *gtk.ApplicationWindow) *HeaderBar {
	container := gtk.NewBox(gtk.OrientationHorizontal, 0)
	container.AddCSSClass("header-container")

	leftHeader := gtk.NewHeaderBar()
	leftHeader.AddCSSClass("headerbar")
	leftHeader.AddCSSClass("left-header")
	leftHeader.SetShowTitleButtons(false)

	rightHeader := gtk.NewHeaderBar()
	rightHeader.AddCSSClass("headerbar")
	rightHeader.AddCSSClass("right-header")
	rightHeader.SetShowTitleButtons(true)

	// Set an empty widget to prevent the default window title from showing in the center
	rightHeader.SetTitleWidget(gtk.NewBox(gtk.OrientationHorizontal, 0))

	searchButton := gtk.NewButtonFromIconName("system-search-symbolic")
	leftHeader.PackStart(searchButton)

	atilganIcon := gtk.NewImageFromIconName("atilgan_icon")
	atilganIcon.SetPixelSize(24)
	leftHeader.PackStart(atilganIcon)

	circularProgressBar := NewCircularProgressBar()
	circularProgressBar.SetVisible(false)
	rightHeader.PackStart(circularProgressBar)

	aboutButton := gtk.NewButtonFromIconName("help-about-symbolic")
	aboutButton.ConnectClicked(func() {
		aboutDialog := gtk.NewAboutDialog()
		aboutDialog.SetProgramName("Atilgan")
		aboutDialog.SetVersion("0.1.0")
		aboutDialog.SetLogoIconName("atilgan_icon")
		aboutDialog.SetCopyright("Copyright © 2025 MrSametBurgazoglu")
		aboutDialog.SetWebsite("https://github.com/MrSametBurgazoglu/AtilganFileManager")
		aboutDialog.SetVisible(true)
	})
	rightHeader.PackEnd(aboutButton)

	shortcutsButton := gtk.NewButtonFromIconName("preferences-desktop-keyboard-shortcuts-symbolic")
	rightHeader.PackEnd(shortcutsButton)

	previewerPanelButton := gtk.NewButtonFromIconName("view-reveal-symbolic")
	rightHeader.PackEnd(previewerPanelButton)

	terminalButton := gtk.NewButtonFromIconName("utilities-terminal-symbolic")
	rightHeader.PackEnd(terminalButton)

	newButton := gtk.NewMenuButton()
	newButton.SetIconName("list-add-symbolic")
	rightHeader.PackEnd(newButton)

	container.Append(leftHeader)
	
	separator := gtk.NewSeparator(gtk.OrientationVertical)
	separator.AddCSSClass("header-separator")
	container.Append(separator)
	
	container.Append(rightHeader)
	rightHeader.SetHExpand(true)

	return &HeaderBar{
		Box:                  container,
		LeftHeader:           leftHeader,
		RightHeader:          rightHeader,
		NewButton:            newButton,
		TerminalButton:       terminalButton,
		ShortcutsButton:      shortcutsButton,
		SearchButton:         searchButton,
		CircularProgressBar:  circularProgressBar,
		PreviewerPanelButton: previewerPanelButton,
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

func (h *HeaderBar) PackStart(widget gtk.Widgetter) {
	h.RightHeader.PackStart(widget)
}
