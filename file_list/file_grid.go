package file_list

import (
	"fmt"
	"os/exec"
	"path/filepath"
	"slices"
	"strings"

	"github.com/MrSametBurgazoglu/atilgan/fileops"
	"github.com/MrSametBurgazoglu/atilgan/preferences"
	"github.com/MrSametBurgazoglu/atilgan/rename_popup"
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
	categoryHeaderWidth = 20
	categoryPadding     = 8
	categoryGap         = 10
)

type Category struct {
	Name     string
	Items    []*types.ListItem
	ItemIdxs []int // Original indices in fl.Items
	Width    int
	Height   int
	X, Y     int
	NumRows  int
	NumCols  int
}

type FileGrid struct {
	*gtk.ScrolledWindow
	Items              []*types.ListItem
	SelectedIDX        int
	SelectedIdxs       map[int]bool
	HoverIDX           int
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
	itemWidth  int
	itemHeight int
	iconSize   int

	SelectionChanged func(index int)
	PathChanged      func(path string)
	PinRequested     func(path string)
	KeyRightPressed  func()
	KeyLeftPressed   func()
}

func (fl *FileGrid) updateDimensions() {
	fl.iconSize = fl.Config.IconSize
	fl.itemWidth = fl.iconSize + 58
	fl.itemHeight = fl.iconSize + 63
}

