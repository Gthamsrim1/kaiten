package node

import (
	"sync"
	"syscall"
	"time"

	"golang.org/x/sys/unix"
)

type FSNode interface {
	GetNode() *Node
}

type AttrUpdate struct {
	Mode  *uint32
	UID   *uint32
	GID   *uint32
	ATime *time.Time
	MTime *time.Time
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

func (n *Node) UpdateAttr(update AttrUpdate) {
	n.mu.Lock()
	defer n.mu.Unlock()

	changed := false

	if update.Mode != nil {
		n.Mode = (n.Mode & syscall.S_IFMT) | (*update.Mode &^ syscall.S_IFMT)
		changed = true
	}

	if update.UID != nil {
		n.UID = *update.UID
		changed = true
	}

	if update.GID != nil {
		n.GID = *update.GID
		changed = true
	}

	if update.ATime != nil {
		n.Atime = *update.ATime
		changed = true
	}

	if update.MTime != nil {
		n.Mtime = *update.MTime
		changed = true
	}

	if changed {
		n.Ctime = time.Now()
	}
}

func (n *Node) CheckAccess(uid, gid uint32, mask uint32) syscall.Errno {
	n.mu.RLock()
	perm := n.Mode & 0777
	owner := n.UID
	group := n.GID
	n.mu.RUnlock()

	if uid == 0 {
		return 0
	}

	var allowed uint32

	switch {
	case uid == owner:
		allowed = (perm >> 6) & 7
	case gid == group:
		allowed = (perm >> 3) & 7
	default:
		allowed = perm & 7
	}

	if mask&unix.R_OK != 0 && allowed&4 == 0 {
		return syscall.EACCES
	}

	if mask&unix.W_OK != 0 && allowed&2 == 0 {
		return syscall.EACCES
	}

	if mask&unix.X_OK != 0 && allowed&1 == 0 {
		return syscall.EACCES
	}

	return 0
}