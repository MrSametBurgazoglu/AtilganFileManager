package viewer

import (
	"fmt"
	"os"
	"path"
	"sort"

	"github.com/MrSametBurgazoglu/atilgan/fileops"
	"github.com/MrSametBurgazoglu/atilgan/preferences"
	"github.com/MrSametBurgazoglu/atilgan/special_path"
	"github.com/MrSametBurgazoglu/atilgan/thumbnail"
	"github.com/MrSametBurgazoglu/atilgan/types"
	"github.com/diamondburned/gotk4-adwaita/pkg/adw"
	"github.com/diamondburned/gotk4/pkg/gtk/v4"
	"github.com/diamondburned/gotk4/pkg/pango"
)

type PictureViewer struct {
	*gtk.Box
	Path               string
	specialPathManager *special_path.SpecialPathManager
	mainWindow         *gtk.Window
	pathChanged        func(string)
	stack              *gtk.Stack

	carousel *adw.Carousel
	flowBox  *gtk.FlowBox

	// Selection state for main.go
	Items            []*types.ListItem
	SelectedIDX      int
	SelectedIdxs     map[int]bool
	SelectionChanged func(int)

	// Internal state mapping GTK widgets to list indices
	widgetToIndex map[gtk.Widgetter]int
}

func NewPictureViewer(mainWindow *gtk.Window, path string, pathChanged func(string), specialPathManager *special_path.SpecialPathManager, config *preferences.Config) *PictureViewer {
	viewer := &PictureViewer{
		Box:                gtk.NewBox(gtk.OrientationVertical, 12),
		Path:               path,
		specialPathManager: specialPathManager,
		mainWindow:         mainWindow,
		pathChanged:        pathChanged,
		SelectedIdxs:       make(map[int]bool),
		widgetToIndex:      make(map[gtk.Widgetter]int),
	}
	viewer.SetVExpand(true)
	viewer.SetHExpand(true)
	viewer.SetMarginTop(12)
	viewer.SetMarginBottom(12)
	viewer.SetMarginStart(12)
	viewer.SetMarginEnd(12)

	viewer.stack = gtk.NewStack()
	viewer.stack.SetTransitionType(gtk.StackTransitionTypeCrossfade)

	// Main content box
	contentBox := gtk.NewBox(gtk.OrientationVertical, 12)
	contentBox.SetVExpand(true)
	contentBox.SetHExpand(true)

	// Carousel for highlights
	viewer.carousel = adw.NewCarousel()
	viewer.carousel.SetHExpand(true)
	viewer.carousel.SetSizeRequest(-1, 250) // Fixed height for highlights

	// FlowBox for everything else
	scrolledWindow := gtk.NewScrolledWindow()
	scrolledWindow.SetPolicy(gtk.PolicyAutomatic, gtk.PolicyAutomatic)
	scrolledWindow.SetVExpand(true)
	scrolledWindow.SetHExpand(true)

	viewer.flowBox = gtk.NewFlowBox()
	viewer.flowBox.SetVAlign(gtk.AlignStart)
	viewer.flowBox.SetMaxChildrenPerLine(10)
	viewer.flowBox.SetSelectionMode(gtk.SelectionMultiple)
	viewer.flowBox.SetColumnSpacing(12)
	viewer.flowBox.SetRowSpacing(12)

	scrolledWindow.SetChild(viewer.flowBox)

	contentBox.Append(viewer.carousel)
	contentBox.Append(scrolledWindow)

	viewer.stack.AddTitled(contentBox, "list", "List")

	emptyLabel := gtk.NewLabel("No Pictures Found")
	emptyLabel.AddCSSClass("preview-title")
	viewer.stack.AddTitled(emptyLabel, "empty", "Empty")

	viewer.Box.Append(viewer.stack)

	// Handle FlowBox selection
	viewer.flowBox.ConnectSelectedChildrenChanged(func() {
		selectedChildren := viewer.flowBox.SelectedChildren()
		viewer.SelectedIdxs = make(map[int]bool)
		viewer.SelectedIDX = -1

		if len(selectedChildren) > 0 {
			// Find the last selected item to act as the primary SelectedIDX
			for _, child := range selectedChildren {
				if idx, ok := viewer.widgetToIndex[child]; ok {
					viewer.SelectedIdxs[idx] = true
					viewer.SelectedIDX = idx
				}
			}
		}

		if viewer.SelectionChanged != nil {
			viewer.SelectionChanged(viewer.SelectedIDX)
		}
	})

	viewer.Refresh()

	return viewer
}