func NewFileGrid(canSelect bool, specialPathManager *special_path.SpecialPathManager, parent *gtk.Window, config *preferences.Config) *FileGrid {
	fl := &FileGrid{
		ScrolledWindow:     gtk.NewScrolledWindow(),
		SelectedIDX:        0,
		SelectedIdxs:       make(map[int]bool),
		HoverIDX:           -1,
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
			// Check for Shift + Key (for jumping to file by letter)
			if state&gdk.ShiftMask != 0 {
				if keyval >= gdk.KEY_A && keyval <= gdk.KEY_Z {
					fl.SetSelectedItemWithLetter(string(rune(keyval)))
					return true
				}
				if keyval >= gdk.KEY_a && keyval <= gdk.KEY_z {
					fl.SetSelectedItemWithLetter(string(rune(keyval)))
					return true
				}
			}

			switch keyval {
			case gdk.KEY_Up:
				if state&gdk.ShiftMask != 0 {
					// Handle Shift+Up for range selection is more complex in navigateUp
					// For now let's just do a simple expansion
					fl.navigateUp()
					fl.SelectedIdxs[fl.SelectedIDX] = true
				} else {
					fl.navigateUp()
				}
				return true

			case gdk.KEY_Down:
				if state&gdk.ShiftMask != 0 {
					fl.navigateDown()
					fl.SelectedIdxs[fl.SelectedIDX] = true
				} else {
					fl.navigateDown()
				}
				return true

			case gdk.KEY_Left:
				if fl.SelectedIDX > 0 {
					if state&gdk.ShiftMask != 0 {
						fl.SelectedIDX--
						fl.SelectedIdxs[fl.SelectedIDX] = true
						fl.ensureVisible()
						fl.DrawingArea.QueueDraw()
					} else {
						fl.selectItem(fl.SelectedIDX - 1)
					}
				}
				return true

			case gdk.KEY_Right:
				if fl.SelectedIDX < len(fl.Items)-1 {
					if state&gdk.ShiftMask != 0 {
						fl.SelectedIDX++
						fl.SelectedIdxs[fl.SelectedIDX] = true
						fl.ensureVisible()
						fl.DrawingArea.QueueDraw()
					} else {
						fl.selectItem(fl.SelectedIDX + 1)
					}
				}
				return true

			case gdk.KEY_BackSpace:
				if fl.KeyLeftPressed != nil {
					fl.KeyLeftPressed()
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
	}

	fl.DrawingArea.AddController(fl.newGestureClick(fl.DrawingArea))

	return fl
}

func (fl *FileGrid) SetItems(items []*types.ListItem) {
	fl.Items = items
	fl.SelectedIDX = 0
	fl.SelectedIdxs = make(map[int]bool)
	if len(items) > 0 {
		fl.SelectedIdxs[0] = true
	}
	fl.HoverIDX = -1
	fl.textureCache = make(map[string]*gdk.Texture)
	fl.DrawingArea.QueueDraw()
}

func (fl *FileGrid) SetCanFocus(canFocus bool) {
	fl.CanFocus = canFocus
}

func (fl *FileGrid) SetKeyLeftPressed(f func()) {
	fl.KeyLeftPressed = f
}

func (fl *FileGrid) SetKeyRightPressed(f func()) {
	fl.KeyRightPressed = f
}

func (fl *FileGrid) SetPinRequested(f func(string)) {
	fl.PinRequested = f
}

func (fl *FileGrid) FocusDrawingArea() {
	fl.DrawingArea.GrabFocus()
}

func (fl *FileGrid) GetItems() []*types.ListItem {
	return fl.Items
}

func (fl *FileGrid) GetSelectedIDX() int {
	return fl.SelectedIDX
}

func (fl *FileGrid) GetSelectedIdxs() map[int]bool {
	return fl.SelectedIdxs
}

func (fl *FileGrid) AddItem(item *types.ListItem) {
	fl.Items = append(fl.Items, item)
	fl.DrawingArea.QueueDraw()
}

func (fl *FileGrid) SetSelectedItemWithLetter(letter string) {
	for i, item := range fl.Items {
		if strings.HasPrefix(strings.ToLower(item.Name), strings.ToLower(letter)) {
			fl.SetItem(i)
			break
		}
	}
}

func (fl *FileGrid) SetItem(index int) {
	if index >= 0 && index < len(fl.Items) {
		fl.SelectedIDX = index
		fl.SelectedIdxs = make(map[int]bool)
		fl.SelectedIdxs[index] = true
		fl.DrawingArea.QueueDraw()
		if fl.SelectionChanged != nil {
			fl.SelectionChanged(fl.SelectedIDX)
		}
	}
}

func (fl *FileGrid) SelectAll() {
	for i := range fl.Items {
		fl.SelectedIdxs[i] = true
	}
	fl.DrawingArea.QueueDraw()
	if fl.SelectionChanged != nil {
		fl.SelectionChanged(fl.SelectedIDX)
	}
}

func (fl *FileGrid) ClearSelection() {
	fl.SelectedIdxs = make(map[int]bool)
	if len(fl.Items) > 0 {
		fl.SelectedIdxs[fl.SelectedIDX] = true
	}
	fl.DrawingArea.QueueDraw()
	if fl.SelectionChanged != nil {
		fl.SelectionChanged(fl.SelectedIDX)
	}
}

func (fl *FileGrid) AddCopyCutItem(path string) bool {
	if slices.Contains(fl.CopyCutPaths, path) {
		return false
	}
	fl.CopyCutPaths = append(fl.CopyCutPaths, path)
	fl.DrawingArea.QueueDraw()
	return true
}

func (fl *FileGrid) CleanCopyCutFiles() {
	fl.CopyCutPaths = make([]string, 0)
	fl.DrawingArea.QueueDraw()
}

func (fl *FileGrid) Refresh(newFilter bool) {
	fl.updateDimensions()
	fl.DrawingArea.QueueDraw()
}

func (fl *FileGrid) SetPathChanged(f func(string)) {
	fl.PathChanged = f
}

func (fl *FileGrid) SetSelectionChanged(f func(int)) {
	fl.SelectionChanged = f
}

func (fl *FileGrid) onDraw(da *gtk.DrawingArea, cr *cairo.Context, w, h int) {
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
		fl.colorTheme.SelectedBgColor = gdk.NewRGBA(accent.Red(), accent.Green(), accent.Blue(), 0.15)
		fl.colorTheme.HoverBgColor = gdk.NewRGBA(accent.Red(), accent.Green(), accent.Blue(), 0.05)
		fl.colorTheme.SelectedTextColor = *accent
	}

	if headerBg, ok := sc.LookupColor("headerbar_bg_color"); ok {
		fl.colorTheme.HeaderBackgroundColor = gdk.NewRGBA(headerBg.Red(), headerBg.Green(), headerBg.Blue(), 0.3)
	}

	categories, totalHeight := fl.layoutCategories(w)

	cr.SetSourceRGBA(float64(fl.colorTheme.BackgroundColor.Red()), float64(fl.colorTheme.BackgroundColor.Green()), float64(fl.colorTheme.BackgroundColor.Blue()), float64(fl.colorTheme.BackgroundColor.Alpha()))
	cr.Rectangle(0, 0, float64(w), float64(totalHeight))
	cr.Fill()

	for _, cat := range categories {
		// Draw Category Card Background
		cr.SetSourceRGBA(float64(fl.colorTheme.HeaderBackgroundColor.Red()), float64(fl.colorTheme.HeaderBackgroundColor.Green()), float64(fl.colorTheme.HeaderBackgroundColor.Blue()), 0.3)
		roundedRectangle(cr, float64(cat.X), float64(cat.Y), float64(cat.Width), float64(cat.Height), 12)
		cr.Fill()

		// Draw Header
		fl.drawHeader(cr, cat.Name, cat.X, cat.Y, cat.Height)
		
		for i, item := range cat.Items {
			row := i / cat.NumCols
			col := i % cat.NumCols
			
			itemX := cat.X + categoryHeaderWidth + categoryPadding + col*fl.itemWidth
			itemY := cat.Y + categoryPadding + row*fl.itemHeight
			
			fl.drawGridItem(cr, cat.ItemIdxs[i], item, itemX, itemY)
		}
	}
}

func (fl *FileGrid) layoutCategories(w int) ([]Category, int) {
	if len(fl.Items) == 0 {
		return nil, 0
	}

	// 1. Group items into categories
	var categories []Category
	var currentCat *Category
	for i, item := range fl.Items {
		if currentCat == nil || item.Group != currentCat.Name {
			categories = append(categories, Category{Name: item.Group})
			currentCat = &categories[len(categories)-1]
		}
		currentCat.Items = append(currentCat.Items, item)
		currentCat.ItemIdxs = append(currentCat.ItemIdxs, i)
	}

	// 2. Calculate dimensions for each category
	for i := range categories {
		cat := &categories[i]
		numItems := len(cat.Items)
		
		availableItemsWidth := w - categoryHeaderWidth - 2*categoryPadding - categoryGap
		if availableItemsWidth < fl.itemWidth {
			availableItemsWidth = fl.itemWidth
		}

		cat.NumCols = availableItemsWidth / fl.itemWidth
		if cat.NumCols > numItems {
			cat.NumCols = numItems
		}
		if cat.NumCols < 1 {
			cat.NumCols = 1
		}

		cat.NumRows = (numItems + cat.NumCols - 1) / cat.NumCols
		cat.Width = categoryHeaderWidth + cat.NumCols*fl.itemWidth + 2*categoryPadding
		cat.Height = cat.NumRows*fl.itemHeight + 2*categoryPadding
		if cat.Height < 60 {
			cat.Height = 60
		}
	}

	// 3. Position categories (Side-by-Side with row justification)
	currentY := 0
	var rows [][]int // Indices of categories in each row
	var currentRow []int
	currentX := 0

	for i := range categories {
		cat := &categories[i]
		if currentX > 0 && currentX+cat.Width > w {
			rows = append(rows, currentRow)
			currentRow = []int{i}
			currentX = cat.Width + categoryGap
		} else {
			currentRow = append(currentRow, i)
			currentX += cat.Width + categoryGap
		}
	}
	if len(currentRow) > 0 {
		rows = append(rows, currentRow)
	}

	for _, rowIdxs := range rows {
		totalCatsWidth := 0
		maxHeight := 0
		for _, idx := range rowIdxs {
			totalCatsWidth += categories[idx].Width
			if categories[idx].Height > maxHeight {
				maxHeight = categories[idx].Height
			}
		}

		// Calculate gap for this row including both sides
		rowGap := (w - totalCatsWidth) / (len(rowIdxs) + 1)
		if rowGap < 0 {
			rowGap = 0
		}

		currentX := rowGap
		for _, idx := range rowIdxs {
			cat := &categories[idx]
			cat.X = currentX
			cat.Y = currentY
			currentX += cat.Width + rowGap
		}
		currentY += maxHeight + categoryGap
	}

	totalHeight := currentY
	fl.DrawingArea.SetContentHeight(totalHeight)
	
	return categories, totalHeight
}

func (fl *FileGrid) drawGridItem(cr *cairo.Context, idx int, item *types.ListItem, x, y int) {
	padding := 6.0
	isSelected := fl.SelectedIdxs[idx] && fl.canSelect
	isFocused := idx == fl.SelectedIDX && fl.canSelect
	isHovered := idx == fl.HoverIDX && fl.canSelect

	if isSelected {
		cr.SetSourceRGBA(float64(fl.colorTheme.SelectedBgColor.Red()), float64(fl.colorTheme.SelectedBgColor.Green()), float64(fl.colorTheme.SelectedBgColor.Blue()), float64(fl.colorTheme.SelectedBgColor.Alpha()*1.5))
		roundedRectangle(cr, float64(x)+padding, float64(y)+padding, float64(fl.itemWidth)-2*padding, float64(fl.itemHeight)-2*padding, 10)
		cr.Fill()

		if isFocused {
			cr.SetSourceRGBA(float64(fl.colorTheme.AccentColor.Red()), float64(fl.colorTheme.AccentColor.Green()), float64(fl.colorTheme.AccentColor.Blue()), 0.8)
			cr.SetLineWidth(2)
			roundedRectangle(cr, float64(x)+padding, float64(y)+padding, float64(fl.itemWidth)-2*padding, float64(fl.itemHeight)-2*padding, 10)
			cr.Stroke()
		}
	} else if isHovered {
		cr.SetSourceRGBA(float64(fl.colorTheme.HoverBgColor.Red()), float64(fl.colorTheme.HoverBgColor.Green()), float64(fl.colorTheme.HoverBgColor.Blue()), 0.3)
		roundedRectangle(cr, float64(x)+padding, float64(y)+padding, float64(fl.itemWidth)-2*padding, float64(fl.itemHeight)-2*padding, 10)
		cr.Fill()
	} else if slices.Contains(fl.CopyCutPaths, item.Path) {
		cr.SetSourceRGBA(float64(fl.colorTheme.CopyCutBgColor.Red()), float64(fl.colorTheme.CopyCutBgColor.Green()), float64(fl.colorTheme.CopyCutBgColor.Blue()), 0.4)
		roundedRectangle(cr, float64(x)+padding, float64(y)+padding, float64(fl.itemWidth)-2*padding, float64(fl.itemHeight)-2*padding, 10)
		cr.Fill()
	}

	// Icon positioning
	var texture *gdk.Texture
	if !item.IsDir && fl.Config.EnableThumbnails {
		if cached, ok := fl.textureCache[item.Path]; ok {
			texture = cached
		} else {
			texture, _ = thumbnail.Generate(item.Path)
			fl.textureCache[item.Path] = texture
		}
	}

	iconY := float64(y) + 12
	if texture != nil {
		pixbuf := gdk.PixbufGetFromTexture(texture)
		if pixbuf != nil {
			w, h := pixbuf.Width(), pixbuf.Height()
			scale := float64(fl.iconSize) / float64(max(w, h))
			newW, newH := int(float64(w)*scale), int(float64(h)*scale)
			pixbuf = pixbuf.ScaleSimple(newW, newH, gdkpixbuf.InterpBilinear)
			gdk.CairoSetSourcePixbuf(cr, pixbuf, float64(x+(fl.itemWidth-newW)/2), iconY+float64(fl.iconSize-newH)/2)
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
				tex, err := gdk.NewTextureFromFile(file)
				if err == nil {
					pixbuf := gdk.PixbufGetFromTexture(tex)
					if pixbuf != nil {
						if pixbuf.Width() != fl.iconSize || pixbuf.Height() != fl.iconSize {
							pixbuf = pixbuf.ScaleSimple(fl.iconSize, fl.iconSize, gdkpixbuf.InterpBilinear)
						}
						gdk.CairoSetSourcePixbuf(cr, pixbuf, float64(x+(fl.itemWidth-fl.iconSize)/2), iconY)
						cr.Paint()
					}
				}
			}
		}
	}

	// Filename
	if isSelected {
		cr.SetSourceRGBA(float64(fl.colorTheme.SelectedTextColor.Red()), float64(fl.colorTheme.SelectedTextColor.Green()), float64(fl.colorTheme.SelectedTextColor.Blue()), 1.0)
	} else {
		cr.SetSourceRGBA(float64(fl.colorTheme.TextColor.Red()), float64(fl.colorTheme.TextColor.Green()), float64(fl.colorTheme.TextColor.Blue()), 1.0)
	}
	
	cr.SelectFontFace("Sans", cairo.FontSlantNormal, cairo.FontWeightNormal)
	cr.SetFontSize(fl.themeConfig.Fonts.FilenameSize)
	
	displayName := fl.truncateText(cr, item.Name, float64(fl.itemWidth)-16)
	
	extents := cr.TextExtents(displayName)
	cr.MoveTo(float64(x)+(float64(fl.itemWidth)-extents.Width)/2, iconY+float64(fl.iconSize)+20)
	cr.ShowText(displayName)

	// Git Status Indicator
	if item.GitStatus != "" && item.GitStatus != "unchanged" {
		dotSize := 4.0
		dotX := float64(x) + (float64(fl.itemWidth)+extents.Width)/2 + 4
		dotY := iconY + float64(fl.iconSize) + 20 - extents.Height/2
		
		switch item.GitStatus {
		case "modified":
			cr.SetSourceRGBA(0.9, 0.6, 0.1, 1.0) // Orange
		case "added":
			cr.SetSourceRGBA(0.2, 0.7, 0.2, 1.0) // Green
		case "untracked":
			cr.SetSourceRGBA(0.5, 0.5, 0.5, 1.0) // Grey
		default:
			cr.SetSourceRGBA(0.1, 0.5, 0.8, 1.0) // Blue for others
		}
		
		cr.Arc(dotX, dotY, dotSize, 0, 2*3.14159)
		cr.Fill()
	}

	// Metadata
	cr.SetSourceRGBA(float64(fl.colorTheme.TextColor.Red()), float64(fl.colorTheme.TextColor.Green()), float64(fl.colorTheme.TextColor.Blue()), 0.5)
	cr.SetFontSize(fl.themeConfig.Fonts.SizeTextSize)
	var meta string
	if item.IsDir {
		itemText := "items"
		if item.ItemCount == 1 {
			itemText = "item"
		}
		meta = fmt.Sprintf("%d %s", item.ItemCount, itemText)
	} else {
		meta = fileops.GetFileSizeAsString(item.Size)
	}
	extents = cr.TextExtents(meta)
	cr.MoveTo(float64(x)+(float64(fl.itemWidth)-extents.Width)/2, iconY+float64(fl.iconSize)+35)
	cr.ShowText(meta)
}

