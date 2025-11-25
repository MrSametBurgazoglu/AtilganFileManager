package thumbnail

import (
	"crypto/md5"
	"encoding/hex"
	"errors"
	"net/url"
	"os"
	"path/filepath"

	"github.com/diamondburned/gotk4/pkg/gdk/v4"
)

func Generate(filePath string) (*gdk.Texture, error) {
	thumbnailPath, err := getThumbnailPath(filePath, "large")
	if err != nil {
		return nil, err
	}

	if _, err := os.Stat(thumbnailPath); os.IsNotExist(err) {
		return nil, errors.New("thumbnail not found")
	}

	pixbuf, err := gdk.NewTextureFromFilename(thumbnailPath)
	if err != nil {
		return nil, err
	}

	return pixbuf, nil
}

func getThumbnailPath(filePath string, size string) (string, error) {
	absPath, err := filepath.Abs(filePath)
	if err != nil {
		return "", err
	}
	u := &url.URL{
		Scheme: "file",
		Path:   absPath,
	}
	uri := u.String()

	hasher := md5.New()
	hasher.Write([]byte(uri))
	md5sum := hex.EncodeToString(hasher.Sum(nil))

	homeDir, err := os.UserHomeDir()
	if err != nil {
		return "", err
	}

	return filepath.Join(homeDir, ".cache", "thumbnails", size, md5sum+".png"), nil
}
