package mountfs

import (
	"errors"
	"fmt"
	"os"

	"github.com/Gthamsrim1/kaiten/internal/persist"
	kfs "github.com/Gthamsrim1/kaiten/internal/tree"
	gofuse "github.com/hanwen/go-fuse/v2/fs"
	"github.com/hanwen/go-fuse/v2/fuse"
)

func Mount(repo, mountPoint string, debug bool) (*kfs.KaitenFS, *fuse.Server, bool, error) {
	createdMountPoint, err := ensureMountPoint(mountPoint)
	if err != nil {
		return nil, nil, false, err
	}

	var kaitenFS *kfs.KaitenFS

	kaitenFS, err = kfs.Restore(repo)
	if err != nil {
		if errors.Is(err, os.ErrNotExist) {
			kaitenFS = kfs.New()
			kaitenFS.Seed()

			snap, err := kaitenFS.Snapshot()
			if err != nil {
				return nil, nil, false, err
			}

			if err := persist.Save(repo, snap); err != nil {
				return nil, nil, false, err
			}
		} else {
			return nil, nil, false, err
		}
	}

	server, err := gofuse.Mount(
		mountPoint,
		kaitenFS.Root,
		&gofuse.Options{
			MountOptions: fuse.MountOptions{
				Debug:      debug,
				AllowOther: true,
			},
		},
	)
	if err != nil {
		if createdMountPoint {
			_ = os.Remove(mountPoint)
		}
		return nil, nil, false, err
	}

	return kaitenFS, server, createdMountPoint, nil
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