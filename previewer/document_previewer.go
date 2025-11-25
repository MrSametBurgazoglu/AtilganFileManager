package previewer

import (
	"fmt"
	"os"
	"path/filepath"

	"github.com/MrSametBurgazoglu/atilgan/cache"
	"github.com/MrSametBurgazoglu/atilgan/fileops"
	"github.com/MrSametBurgazoglu/atilgan/previewer/doc_extractor"
	"github.com/MrSametBurgazoglu/atilgan/previewer/pdf_extractor"
	"github.com/MrSametBurgazoglu/atilgan/thumbnail"
	"github.com/diamondburned/gotk4/pkg/core/glib"
	"github.com/diamondburned/gotk4/pkg/gdk/v4"
	"github.com/diamondburned/gotk4/pkg/gtk/v4"
)

type DocumentPreviewer struct {
	*gtk.Box
	Stack              *gtk.Stack
	Picture            *gtk.Picture
	Spinner            *gtk.Spinner
	Thumbnail          *gtk.Image
	LoadingPlaceholder *gtk.Box
	PictureList        *gtk.Box
	PageLabel          *gtk.Label
	PrevButton         *gtk.Button
	NextButton         *gtk.Button
	currentPage        int
	pageCount          int
	filePath           string
	images             []*gdk.Texture
	tempDir            string
}

func NewDocumentPreviewer() *DocumentPreviewer {
	box := gtk.NewBox(gtk.OrientationVertical, 0)
	picture := gtk.NewPicture()
	picture.SetContentFit(gtk.ContentFitContain)

	placeholder := gtk.NewBox(gtk.OrientationVertical, 0)
	thumbnail := gtk.NewImageFromIconName("text-x-generic-symbolic")
	thumbnail.SetPixelSize(256)
	placeholder.Append(thumbnail)
	spinner := gtk.NewSpinner()
	placeholder.Append(spinner)

	pictureList := gtk.NewBox(gtk.OrientationHorizontal, 0)
	pageLabel := gtk.NewLabel("")
	prevButton := gtk.NewButtonWithLabel("Prev")
	nextButton := gtk.NewButtonWithLabel("Next")

	controlBox := gtk.NewBox(gtk.OrientationHorizontal, 0)
	controlBox.SetHAlign(gtk.AlignCenter)
	controlBox.Append(prevButton)
	controlBox.Append(pageLabel)
	controlBox.Append(nextButton)

	stack := gtk.NewStack()
	stack.AddTitled(picture, "picture", "Picture")
	stack.AddTitled(placeholder, "placeholder", "Placeholder")
	stack.SetVExpand(true)

	box.Append(stack)
	box.Append(pictureList)
	box.Append(controlBox)

	tempDir, err := os.MkdirTemp("", "doc_preview")
	if err != nil {
		println("couldn't create doc preview")
	}

	documentPreviewer := &DocumentPreviewer{
		Box:         box,
		Picture:     picture,
		Stack:       stack,
		Spinner:     spinner,
		Thumbnail:   thumbnail,
		PictureList: pictureList,
		PageLabel:   pageLabel,
		PrevButton:  prevButton,
		NextButton:  nextButton,
		currentPage: 0,
		pageCount:   0,
		tempDir:     tempDir,
	}

	prevButton.ConnectClicked(documentPreviewer.prevPage)
	nextButton.ConnectClicked(documentPreviewer.nextPage)

	return documentPreviewer
}

func (dp *DocumentPreviewer) SetDocument(filePath string, fileInfo os.FileInfo) {
	dp.Spinner.Start()
	dp.Stack.SetVisibleChildName("placeholder")
	if fileInfo.IsDir() {
		return
	}
	if cachedImages, found := cache.Get(filePath); found {
		dp.setImages(cachedImages)
		return
	}
	paintable, err := thumbnail.Generate(filePath)
	if err == nil {
		dp.Thumbnail.SetFromPaintable(paintable)
	} else {
		dp.Thumbnail.SetFromIconName(fileops.GetIconForFile(filePath))
	}

	go func() {
		images, err := dp.GetImages(filePath, fileInfo)
		if err != nil {
			glib.IdleAdd(func() {
				dp.Spinner.Stop()
			})
			return
		}
		glib.IdleAdd(func() {
			dp.setImages(images)
		})
	}()
}

func (dp *DocumentPreviewer) setImages(images []*gdk.Texture) {
	dp.images = images
	dp.currentPage = 0
	dp.pageCount = len(images)
	dp.SetPicture(images[0])
	dp.Spinner.Stop()
	dp.Stack.SetVisibleChildName("picture")
}

func (dp *DocumentPreviewer) GetImages(filePath string, fileInfo os.FileInfo) ([]*gdk.Texture, error) {
	if cachedImages, found := cache.Get(filePath); found {
		return cachedImages, nil
	}

	var imagePaths []string
	var err error

	switch ext := filepath.Ext(filePath); ext {
	case ".pdf":
		imagePaths, err = pdf_extractor.PdfToImages(dp.tempDir, filePath, fileInfo.Name())
	case ".doc", ".docx", ".xls", ".xlsx", ".ppt", ".pptx":
		imagePaths, err = doc_extractor.DocumentToImages(dp.tempDir, filePath, fileInfo.Name())
	default:
		return nil, fmt.Errorf("unsupported file type")
	}
	if err != nil {
		return nil, err
	}

	var textures []*gdk.Texture
	for _, path := range imagePaths {
		texture, err := gdk.NewTextureFromFilename(path)
		if err != nil {
			return nil, err
		}
		textures = append(textures, texture)
	}

	cache.Add(filePath, textures)
	return textures, nil
}

func (dp *DocumentPreviewer) SetPicture(texture *gdk.Texture) {
	dp.Picture.SetPaintable(texture)
}

func (dp *DocumentPreviewer) SetPictureList(textures []*gdk.Texture) {
	for child := dp.PictureList.FirstChild(); child != nil; child = dp.PictureList.FirstChild() {
		dp.PictureList.Remove(child)
	}
	startIndex := dp.currentPage - 2
	if startIndex < 0 {
		startIndex = 0
	}
	endIndex := startIndex + 2
	if endIndex > len(textures) {
		endIndex = len(textures)
	}
	for _, texture := range textures[startIndex:endIndex] {
		newPicture := gtk.NewPictureForPaintable(texture)
		dp.PictureList.Append(newPicture)
	}
}

func (dp *DocumentPreviewer) prevPage() {
	if dp.currentPage > 0 {
		dp.currentPage--
		dp.SetPicture(dp.images[dp.currentPage])
	}
}

func (dp *DocumentPreviewer) nextPage() {
	if dp.currentPage < dp.pageCount-1 {
		dp.currentPage++
		dp.SetPicture(dp.images[dp.currentPage])
	}
}

func (dp *DocumentPreviewer) Close() {
	entries, err := os.ReadDir(dp.tempDir)
	if err != nil {
		return
	}
	for _, entry := range entries {
		os.RemoveAll(filepath.Join(dp.tempDir, entry.Name()))
	}
}
