package previewer

import (
	"fmt"
	"os"

	"github.com/MrSametBurgazoglu/atilgan/fileops"
	"github.com/diamondburned/gotk4/pkg/gio/v2"
	"github.com/diamondburned/gotk4/pkg/gtk/v4"
)

type VideoDetailsPreviewer struct {
	*gtk.Box
	video       *gtk.Video
	mediaFile   *gtk.MediaFile
	playButton  *gtk.Button
	stopButton  *gtk.Button
	nameLabel   *gtk.Label
	sizeLabel   *gtk.Label
	pathLabel   *gtk.Label
}

func NewVideoDetailsPreviewer() *VideoDetailsPreviewer {
	box := gtk.NewBox(gtk.OrientationVertical, 10)
	box.SetMarginTop(12)
	box.SetMarginBottom(12)
	box.SetMarginStart(12)
	box.SetMarginEnd(12)

	video := gtk.NewVideo()
	video.SetSizeRequest(250, 200)
	video.SetHExpand(true)

	controlBox := gtk.NewBox(gtk.OrientationHorizontal, 6)
	controlBox.SetHAlign(gtk.AlignCenter)
	playButton := gtk.NewButtonFromIconName("media-playback-start-symbolic")
	stopButton := gtk.NewButtonFromIconName("media-playback-stop-symbolic")
	controlBox.Append(playButton)
	controlBox.Append(stopButton)

	infoBox := gtk.NewBox(gtk.OrientationVertical, 4)
	nameLabel := gtk.NewLabel("")
	nameLabel.SetWrap(true)
	nameLabel.AddCSSClass("title-2")
	
	sizeLabel := gtk.NewLabel("")
	pathLabel := gtk.NewLabel("")
	pathLabel.SetWrap(true)
	pathLabel.AddCSSClass("caption")

	infoBox.Append(nameLabel)
	infoBox.Append(sizeLabel)
	infoBox.Append(pathLabel)

	box.Append(video)
	box.Append(controlBox)
	box.Append(infoBox)

	vdp := &VideoDetailsPreviewer{
		Box:        box,
		video:      video,
		playButton: playButton,
		stopButton: stopButton,
		nameLabel:  nameLabel,
		sizeLabel:  sizeLabel,
		pathLabel:  pathLabel,
	}

	playButton.ConnectClicked(func() {
		if vdp.mediaFile != nil {
			vdp.mediaFile.Play()
		}
	})
	stopButton.ConnectClicked(func() {
		if vdp.mediaFile != nil {
			vdp.mediaFile.SetPlaying(false)
		}
	})

	return vdp
}

func (vdp *VideoDetailsPreviewer) SetVideo(filePath string, info os.FileInfo) {
	if vdp.mediaFile != nil {
		vdp.mediaFile.SetPlaying(false)
		vdp.mediaFile.Clear()
	}

	vdp.mediaFile = gtk.NewMediaFile()
	vdp.mediaFile.SetFile(gio.NewFileForPath(filePath))
	vdp.video.SetMediaStream(vdp.mediaFile)

	vdp.nameLabel.SetLabel(info.Name())
	vdp.sizeLabel.SetLabel(fmt.Sprintf("Size: %s", fileops.GetFileSizeAsString(info.Size())))
	vdp.pathLabel.SetLabel(filePath)
}

func (vdp *VideoDetailsPreviewer) Close() {
	if vdp.mediaFile != nil {
		vdp.mediaFile.SetPlaying(false)
		vdp.mediaFile.Clear()
	}
}
