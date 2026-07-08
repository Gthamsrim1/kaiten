package tree

import (
	"sync"
	"sync/atomic"
	"syscall"

	"github.com/Gthamsrim1/kaiten/internal/content"
	"github.com/Gthamsrim1/kaiten/internal/node"
	gofuse "github.com/hanwen/go-fuse/v2/fs"
)

type KaitenFS struct {
	Root    *Directory
	ID      atomic.Uint64
	dirty   atomic.Bool
	mu      sync.Mutex
	mounted map[uint64]*gofuse.Inode
}

func New() *KaitenFS {
	fs := &KaitenFS{mounted: make(map[uint64]*gofuse.Inode)}
	fs.Root = fs.newRoot()

	return fs
}

func (k *KaitenFS) Seed() {
	root := k.Root
	_, err := root.CreateFile("hello", content.Memory([]byte("Hello from KaitenFS!\n")), 0644)
	if err != nil {
		panic(err)
	}

	_, err = root.CreateFile("readme", content.Memory([]byte("Homura did nothing wrong!\n")), 0644)
	if err != nil {
		panic(err)
	}

	_, err = root.CreateDirectory("bin", 0755)
	if err != nil {
		panic(err)
	}

	_, err = root.CreateDirectory("lib", 0755)
	if err != nil {
		panic(err)
	}

	_, err = root.CreateDirectory("usr", 0755)
	if err != nil {
		panic(err)
	}
}

func (k *KaitenFS) newRoot() *Directory {
	return &Directory{
		Node:     newNode(k, "/", nil, syscall.S_IFDIR, 0755),
		FS:       k,
		Children: make(map[string]node.FSNode),
	}
}

func (k *KaitenFS) nextID() uint64 {
	return k.ID.Add(1)
}

func (k *KaitenFS) CurrentID() uint64 {
	return k.ID.Load()
}