func (viewer *PictureViewer) SetPath(path string) {
	viewer.Path = path
	viewer.Refresh()
}

func (viewer *PictureViewer) Refresh() {
	if viewer.Path == "" {
		return
	}

	entries, err := os.ReadDir(viewer.Path)
	if err != nil {
		fmt.Println("Error reading directory:", err)
		return
	}

	// Reset states
	viewer.flowBox.RemoveAll()
	viewer.Items = make([]*types.ListItem, 0)
	viewer.SelectedIdxs = make(map[int]bool)
	viewer.SelectedIDX = -1
	viewer.widgetToIndex = make(map[gtk.Widgetter]int)
	
	// Since there's no RemoveAll for Carousel in older gotk4 bindings,
	// we might need to iterate and remove if needed. Wait, gotk4-adwaita might not have RemoveAll.
	// We will create a new Carousel if needed, or rely on removing children.
	if viewer.carousel != nil && viewer.carousel.Parent() != nil {
		parent := viewer.carousel.Parent().(*gtk.Box)
		parent.Remove(viewer.carousel)
		viewer.carousel = adw.NewCarousel()
		viewer.carousel.SetHExpand(true)
		viewer.carousel.SetSizeRequest(-1, 250)
		parent.Prepend(viewer.carousel)
	}

	var directories []*types.ListItem
	var pictures []*types.ListItem

	for _, entry := range entries {
		if !entry.IsDir() && !fileops.IsImage(entry.Name()) {
			continue
		}

		fullPath := path.Join(viewer.Path, entry.Name())
		listItem := &types.ListItem{
			Name:  entry.Name(),
			Path:  fullPath,
			Group: "Pictures",
			IsDir: entry.IsDir(),
		}

		info, err := entry.Info()
		if err == nil {
			listItem.Size = info.Size()
		}

		if listItem.IsDir {
			listItem.Group = "Directories"
			listItem.ItemCount = getDirItemCount(fullPath)
			directories = append(directories, listItem)
		} else {
			pictures = append(pictures, listItem)
		}
	}

	// Sort pictures (e.g. by name, ideally by modification time for "recent")
	sort.Slice(pictures, func(i, j int) bool {
		infoI, _ := os.Stat(pictures[i].Path)
		infoJ, _ := os.Stat(pictures[j].Path)
		if infoI != nil && infoJ != nil {
			return infoI.ModTime().After(infoJ.ModTime())
		}
		return pictures[i].Name < pictures[j].Name
	})

	// Combine into Items list
	viewer.Items = append(viewer.Items, directories...)
	viewer.Items = append(viewer.Items, pictures...)

	if len(viewer.Items) == 0 {
		viewer.stack.SetVisibleChildName("empty")
		return
	}

	viewer.stack.SetVisibleChildName("list")

	// Add highlights to Carousel
	highlightCount := 5
	if len(pictures) < highlightCount {
		highlightCount = len(pictures)
	}

	for i := 0; i < highlightCount; i++ {
		picItem := pictures[i]
		widget := viewer.createHighlightWidget(picItem, len(directories)+i)
		viewer.carousel.Append(widget)
	}

	// Add all to FlowBox (directories first, then pictures)
	for i, item := range viewer.Items {
		var widget gtk.Widgetter
		if item.IsDir {
			widget = viewer.createAlbumWidget(item, i)
		} else {
			widget = viewer.createThumbWidget(item, i)
		}
		viewer.flowBox.Append(widget)
		flowChild := viewer.flowBox.ChildAtIndex(i)
		viewer.widgetToIndex[flowChild] = i
	}
}

