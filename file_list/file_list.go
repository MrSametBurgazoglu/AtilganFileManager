package file_list

import (
	"fmt"
	"os/exec"
	"slices"
	"strings"

	"github.com/MrSametBurgazoglu/atilgan/fileops"
	"github.com/MrSametBurgazoglu/atilgan/special_path"
	"github.com/MrSametBurgazoglu/atilgan/tag_popup"
	"github.com/MrSametBurgazoglu/atilgan/types"
	"github.com/diamondburned/gotk4/pkg/cairo"
	"github.com/diamondburned/gotk4/pkg/gdk/v4"
	"github.com/diamondburned/gotk4/pkg/glib/v2"
	"github.com/diamondburned/gotk4/pkg/gtk/v4"
)

type FileListTheme struct {
	BackgroundColor       gdk.RGBA
	TextColor             gdk.RGBA
	SelectedBgColor       gdk.RGBA
	SelectedTextColor     gdk.RGBA
	HeaderBackgroundColor gdk.RGBA
	HeaderTextColor       gdk.RGBA
	CopyCutBgColor        gdk.RGBA
	HoverBgColor          gdk.RGBA
}

func NewFileListTheme() *FileListTheme {
	return &FileListTheme{
		BackgroundColor:       gdk.NewRGBA(45.0/255, 45.0/255, 45.0/255, 1),
		TextColor:             gdk.NewRGBA(245.0/255, 245.0/255, 245.0/255, 1),
		SelectedBgColor:       gdk.NewRGBA(64.0/255, 64.0/255, 64.0/255, 1),
		SelectedTextColor:     gdk.NewRGBA(245.0/255, 245.0/255, 245.0/255, 1),
		HeaderBackgroundColor: gdk.NewRGBA(36.0/255, 36.0/255, 36.0/255, 1),
		HeaderTextColor:       gdk.NewRGBA(245.0/255, 245.0/255, 245.0/255, 1),
		CopyCutBgColor:        gdk.NewRGBA(50.0/255, 70.0/255, 90.0/255, 1),
		HoverBgColor:          gdk.NewRGBA(55.0/255, 55.0/255, 55.0/255, 1),
	}
}

const (
	rowHeight    = 36
	headerHeight = 20
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
	theme              *FileListTheme
	specialPathManager *special_path.SpecialPathManager
	parent             *gtk.Window

	SelectionChanged func(index int)
	PathChanged      func(path string)
	KeyRightPressed  func()
	KeyLeftPressed   func()
}

func NewFileList(canSelect bool, specialPathManager *special_path.SpecialPathManager, parent *gtk.Window) *FileList {
	fl := &FileList{
		ScrolledWindow:     gtk.NewScrolledWindow(),
		SelectedIDX:        0,
		DrawingArea:        gtk.NewDrawingArea(),
		iconTheme:          gtk.IconThemeGetForDisplay(gdk.DisplayGetDefault()),
		canSelect:          canSelect,
		CanFocus:           true,
		theme:              NewFileListTheme(),
		specialPathManager: specialPathManager,
		parent:             parent,
	}

	fl.DrawingArea.SetDrawFunc(fl.onDraw)

	fl.SetChild(fl.DrawingArea)
	fl.SetVExpand(true)

	fl.SetMinContentWidth(600)
	fl.SetMaxContentHeight(600)
	fl.SetPolicy(gtk.PolicyAlways, gtk.PolicyAlways)

	if canSelect {
		key := gtk.NewEventControllerKey()
		key.ConnectKeyPressed(func(keyval, keycode uint, state gdk.ModifierType) bool {
			switch keyval {
			case gdk.KEY_Up:
				if fl.SelectedIDX > 0 {
					fl.SelectedIDX--
					fl.DrawingArea.QueueDraw()
					fl.SelectionChanged(fl.SelectedIDX)
				}
				return true

			case gdk.KEY_Down:
				if fl.SelectedIDX < len(fl.Items)-1 {
					fl.SelectedIDX++
					fl.DrawingArea.QueueDraw()
					fl.SelectionChanged(fl.SelectedIDX)
				}
				return true

			case gdk.KEY_Left:
				fl.KeyLeftPressed()
				return true

			case gdk.KEY_Right:
				fl.KeyRightPressed()
				return true

			case gdk.KEY_Return:
				if fl.SelectedIDX >= 0 && fl.Items[fl.SelectedIDX] != nil {
					if !fl.Items[fl.SelectedIDX].IsDir {
						cmd := exec.Command("xdg-open", fl.Items[fl.SelectedIDX].Path)
						cmd.Start()
					} else {
						fl.KeyRightPressed()
					}
				}
				return true
			}
			return false
		})

		fl.DrawingArea.AddController(key)

		fl.DrawingArea.SetFocusable(true)
		fl.DrawingArea.AddController(fl.newMouseController(fl.DrawingArea))
		fl.DrawingArea.AddController(fl.newGestureClick(fl.DrawingArea))

		fl.DrawingArea.AddController(fl.newContextMenuController(fl.DrawingArea))

		dragSource := gtk.NewDragSource()
		dragSource.SetActions(gdk.ActionCopy)
		dragSource.ConnectPrepare(func(x, y float64) *gdk.ContentProvider {
			idx := fl.ItemAt(int(y))
			if idx < 0 {
				return nil
			}
			item := fl.Items[idx]

			iconSize := 32
			iconName := fileops.GetIconForFile(item.Name)
			if item.IsDir {
				iconName = "folder"
			}
			paintable := fl.iconTheme.LookupIcon(iconName, nil, iconSize, 1, gtk.TextDirNone, 0)
			if paintable != nil {
				dragSource.SetIcon(paintable, 0, 0)
			}

			uri := "file://" + item.Path + "\r\n"
			return gdk.NewContentProviderForBytes("text/uri-list", glib.NewBytes([]byte(uri)))
		})
		fl.DrawingArea.AddController(dragSource)
	}

	return fl
}

