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
	kaitenFS, err := kfs.Restore(repo)
	if err != nil {
		if errors.Is(err, os.ErrNotExist) {
			kaitenFS = kfs.New()
			kaitenFS.Seed()

			if err := persist.Commit(repo, kaitenFS); err != nil {
				return nil, nil, false, err
			}
		} else {
			return nil, nil, false, err
		}
	}

	server, createdMountPoint, err := MountFS(kaitenFS, mountPoint, debug)
	if err != nil {
		return nil, nil, false, err
	}

	return kaitenFS, server, createdMountPoint, nil
}

func MountFS(fs *kfs.KaitenFS, mountPoint string, debug bool) (*fuse.Server, bool, error) {
	createdMountPoint, err := ensureMountPoint(mountPoint)
	if err != nil {
		return nil, false, err
	}

	server, err := gofuse.Mount(
		mountPoint,
		fs.Root,
		&gofuse.Options{
			MountOptions: fuse.MountOptions{
				Debug:      debug,
				AllowOther: true,
			},
		},
	)
	if err != nil {
		if createdMountPoint {
			_ = os.RemoveAll(mountPoint)
		}
		return nil, false, err
	}

	if err := server.WaitMount(); err != nil {
		server.Unmount()

		if createdMountPoint {
			_ = os.RemoveAll(mountPoint)
		}

		return nil, false, fmt.Errorf("mount failed: %w", err)
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