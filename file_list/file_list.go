package file_list

import (
	"fmt"
	"os/exec"
	"slices"
	"strings"

	"github.com/MrSametBurgazoglu/atilgan/fileops"
	"github.com/MrSametBurgazoglu/atilgan/preferences"
	"github.com/MrSametBurgazoglu/atilgan/special_path"
	"github.com/MrSametBurgazoglu/atilgan/tag_popup"
	"github.com/MrSametBurgazoglu/atilgan/theme"
	"github.com/MrSametBurgazoglu/atilgan/thumbnail"
	"github.com/MrSametBurgazoglu/atilgan/types"
	"github.com/diamondburned/gotk4-adwaita/pkg/adw"
	"github.com/diamondburned/gotk4/pkg/cairo"
	"github.com/diamondburned/gotk4/pkg/gdk/v4"
	"github.com/diamondburned/gotk4/pkg/gdkpixbuf/v2"
	"github.com/diamondburned/gotk4/pkg/glib/v2"
	"github.com/diamondburned/gotk4/pkg/gtk/v4"
)

const (
	headerHeight = 28
)

type FileList struct {
	*gtk.ScrolledWindow
	Items              []*types.ListItem
	SelectedIDX        int
	DrawingArea        *gtk.DrawingArea
	iconTheme          *gtk.IconTheme
	canSelect          bool
	CanFocus           bool
	CopyCutPaths       []string
	colorTheme         *theme.ColorTheme
	themeConfig        *theme.SpacingConfig
	specialPathManager *special_path.SpecialPathManager
	parent             *gtk.Window
	textureCache       map[string]*gdk.Texture
	Config             *preferences.Config

	// Dynamic dimensions
	rowHeight int
	iconSize  int

	SelectionChanged func(index int)
	PathChanged      func(path string)
	KeyRightPressed  func()
	KeyLeftPressed   func()
}

func (fl *FileList) updateDimensions() {
	fl.iconSize = max(16, fl.Config.IconSize/2)
	fl.rowHeight = fl.themeConfig.GetRowHeight(fl.iconSize)
}

func NewFileList(canSelect bool, specialPathManager *special_path.SpecialPathManager, parent *gtk.Window, config *preferences.Config) *FileList {
	fl := &FileList{
		ScrolledWindow:     gtk.NewScrolledWindow(),
		SelectedIDX:        0,
		DrawingArea:        gtk.NewDrawingArea(),
		iconTheme:          gtk.IconThemeGetForDisplay(gdk.DisplayGetDefault()),
		canSelect:          canSelect,
		CanFocus:           true,
		colorTheme:         theme.NewColorTheme(),
		themeConfig:        theme.NewTheme(),
		specialPathManager: specialPathManager,
		parent:             parent,
		textureCache:       make(map[string]*gdk.Texture),
		Config:             config,
	}
	fl.updateDimensions()

	fl.DrawingArea.SetHExpand(true)
	fl.DrawingArea.SetDrawFunc(fl.onDraw)

	fl.SetChild(fl.DrawingArea)
	fl.SetVExpand(true)
	fl.SetHExpand(true)

	fl.SetMinContentWidth(200)
	fl.SetMarginStart(0)
	fl.SetMarginEnd(0)
	fl.SetPolicy(gtk.PolicyAutomatic, gtk.PolicyAlways)

	if canSelect {
		key := gtk.NewEventControllerKey()
		key.ConnectKeyPressed(func(keyval, keycode uint, state gdk.ModifierType) bool {
			switch keyval {
			case gdk.KEY_Up:
				if fl.SelectedIDX > 0 {
					fl.selectItem(fl.SelectedIDX - 1)
				}
				return true

			case gdk.KEY_Down:
				if fl.SelectedIDX < len(fl.Items)-1 {
					fl.selectItem(fl.SelectedIDX + 1)
				}
				return true

			case gdk.KEY_Left:
				if fl.KeyLeftPressed != nil {
					fl.KeyLeftPressed()
				}
				return true

			case gdk.KEY_Right:
				if fl.KeyRightPressed != nil {
					fl.KeyRightPressed()
				}
				return true

			case gdk.KEY_Return:
				if fl.SelectedIDX >= 0 && fl.Items[fl.SelectedIDX] != nil {
					if !fl.Items[fl.SelectedIDX].IsDir {
						cmd := exec.Command("xdg-open", fl.Items[fl.SelectedIDX].Path)
						cmd.Start()
					} else {
						if fl.KeyRightPressed != nil {
							fl.KeyRightPressed()
						}
					}
				}
				return true
			}
			return false
		})

		fl.DrawingArea.AddController(key)

		fl.DrawingArea.SetFocusable(true)
		fl.DrawingArea.AddController(fl.newMouseController(fl.DrawingArea))
		fl.DrawingArea.AddController(fl.newContextMenuController(fl.DrawingArea))

		dragSource := gtk.NewDragSource()
		dragSource.SetActions(gdk.ActionCopy)
		dragSource.ConnectPrepare(func(x, y float64) *gdk.ContentProvider {
			idx := fl.ItemAt(int(y))
			if idx < 0 {
				return nil
			}
			item := fl.Items[idx]

			iconSize := 16
			iconName := fileops.GetIconForFile(item.Name)
			if item.IsDir {
				iconName = fileops.GetIconForFolder(item.Path)
			}
			
			// Try to lookup icon, fallback to default if not found
			paintable := fl.iconTheme.LookupIcon(iconName, nil, iconSize, 1, gtk.TextDirNone, 0)
			if paintable == nil && item.IsDir {
				paintable = fl.iconTheme.LookupIcon("folder", nil, iconSize, 1, gtk.TextDirNone, 0)
			}
			if paintable == nil && !item.IsDir {
				paintable = fl.iconTheme.LookupIcon("text-x-generic", nil, iconSize, 1, gtk.TextDirNone, 0)
			}
			if paintable != nil {
				dragSource.SetIcon(paintable, 0, 0)
			}

			uri := "file://" + item.Path + "\r\n"
			return gdk.NewContentProviderForBytes("text/uri-list", glib.NewBytes([]byte(uri)))
		})
		fl.DrawingArea.AddController(dragSource)
	}

	fl.DrawingArea.AddController(fl.newGestureClick(fl.DrawingArea))

	return fl
}

