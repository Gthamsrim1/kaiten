package fs

import (
	"sync"
	"sync/atomic"

	gofuse "github.com/hanwen/go-fuse/v2/fs"
)

type KaitenFS struct {
	Root    *Directory
	ID      atomic.Uint64
	mu      sync.Mutex
	mounted map[uint64]*gofuse.Inode
}

func New() *KaitenFS {
	fs := &KaitenFS{ mounted: make(map[uint64]*gofuse.Inode) }
	fs.Root = fs.newRoot()
	
	return fs
}

func (k *KaitenFS) Seed() {
	root := k.Root
	_, err := root.CreateFile("hello", Memory([]byte("Hello from KaitenFS!\n")))
	if err != nil {
		panic(err)
	}

	_, err = root.CreateFile("readme", Memory([]byte("Welcome to KaitenFS!\n")))
	if err != nil {
		panic(err)
	}

	_, err = root.CreateDirectory("bin")
	if err != nil {
		panic(err)
	}

	_, err = root.CreateDirectory("lib")
	if err != nil {
		panic(err)
	}

	_, err = root.CreateDirectory("usr")
	if err != nil {
		panic(err)
	}
}

func (k *KaitenFS) newRoot() *Directory {
	return &Directory{
		Node: Node{
			ID:   k.nextID(),
			Name: "/",
		},
		FS:       k,
		Children: make(map[string]FSNode),
	}
}

func (k *KaitenFS) nextID() uint64 {
	return k.ID.Add(1)
}
