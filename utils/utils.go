package utils

import (
	"fmt"

	"github.com/diamondburned/gotk4/pkg/gtk/v4"
)

func ShowError(parent *gtk.Window, err error) {
	if err == nil {
		return
	}
	
	dialog := gtk.NewMessageDialog(
		parent,
		gtk.DialogModal,
		gtk.MessageError,
		gtk.ButtonsOK,
	)
	dialog.SetMarkup("<b>An error occurred</b>")
	dialog.SetObjectProperty("secondary-text", err.Error())
	dialog.ConnectResponse(func(responseID int) {
		dialog.Destroy()
	})
	dialog.Show()
	fmt.Printf("Error: %v\n", err)
}

func LogError(err error) {
	if err != nil {
		fmt.Printf("Error: %v\n", err)
	}
}
