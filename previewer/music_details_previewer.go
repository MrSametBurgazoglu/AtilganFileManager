package previewer

import (
	"fmt"
	"os"

	"github.com/MrSametBurgazoglu/atilgan/fileops"
	"github.com/diamondburned/gotk4/pkg/gio/v2"
	"github.com/diamondburned/gotk4/pkg/gtk/v4"
)

type MusicDetailsPreviewer struct {
	*gtk.Box
	audio       *gtk.Video // Using Video for simple audio playback in GTK4
	mediaFile   *gtk.MediaFile
	playButton  *gtk.Button
	stopButton  *gtk.Button
	nameLabel   *gtk.Label
	sizeLabel   *gtk.Label
	pathLabel   *gtk.Label
}

func NewMusicDetailsPreviewer() *MusicDetailsPreviewer {
	box := gtk.NewBox(gtk.OrientationVertical, 10)
	box.SetMarginTop(12)
	box.SetMarginBottom(12)
	box.SetMarginStart(12)
	box.SetMarginEnd(12)

	icon := gtk.NewImageFromIconName("audio-x-generic-symbolic")
	icon.SetPixelSize(64)
	icon.SetMarginTop(20)
	icon.SetMarginBottom(20)

	audio := gtk.NewVideo()
	audio.SetVisible(false) // Keep it hidden for audio

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

	box.Append(icon)
	box.Append(audio)
	box.Append(controlBox)
	box.Append(infoBox)

	mdp := &MusicDetailsPreviewer{
		Box:        box,
		audio:      audio,
		playButton: playButton,
		stopButton: stopButton,
		nameLabel:  nameLabel,
		sizeLabel:  sizeLabel,
		pathLabel:  pathLabel,
	}

	playButton.ConnectClicked(func() {
		if mdp.mediaFile != nil {
			mdp.mediaFile.Play()
		}
	})
	stopButton.ConnectClicked(func() {
		if mdp.mediaFile != nil {
			mdp.mediaFile.SetPlaying(false)
		}
	})

	return mdp
}

func (mdp *MusicDetailsPreviewer) SetMusic(filePath string, info os.FileInfo) {
	if mdp.mediaFile != nil {
		mdp.mediaFile.SetPlaying(false)
		mdp.mediaFile.Clear()
	}

	mdp.mediaFile = gtk.NewMediaFile()
	mdp.mediaFile.SetFile(gio.NewFileForPath(filePath))
	mdp.audio.SetMediaStream(mdp.mediaFile)

	mdp.nameLabel.SetLabel(info.Name())
	mdp.sizeLabel.SetLabel(fmt.Sprintf("Size: %s", fileops.GetFileSizeAsString(info.Size())))
	mdp.pathLabel.SetLabel(filePath)
}

func (mdp *MusicDetailsPreviewer) Close() {
	if mdp.mediaFile != nil {
		mdp.mediaFile.SetPlaying(false)
		mdp.mediaFile.Clear()
	}
}
