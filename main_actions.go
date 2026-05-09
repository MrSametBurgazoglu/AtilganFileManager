package main

import (
	"github.com/MrSametBurgazoglu/atilgan/create_popup"
	"github.com/MrSametBurgazoglu/atilgan/header"
	"github.com/MrSametBurgazoglu/atilgan/preferences"
	"github.com/MrSametBurgazoglu/atilgan/shortcut_popup"
	"github.com/MrSametBurgazoglu/atilgan/types"
	"github.com/diamondburned/gotk4-adwaita/pkg/adw"
	"github.com/diamondburned/gotk4/pkg/gtk/v4"
)

func (m *MainBox) setupActionsMenu(mainWindow *adw.ApplicationWindow, headerBar *header.HeaderBar) {
	actionsPopover := gtk.NewPopover()
	headerBar.ActionsButton.SetPopover(actionsPopover)

	actionsBox := gtk.NewBox(gtk.OrientationVertical, 6)
	actionsBox.SetMarginTop(8)
	actionsBox.SetMarginBottom(8)
	actionsBox.SetMarginStart(8)
	actionsBox.SetMarginEnd(8)
	actionsPopover.SetChild(actionsBox)

	createActionBtn := func(iconName, labelText string, action func()) *gtk.Button {
		btn := gtk.NewButton()
		btn.AddCSSClass("flat")
		
		btnBox := gtk.NewBox(gtk.OrientationHorizontal, 8)
		icon := gtk.NewImageFromIconName(iconName)
		label := gtk.NewLabel(labelText)
		label.SetHAlign(gtk.AlignStart)
		
		btnBox.Append(icon)
		btnBox.Append(label)
		btn.SetChild(btnBox)
		
		btn.ConnectClicked(func() {
			action()
			actionsPopover.Popdown()
		})
		return btn
	}

	actionsLabel := gtk.NewLabel("Actions")
	actionsLabel.AddCSSClass("caption")
	actionsLabel.SetHAlign(gtk.AlignStart)
	actionsBox.Append(actionsLabel)

	createBox := gtk.NewBox(gtk.OrientationHorizontal, 4)
	createBox.SetHomogeneous(true)

	createBox.Append(createActionBtn("document-new-symbolic", "New File", func() {
		fs := create_popup.NewFileSelector(m.Path, m.pathChanged)
		fs.SetVisible(true)
		fs.SetTransientFor(&mainWindow.Window)
		fs.SetModal(true)
	}))
	createBox.Append(createActionBtn("folder-new-symbolic", "New Folder", func() {
		ds := create_popup.NewDirectorySelector(m.Path, m.pathChanged)
		ds.SetVisible(true)
		ds.SetTransientFor(&mainWindow.Window)
		ds.SetModal(true)
	}))
	actionsBox.Append(createBox)
	actionsBox.Append(gtk.NewSeparator(gtk.OrientationHorizontal))

	actionsBox.Append(createActionBtn("utilities-terminal-symbolic", "Open Terminal", func() {
		activePanel := m.getActivePanel()
		if activePanel != nil {
			activePanel.FileViewer.OpenTerminal()
		}
	}))

	actionsBox.Append(createActionBtn("tab-new-symbolic", "New Tab", func() {
		m.Workspace.NewTab(m.Path)
	}))

	actionsBox.Append(gtk.NewSeparator(gtk.OrientationHorizontal))

	actionsBox.Append(createActionBtn("preferences-desktop-keyboard-shortcuts-symbolic", "Shortcuts", func() {
		shortcut_popup.NewShortcutPopup(&mainWindow.Window)
	}))
	actionsBox.Append(createActionBtn("preferences-system-symbolic", "Preferences", func() {
		dialog := preferences.NewPreferencesDialog(&mainWindow.Window, m.Config)
		dialog.OnChanged = m.applyPreferences
		dialog.Show()
	}))
	actionsBox.Append(createActionBtn("help-about-symbolic", "About", func() {
		aboutDialog := adw.NewAboutWindow()
		aboutDialog.SetApplicationName("Atilgan")
		aboutDialog.SetVersion("0.2.0")
		aboutDialog.SetApplicationIcon("io.github.mrsametburgazoglu.AtilganFileManager")
		aboutDialog.SetCopyright("Copyright © 2025 MrSametBurgazoglu")
		aboutDialog.SetWebsite("https://github.com/MrSametBurgazoglu/AtilganFileManager")
		aboutDialog.SetTransientFor(&mainWindow.Window)
		aboutDialog.SetVisible(true)
	}))

	actionsBox.Append(gtk.NewSeparator(gtk.OrientationHorizontal))

	sortLabel := gtk.NewLabel("Sort By")
	sortLabel.AddCSSClass("caption")
	sortLabel.SetHAlign(gtk.AlignStart)
	actionsBox.Append(sortLabel)

	sortBox := gtk.NewBox(gtk.OrientationHorizontal, 0)
	sortBox.AddCSSClass("linked")
	
	sortNameBtn := gtk.NewButtonWithLabel("Name")
	sortTimeBtn := gtk.NewButtonWithLabel("Time")
	sortSizeBtn := gtk.NewButtonWithLabel("Size")
	sortTypeBtn := gtk.NewButtonWithLabel("Type")

	sortNameBtn.ConnectClicked(func() {
		activePanel := m.getActivePanel()
		if activePanel != nil {
			activePanel.FileViewer.SortOrder = types.SortByName
			activePanel.FileViewer.Refresh(false)
		}
		actionsPopover.Popdown()
	})
	sortTimeBtn.ConnectClicked(func() {
		activePanel := m.getActivePanel()
		if activePanel != nil {
			activePanel.FileViewer.SortOrder = types.SortByTime
			activePanel.FileViewer.Refresh(false)
		}
		actionsPopover.Popdown()
	})
	sortSizeBtn.ConnectClicked(func() {
		activePanel := m.getActivePanel()
		if activePanel != nil {
			activePanel.FileViewer.SortOrder = types.SortBySize
			activePanel.FileViewer.Refresh(false)
		}
		actionsPopover.Popdown()
	})
	sortTypeBtn.ConnectClicked(func() {
		activePanel := m.getActivePanel()
		if activePanel != nil {
			activePanel.FileViewer.SortOrder = types.SortByType
			activePanel.FileViewer.Refresh(false)
		}
		actionsPopover.Popdown()
	})

	sortBox.Append(sortNameBtn)
	sortBox.Append(sortTimeBtn)
	sortBox.Append(sortSizeBtn)
	sortBox.Append(sortTypeBtn)
	sortNameBtn.SetHExpand(true)
	sortTimeBtn.SetHExpand(true)
	sortSizeBtn.SetHExpand(true)
	sortTypeBtn.SetHExpand(true)
	actionsBox.Append(sortBox)

	actionsBox.Append(gtk.NewSeparator(gtk.OrientationHorizontal))

	filterLabel := gtk.NewLabel("Filters")
	filterLabel.AddCSSClass("caption")
	filterLabel.SetHAlign(gtk.AlignStart)
	actionsBox.Append(filterLabel)
	
	filterWrapper := gtk.NewBox(gtk.OrientationVertical, 0)
	actionsBox.Append(filterWrapper)

	actionsPopover.Connect("map", func() {
		activePanel := m.getActivePanel()
		if activePanel != nil {
			for child := filterWrapper.FirstChild(); child != nil; {
				next := gtk.BaseWidget(child).NextSibling()
				filterWrapper.Remove(child)
				child = next
			}
			filterWrapper.Append(activePanel.FileViewer.FilterBox)
		}
	})
}
