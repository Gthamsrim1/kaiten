package persist

import (
	"time"

	"github.com/Gthamsrim1/kaiten/internal/store"
)

type NodeType uint8

const (
	TypeFile NodeType = iota
	TypeDirectory
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
}

type Filesystem struct {
	NextID  uint64
	Nodes   []Node
	Objects []Object
}

type Metadata struct {
	NextID  uint64
	Nodes   []Node
	Objects []ObjectRef `json:"objects"`
}

type ObjectRef struct {
	ID [32]byte `json:"id"`
}

type Object struct {
	ID   [32]byte
	Data []byte
}
