package tree

import (
	gofuse "github.com/hanwen/go-fuse/v2/fs"
)

// newTestFS returns a KaitenFS whose Root is wired into a go-fuse bridge,
// so Lookup/Mount/Create/Mkdir work without an actual kernel mount.
func newTestFS() *KaitenFS {
	k := New()
	_ = gofuse.NewNodeFS(k.Root, &gofuse.Options{})
	return k
}