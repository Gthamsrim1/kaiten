package mountfs

import (
	"fmt"
	"os"

	kfs "github.com/Gthamsrim1/kaiten/internal/tree"
	gofuse "github.com/hanwen/go-fuse/v2/fs"
	"github.com/hanwen/go-fuse/v2/fuse"
)

// Mount creates a fresh kaiten filesystem, seeds it, and mounts it at mountPoint.
// It returns the server (for Unmount/Wait) and whether the mountpoint dir was created.
func Mount(mountPoint string, debug bool) (server *fuse.Server, createdMountPoint bool, err error) {
	createdMountPoint, err = ensureMountPoint(mountPoint)
	if err != nil {
		return nil, false, err
	}

	kaitenFS := kfs.New()
	kaitenFS.Seed()

	server, err = gofuse.Mount(
		mountPoint,
		kaitenFS.Root,
		&gofuse.Options{
			MountOptions: fuse.MountOptions{
				Debug: debug,
			},
		},
	)
	if err != nil {
		if createdMountPoint {
			os.Remove(mountPoint)
		}
		return nil, false, err
	}

	return server, createdMountPoint, nil
}

func ensureMountPoint(path string) (created bool, err error) {
	info, err := os.Stat(path)
	switch {
	case os.IsNotExist(err):
		if err := os.MkdirAll(path, 0755); err != nil {
			return false, fmt.Errorf("creating mountpoint: %w", err)
		}
		return true, nil

	case err != nil:
		return false, fmt.Errorf("checking mountpoint: %w", err)

	case !info.IsDir():
		return false, fmt.Errorf("mountpoint %q exists and is not a directory", path)
	}

	entries, err := os.ReadDir(path)
	if err != nil {
		return false, fmt.Errorf("reading mountpoint: %w", err)
	}
	if len(entries) > 0 {
		return false, fmt.Errorf("mountpoint %q is not empty", path)
	}

	return false, nil
}