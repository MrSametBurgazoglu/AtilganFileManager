package preferences

import (
	"github.com/diamondburned/gotk4-adwaita/pkg/adw"
	"github.com/diamondburned/gotk4/pkg/gtk/v4"
)

type PreferencesDialog struct {
	*adw.PreferencesWindow
	Config    *Config
	OnChanged func()
}

func NewPreferencesDialog(parent *gtk.Window, config *Config) *PreferencesDialog {
	win := adw.NewPreferencesWindow()
	win.SetTransientFor(parent)
	win.SetTitle("Preferences")
	win.SetDefaultSize(500, 400)

	pd := &PreferencesDialog{
		PreferencesWindow: win,
		Config:            config,
	}

	// General Page
	generalPage := adw.NewPreferencesPage()
	generalPage.SetTitle("General")
	generalPage.SetIconName("preferences-system-symbolic")
	win.Add(generalPage)

	// View Group
	viewGroup := adw.NewPreferencesGroup()
	viewGroup.SetTitle("View")
	generalPage.Add(viewGroup)

	sortOrderRow := adw.NewComboRow()
	sortOrderRow.SetTitle("Default Sort Order")
	sortOrderRow.SetModel(gtk.NewStringList([]string{"Name", "Size", "Time", "Type"}))
	sortOrderIdx := uint(0)
	switch config.DefaultSortOrder {
	case "size":
		sortOrderIdx = 1
	case "time":
		sortOrderIdx = 2
	case "type":
		sortOrderIdx = 3
	}
	sortOrderRow.SetSelected(sortOrderIdx)
	sortOrderRow.Connect("notify::selected", func() {
		switch sortOrderRow.Selected() {
		case 0:
			config.DefaultSortOrder = "name"
		case 1:
			config.DefaultSortOrder = "size"
		case 2:
			config.DefaultSortOrder = "time"
		case 3:
			config.DefaultSortOrder = "type"
		}
		pd.notifyChanged()
	})
	viewGroup.Add(sortOrderRow)

	iconSizeRow := adw.NewActionRow()
	iconSizeRow.SetTitle("Icon Size")
	iconSizeScale := gtk.NewScaleWithRange(gtk.OrientationHorizontal, 32, 128, 8)
	iconSizeScale.SetDrawValue(true)
	iconSizeScale.SetValue(float64(config.IconSize))
	iconSizeScale.SetSizeRequest(200, -1)
	iconSizeScale.Connect("value-changed", func() {
		config.IconSize = int(iconSizeScale.Value())
		pd.notifyChanged()
	})
	iconSizeRow.AddSuffix(iconSizeScale)
	viewGroup.Add(iconSizeRow)

	// File Operations Group
	fileGroup := adw.NewPreferencesGroup()
	fileGroup.SetTitle("File Operations")
	generalPage.Add(fileGroup)

	showHiddenRow := adw.NewSwitchRow()
	showHiddenRow.SetTitle("Show Hidden Files")
	showHiddenRow.SetSubtitle("Display files starting with a dot")
	showHiddenRow.SetActive(config.ShowHidden)
	showHiddenRow.Connect("notify::active", func() {
		config.ShowHidden = showHiddenRow.Active()
		pd.notifyChanged()
	})
	fileGroup.Add(showHiddenRow)

	confirmDeleteRow := adw.NewSwitchRow()
	confirmDeleteRow.SetTitle("Confirm Deletion")
	confirmDeleteRow.SetActive(config.ConfirmDelete)
	confirmDeleteRow.Connect("notify::active", func() {
		config.ConfirmDelete = confirmDeleteRow.Active()
		pd.notifyChanged()
	})
	fileGroup.Add(confirmDeleteRow)

	thumbnailsRow := adw.NewSwitchRow()
	thumbnailsRow.SetTitle("Enable Thumbnails")
	thumbnailsRow.SetActive(config.EnableThumbnails)
	thumbnailsRow.Connect("notify::active", func() {
		config.EnableThumbnails = thumbnailsRow.Active()
		pd.notifyChanged()
	})
	fileGroup.Add(thumbnailsRow)

	// Appearance Page
	appearancePage := adw.NewPreferencesPage()
	appearancePage.SetTitle("Appearance")
	appearancePage.SetIconName("preferences-desktop-theme-symbolic")
	win.Add(appearancePage)

	// Theme Group
	themeGroup := adw.NewPreferencesGroup()
	themeGroup.SetTitle("Theme")
	appearancePage.Add(themeGroup)

	themeRow := adw.NewComboRow()
	themeRow.SetTitle("Application Theme")
	themeRow.SetModel(gtk.NewStringList([]string{"System", "Light", "Dark"}))

	initialSelected := uint(0)
	switch config.Theme {
	case "light":
		initialSelected = 1
	case "dark":
		initialSelected = 2
	}
	themeRow.SetSelected(initialSelected)

	themeRow.Connect("notify::selected", func() {
		switch themeRow.Selected() {
		case 0:
			config.Theme = "system"
		case 1:
			config.Theme = "light"
		case 2:
			config.Theme = "dark"
		}
		pd.notifyChanged()
	})
	themeGroup.Add(themeRow)

	// Layout Group
	layoutGroup := adw.NewPreferencesGroup()
	layoutGroup.SetTitle("Layout")
	appearancePage.Add(layoutGroup)

	previewPaneRow := adw.NewSwitchRow()
	previewPaneRow.SetTitle("Enable Preview Pane")
	previewPaneRow.SetActive(config.EnablePreviewPane)
	previewPaneRow.Connect("notify::active", func() {
		config.EnablePreviewPane = previewPaneRow.Active()
		pd.notifyChanged()
	})
	layoutGroup.Add(previewPaneRow)

	sidebarWidthRow := adw.NewActionRow()
	sidebarWidthRow.SetTitle("Sidebar Width")
	sidebarWidthScale := gtk.NewScaleWithRange(gtk.OrientationHorizontal, 100, 400, 10)
	sidebarWidthScale.SetDrawValue(true)
	sidebarWidthScale.SetValue(float64(config.SidebarWidth))
	sidebarWidthScale.SetSizeRequest(200, -1)
	sidebarWidthScale.Connect("value-changed", func() {
		config.SidebarWidth = int(sidebarWidthScale.Value())
		pd.notifyChanged()
	})
	sidebarWidthRow.AddSuffix(sidebarWidthScale)
	layoutGroup.Add(sidebarWidthRow)

	// System Page
	systemPage := adw.NewPreferencesPage()
	systemPage.SetTitle("System")
	systemPage.SetIconName("emblem-system-symbolic")
	win.Add(systemPage)

	// External Group
	externalGroup := adw.NewPreferencesGroup()
	externalGroup.SetTitle("External Applications")
	systemPage.Add(externalGroup)

	terminalRow := adw.NewEntryRow()
	terminalRow.SetTitle("Terminal Emulator")
	terminalRow.SetText(config.TerminalEmulator)
	terminalRow.Connect("notify::text", func() {
		config.TerminalEmulator = terminalRow.Text()
		pd.notifyChanged()
	})
	externalGroup.Add(terminalRow)

	return pd
}

func (pd *PreferencesDialog) notifyChanged() {
	pd.Config.Save()
	if pd.OnChanged != nil {
		pd.OnChanged()
	}
}

func (pd *PreferencesDialog) Show() {
	pd.Present()
}
