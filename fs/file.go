package fs

import (
	"context"
	"syscall"

	gofuse "github.com/hanwen/go-fuse/v2/fs"
	"github.com/hanwen/go-fuse/v2/fuse"
)

type Content interface {
	Read(offset int64, p []byte) (int, error)
	Write(offset int64, p []byte) (int, error)
	Size() uint64
}

type MemoryContent struct {
	data []byte
}

type File struct {
	gofuse.Inode

	Node
	Content Content
}

// FILE
func (f *File) Getattr(ctx context.Context, fh gofuse.FileHandle, out *fuse.AttrOut) syscall.Errno {
	out.Mode = syscall.S_IFREG | 0644
	out.Size = f.Content.Size()
	return 0
}

func (f *File) Open(ctx context.Context, flags uint32) (gofuse.FileHandle, uint32, syscall.Errno) {
	return nil, fuse.FOPEN_DIRECT_IO, 0
}

func (f *File) Write(data []byte, off int64) (int, error) {
	return f.Content.Write(off, data)
}

func (f *File) Read(ctx context.Context, fh gofuse.FileHandle, dest []byte, off int64) (fuse.ReadResult, syscall.Errno) {
	n, err := f.Content.Read(off, dest)
	if err != nil {
		return nil, syscall.EIO
	}

	return fuse.ReadResultData(dest[:n]), 0
}