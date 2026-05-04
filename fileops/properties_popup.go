package fileops

import (
	"crypto/md5"
	"crypto/sha256"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strconv"

	"github.com/diamondburned/gotk4/pkg/glib/v2"
	"github.com/diamondburned/gotk4/pkg/gtk/v4"
)

type PropertiesWindow struct {
	*gtk.Window
	Path string
}

func NewPropertiesWindow(path string) *PropertiesWindow {
	win := &PropertiesWindow{
		Window: gtk.NewWindow(),
		Path:   path,
	}

	win.SetTitle("Properties - " + path)
	win.SetDefaultSize(400, 450)

	mainBox := gtk.NewBox(gtk.OrientationVertical, 12)
	mainBox.SetMarginTop(12)
	mainBox.SetMarginBottom(12)
	mainBox.SetMarginStart(12)
	mainBox.SetMarginEnd(12)
	win.SetChild(mainBox)

	info, err := os.Stat(path)
	if err != nil {
		mainBox.Append(gtk.NewLabel("Error: " + err.Error()))
		return win
	}

	addRow := func(label, value string) *gtk.Label {
		row := gtk.NewBox(gtk.OrientationHorizontal, 12)
		l := gtk.NewLabel(label)
		l.SetHAlign(gtk.AlignStart)
		l.SetSizeRequest(100, -1)
		v := gtk.NewLabel(value)
		v.SetHAlign(gtk.AlignStart)
		v.SetSelectable(true)
		v.SetEllipsize(1) // pango.EllipsizeEnd
		row.Append(l)
		row.Append(v)
		mainBox.Append(row)
		return v
	}

	addRow("Name:", info.Name())

	sizeLabel := addRow("Size:", GetFileSizeAsString(info.Size()))
	if info.IsDir() {
		sizeLabel.SetText("Calculating...")
		go func() {
			var totalSize int64
			filepath.Walk(path, func(_ string, info os.FileInfo, err error) error {
				if err == nil && !info.IsDir() {
					totalSize += info.Size()
				}
				return nil
			})
			glib.IdleAdd(func() {
				sizeLabel.SetText(GetFileSizeAsString(totalSize))
			})
		}()
	}

	addRow("Modified:", info.ModTime().Format("2006-01-02 15:04:05"))
	addRow("Permissions:", info.Mode().String())

	if !info.IsDir() {
		mainBox.Append(gtk.NewSeparator(gtk.OrientationHorizontal))

		md5Label := addRow("MD5:", "Calculating...")
		go func() {
			h := md5.New()
			f, err := os.Open(path)
			if err == nil {
				defer f.Close()
				io.Copy(h, f)
				sum := fmt.Sprintf("%x", h.Sum(nil))
				glib.IdleAdd(func() {
					md5Label.SetText(sum)
				})
			}
		}()

		shaLabel := addRow("SHA-256:", "Calculating...")
		go func() {
			h := sha256.New()
			f, err := os.Open(path)
			if err == nil {
				defer f.Close()
				io.Copy(h, f)
				sum := fmt.Sprintf("%x", h.Sum(nil))
				glib.IdleAdd(func() {
					shaLabel.SetText(sum)
				})
			}
		}()
	}

	mainBox.Append(gtk.NewSeparator(gtk.OrientationHorizontal))

	permLabel := gtk.NewLabel("Change Permissions (Octal):")
	permLabel.SetHAlign(gtk.AlignStart)
	mainBox.Append(permLabel)

	permEntry := gtk.NewEntry()
	permEntry.SetText(fmt.Sprintf("%o", info.Mode().Perm()))
	mainBox.Append(permEntry)

	applyBtn := gtk.NewButtonWithLabel("Apply")
	applyBtn.ConnectClicked(func() {
		permStr := permEntry.Text()
		perm, err := strconv.ParseUint(permStr, 8, 32)
		if err != nil {
			fmt.Println("Invalid permission string:", err)
			return
		}
		err = os.Chmod(path, os.FileMode(perm))
		if err != nil {
			fmt.Println("Error changing permissions:", err)
		}
		win.Close()
	})
	mainBox.Append(applyBtn)

	return win
}