func (fl *FileGrid) truncateText(cr *cairo.Context, text string, maxWidth float64) string {
	extents := cr.TextExtents(text)
	if extents.Width <= maxWidth {
		return text
	}

	for len(text) > 0 {
		text = text[:len(text)-1]
		if cr.TextExtents(text+"...").Width <= maxWidth {
			return text + "..."
		}
	}
	return "..."
}

func (fl *FileGrid) drawHeader(cr *cairo.Context, text string, x, y, h int) {
	cr.SetSourceRGBA(float64(fl.colorTheme.AccentColor.Red()), float64(fl.colorTheme.AccentColor.Green()), float64(fl.colorTheme.AccentColor.Blue()), 0.7)
	cr.SelectFontFace("Sans", cairo.FontSlantNormal, cairo.FontWeightBold)
	cr.SetFontSize(fl.themeConfig.Fonts.HeaderSize - 1)
	
	displayName := strings.ToUpper(text)
	runes := []rune(displayName)
	fontSize := fl.themeConfig.Fonts.HeaderSize - 1
	lineHeight := fontSize + 2
	totalTextHeight := float64(len(runes)) * lineHeight
	
	startY := float64(y) + (float64(h)-totalTextHeight)/2 + fontSize
	
	for i, r := range runes {
		char := string(r)
		extents := cr.TextExtents(char)
		
		// Center character horizontally in the header width
		charX := float64(x) + (float64(categoryHeaderWidth)-extents.Width)/2 - extents.XBearing
		charY := startY + float64(i)*lineHeight
		
		cr.MoveTo(charX, charY)
		cr.ShowText(char)
	}
	
	// Subtle separator line
	cr.SetSourceRGBA(float64(fl.colorTheme.AccentColor.Red()), float64(fl.colorTheme.AccentColor.Green()), float64(fl.colorTheme.AccentColor.Blue()), 0.1)
	cr.SetLineWidth(1)
	cr.MoveTo(float64(x+categoryHeaderWidth), float64(y)+10)
	cr.LineTo(float64(x+categoryHeaderWidth), float64(y+h)-10)
	cr.Stroke()
}

