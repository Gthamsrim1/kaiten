package tree

import (
	"context"
	"fmt"
	"sync"
	"syscall"

	"github.com/Gthamsrim1/kaiten/internal/content"
	"github.com/Gthamsrim1/kaiten/internal/errs"
	"github.com/Gthamsrim1/kaiten/internal/fuseutil"
	"github.com/Gthamsrim1/kaiten/internal/node"
	gofuse "github.com/hanwen/go-fuse/v2/fs"
	"github.com/hanwen/go-fuse/v2/fuse"
	"golang.org/x/sys/unix"
)

type Directory struct {
	gofuse.Inode

	node.Node
	FS *KaitenFS

	mu       sync.RWMutex
	Children map[string]node.FSNode
}

func (d *Directory) GetNode() *node.Node {
	return &d.Node
}

func (d *Directory) CreateFile(name string, content content.Content, perm uint32) (*File, error) {
	d.FS.MarkDirty()
	return d.FS.createFile(name, d, content, perm)
}

func (d *Directory) DeleteFile(name string) error {
	d.FS.MarkDirty()
	return d.FS.deleteFile(name, d)
}

func (d *Directory) CreateDirectory(name string, perm uint32) (*Directory, error) {
	d.FS.MarkDirty()
	return d.FS.createDirectory(name, d, perm)
}

func (d *Directory) DeleteDirectory(name string) error {
	d.FS.MarkDirty()
	return d.FS.deleteDirectory(name, d)
}

func (d *Directory) Mount(ctx context.Context, node node.FSNode) *gofuse.Inode {
	id := node.GetNode().ID

	d.FS.mu.Lock()
	defer d.FS.mu.Unlock()

	if inode, ok := d.FS.mounted[id]; ok {
		return inode
	}

	var embed gofuse.InodeEmbedder

	switch v := node.(type) {
	case *File:
		embed = v
	case *Directory:
		embed = v
	default:
		panic(fmt.Sprintf("unsupported node type %T", node))
	}

	inode := d.NewPersistentInode(ctx, embed, gofuse.StableAttr{
		Mode: node.GetNode().Mode,
	})

	d.FS.mounted[id] = inode
	d.AddChild(node.GetNode().Name, inode, true)

	return inode
}

func (d *Directory) Create(ctx context.Context, name string, flags uint32, mode uint32, out *fuse.EntryOut) (node *gofuse.Inode, fh gofuse.FileHandle, fuseFlags uint32, errno syscall.Errno) {
	if errno := fuseutil.RequireAccess(ctx, &d.Node, unix.W_OK | unix.X_OK); errno != 0 {
		return nil, nil, 0, errno
	}
	
	file, err := d.CreateFile(name, content.Memory(nil), 0644)
	if err != nil {
		return nil, nil, 0, errs.ToErrno(err)
	}

	file.Node.Mode = syscall.S_IFREG | (mode &^ syscall.S_IFMT)

	inode := d.Mount(ctx, file)

	var attrOut fuse.AttrOut
	file.Getattr(ctx, nil, &attrOut)
	out.Attr = attrOut.Attr

	return inode, nil, fuse.FOPEN_DIRECT_IO, 0
}

func (d *Directory) Mkdir(ctx context.Context, name string, mode uint32, out *fuse.EntryOut) (node *gofuse.Inode, errno syscall.Errno) {
	if errno := fuseutil.RequireAccess(ctx, &d.Node, unix.W_OK | unix.X_OK); errno != 0 {
		return nil, errno
	}

	dir, err := d.CreateDirectory(name, 0755)
	if err != nil {
		return nil, errs.ToErrno(err)
	}

	dir.Node.Mode = syscall.S_IFDIR | (mode &^ syscall.S_IFMT)

	inode := d.Mount(ctx, dir)

	var attrOut fuse.AttrOut
	dir.Getattr(ctx, inode, &attrOut)
	out.Attr = attrOut.Attr

	return inode, 0
}

func (d *Directory) Getattr(ctx context.Context, f gofuse.FileHandle, out *fuse.AttrOut) syscall.Errno {
	out.Mode = d.Node.Mode
	out.Uid = d.Node.UID
	out.Gid = d.Node.GID
	out.SetTimes(&d.Node.Atime, &d.Node.Mtime, &d.Node.Ctime)
	return 0
}

func (d *Directory) Lookup(ctx context.Context, name string, out *fuse.EntryOut) (*gofuse.Inode, syscall.Errno) {
	if errno := fuseutil.RequireAccess(ctx, &d.Node, unix.X_OK); errno != 0 {
		return nil, errno
	}

	d.mu.RLock()
	node, ok := d.Children[name]
	d.mu.RUnlock()

	if !ok {
		return nil, syscall.ENOENT
	}

	inode := d.Mount(ctx, node)

	var attrOut fuse.AttrOut
	switch n := node.(type) {
	case *File:
		n.Getattr(ctx, nil, &attrOut)
	case *Directory:
		n.Getattr(ctx, inode, &attrOut)
	}
	out.Attr = attrOut.Attr

	return inode, 0
}

