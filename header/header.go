package header

import (
	"github.com/diamondburned/gotk4-adwaita/pkg/adw"
	"github.com/diamondburned/gotk4/pkg/gtk/v4"
)

type HeaderBar struct {
	LeftHeader          *adw.HeaderBar
	RightHeader         *adw.HeaderBar
	ActionsButton       *gtk.MenuButton
	SearchButton        *gtk.Button
	CircularProgressBar *CircularProgressBar
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

	logo := gtk.NewImageFromFile("atilgan_icon.svg")
	logo.SetPixelSize(24)
	logo.SetMarginStart(6)
	leftHeader.PackStart(logo)

	actionsButton := gtk.NewMenuButton()
	actionsButton.SetIconName("open-menu-symbolic")
	leftHeader.PackEnd(actionsButton)

	searchButton := gtk.NewButtonFromIconName("system-search-symbolic")
	rightHeader.PackStart(searchButton)

	circularProgressBar := NewCircularProgressBar()
	circularProgressBar.SetVisible(false)
	rightHeader.PackStart(circularProgressBar)

	return &HeaderBar{
		LeftHeader:          leftHeader,
		RightHeader:         rightHeader,
		ActionsButton:       actionsButton,
		SearchButton:        searchButton,
		CircularProgressBar: circularProgressBar,
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