func (fl *FileList) updateHeight() {
	y := 0
	currentGroup := ""
	for _, item := range fl.Items {
		if item.Group != currentGroup {
			y += headerHeight
			currentGroup = item.Group
		}
		y += fl.rowHeight
	}
	fl.DrawingArea.SetContentHeight(y)
}

func (fl *FileList) SetItems(items []*types.ListItem) {
	fl.Items = items
	fl.SelectedIDX = 0
	fl.textureCache = make(map[string]*gdk.Texture)
	fl.updateHeight()
	fl.DrawingArea.QueueDraw()
}

type FileListView interface {
	gtk.Widgetter
	SetItems([]*types.ListItem)
	GetItems() []*types.ListItem
	GetSelectedIDX() int
	GetSelectedIdxs() map[int]bool
	CleanCopyCutFiles()
	AddCopyCutItem(string) bool
	Refresh(bool)
	SetPathChanged(func(string))
	SetSelectionChanged(func(int))
	SetSelectedItemWithLetter(string)
	SetItem(int)
	SelectAll()
	ClearSelection()
	AddItem(*types.ListItem)
	SetCanFocus(bool)
	SetKeyLeftPressed(func())
	SetKeyRightPressed(func())
	SetPinRequested(func(string))
	FocusDrawingArea()
}

func (fl *FileList) FocusDrawingArea() {
	fl.DrawingArea.GrabFocus()
}

func (fl *FileList) SetCanFocus(canFocus bool) {
	fl.CanFocus = canFocus
}

func (fl *FileList) SetKeyLeftPressed(f func()) {
	fl.KeyLeftPressed = f
}

func (fl *FileList) SetKeyRightPressed(f func()) {
	fl.KeyRightPressed = f
}

func (fl *FileList) SetPinRequested(f func(string)) {
	// FileList might not implement PinRequested yet, but we fulfill the interface
}

func (fl *FileList) SetSelectedItemWithLetter(letter string) {
	for i, item := range fl.Items {
		if strings.HasPrefix(strings.ToLower(item.Name), strings.ToLower(letter)) {
			fl.SetItem(i)
			break
		}
	}
}

func (fl *FileList) SetItem(index int) {
	if index >= 0 && index < len(fl.Items) {
		fl.SelectedIDX = index
		fl.DrawingArea.QueueDraw()
	}
}

