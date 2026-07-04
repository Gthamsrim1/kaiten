package tree

import (
	"context"
	"syscall"
	"time"

	"github.com/Gthamsrim1/kaiten/internal/content"
	"github.com/Gthamsrim1/kaiten/internal/errs"
	"github.com/Gthamsrim1/kaiten/internal/node"
	gofuse "github.com/hanwen/go-fuse/v2/fs"
	"github.com/hanwen/go-fuse/v2/fuse"
)

type File struct {
	gofuse.Inode

	node.Node
	Content content.Content
}

func (f *File) GetNode() *node.Node {
	return &f.Node
}

func (f *File) Getattr(ctx context.Context, fh gofuse.FileHandle, out *fuse.AttrOut) syscall.Errno {
	out.Mode = f.Node.Mode
	out.Size = f.Content.Size()
	out.Uid = f.Node.UID
	out.Gid = f.Node.GID
	out.SetTimes(&f.Node.Atime, &f.Node.Mtime, &f.Node.Ctime)
	return 0
}

func (f *File) Open(ctx context.Context, flags uint32) (gofuse.FileHandle, uint32, syscall.Errno) {
	return nil, fuse.FOPEN_DIRECT_IO, 0
}

func (f *File) Write(ctx context.Context, fh gofuse.FileHandle, data []byte, off int64) (uint32, syscall.Errno) {
	n, err := f.Content.Write(off, data)
	if err != nil {
		return 0, errs.ToErrno(err)
	}

	now := time.Now()
	f.Node.Mtime = now
	f.Node.Ctime = now

	return uint32(n), 0
}

func (f *File) Read(ctx context.Context, fh gofuse.FileHandle, dest []byte, off int64) (fuse.ReadResult, syscall.Errno) {
	n, err := f.Content.Read(off, dest)
	if err != nil {
		return nil, syscall.EIO
	}
	f.Node.Atime = time.Now()

	return fuse.ReadResultData(dest[:n]), 0
}

var _ gofuse.NodeGetattrer = (*File)(nil)
var _ gofuse.NodeOpener = (*File)(nil)
var _ gofuse.NodeReader = (*File)(nil)
var _ gofuse.NodeWriter = (*File)(nil)