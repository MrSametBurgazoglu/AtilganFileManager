package previewer

import (
	"fmt"
	"image"
	_ "image/gif"
	_ "image/jpeg"
	_ "image/png"
	"os"
	"path/filepath"

	"github.com/MrSametBurgazoglu/atilgan/cache"
	"github.com/diamondburned/gotk4/pkg/gtk/v4"
)

type ImagePreviewer struct {
	*gtk.Box
	Picture        *gtk.Picture
	DimensionLabel *gtk.Label
	TypeLabel      *gtk.Label
	fileInfoCache  *cache.FileInfoCache
}

func NewImagePreviewer() *ImagePreviewer {
	box := gtk.NewBox(gtk.OrientationVertical, 0)
	picture := gtk.NewPicture()
	dimensionLabel := gtk.NewLabel("")
	typeLabel := gtk.NewLabel("")

	box.Append(picture)
	box.Append(dimensionLabel)
	box.Append(typeLabel)

	return &ImagePreviewer{
		Box:            box,
		Picture:        picture,
		DimensionLabel: dimensionLabel,
		TypeLabel:      typeLabel,
		fileInfoCache:  cache.NewFileInfoCache(),
	}
}

func (ip *ImagePreviewer) SetImage(filePath string, fileInfo os.FileInfo) {
	ip.Picture.SetFilename(filePath)

	if info, found := ip.fileInfoCache.Get(filePath); found {
		ip.DimensionLabel.SetText(fmt.Sprintf("%d x %d", info.Width, info.Height))
		ip.TypeLabel.SetText(info.Type)
		return
	}

	file, err := os.Open(filePath)
	if err != nil {
		return
	}
	defer file.Close()

	config, _, err := image.DecodeConfig(file)
	if err != nil {
		return
	}

	info := &cache.FileInfo{
		Width:  config.Width,
		Height: config.Height,
		Type:   filepath.Ext(fileInfo.Name()),
	}
	ip.fileInfoCache.Set(filePath, info)

	ip.DimensionLabel.SetText(fmt.Sprintf("%d x %d", info.Width, info.Height))
	ip.TypeLabel.SetText(info.Type)
}
