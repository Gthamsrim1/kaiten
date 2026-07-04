package fs

import (
	"context"
	"fmt"
	"syscall"

	gofuse "github.com/hanwen/go-fuse/v2/fs"
	"github.com/hanwen/go-fuse/v2/fuse"
)

type Directory struct {
	gofuse.Inode

	Node
	FS *KaitenFS

	Children map[string]FSNode
}

func (d *Directory) CreateFile(name string, content Content) (*File, error) {
	return d.FS.createFile(name, d, content)
}

func (d *Directory) CreateDirectory(name string) (*Directory, error) {
	return d.FS.createDirectory(name, d)
}

func (d *Directory) Mount(ctx context.Context, node FSNode) *gofuse.Inode  {
	id := node.GetNode().ID

	if inode, ok := d.FS.mounted[id]; ok {
		return inode
	}

	var (
		embed gofuse.InodeEmbedder
		mode  uint32
	)

	switch n := node.(type) {
	case *File:
		embed = n
		mode = syscall.S_IFREG

	case *Directory:
		embed = n
		mode = syscall.S_IFDIR

	default:
		panic(fmt.Sprintf("unsupported node type %T", node))
	}

	inode := d.NewPersistentInode(ctx, embed, gofuse.StableAttr{
		Mode: mode,
	})

	d.FS.mounted[id] = inode
	d.AddChild(node.GetNode().Name, inode, true)

	return inode
}

func (d *Directory) Create(ctx context.Context, name string, flags uint32, mode uint32, out *fuse.EntryOut) (node *gofuse.Inode, fh gofuse.FileHandle, fuseFlags uint32, errno syscall.Errno) {
	file, err := d.CreateFile(name, Memory(nil))
	if err != nil {
		return nil, nil, 0, ToErrno(err)
	}

	inode := d.Mount(ctx, file)

	return inode, nil, fuse.FOPEN_DIRECT_IO, 0
}

func (d *Directory) Mkdir(ctx context.Context, name string, mode uint32, out *fuse.EntryOut) (node *gofuse.Inode, errno syscall.Errno) {
	dir, err := d.CreateDirectory(name)
	if err != nil {
		return nil, ToErrno(err)
	}

	inode := d.Mount(ctx, dir)

	return inode, 0
}

func (d *Directory) Lookup(ctx context.Context, name string, out *fuse.EntryOut) (*gofuse.Inode, syscall.Errno) {
	node, ok := d.Children[name]
	if !ok {
		return nil, syscall.ENOENT
	}

	return d.Mount(ctx, node), 0
}

func (d *Directory) Readdir(ctx context.Context) (gofuse.DirStream, syscall.Errno) {
	entries := make([]fuse.DirEntry, 0, len(d.Children))

	for name, node := range d.Children {
		mode := uint32(0)

		switch node.(type) {
		case *Directory:
			mode = syscall.S_IFDIR

		case *File:
			mode = syscall.S_IFREG
		}

		entries = append(entries, fuse.DirEntry{
			Name: name,
			Mode: mode,
			Ino:  node.GetNode().ID,
		})
	}

	return gofuse.NewListDirStream(entries), 0
}

var _ gofuse.NodeCreater = (*Directory)(nil)
var _ gofuse.NodeMkdirer = (*Directory)(nil)
var _ gofuse.NodeLookuper = (*Directory)(nil)
var _ gofuse.NodeReaddirer = (*Directory)(nil)
