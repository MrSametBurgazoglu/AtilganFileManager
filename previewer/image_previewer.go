package previewer

import (
	"fmt"
	"image"
	_ "image/gif"
	_ "image/jpeg"
	_ "image/png"
	"os"
	"path/filepath"
	"strings"

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
		if info.Width > 0 && info.Height > 0 {
			ip.DimensionLabel.SetText(fmt.Sprintf("%d x %d", info.Width, info.Height))
		} else {
			ip.DimensionLabel.SetText("")
		}
		ip.TypeLabel.SetText(info.Type)
		return
	}

	ext := strings.ToLower(filepath.Ext(fileInfo.Name()))
	if ext == ".svg" {
		info := &cache.FileInfo{
			Width:  0,
			Height: 0,
			Type:   "Scalable Vector Graphics",
		}
		ip.fileInfoCache.Set(filePath, info)
		ip.DimensionLabel.SetText("")
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
		// Fallback for other formats that might not be supported by image.DecodeConfig
		ip.DimensionLabel.SetText("")
		ip.TypeLabel.SetText(strings.ToUpper(strings.TrimPrefix(ext, ".")))
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
