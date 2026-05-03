package types

type ListItem struct {
	Name        string
	IsDir       bool
	Path        string // full path include file name and extension
	Group       string
	ItemCount   int
	Size        int64  //as byte
	SpecialInfo string // for special paths
}

type FileType int

const (
	TypeDir FileType = iota
	TypeExec
	TypeHidden
	TypeTemp
	TypeOther
	TypeDoc
	TypeMedia
)
