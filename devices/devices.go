package devices

import (
	"github.com/diamondburned/gotk4/pkg/gio/v2"
	"github.com/diamondburned/gotk4/pkg/gtk/v4"
	"github.com/diamondburned/gotk4/pkg/pango"
)

type DeviceManager struct {
	*gtk.Box
	pathChanged func(string)
	monitor     *gio.VolumeMonitor
	buttons     map[string]*gtk.Button
}

func NewDeviceManager(pathChanged func(string)) *DeviceManager {
	dm := &DeviceManager{
		Box:         gtk.NewBox(gtk.OrientationVertical, 8),
		pathChanged: pathChanged,
		monitor:     gio.VolumeMonitorGet(),
		buttons:     make(map[string]*gtk.Button),
	}

	dm.setupMonitor()
	dm.Update()

	return dm
}

func (dm *DeviceManager) setupMonitor() {
	dm.monitor.ConnectMountAdded(func(_ gio.Mounter) { dm.Update() })
	dm.monitor.ConnectMountRemoved(func(_ gio.Mounter) { dm.Update() })
	dm.monitor.ConnectVolumeAdded(func(_ gio.Volumer) { dm.Update() })
	dm.monitor.ConnectVolumeRemoved(func(_ gio.Volumer) { dm.Update() })
	dm.monitor.ConnectDriveConnected(func(_ gio.Driver) { dm.Update() })
	dm.monitor.ConnectDriveDisconnected(func(_ gio.Driver) { dm.Update() })
}

func (dm *DeviceManager) Update() {
	// Clear existing widgets
	for child := dm.FirstChild(); child != nil; {
		next := gtk.BaseWidget(child).NextSibling()
		dm.Remove(child)
		child = next
	}

	// Clear buttons map
	for k := range dm.buttons {
		delete(dm.buttons, k)
	}

	mounts := dm.monitor.Mounts()
	volumes := dm.monitor.Volumes()

	if len(mounts) > 0 || len(volumes) > 0 {
		label := gtk.NewLabel("Devices")
		label.SetHAlign(gtk.AlignStart)
		label.AddCSSClass("sidebar-header")
		dm.Append(label)
	}

	// Add mounted devices
	for _, mount := range mounts {
		dm.addMountRow(mount)
	}

	// Add unmounted volumes
	for _, volume := range volumes {
		if volume.GetMount() == nil {
			dm.addVolumeRow(volume)
		}
	}
}

func (dm *DeviceManager) addMountRow(mount gio.Mounter) {
	root := mount.Root()
	if root == nil {
		return
	}
	path := root.Path()
	name := mount.Name()
	icon := mount.SymbolicIcon()

	row := gtk.NewBox(gtk.OrientationHorizontal, 0)
	row.AddCSSClass("sidebar-row")

	btn := gtk.NewButton()
	box := gtk.NewBox(gtk.OrientationHorizontal, 8)
	box.SetHAlign(gtk.AlignStart)

	img := gtk.NewImageFromGIcon(icon)
	lbl := gtk.NewLabel(name)
	lbl.SetEllipsize(pango.EllipsizeEnd)

	box.Append(img)
	box.Append(lbl)

	btn.SetChild(box)
	btn.AddCSSClass("sidebar-button")
	btn.SetHExpand(true)
	btn.SetHAlign(gtk.AlignFill)

	btn.ConnectClicked(func() {
		dm.pathChanged(path)
	})

	row.Append(btn)

	if mount.CanEject() || mount.CanUnmount() {
		ejectBtn := gtk.NewButtonFromIconName("media-eject-symbolic")
		ejectBtn.AddCSSClass("sidebar-eject-button")
		ejectBtn.SetTooltipText("Eject")
		ejectBtn.ConnectClicked(func() {
			if mount.CanEject() {
				mount.Eject(nil, gio.MountUnmountNone, nil)
			} else {
				mount.Unmount(nil, gio.MountUnmountNone, nil)
			}
		})
		row.Append(ejectBtn)
	}

	dm.Append(row)
	dm.buttons[path] = btn
}

func (dm *DeviceManager) addVolumeRow(volume gio.Volumer) {
	name := volume.Name()
	icon := volume.SymbolicIcon()

	row := gtk.NewBox(gtk.OrientationHorizontal, 0)
	row.AddCSSClass("sidebar-row")

	btn := gtk.NewButton()
	box := gtk.NewBox(gtk.OrientationHorizontal, 8)
	box.SetHAlign(gtk.AlignStart)

	img := gtk.NewImageFromGIcon(icon)
	lbl := gtk.NewLabel(name)
	lbl.SetEllipsize(pango.EllipsizeEnd)

	box.Append(img)
	box.Append(lbl)

	btn.SetChild(box)
	btn.AddCSSClass("sidebar-button")
	btn.SetHExpand(true)
	btn.SetHAlign(gtk.AlignFill)

	btn.ConnectClicked(func() {
		volume.Mount(nil, gio.MountMountNone, nil, func(res gio.AsyncResulter) {
			err := volume.MountFinish(res)
			if err == nil {
				mount := volume.GetMount()
				if mount != nil {
					dm.pathChanged(mount.Root().Path())
				}
			}
		})
	})

	row.Append(btn)
	dm.Append(row)
}

func (dm *DeviceManager) SetPath(path string) {
	for btnPath, btn := range dm.buttons {
		if btnPath == path {
			btn.AddCSSClass("selected")
		} else {
			btn.RemoveCSSClass("selected")
		}
	}
}