func (fl *FileGrid) navigateUp() {
	w := fl.DrawingArea.Width()
	if w <= 0 {
		return
	}
	categories, _ := fl.layoutCategories(w)
	
	for _, cat := range categories {
		for i, idx := range cat.ItemIdxs {
			if idx == fl.SelectedIDX {
				col := i % cat.NumCols
				row := i / cat.NumCols
				
				if row > 0 {
					fl.selectItem(cat.ItemIdxs[(row-1)*cat.NumCols+col])
					return
				} else {
					if fl.SelectedIDX > 0 {
						fl.selectItem(fl.SelectedIDX - 1)
					}
					return
				}
			}
		}
	}
}

func (fl *FileGrid) navigateDown() {
	w := fl.DrawingArea.Width()
	if w <= 0 {
		return
	}
	categories, _ := fl.layoutCategories(w)
	
	for _, cat := range categories {
		for i, idx := range cat.ItemIdxs {
			if idx == fl.SelectedIDX {
				col := i % cat.NumCols
				row := i / cat.NumCols
				
				if row < cat.NumRows-1 {
					nextIdx := (row+1)*cat.NumCols+col
					if nextIdx < len(cat.ItemIdxs) {
						fl.selectItem(cat.ItemIdxs[nextIdx])
					} else {
						fl.selectItem(cat.ItemIdxs[len(cat.ItemIdxs)-1])
					}
					return
				} else {
					if fl.SelectedIDX < len(fl.Items)-1 {
						fl.selectItem(fl.SelectedIDX + 1)
					}
					return
				}
			}
		}
	}
}

