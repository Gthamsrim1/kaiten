package tree

import (
	"context"
	"syscall"

	"github.com/Gthamsrim1/kaiten/internal/node"
	gofuse "github.com/hanwen/go-fuse/v2/fs"
	"github.com/hanwen/go-fuse/v2/fuse"
)

type Symlink struct {
	gofuse.Inode

	node.Node
	FS *KaitenFS

	Target string
}

func (s *Symlink) GetNode() *node.Node {
	return &s.Node
}

func (s *Symlink) Getattr(ctx context.Context, f gofuse.FileHandle, out *fuse.AttrOut) syscall.Errno {
	out.Mode = s.Node.Mode
	out.Uid = s.Node.UID
	out.Gid = s.Node.GID
	out.SetTimes(&s.Node.Atime, &s.Node.Mtime, &s.Node.Ctime)
	return 0
}

func (s *Symlink) Readlink(ctx context.Context) ([]byte, syscall.Errno) {
	return []byte(s.Target), 0
}

var _ gofuse.NodeReadlinker = (*Symlink)(nil)
var _ gofuse.NodeGetattrer = (*Symlink)(nil)