func (fl *FileList) SelectAll() {
	// FileList selection is simpler (one item usually, but let's be consistent if needed)
	// Actually FileList only supports single selection currently based on implementation
	// But we implement the interface anyway.
	fl.DrawingArea.QueueDraw()
}

func (fl *FileList) ClearSelection() {
	fl.DrawingArea.QueueDraw()
}

func (fl *FileList) GetItems() []*types.ListItem {
	return fl.Items
}

func (fl *FileList) AddItem(item *types.ListItem) {
	fl.Items = append(fl.Items, item)
	fl.updateHeight()
	fl.DrawingArea.QueueDraw()
}

func (fl *FileList) GetSelectedIDX() int {
	return fl.SelectedIDX
}

func (fl *FileList) GetSelectedIdxs() map[int]bool {
	return map[int]bool{fl.SelectedIDX: true}
}

func (fl *FileList) CleanCopyCutFiles() {
	fl.CopyCutPaths = []string{}
	fl.DrawingArea.QueueDraw()
}

func (fl *FileList) AddCopyCutItem(path string) bool {
	if slices.Contains(fl.CopyCutPaths, path) {
		return false
	}
	fl.CopyCutPaths = append(fl.CopyCutPaths, path)
	fl.DrawingArea.QueueDraw()
	return true
}

func (fl *FileList) Refresh(newFilter bool) {
	fl.updateDimensions()
	fl.DrawingArea.QueueDraw()
}

func (fl *FileList) SetPathChanged(f func(string)) {
	fl.PathChanged = f
}

func (fl *FileList) SetSelectionChanged(f func(int)) {
	fl.SelectionChanged = f
}

func (fl *FileList) onDraw(da *gtk.DrawingArea, cr *cairo.Context, w, h int) {
	// Try to get system colors from StyleContext
	sc := da.StyleContext()
	
	if bg, ok := sc.LookupColor("view_bg_color"); ok {
		fl.colorTheme.BackgroundColor = *bg
	} else if bg, ok := sc.LookupColor("window_bg_color"); ok {
		fl.colorTheme.BackgroundColor = *bg
	}
	
	if fg, ok := sc.LookupColor("view_fg_color"); ok {
		fl.colorTheme.TextColor = *fg
	} else if fg, ok := sc.LookupColor("window_fg_color"); ok {
		fl.colorTheme.TextColor = *fg
	}

	if accent, ok := sc.LookupColor("accent_bg_color"); ok {
		fl.colorTheme.AccentColor = *accent
		fl.colorTheme.SelectedBgColor = gdk.NewRGBA(accent.Red(), accent.Green(), accent.Blue(), 0.2)
		fl.colorTheme.SelectedTextColor = *accent
	}
	
	if headerBg, ok := sc.LookupColor("headerbar_bg_color"); ok {
		fl.colorTheme.HeaderBackgroundColor = gdk.NewRGBA(headerBg.Red(), headerBg.Green(), headerBg.Blue(), 0.5)
	}

	// Pre-calculate height to fill background accurately
	contentHeight := 0
	tempGroup := ""
	for _, item := range fl.Items {
		if item.Group != tempGroup {
			contentHeight += headerHeight
			tempGroup = item.Group
		}
		contentHeight += fl.rowHeight
	}

	cr.SetSourceRGBA(float64(fl.colorTheme.BackgroundColor.Red()), float64(fl.colorTheme.BackgroundColor.Green()), float64(fl.colorTheme.BackgroundColor.Blue()), float64(fl.colorTheme.BackgroundColor.Alpha()))
	cr.Rectangle(0, 0, float64(w), float64(contentHeight))
	cr.Fill()

	y := 0
	currentGroup := ""

	for i, item := range fl.Items {
		if item.Group != currentGroup {
			fl.drawHeader(cr, item.Group, y, w)
			y += headerHeight
			currentGroup = item.Group
		}

		fl.drawRow(cr, i, item, y, w)
		y += fl.rowHeight
	}
}

func (fl *FileList) selectItem(idx int) {
	if idx < 0 || idx >= len(fl.Items) {
		return
	}
	fl.SelectedIDX = idx
	fl.ensureVisible()
	fl.DrawingArea.QueueDraw()
	if fl.SelectionChanged != nil {
		fl.SelectionChanged(fl.SelectedIDX)
	}
}