func (fl *FileGrid) selectItem(idx int) {
	if idx < 0 || idx >= len(fl.Items) {
		return
	}
	fl.SelectedIDX = idx
	fl.SelectedIdxs = make(map[int]bool)
	fl.SelectedIdxs[idx] = true
	fl.ensureVisible()
	fl.DrawingArea.QueueDraw()
	if fl.SelectionChanged != nil {
		fl.SelectionChanged(fl.SelectedIDX)
	}
}

func (fl *FileGrid) ensureVisible() {
	adj := fl.VAdjustment()
	scrollPos := adj.Value()
	visibleHeight := float64(fl.Height())
	w := fl.DrawingArea.Width()
	if w <= 0 {
		return
	}

	itemTop, itemBottom := fl.getItemBounds(fl.SelectedIDX, w)

	if float64(itemTop) >= scrollPos && float64(itemBottom) <= scrollPos+visibleHeight {
		return
	}

	if float64(itemTop) < scrollPos {
		adj.SetValue(float64(itemTop))
	} else {
		adj.SetValue(float64(itemBottom) - visibleHeight)
	}
}

func (fl *FileGrid) getItemBounds(idx int, w int) (top, bottom int) {
	if idx < 0 || idx >= len(fl.Items) {
		return 0, 0
	}

	categories, _ := fl.layoutCategories(w)
	for _, cat := range categories {
		for i, itemIdx := range cat.ItemIdxs {
			if itemIdx == idx {
				row := i / cat.NumCols
				itemY := cat.Y + categoryPadding + row*fl.itemHeight
				return itemY, itemY + fl.itemHeight
			}
		}
	}
	
	return 0, 0
}

