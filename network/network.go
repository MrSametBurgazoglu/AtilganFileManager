package network

import (
	"context"
	"encoding/json"
	"os"
	"path/filepath"
	"strings"

	"github.com/MrSametBurgazoglu/atilgan/types"
	"github.com/diamondburned/gotk4/pkg/gio/v2"
)

const NetworkPath = "network://"

type Connection struct {
	Name     string `json:"name"`
	Protocol string `json:"protocol"` // ftp, smb, sftp
	Host     string `json:"host"`
	Port     string `json:"port"`
	User     string `json:"user"`
	Password string `json:"password"`
	Path     string `json:"path"`
}

type Network struct {
	Connections []Connection
}

func NewNetwork() *Network {
	n := &Network{}
	n.LoadConnections()
	return n
}

func (n *Network) GetItems() []*types.ListItem {
	n.LoadConnections()
	items := make([]*types.ListItem, 0)

	// Add "Connect to Server" item
	items = append(items, &types.ListItem{
		Name:  "Connect to Server...",
		Path:  NetworkPath + "connect",
		IsDir: true,
	})

	for _, conn := range n.Connections {
		items = append(items, &types.ListItem{
			Name:  conn.Name,
			Path:  conn.GetURI(),
			IsDir: true,
		})
	}

	return items
}

func (c *Connection) GetURI() string {
	protocol := c.Protocol
	if protocol == "ssh" {
		protocol = "sftp"
	}
	uri := protocol + "://"
	if c.User != "" {
		uri += c.User
		if c.Password != "" {
			uri += ":" + c.Password
		}
		uri += "@"
	}
	uri += c.Host
	if c.Port != "" {
		uri += ":" + c.Port
	}
	if c.Path != "" {
		if c.Path[0] != '/' {
			uri += "/"
		}
		uri += c.Path
	} else {
		uri += "/"
	}
	return uri
}

func (n *Network) GetPath() string {
	return NetworkPath
}

func (n *Network) GetParentPath() string {
	return ""
}

func (n *Network) GetName() string {
	return "Network"
}

func (n *Network) LoadConnections() {
	configDir, _ := os.UserConfigDir()
	atilganDir := filepath.Join(configDir, "atilgan")
	_ = os.MkdirAll(atilganDir, 0755)
	
	connFile := filepath.Join(atilganDir, "network_connections.json")
	data, err := os.ReadFile(connFile)
	if err != nil {
		return
	}
	
	_ = json.Unmarshal(data, &n.Connections)
}

func (n *Network) SaveConnections() {
	configDir, _ := os.UserConfigDir()
	atilganDir := filepath.Join(configDir, "atilgan")
	connFile := filepath.Join(atilganDir, "network_connections.json")
	
	data, _ := json.Marshal(n.Connections)
	_ = os.WriteFile(connFile, data, 0644)
}

func (n *Network) AddConnection(conn Connection) {
	n.Connections = append(n.Connections, conn)
	n.SaveConnections()
}

type RemotePath struct {
	URI string
}

func NewRemotePath(uri string) *RemotePath {
	return &RemotePath{URI: uri}
}

func (rp *RemotePath) GetItems() []*types.ListItem {
	file := gio.NewFileForURI(rp.URI)
	// We use "standard::*" to get name, type, etc.
	enumerator, err := file.EnumerateChildren(context.Background(), "standard::*", gio.FileQueryInfoNone)
	if err != nil {
		println("Error enumerating children:", err.Error())
		return nil
	}
	defer enumerator.Close(nil)

	var items []*types.ListItem
	for {
		info, err := enumerator.NextFile(nil)
		if err != nil || info == nil {
			break
		}

		name := info.Name()
		isDir := info.FileType() == gio.FileTypeDirectory
		
		path := rp.URI
		if !strings.HasSuffix(path, "/") {
			path += "/"
		}
		path += name

		items = append(items, &types.ListItem{
			Name:  name,
			IsDir: isDir,
			Path:  path,
			Group: "Remote Files",
		})
	}
	return items
}

func (rp *RemotePath) GetPath() string {
	return rp.URI
}

func (rp *RemotePath) GetParentPath() string {
	// Simple parent path calculation for URI
	u := rp.URI
	u = strings.TrimSuffix(u, "/")
	lastSlash := strings.LastIndex(u, "/")
	if lastSlash > strings.Index(u, "://")+2 {
		return u[:lastSlash]
	}
	return NetworkPath
}

func (rp *RemotePath) GetName() string {
	return rp.URI
}