func (fl *FileList) drawHeader(cr *cairo.Context, text string, y int, w int) {
	cr.SetSourceRGBA(float64(fl.colorTheme.HeaderBackgroundColor.Red()), float64(fl.colorTheme.HeaderBackgroundColor.Green()), float64(fl.colorTheme.HeaderBackgroundColor.Blue()), float64(fl.colorTheme.HeaderBackgroundColor.Alpha()))
	cr.Rectangle(0, float64(y), float64(w), float64(headerHeight))
	cr.Fill()

	cr.SetSourceRGBA(float64(fl.colorTheme.HeaderTextColor.Red()), float64(fl.colorTheme.HeaderTextColor.Green()), float64(fl.colorTheme.HeaderTextColor.Blue()), float64(fl.colorTheme.HeaderTextColor.Alpha()))
	cr.SelectFontFace("Sans", cairo.FontSlantNormal, cairo.FontWeightBold)
	cr.SetFontSize(fl.themeConfig.Fonts.HeaderSize)
	cr.MoveTo(12, float64(y+18))
	cr.ShowText(text)
}

func (fl *FileList) drawRow(cr *cairo.Context, idx int, item *types.ListItem, y int, w int) {
	padding := 4.0
	if idx == fl.SelectedIDX && fl.canSelect {
		cr.SetSourceRGBA(float64(fl.colorTheme.SelectedBgColor.Red()), float64(fl.colorTheme.SelectedBgColor.Green()), float64(fl.colorTheme.SelectedBgColor.Blue()), float64(fl.colorTheme.SelectedBgColor.Alpha()))
		roundedRectangle(cr, padding, float64(y)+padding, float64(w)-2*padding, float64(fl.rowHeight)-2*padding, 8)
		cr.Fill()

		// Border for selection
		cr.SetSourceRGBA(float64(fl.colorTheme.AccentColor.Red()), float64(fl.colorTheme.AccentColor.Green()), float64(fl.colorTheme.AccentColor.Blue()), 0.5)
		cr.SetLineWidth(1)
		roundedRectangle(cr, padding, float64(y)+padding, float64(w)-2*padding, float64(fl.rowHeight)-2*padding, 8)
		cr.Stroke()
	} else if slices.Contains(fl.CopyCutPaths, item.Path) {
		cr.SetSourceRGBA(float64(fl.colorTheme.CopyCutBgColor.Red()), float64(fl.colorTheme.CopyCutBgColor.Green()), float64(fl.colorTheme.CopyCutBgColor.Blue()), float64(fl.colorTheme.CopyCutBgColor.Alpha()))
		roundedRectangle(cr, padding, float64(y)+padding, float64(w)-2*padding, float64(fl.rowHeight)-2*padding, 8)
		cr.Fill()
	} else {
		cr.SetSourceRGBA(float64(fl.colorTheme.BackgroundColor.Red()), float64(fl.colorTheme.BackgroundColor.Green()), float64(fl.colorTheme.BackgroundColor.Blue()), float64(fl.colorTheme.BackgroundColor.Alpha()))
		cr.Rectangle(0, float64(y), float64(w), float64(fl.rowHeight))
		cr.Fill()
	}

	var texture *gdk.Texture

	if !item.IsDir && fl.Config.EnableThumbnails {
		if cached, ok := fl.textureCache[item.Path]; ok {
			texture = cached
		} else {
			texture, _ = thumbnail.Generate(item.Path)
			fl.textureCache[item.Path] = texture
		}
	}

	if texture != nil {
		pixbuf := gdk.PixbufGetFromTexture(texture)
		if pixbuf != nil {
			if pixbuf.Width() != fl.iconSize || pixbuf.Height() != fl.iconSize {
				pixbuf = pixbuf.ScaleSimple(fl.iconSize, fl.iconSize, gdkpixbuf.InterpBilinear)
			}
			gdk.CairoSetSourcePixbuf(cr, pixbuf, 12, float64(y+(fl.rowHeight-fl.iconSize)/2))
			cr.Paint()
		}
	} else {
		iconName := fileops.GetIconForFile(item.Name)
		if item.IsDir {
			iconName = fileops.GetIconForFolder(item.Path)
		}

		paintable := fl.iconTheme.LookupIcon(iconName, nil, fl.iconSize, 1, gtk.TextDirNone, 0)
		if paintable == nil && item.IsDir {
			paintable = fl.iconTheme.LookupIcon("folder", nil, fl.iconSize, 1, gtk.TextDirNone, 0)
		}
		if paintable == nil && !item.IsDir {
			paintable = fl.iconTheme.LookupIcon("text-x-generic", nil, fl.iconSize, 1, gtk.TextDirNone, 0)
		}
		if paintable != nil {
			file := paintable.File()
			if file != nil {
				texture, err := gdk.NewTextureFromFile(file)
				if err == nil {
					pixbuf := gdk.PixbufGetFromTexture(texture)
					if pixbuf != nil {
						if pixbuf.Width() != fl.iconSize || pixbuf.Height() != fl.iconSize {
							pixbuf = pixbuf.ScaleSimple(fl.iconSize, fl.iconSize, gdkpixbuf.InterpBilinear)
						}
						gdk.CairoSetSourcePixbuf(cr, pixbuf, 12, float64(y+(fl.rowHeight-fl.iconSize)/2))
						cr.Paint()
					}
				}
			}
		}
	}

	if idx == fl.SelectedIDX && fl.canSelect {
		cr.SetSourceRGBA(float64(fl.colorTheme.SelectedTextColor.Red()), float64(fl.colorTheme.SelectedTextColor.Green()), float64(fl.colorTheme.SelectedTextColor.Blue()), float64(fl.colorTheme.SelectedTextColor.Alpha()))
	} else {
		cr.SetSourceRGBA(float64(fl.colorTheme.TextColor.Red()), float64(fl.colorTheme.TextColor.Green()), float64(fl.colorTheme.TextColor.Blue()), float64(fl.colorTheme.TextColor.Alpha()))
	}
	cr.SelectFontFace("Sans", cairo.FontSlantNormal, cairo.FontWeightNormal)
	cr.SetFontSize(fl.themeConfig.Fonts.FilenameSize)
	yOffset := fl.themeConfig.GetFilenameYOffset(fl.rowHeight)
	cr.MoveTo(40, float64(y)+yOffset)
	cr.ShowText(item.Name)

	if idx == fl.SelectedIDX && fl.canSelect {
		cr.SetSourceRGBA(float64(fl.colorTheme.SelectedTextColor.Red()), float64(fl.colorTheme.SelectedTextColor.Green()), float64(fl.colorTheme.SelectedTextColor.Blue()), float64(fl.colorTheme.SelectedTextColor.Alpha()))
	} else {
		cr.SetSourceRGBA(float64(fl.colorTheme.TextColor.Red()), float64(fl.colorTheme.TextColor.Green()), float64(fl.colorTheme.TextColor.Blue()), float64(fl.colorTheme.TextColor.Alpha()*0.8))
	}
	cr.SelectFontFace("Sans", cairo.FontSlantNormal, cairo.FontWeightNormal)
	cr.SetFontSize(fl.themeConfig.Fonts.SizeTextSize)
	sizeYOffset := fl.themeConfig.GetSizeTextYOffset(fl.rowHeight)
	cr.MoveTo(float64(w)-100, float64(y)+sizeYOffset)
	if item.IsDir {
		itemText := "items"
		if item.ItemCount == 1 {
			itemText = "item"
		}
		cr.ShowText(fmt.Sprintf("%d %s", item.ItemCount, itemText))
	} else if item.Size > 0 {
		cr.ShowText(fileops.GetFileSizeAsString(item.Size))
	}
}

