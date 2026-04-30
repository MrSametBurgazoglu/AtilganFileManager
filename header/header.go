package header

import (
	"github.com/diamondburned/gotk4/pkg/gtk/v4"
)

type HeaderBar struct {
	*gtk.HeaderBar
	NewButton            *gtk.MenuButton
	TerminalButton       *gtk.Button
	ShortcutsButton      *gtk.Button
	SearchButton         *gtk.Button
	PreviewerPanelButton *gtk.Button
	CircularProgressBar  *CircularProgressBar
}

func NewHeaderBar(mainWindow *gtk.ApplicationWindow) *HeaderBar {
	headerBar := gtk.NewHeaderBar()
	headerBar.AddCSSClass("headerbar")

	atilganIcon := gtk.NewImageFromIconName("atilgan_icon")
	atilganIcon.SetPixelSize(32)

	searchButton := gtk.NewButtonFromIconName("system-search-symbolic")
	headerBar.PackStart(searchButton)

	circularProgressBar := NewCircularProgressBar()
	circularProgressBar.SetVisible(false)
	headerBar.PackStart(circularProgressBar)

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
	headerBar.PackEnd(aboutButton)

	shortcutsButton := gtk.NewButtonFromIconName("preferences-desktop-keyboard-shortcuts-symbolic")
	headerBar.PackEnd(shortcutsButton)

	previewerPanelButton := gtk.NewButtonFromIconName("view-reveal-symbolic")
	headerBar.PackEnd(previewerPanelButton)

	terminalButton := gtk.NewButtonFromIconName("utilities-terminal-symbolic")
	headerBar.PackEnd(terminalButton)

	newButton := gtk.NewMenuButton()
	newButton.SetIconName("list-add-symbolic")
	headerBar.PackEnd(newButton)

	return &HeaderBar{
		HeaderBar:            headerBar,
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
