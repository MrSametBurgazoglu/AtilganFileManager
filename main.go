package main

import (
	"embed"
	"os"
	"path/filepath"

	"github.com/MrSametBurgazoglu/atilgan/header"
	"github.com/diamondburned/gotk4-adwaita/pkg/adw"
	"github.com/diamondburned/gotk4/pkg/gdk/v4"
	"github.com/diamondburned/gotk4/pkg/gio/v2"
	"github.com/diamondburned/gotk4/pkg/gtk/v4"
)

//go:embed style.css
var styleCSS embed.FS

//go:embed atilgan_icon.svg
var atilganIcon []byte

func main() {
	app := adw.NewApplication("io.github.mrsametburgazoglu.AtilganFileManager", gio.ApplicationFlagsNone)
	app.ConnectActivate(func() {
		activate(app)
	})
	if code := app.Run(os.Args); code > 0 {
		os.Exit(code)
	}
}

func activate(app *adw.Application) {
	window := adw.NewApplicationWindow(&app.Application)
	window.SetTitle("Atilgan")
	window.SetDefaultSize(1200, 800)

	display := gdk.DisplayGetDefault()

	cssProvider := gtk.NewCSSProvider()
	cssBytes, err := styleCSS.ReadFile("style.css")
	if err == nil {
		cssProvider.LoadFromData(string(cssBytes))
	}
	gtk.StyleContextAddProviderForDisplay(display, cssProvider, gtk.STYLE_PROVIDER_PRIORITY_APPLICATION)

	iconTheme := gtk.IconThemeGetForDisplay(display)

	exePath, err := os.Executable()
	if err == nil {
		exeDir := filepath.Dir(exePath)
		iconTheme.AddSearchPath(exeDir)
		iconTheme.AddSearchPath(filepath.Join(filepath.Dir(exeDir), "share", "icons"))
	}
	curDir, err := os.Getwd()
	if err == nil {
		iconTheme.AddSearchPath(curDir)
	}
	window.SetIconName("io.github.mrsametburgazoglu.AtilganFileManager")

	headerBar := header.NewHeaderBar(window, atilganIcon)
	
	initialPath := ""
	if len(os.Args) > 1 {
		initialPath = os.Args[1]
	}
	
	mainBox := NewMainBox(window, headerBar, initialPath)
	window.SetContent(mainBox)

	window.SetVisible(true)
	activePanel := mainBox.getActivePanel()
	if activePanel != nil {
		activePanel.FileViewer.FileViewerList.FocusDrawingArea()
	}
}
