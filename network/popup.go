package network

import (
	"github.com/diamondburned/gotk4/pkg/gtk/v4"
)

type ConnectWindow struct {
	*gtk.Window
	ProtocolCombo *gtk.DropDown
	HostEntry     *gtk.Entry
	PortEntry     *gtk.Entry
	UserEntry     *gtk.Entry
	PassEntry     *gtk.Entry
	PathEntry     *gtk.Entry
	NameEntry     *gtk.Entry
	OnConnect     func(Connection)
}

func NewConnectWindow(onConnect func(Connection)) *ConnectWindow {
	cw := &ConnectWindow{
		Window:    gtk.NewWindow(),
		OnConnect: onConnect,
	}

	cw.SetTitle("Connect to Server")
	cw.SetResizable(false)
	cw.SetDefaultSize(400, -1)
	cw.SetModal(true)

	headerBar := gtk.NewHeaderBar()
	headerBar.SetShowTitleButtons(false)

	titleLabel := gtk.NewLabel("Connect to Server")
	titleLabel.AddCSSClass("title-4")
	headerBar.SetTitleWidget(titleLabel)

	cancelBtn := gtk.NewButtonWithLabel("Cancel")
	cancelBtn.ConnectClicked(func() {
		cw.Destroy()
	})
	headerBar.PackStart(cancelBtn)

	connectBtn := gtk.NewButtonWithLabel("Connect")
	connectBtn.AddCSSClass("suggested-action")
	headerBar.PackEnd(connectBtn)
	cw.SetTitlebar(headerBar)

	mainBox := gtk.NewBox(gtk.OrientationVertical, 12)
	mainBox.SetMarginTop(16)
	mainBox.SetMarginBottom(16)
	mainBox.SetMarginStart(16)
	mainBox.SetMarginEnd(16)

	protocols := []string{"ftp", "smb", "sftp", "ssh"}
	cw.ProtocolCombo = gtk.NewDropDownFromStrings(protocols)
	
	cw.NameEntry = gtk.NewEntry()
	cw.NameEntry.SetPlaceholderText("Connection Name (optional)")
	
	cw.HostEntry = gtk.NewEntry()
	cw.HostEntry.SetPlaceholderText("Host (e.g. 192.168.1.1)")
	
	cw.PortEntry = gtk.NewEntry()
	cw.PortEntry.SetPlaceholderText("Port (optional)")
	
	cw.UserEntry = gtk.NewEntry()
	cw.UserEntry.SetPlaceholderText("User (optional)")

	cw.PassEntry = gtk.NewEntry()
	cw.PassEntry.SetPlaceholderText("Password (optional)")
	cw.PassEntry.SetVisibility(false)
	
	cw.PathEntry = gtk.NewEntry()
	cw.PathEntry.SetPlaceholderText("Path (optional)")

	addRow := func(label string, widget gtk.Widgetter) {
		box := gtk.NewBox(gtk.OrientationHorizontal, 12)
		l := gtk.NewLabel(label)
		l.SetSizeRequest(80, -1)
		l.SetHAlign(gtk.AlignStart)
		box.Append(l)
		box.Append(widget)
		if w, ok := widget.(interface{ SetHExpand(bool) }); ok {
			w.SetHExpand(true)
		}
		mainBox.Append(box)
	}

	addRow("Protocol", cw.ProtocolCombo)
	addRow("Name", cw.NameEntry)
	addRow("Host", cw.HostEntry)
	addRow("Port", cw.PortEntry)
	addRow("User", cw.UserEntry)
	addRow("Password", cw.PassEntry)
	addRow("Path", cw.PathEntry)

	connectAction := func() {
		if cw.HostEntry.Text() == "" {
			return
		}
		
		protocol := protocols[cw.ProtocolCombo.Selected()]
		name := cw.NameEntry.Text()
		if name == "" {
			name = protocol + "://" + cw.HostEntry.Text()
		}
		
		conn := Connection{
			Name:     name,
			Protocol: protocol,
			Host:     cw.HostEntry.Text(),
			Port:     cw.PortEntry.Text(),
			User:     cw.UserEntry.Text(),
			Password: cw.PassEntry.Text(),
			Path:     cw.PathEntry.Text(),
		}
		
		cw.OnConnect(conn)
		cw.Destroy()
	}

	connectBtn.ConnectClicked(connectAction)
	cw.HostEntry.Connect("activate", connectAction)

	cw.SetChild(mainBox)
	cw.HostEntry.GrabFocus()

	return cw
}
