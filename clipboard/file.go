package clipboard

import (
	"github.com/diamondburned/gotk4/pkg/gdk/v4"
	"github.com/diamondburned/gotk4/pkg/gio/v2"
	"github.com/diamondburned/gotk4/pkg/glib/v2"
)

func CopyFileToClipboard(file *gio.File) {
	display := gdk.DisplayGetDefault()
	clipboard := display.Clipboard()

	path := file.Path()
	valText := glib.NewValue(path)
	providerText := gdk.NewContentProviderForValue(valText)

	valFile := glib.NewValue(file)
	providerFile := gdk.NewContentProviderForValue(valFile)

	providers := []*gdk.ContentProvider{providerFile, providerText}
	unionProvider := gdk.NewContentProviderUnion(providers)

	clipboard.SetContent(unionProvider)
}
