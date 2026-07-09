package persist

import (
	"time"

	"github.com/Gthamsrim1/kaiten/internal/store"
)

type NodeType uint8

const (
	TypeFile NodeType = iota
	TypeDirectory
	TypeSymlink
)

type Node struct {
	ID       uint64
	ParentID uint64

	Name string

	Type NodeType

	Chunks []store.ChunkRef

	Mode  uint32
	UID   uint32
	GID   uint32
	Nlink uint32

	Atime time.Time
	Mtime time.Time
	Ctime time.Time

	Target string `json:",omitempty"`
}

type Snapshot struct {
	ID string

	ParentID *string

	NextID uint64
	Nodes  []Node

	Objects []Object `json:"-"`
}

type ObjectRef struct {
	ID [32]byte `json:"id"`
}

type Object struct {
	ID   [32]byte
	Data []byte
}