func (viewer *PictureViewer) createHighlightWidget(item *types.ListItem, index int) gtk.Widgetter {
	box := gtk.NewBox(gtk.OrientationVertical, 0)
	box.SetSizeRequest(400, 250)
	box.SetHAlign(gtk.AlignCenter)

	pic := gtk.NewPicture()
	pic.SetCanShrink(true)
	pic.SetContentFit(gtk.ContentFitCover)
	pic.SetSizeRequest(400, 250)
	pic.AddCSSClass("card") // Rounded corners if available

	tex, err := thumbnail.Generate(item.Path)
	if err == nil {
		pic.SetPaintable(tex)
	}

	overlay := gtk.NewOverlay()
	overlay.SetChild(pic)

	// Add click gesture to carousel items
	gesture := gtk.NewGestureClick()
	gesture.ConnectPressed(func(nPress int, x, y float64) {
		viewer.flowBox.UnselectAll()
		viewer.SelectedIdxs = map[int]bool{index: true}
		viewer.SelectedIDX = index
		if viewer.SelectionChanged != nil {
			viewer.SelectionChanged(index)
		}
	})
	box.AddController(gesture)

	overlay.SetChild(pic)
	box.Append(overlay)
	return box
}

func (viewer *PictureViewer) createAlbumWidget(item *types.ListItem, index int) gtk.Widgetter {
	box := gtk.NewBox(gtk.OrientationVertical, 4)
	box.SetSizeRequest(160, 200)
	box.SetVAlign(gtk.AlignStart)
	box.SetHAlign(gtk.AlignCenter)

	// Album cover background
	cover := gtk.NewBox(gtk.OrientationVertical, 0)
	cover.SetSizeRequest(160, 160)
	cover.AddCSSClass("card")
	cover.SetHAlign(gtk.AlignCenter)
	cover.SetVAlign(gtk.AlignCenter)

	// Icon for directory
	img := gtk.NewImageFromIconName("folder-pictures-symbolic")
	img.SetPixelSize(64)
	cover.Append(img)

	label := gtk.NewLabel(item.Name)
	label.SetEllipsize(pango.EllipsizeEnd)
	label.SetMaxWidthChars(15)

	countLabel := gtk.NewLabel(fmt.Sprintf("%d items", item.ItemCount))
	countLabel.AddCSSClass("dim-label")

	box.Append(cover)
	box.Append(label)
	box.Append(countLabel)

	// Double-click to open album
	gesture := gtk.NewGestureClick()
	gesture.ConnectPressed(func(nPress int, x, y float64) {
		if nPress == 2 && viewer.pathChanged != nil {
			viewer.pathChanged(item.Path)
		}
	})
	box.AddController(gesture)

	return box
}

func (viewer *PictureViewer) createThumbWidget(item *types.ListItem, index int) gtk.Widgetter {
	box := gtk.NewBox(gtk.OrientationVertical, 4)
	box.SetSizeRequest(160, 200)
	box.SetVAlign(gtk.AlignStart)
	box.SetHAlign(gtk.AlignCenter)

	pic := gtk.NewPicture()
	pic.SetCanShrink(true)
	pic.SetContentFit(gtk.ContentFitCover)
	pic.SetSizeRequest(160, 160)
	pic.AddCSSClass("card")

	tex, err := thumbnail.Generate(item.Path)
	if err == nil {
		pic.SetPaintable(tex)
	}

	label := gtk.NewLabel(item.Name)
	label.SetEllipsize(pango.EllipsizeEnd)
	label.SetMaxWidthChars(15)

	box.Append(pic)
	box.Append(label)

	return box
}
