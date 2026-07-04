package fs

import (
	"time"
)

type FSNode interface {
	GetNode() *Node
}

type Node struct {
	ID       uint64
	Name     string
	Parent   *Directory
	ObjectID *string

	Mode  uint32
	UID   uint32
	GID   uint32
	Nlink uint32

	Atime time.Time
	Mtime time.Time
	Ctime time.Time
}

func (f *File) GetNode() *Node {
	return &f.Node
}

func (d *Directory) GetNode() *Node {
	return &d.Node
}
