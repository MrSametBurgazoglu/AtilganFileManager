package previewer

import (
	"os"

	"github.com/diamondburned/gotk4/pkg/gdk/v4"
	"github.com/diamondburned/gotk4/pkg/gio/v2"
	"github.com/diamondburned/gotk4/pkg/gtk/v4"
)

type MediaPreviewer struct {
	*gtk.Box
	video         *gtk.Video
	mediaFile     *gtk.MediaFile
	playButton    *gtk.Button
	stopButton    *gtk.Button
	durationLabel *gtk.Label
}

func NewMediaPreviewer() *MediaPreviewer {
	box := gtk.NewBox(gtk.OrientationVertical, 0)
	video := gtk.NewVideo()
	video.SetHExpandSet(true)
	video.SetVisible(true)
	video.SetHExpand(false)
	video.SetVExpand(false)
	video.SetSizeRequest(200, 225)
	box.SetVExpand(false)
	box.SetHExpand(false)
	box.SetSizeRequest(200, 225)

	video.SetName("player")
	provider := gtk.NewCSSProvider()

	gtk.StyleContextAddProviderForDisplay(
		gdk.DisplayGetDefault(),
		provider,
		gtk.STYLE_PROVIDER_PRIORITY_APPLICATION,
	)

	playButton := gtk.NewButtonWithLabel("Play")
	stopButton := gtk.NewButtonWithLabel("Stop")

	controlBox := gtk.NewBox(gtk.OrientationHorizontal, 0)
	controlBox.SetHAlign(gtk.AlignCenter)
	controlBox.Append(playButton)
	controlBox.Append(stopButton)

	box.Append(video)
	box.Append(controlBox)

	mediaPreviewer := &MediaPreviewer{
		Box:        box,
		video:      video,
		playButton: playButton,
		stopButton: stopButton,
	}

	playButton.ConnectClicked(mediaPreviewer.play)
	stopButton.ConnectClicked(mediaPreviewer.stop)

	return mediaPreviewer
}

func (mp *MediaPreviewer) SetMedia(filePath string, fileInfo os.FileInfo) {
	mp.stop()
	mp.mediaFile = gtk.NewMediaFile()
	mp.mediaFile.SetFile(gio.NewFileForPath(filePath))
	mp.video.SetMediaStream(mp.mediaFile)
}

func (mp *MediaPreviewer) play() {
	if mp.mediaFile == nil {
		return
	}
	mp.mediaFile.Play()
}

func (mp *MediaPreviewer) stop() {
	if mp.mediaFile == nil {
		return
	}
	mp.mediaFile.SetPlaying(false)
}

func (mp *MediaPreviewer) Close() {
	if mp.mediaFile != nil {
		mp.stop()
		mp.mediaFile.Clear()
	}
}
