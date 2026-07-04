package fs

import (
	"context"
	"syscall"
	"time"

	gofuse "github.com/hanwen/go-fuse/v2/fs"
	"github.com/hanwen/go-fuse/v2/fuse"
)

type File struct {
	gofuse.Inode

	Node
	Content Content
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

func (f *File) Write(data []byte, off int64) (int, error) {
	n, err := f.Content.Write(off, data)
	if err == nil {
		now := time.Now()
		f.Node.Mtime = now
		f.Node.Ctime = now
	}
	return n, err
}

func (f *File) Read(ctx context.Context, fh gofuse.FileHandle, dest []byte, off int64) (fuse.ReadResult, syscall.Errno) {
	n, err := f.Content.Read(off, dest)
	if err != nil {
		return nil, syscall.EIO
	}
	f.Node.Atime = time.Now()

	return fuse.ReadResultData(dest[:n]), 0
}