func (d *Directory) Readdir(ctx context.Context) (gofuse.DirStream, syscall.Errno) {
	if errno := fuseutil.RequireAccess(ctx, &d.Node, unix.R_OK | unix.X_OK); errno != 0 {
		return nil, errno
	}

	d.mu.RLock()
	defer d.mu.RUnlock()

	entries := make([]fuse.DirEntry, 0, len(d.Children))

	for name, node := range d.Children {
		entries = append(entries, fuse.DirEntry{
			Name: name,
			Mode: node.GetNode().Mode,
			Ino:  node.GetNode().ID,
		})
	}

	return gofuse.NewListDirStream(entries), 0
}

func (d *Directory) Unlink(ctx context.Context, name string) syscall.Errno {
	if errno := fuseutil.RequireAccess(ctx, &d.Node, unix.W_OK | unix.X_OK); errno != 0 {
		return errno
	}

	d.mu.RLock()
	node, ok := d.Children[name]
	d.mu.RUnlock()
	if !ok {
		return syscall.ENOENT
	}

	if err := d.DeleteFile(name); err != nil {
		return errs.ToErrno(err)
	}

	id := node.GetNode().ID

	d.RmChild(name)

	d.FS.mu.Lock()
	delete(d.FS.mounted, id)
	d.FS.mu.Unlock()

	return 0
}

func (d *Directory) Setattr(ctx context.Context, f gofuse.FileHandle, in *fuse.SetAttrIn, out *fuse.AttrOut) syscall.Errno {
	if errno := fuseutil.RequireAccess(ctx, &d.Node, unix.W_OK); errno != 0 {
		return errno
	}

	changed := d.Node.UpdateAttr(fuseutil.UpdateAttributes(in))

	out.Mode = d.Node.Mode
	out.Uid = d.Node.UID
	out.Gid = d.Node.GID
	out.SetTimes(&d.Node.Atime, &d.Node.Mtime, &d.Node.Ctime)

	if changed {
		d.FS.MarkDirty()
	}

	return 0
}

func (d *Directory) Rmdir(ctx context.Context, name string) syscall.Errno {
	if errno := fuseutil.RequireAccess(ctx, &d.Node, unix.W_OK | unix.X_OK); errno != 0 {
		return errno
	}

	d.mu.RLock()
	node, ok := d.Children[name]
	d.mu.RUnlock()
	if !ok {
		return syscall.ENOENT
	}

	if err := d.DeleteDirectory(name); err != nil {
		return errs.ToErrno(err)
	}

	id := node.GetNode().ID

	d.RmChild(name)

	d.FS.mu.Lock()
	delete(d.FS.mounted, id)
	d.FS.mu.Unlock()

	return 0
}

func (d *Directory) Rename(ctx context.Context, name string, newParent gofuse.InodeEmbedder, newName string, flags uint32) syscall.Errno {
	if errno := fuseutil.RequireAccess(ctx, &d.Node, unix.W_OK | unix.X_OK); errno != 0 {
		return errno
	}

	dir, ok := newParent.(*Directory)
	if !ok {
		return syscall.EIO
	}

	if errno := fuseutil.RequireAccess(ctx, &dir.Node, unix.W_OK | unix.X_OK); errno != 0 {
		return errno
	}

	if err := d.FS.rename(d, dir, name, newName); err != nil {
		return errs.ToErrno(err)
	}

	d.RmChild(name)
	if child := dir.GetChild(newName); child != nil {
		dir.RmChild(newName)
	}

	dir.mu.RLock()
	node, ok := dir.Children[newName]
	dir.mu.RUnlock()
	if !ok {
		return syscall.EIO
	}

	d.FS.mu.Lock()
	mounted, exists := d.FS.mounted[node.GetNode().ID]
	d.FS.mu.Unlock()

	if exists {
		dir.AddChild(newName, mounted, true)
	}

	d.FS.MarkDirty()

	return 0
}

func (d *Directory) Access(ctx context.Context, mask uint32) syscall.Errno {
	return fuseutil.RequireAccess(ctx, &d.Node, mask)
}

var _ gofuse.NodeCreater = (*Directory)(nil)
var _ gofuse.NodeMkdirer = (*Directory)(nil)
var _ gofuse.NodeGetattrer = (*Directory)(nil)
var _ gofuse.NodeSetattrer = (*Directory)(nil)
var _ gofuse.NodeLookuper = (*Directory)(nil)
var _ gofuse.NodeReaddirer = (*Directory)(nil)
var _ gofuse.NodeUnlinker = (*Directory)(nil)
var _ gofuse.NodeRmdirer = (*Directory)(nil)
var _ gofuse.NodeRenamer = (*Directory)(nil)
var _ gofuse.NodeAccesser = (*Directory)(nil)
