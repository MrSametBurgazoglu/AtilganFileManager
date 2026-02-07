package thumbnail

import (
	"crypto/md5"
	"encoding/hex"
	"errors"
	"net/url"
	"os"
	"path"
	"path/filepath"
	"strings"

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
	
	// Convert backslashes to forward slashes for POSIX-compatible URL
	posixPath := strings.ReplaceAll(absPath, "\\", "/")
	
	// On Windows, filepath.Abs returns C:\path, but URL path should be /C:/path
	u := &url.URL{
		Scheme: "file",
		Path:   posixPath,
	}
	uri := u.String()

	hasher := md5.New()
	hasher.Write([]byte(uri))
	md5sum := hex.EncodeToString(hasher.Sum(nil))

	homeDir, err := os.UserHomeDir()
	if err != nil {
		return "", err
	}

	cacheDir := path.Join(homeDir, ".cache", "thumbnails", size)
	// Use filepath for the final join to be OS-compatible
	return filepath.Join(cacheDir, md5sum+".png"), nil
}