func (fl *FileGrid) ItemAt(x, y int) int {
	w := fl.DrawingArea.Width()
	if w <= 0 {
		return -1
	}

	categories, _ := fl.layoutCategories(w)
	for _, cat := range categories {
		if x >= cat.X && x < cat.X+cat.Width && y >= cat.Y && y < cat.Y+cat.Height {
			relX := x - cat.X - categoryHeaderWidth - categoryPadding
			relY := y - cat.Y - categoryPadding
			
			if relX < 0 || relY < 0 {
				continue
			}
			
			col := relX / fl.itemWidth
			row := relY / fl.itemHeight
			
			if col >= cat.NumCols {
				continue
			}

			idx := row*cat.NumCols + col
			if idx < len(cat.ItemIdxs) {
				return cat.ItemIdxs[idx]
			}
		}
	}
	
	return -1
}

func (fl *FileGrid) newGestureClick(da *gtk.DrawingArea) *gtk.GestureClick {
	click := gtk.NewGestureClick()
	click.ConnectPressed(func(n int, x, y float64) {
		idx := fl.ItemAt(int(x), int(y))
		if idx >= 0 {
			state := click.CurrentEventState()
			
			if state&gdk.ControlMask != 0 {
				// Toggle selection
				fl.SelectedIDX = idx
				fl.SelectedIdxs[idx] = !fl.SelectedIdxs[idx]
				if fl.SelectionChanged != nil {
					fl.SelectionChanged(fl.SelectedIDX)
				}
			} else if state&gdk.ShiftMask != 0 {
				// Range selection
				start := fl.SelectedIDX
				end := idx
				if start > end {
					start, end = end, start
				}
				for i := start; i <= end; i++ {
					fl.SelectedIdxs[i] = true
				}
				fl.SelectedIDX = idx
				if fl.SelectionChanged != nil {
					fl.SelectionChanged(fl.SelectedIDX)
				}
			} else {
				// Single selection
				fl.selectItem(idx)
			}

			if click.CurrentButton() == gdk.BUTTON_PRIMARY && n == 2 {
				if !fl.Items[idx].IsDir {
					cmd := exec.Command("xdg-open", fl.Items[idx].Path)
					cmd.Start()
				} else {
					if fl.PathChanged != nil {
						fl.PathChanged(fl.Items[idx].Path)
					}
				}
			}
			fl.DrawingArea.QueueDraw()
		} else {
			// Clicked on empty space
			fl.ClearSelection()
		}
	})
	return click
}

