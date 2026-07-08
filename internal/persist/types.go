package persist

import "time"

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

    Mode  uint32
    UID   uint32
    GID   uint32
    Nlink uint32

    Atime time.Time
    Mtime time.Time
    Ctime time.Time

    ObjectID *string
}

type Filesystem struct {
    NextID uint64
    Nodes []Node
    Objects []Object
}

type Metadata struct {
	NextID uint64
	Nodes   []Node
	Objects []ObjectRef `json:"objects"`
}

type ObjectRef struct {
	ID string `json:"id"`
}

type Object struct {
    ID   string
    Data []byte
}