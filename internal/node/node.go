package node

import (
	"sync"
	"time"
)

type FSNode interface {
	GetNode() *Node
}

type Node struct {
	mu sync.RWMutex

	ID       uint64
	Name     string
	Parent   FSNode
	ObjectID *string

	Mode  uint32
	UID   uint32
	GID   uint32
	Nlink uint32

	Atime time.Time
	Mtime time.Time
	Ctime time.Time
}