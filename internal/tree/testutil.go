package tree

import (
	"context"

	gofuse "github.com/hanwen/go-fuse/v2/fs"
	"github.com/hanwen/go-fuse/v2/fuse"
)

// newTestFS returns a KaitenFS whose Root is wired into a go-fuse bridge,
// so Lookup/Mount/Create/Mkdir work without an actual kernel mount.
func newTestFS() *KaitenFS {
	k := New()
	_ = gofuse.NewNodeFS(k.Root, &gofuse.Options{})
	return k
}

func testContext() context.Context {
	return fuse.NewContext(context.Background(), &fuse.Caller{
		Owner: fuse.Owner{
			Uid: 0,
			Gid: 0,
		},
		Pid: 1,
	})
}