func (fl *FileList) getItemBounds(idx int) (top, bottom int) {
	if idx < 0 || idx >= len(fl.Items) {
		return 0, 0
	}

	pos := 0
	currentGroup := ""

	for i, item := range fl.Items {
		if item.Group != currentGroup {
			pos += headerHeight
			currentGroup = item.Group
		}

		if i == idx {
			return pos, pos + fl.rowHeight
		}
		pos += fl.rowHeight
	}
	return 0, 0
}

func (fl *FileList) ensureVisible() {
	adj := fl.VAdjustment()
	scrollPos := adj.Value()
	visibleHeight := float64(fl.Height())

	itemTop, itemBottom := fl.getItemBounds(fl.SelectedIDX)

	if float64(itemTop) >= scrollPos && float64(itemBottom) <= scrollPos+visibleHeight {
		return
	}

	if float64(itemTop) < scrollPos {
		targetIdx := fl.SelectedIDX - 5
		if targetIdx < 0 {
			targetIdx = 0
		}
		targetTop, _ := fl.getItemBounds(targetIdx)
		adj.SetValue(float64(targetTop))
	} else {
		targetIdx := fl.SelectedIDX + 5
		if targetIdx >= len(fl.Items) {
			targetIdx = len(fl.Items) - 1
		}
		_, targetBottom := fl.getItemBounds(targetIdx)
		newValue := float64(targetBottom) - visibleHeight
		if newValue < 0 {
			newValue = 0
		}
		adj.SetValue(newValue)
	}
}