func (fl *FileList) SetItems(items []*types.ListItem) {
	fl.Items = items
	fl.SelectedIDX = 0
	fl.DrawingArea.QueueDraw()
}

func (fl *FileList) AddItem(item *types.ListItem) {
	fl.Items = append(fl.Items, item)
	fl.DrawingArea.QueueDraw()
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

func (fl *FileList) AddCopyCutItem(path string) bool {
	if slices.Contains(fl.CopyCutPaths, path) {
		return false
	}
	fl.CopyCutPaths = append(fl.CopyCutPaths, path)
	fl.DrawingArea.QueueDraw()
	return true
}

func (fl *FileList) CleanCopyCutItems() {
	fl.CopyCutPaths = make([]string, 0)
	fl.DrawingArea.QueueDraw()
}

func (fl *FileList) onDraw(da *gtk.DrawingArea, cr *cairo.Context, w, h int) {
	y := 0
	currentGroup := ""

	for i, item := range fl.Items {
		if item.Group != currentGroup {
			fl.drawHeader(cr, item.Group, y)
			y += headerHeight
			currentGroup = item.Group
		}

		fl.drawRow(cr, i, item, y)
		y += rowHeight
	}
	fl.DrawingArea.SetContentHeight(y)
	fl.ensureVisible()
}

func (fl *FileList) drawHeader(cr *cairo.Context, text string, y int) {
	cr.SetSourceRGBA(float64(fl.theme.HeaderBackgroundColor.Red()), float64(fl.theme.HeaderBackgroundColor.Green()), float64(fl.theme.HeaderBackgroundColor.Blue()), float64(fl.theme.HeaderBackgroundColor.Alpha()))
	cr.Rectangle(0, float64(y), 1200, float64(headerHeight))
	cr.Fill()

	cr.SetSourceRGBA(float64(fl.theme.HeaderTextColor.Red()), float64(fl.theme.HeaderTextColor.Green()), float64(fl.theme.HeaderTextColor.Blue()), float64(fl.theme.HeaderTextColor.Alpha()))
	cr.SelectFontFace("Sans", cairo.FontSlantNormal, cairo.FontWeightBold)
	cr.SetFontSize(10)
	cr.MoveTo(8, float64(y+15))
	cr.ShowText(text)
}

func (fl *FileList) drawRow(cr *cairo.Context, idx int, item *types.ListItem, y int) {
	if idx == fl.SelectedIDX && fl.canSelect {
		cr.SetSourceRGBA(float64(fl.theme.SelectedBgColor.Red()), float64(fl.theme.SelectedBgColor.Green()), float64(fl.theme.SelectedBgColor.Blue()), float64(fl.theme.SelectedBgColor.Alpha()))
		cr.Rectangle(0, float64(y), 1200, float64(rowHeight))
		cr.Fill()
	} else if slices.Contains(fl.CopyCutPaths, item.Path) {
		cr.SetSourceRGBA(float64(fl.theme.CopyCutBgColor.Red()), float64(fl.theme.CopyCutBgColor.Green()), float64(fl.theme.CopyCutBgColor.Blue()), float64(fl.theme.CopyCutBgColor.Alpha()))
		cr.Rectangle(0, float64(y), 1200, float64(rowHeight))
		cr.Fill()
	} else {
		cr.SetSourceRGBA(float64(fl.theme.BackgroundColor.Red()), float64(fl.theme.BackgroundColor.Green()), float64(fl.theme.BackgroundColor.Blue()), float64(fl.theme.BackgroundColor.Alpha()))
		cr.Rectangle(0, float64(y), 1200, float64(rowHeight))
		cr.Fill()
	}

	iconSize := 24
	iconName := fileops.GetIconForFile(item.Name)
	if item.IsDir {
		iconName = fileops.GetIconForFolder(item.Path)
	}

	paintable := fl.iconTheme.LookupIcon(iconName, nil, iconSize, 1, gtk.TextDirNone, 0)
	if paintable != nil {
		file := paintable.File()
		if file != nil {
			path := file.Path()
			if path != "" {
				texture, err := gdk.NewTextureFromFile(file)
				if err == nil {
					pixbuf := gdk.PixbufGetFromTexture(texture)
					if pixbuf != nil {
						gdk.CairoSetSourcePixbuf(cr, pixbuf, 8, float64(y+(rowHeight-iconSize)/2))
						cr.Paint()
					}
				}
			}
		}
	}

	if idx == fl.SelectedIDX && fl.canSelect {
		cr.SetSourceRGBA(float64(fl.theme.SelectedTextColor.Red()), float64(fl.theme.SelectedTextColor.Green()), float64(fl.theme.SelectedTextColor.Blue()), float64(fl.theme.SelectedTextColor.Alpha()))
	} else {
		cr.SetSourceRGBA(float64(fl.theme.TextColor.Red()), float64(fl.theme.TextColor.Green()), float64(fl.theme.TextColor.Blue()), float64(fl.theme.TextColor.Alpha()))
	}
	cr.SelectFontFace("Sans", cairo.FontSlantNormal, cairo.FontWeightBold)
	cr.SetFontSize(14)
	cr.MoveTo(40, float64(y+23))
	cr.ShowText(item.Name)

	if idx == fl.SelectedIDX && fl.canSelect {
		cr.SetSourceRGBA(float64(fl.theme.SelectedTextColor.Red()), float64(fl.theme.SelectedTextColor.Green()), float64(fl.theme.SelectedTextColor.Blue()), float64(fl.theme.SelectedTextColor.Alpha()))
	} else {
		cr.SetSourceRGBA(float64(fl.theme.TextColor.Red()), float64(fl.theme.TextColor.Green()), float64(fl.theme.TextColor.Blue()), float64(fl.theme.TextColor.Alpha()))
	}
	cr.SelectFontFace("Sans", cairo.FontSlantNormal, cairo.FontWeightNormal)
	cr.SetFontSize(11)
	cr.MoveTo(520, float64(y+20))
	if item.IsDir {
		cr.ShowText(fmt.Sprintf("%d item", item.ItemCount))
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
			return pos, pos + rowHeight
		}
		pos += rowHeight
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
			fl.SelectedIDX = idx
			fl.SelectionChanged(fl.SelectedIDX)
			da.QueueDraw()

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
			print("Delete clicked")
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

		if y >= pos && y < pos+rowHeight {
			return i
		}
		pos += rowHeight
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

func getFileSizeAsString(size int) string {
	if size < 1024 {
		return fmt.Sprintf("%d bytes", size)
	} else if size < 1024*1024 {
		return fmt.Sprintf("%.2f KB", float64(size)/1024)
	} else if size < 1024*1024*1024 {
		return fmt.Sprintf("%.2f MB", float64(size)/(1024*1024))
	} else {
		return fmt.Sprintf("%.2f GB", float64(size)/(1024*1024*1024))
	}
}
