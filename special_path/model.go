package special_path

import "github.com/MrSametBurgazoglu/atilgan/types"

type IPath interface {
	GetItems() []*types.ListItem
	GetPath() string
	GetParentPath() string
	GetName() string
}
