package network

import (
	"github.com/diamondburned/gotk4/pkg/gtk/v4"
)

type NetworkViewer struct {
	*gtk.Box
	network     *Network
	pathChanged func(string)
	flowBox     *gtk.FlowBox
}

func NewNetworkViewer(network *Network, pathChanged func(string)) *NetworkViewer {
	nv := &NetworkViewer{
		Box:         gtk.NewBox(gtk.OrientationVertical, 32),
		network:     network,
		pathChanged: pathChanged,
	}

	nv.AddCSSClass("network-viewer")
	nv.SetMarginTop(40)
	nv.SetMarginBottom(40)
	nv.SetMarginStart(40)
	nv.SetMarginEnd(40)

	// --- Connect Section ---
	connectBox := gtk.NewBox(gtk.OrientationVertical, 12)
	connectBox.SetHAlign(gtk.AlignCenter)

	sectionTitle := gtk.NewLabel("Server Connections")
	sectionTitle.AddCSSClass("network-section-title")
	connectBox.Append(sectionTitle)

	formCard := gtk.NewBox(gtk.OrientationVertical, 16)
	formCard.AddCSSClass("network-form-card")
	formCard.SetSizeRequest(450, -1)

	protocols := []string{"ftp", "smb", "sftp", "ssh"}
	protocolCombo := gtk.NewDropDownFromStrings(protocols)
	
	hostEntry := gtk.NewEntry()
	hostEntry.SetPlaceholderText("Hostname or IP Address")
	
	userEntry := gtk.NewEntry()
	userEntry.SetPlaceholderText("Username")
	
	passEntry := gtk.NewEntry()
	passEntry.SetPlaceholderText("Password")
	passEntry.SetVisibility(false)

	connectBtn := gtk.NewButtonWithLabel("Establish Connection")
	connectBtn.AddCSSClass("suggested-action")
	connectBtn.AddCSSClass("network-connect-button")

	addRow := func(label string, widget gtk.Widgetter) {
		row := gtk.NewBox(gtk.OrientationHorizontal, 12)
		l := gtk.NewLabel(label)
		l.SetSizeRequest(90, -1)
		l.SetHAlign(gtk.AlignStart)
		l.AddCSSClass("caption")
		row.Append(l)
		row.Append(widget)
		if w, ok := widget.(interface{ SetHExpand(bool) }); ok {
			w.SetHExpand(true)
		}
		formCard.Append(row)
	}

	addRow("Protocol", protocolCombo)
	addRow("Host", hostEntry)
	addRow("User", userEntry)
	addRow("Password", passEntry)
	formCard.Append(connectBtn)

	connectBox.Append(formCard)
	nv.Append(connectBox)

	connectAction := func() {
		if hostEntry.Text() == "" {
			return
		}
		protocol := protocols[protocolCombo.Selected()]
		conn := Connection{
			Name:     protocol + "://" + hostEntry.Text(),
			Protocol: protocol,
			Host:     hostEntry.Text(),
			User:     userEntry.Text(),
			Password: passEntry.Text(),
		}
		nv.network.AddConnection(conn)
		nv.Refresh()
		nv.pathChanged(conn.GetURI())
	}

	connectBtn.ConnectClicked(connectAction)
	hostEntry.Connect("activate", connectAction)

	// --- Saved Connections Section ---
	savedHeader := gtk.NewBox(gtk.OrientationHorizontal, 8)
	savedTitle := gtk.NewLabel("Saved Locations")
	savedTitle.AddCSSClass("network-section-title")
	savedHeader.Append(savedTitle)
	nv.Append(savedHeader)

	nv.flowBox = gtk.NewFlowBox()
	nv.flowBox.SetSelectionMode(gtk.SelectionNone)
	nv.flowBox.SetColumnSpacing(16)
	nv.flowBox.SetRowSpacing(16)
	nv.flowBox.SetMaxChildrenPerLine(10)
	nv.flowBox.SetHAlign(gtk.AlignStart)
	
	scrolled := gtk.NewScrolledWindow()
	scrolled.SetChild(nv.flowBox)
	scrolled.SetVExpand(true)
	nv.Append(scrolled)

	nv.Refresh()

	return nv
}

func (nv *NetworkViewer) Refresh() {
	nv.network.LoadConnections()
	
	for child := nv.flowBox.FirstChild(); child != nil; {
		next := gtk.BaseWidget(child).NextSibling()
		nv.flowBox.Remove(child)
		child = next
	}

	for _, conn := range nv.network.Connections {
		nv.flowBox.Append(nv.createConnectionCard(conn))
	}
}

func (nv *NetworkViewer) createConnectionCard(conn Connection) *gtk.Box {
	card := gtk.NewBox(gtk.OrientationVertical, 8)
	card.AddCSSClass("network-card")
	card.SetSizeRequest(180, 120)

	topRow := gtk.NewBox(gtk.OrientationHorizontal, 0)
	
	iconName := "network-server-symbolic"
	if conn.Protocol == "ssh" || conn.Protocol == "sftp" {
		iconName = "network-workgroup-symbolic"
	}
	
	icon := gtk.NewImageFromIconName(iconName)
	icon.SetPixelSize(48)
	icon.SetHAlign(gtk.AlignCenter)
	card.Append(icon)

	title := gtk.NewLabel(conn.Name)
	title.AddCSSClass("network-card-title")
	title.SetEllipsize(1) // pango.EllipsizeEnd
	card.Append(title)

	subtitle := gtk.NewLabel(conn.Host)
	subtitle.AddCSSClass("network-card-subtitle")
	subtitle.SetEllipsize(1)
	card.Append(subtitle)

	gesture := gtk.NewGestureClick()
	gesture.ConnectPressed(func(n int, x, y float64) {
		nv.pathChanged(conn.GetURI())
	})
	card.AddController(gesture)

	removeBtn := gtk.NewButtonFromIconName("user-trash-symbolic")
	removeBtn.AddCSSClass("flat")
	removeBtn.AddCSSClass("network-remove-button")
	removeBtn.SetTooltipText("Remove Connection")
	removeBtn.SetHAlign(gtk.AlignEnd)
	removeBtn.SetVAlign(gtk.AlignStart)
	
	overlay := gtk.NewOverlay()
	overlay.SetChild(card)
	overlay.AddOverlay(removeBtn)
	
	removeBtn.ConnectClicked(func() {
		nv.removeConnection(conn)
	})

	container := gtk.NewBox(gtk.OrientationVertical, 0)
	container.Append(overlay)

	return container
}

func (nv *NetworkViewer) removeConnection(conn Connection) {
	newConns := make([]Connection, 0)
	for _, c := range nv.network.Connections {
		if c.GetURI() != conn.GetURI() {
			newConns = append(newConns, c)
		}
	}
	nv.network.Connections = newConns
	nv.network.SaveConnections()
	nv.Refresh()
}