func (fl *FileList) newGestureClick(da *gtk.DrawingArea) *gtk.GestureClick {
	click := gtk.NewGestureClick()
	click.ConnectPressed(func(n int, x, y float64) {
		idx := fl.ItemAt(int(y))
		if idx >= 0 {
			fl.selectItem(idx)

			if click.CurrentButton() == gdk.BUTTON_PRIMARY && n == 2 {
				if !fl.Items[idx].IsDir {
					cmd := exec.Command("xdg-open", fl.Items[idx].Path)
					cmd.Start()
				} else {
					fl.PathChanged(fl.Items[idx].Path)
				}
			}
		}
	})
	return click
}

func (fl *FileList) newContextMenuController(da *gtk.DrawingArea) *gtk.GestureClick {
	click := gtk.NewGestureClick()
	click.SetButton(gdk.BUTTON_SECONDARY)
	click.ConnectPressed(func(n int, x, y float64) {
		idx := fl.ItemAt(int(y))
		if idx < 0 {
			return
		}

		pop := gtk.NewPopover()
		popoverBox := gtk.NewBox(gtk.OrientationVertical, 6)
		pop.SetChild(popoverBox)

		open := gtk.NewButtonWithLabel("Open")
		open.Connect("clicked", func() {
			if !fl.Items[idx].IsDir {
				cmd := exec.Command("xdg-open", fl.Items[idx].Path)
				cmd.Start()
			} else {
				fl.PathChanged(fl.Items[idx].Path)
				pop.Popdown()
			}
		})

		delete := gtk.NewButtonWithLabel("Delete")
		delete.Connect("clicked", func() {
			performDelete := func() {
				cmd := exec.Command("gio", "trash", fl.Items[idx].Path)
				err := cmd.Start()
				if err != nil {
					println("couldn't delete file")
					return
				}
				go func() {
					cmd.Wait()
					glib.IdleAdd(func() {
						fl.PathChanged("")
						pop.Popdown()
					})
				}()
			}

			if fl.Config.ConfirmDelete {
				dialog := adw.NewMessageDialog(fl.parent, "Confirm Deletion", "Are you sure you want to move this item to trash?")
				dialog.AddResponse("cancel", "Cancel")
				dialog.AddResponse("delete", "Delete")
				dialog.SetResponseAppearance("delete", adw.ResponseDestructive)
				dialog.SetDefaultResponse("cancel")
				dialog.SetCloseResponse("cancel")

				dialog.ConnectResponse(func(response string) {
					if response == "delete" {
						performDelete()
					}
					dialog.Destroy()
				})
				dialog.Present()
			} else {
				performDelete()
			}
		})

		addTag := gtk.NewButtonWithLabel("Add Tag")
		addTag.Connect("clicked", func() {
			tagPopup := tag_popup.NewTagPopup(fl.parent, fl.specialPathManager.GetTagManager(), fl.Items[idx].Path)
			tagPopup.Show()
			pop.Popdown()
		})

		popoverBox.Append(open)
		popoverBox.Append(delete)
		popoverBox.Append(addTag)
		pop.SetHasArrow(true)
		rect := gdk.NewRectangle(int(x), int(y), 1, 1)
		pop.SetPointingTo(&rect)

		pop.SetParent(da)
		pop.Popup()
	})
	return click
}

func (fl *FileList) ItemAt(y int) int {
	currentGroup := ""
	pos := 0

	for i, item := range fl.Items {
		if item.Group != currentGroup {
			if y >= pos && y < pos+headerHeight {
				return -1
			}
			pos += headerHeight
			currentGroup = item.Group
		}

		if y >= pos && y < pos+fl.rowHeight {
			return i
		}
		pos += fl.rowHeight
	}
	return -1
}

func (fl *FileList) newMouseController(da *gtk.DrawingArea) *gtk.EventControllerMotion {
	ctrl := gtk.NewEventControllerMotion()
	ctrl.ConnectMotion(func(x, y float64) {
		if fl.CanFocus {
			fl.DrawingArea.GrabFocus()
		}
	})
	return ctrl
}