func (fl *FileGrid) newContextMenuController(da *gtk.DrawingArea) *gtk.GestureClick {
	click := gtk.NewGestureClick()
	click.SetButton(gdk.BUTTON_SECONDARY)
	click.ConnectPressed(func(n int, x, y float64) {
		idx := fl.ItemAt(int(x), int(y))

		pop := gtk.NewPopover()
		popoverBox := gtk.NewBox(gtk.OrientationVertical, 6)
		pop.SetChild(popoverBox)

		if idx >= 0 {
			if !fl.SelectedIdxs[idx] {
				fl.selectItem(idx)
			}

			open := gtk.NewButtonWithLabel("Open")
			open.Connect("clicked", func() {
				if !fl.Items[idx].IsDir {
					cmd := exec.Command("xdg-open", fl.Items[idx].Path)
					cmd.Start()
				} else {
					if fl.PathChanged != nil {
						fl.PathChanged(fl.Items[idx].Path)
					}
					pop.Popdown()
				}
			})

			delete := gtk.NewButtonWithLabel("Delete Selected")
			delete.Connect("clicked", func() {
				performDelete := func() {
					for i, selected := range fl.SelectedIdxs {
						if selected {
							cmd := exec.Command("gio", "trash", fl.Items[i].Path)
							cmd.Run()
						}
					}
					glib.IdleAdd(func() {
						if fl.PathChanged != nil {
							fl.PathChanged("")
						}
						pop.Popdown()
					})
				}

				if fl.Config.ConfirmDelete {
					dialog := adw.NewMessageDialog(fl.parent, "Confirm Deletion", "Are you sure you want to move the selected items to trash?")
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

			// Bulk Rename option
			selectedCount := 0
			for _, selected := range fl.SelectedIdxs {
				if selected {
					selectedCount++
				}
			}
			if selectedCount > 1 {
				bulkRename := gtk.NewButtonWithLabel("Bulk Rename")
				bulkRename.Connect("clicked", func() {
					var paths []string
					for i, selected := range fl.SelectedIdxs {
						if selected {
							paths = append(paths, fl.Items[i].Path)
						}
					}
					dir := filepath.Dir(paths[0])
					bulkWin := rename_popup.NewBulkRenameWindow(dir, paths, fl.PathChanged)
					bulkWin.SetTransientFor(fl.parent)
					bulkWin.SetVisible(true)
					pop.Popdown()
				})
				popoverBox.Append(bulkRename)
			}

			// Archive options
			isArchive := false
			ext := strings.ToLower(filepath.Ext(fl.Items[idx].Path))
			if ext == ".zip" || ext == ".tar" || ext == ".gz" || ext == ".xz" || ext == ".7z" {
				isArchive = true
			}

			if isArchive {
				extract := gtk.NewButtonWithLabel("Extract Here")
				extract.Connect("clicked", func() {
					dir := filepath.Dir(fl.Items[idx].Path)
					var cmd *exec.Cmd
					if ext == ".zip" {
						cmd = exec.Command("unzip", fl.Items[idx].Path, "-d", dir)
					} else {
						cmd = exec.Command("tar", "-xf", fl.Items[idx].Path, "-C", dir)
					}
					go func() {
						cmd.Run()
						glib.IdleAdd(func() {
							fl.PathChanged("")
						})
					}()
					pop.Popdown()
				})
				popoverBox.Append(extract)
			}

			compress := gtk.NewButtonWithLabel("Compress to .zip")
			compress.Connect("clicked", func() {
				var paths []string
				for i, selected := range fl.SelectedIdxs {
					if selected {
						paths = append(paths, fl.Items[i].Path)
					}
				}
				zipName := filepath.Base(paths[0]) + ".zip"
				zipPath := filepath.Join(filepath.Dir(paths[0]), zipName)
				
				args := []string{"-r", zipPath}
				for _, p := range paths {
					args = append(args, filepath.Base(p))
				}
				
				cmd := exec.Command("zip", args...)
				cmd.Dir = filepath.Dir(paths[0])
				go func() {
					cmd.Run()
					glib.IdleAdd(func() {
						fl.PathChanged("")
					})
				}()
				pop.Popdown()
			})
			popoverBox.Append(compress)

			properties := gtk.NewButtonWithLabel("Properties")
			properties.Connect("clicked", func() {
				propWin := fileops.NewPropertiesWindow(fl.Items[idx].Path)
				propWin.SetTransientFor(fl.parent)
				propWin.SetVisible(true)
				pop.Popdown()
			})
			popoverBox.Append(properties)

			if fl.Items[idx].IsDir {
				pin := gtk.NewButtonWithLabel("Pin to Sidebar")
				pin.Connect("clicked", func() {
					if fl.PinRequested != nil {
						fl.PinRequested(fl.Items[idx].Path)
					}
					pop.Popdown()
				})
				popoverBox.Append(pin)
			}

			popoverBox.Append(gtk.NewSeparator(gtk.OrientationHorizontal))

		}

		selectAll := gtk.NewButtonWithLabel("Select All")
		selectAll.Connect("clicked", func() {
			fl.SelectAll()
			pop.Popdown()
		})
		popoverBox.Append(selectAll)

		pop.SetHasArrow(true)
		rect := gdk.NewRectangle(int(x), int(y), 1, 1)
		pop.SetPointingTo(&rect)

		pop.SetParent(da)
		pop.Popup()
	})
	return click
}

func (fl *FileGrid) newMouseController(da *gtk.DrawingArea) *gtk.EventControllerMotion {
	ctrl := gtk.NewEventControllerMotion()
	ctrl.ConnectMotion(func(x, y float64) {
		if fl.CanFocus {
			fl.DrawingArea.GrabFocus()
		}
		
		newHover := fl.ItemAt(int(x), int(y))
		if newHover != fl.HoverIDX {
			fl.HoverIDX = newHover
			fl.DrawingArea.QueueDraw()
		}
	})
	ctrl.ConnectLeave(func() {
		if fl.HoverIDX != -1 {
			fl.HoverIDX = -1
			fl.DrawingArea.QueueDraw()
		}
	})
	return ctrl
}

func max(a, b int) int {
	if a > b {
		return a
	}
	return b
}
