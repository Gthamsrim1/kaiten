// Copyright (c) 2026 Gautham Sriram All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package runtime

import (
	"os"
	"path/filepath"

	"github.com/Gthamsrim1/kaiten/internal/mountfs"
	"github.com/Gthamsrim1/kaiten/internal/tree"
)

func mountSnapshot(cfg Config) (string, func() error, error) {
	base := filepath.Join(os.TempDir(), "kaiten", "runtime")

	if err := os.MkdirAll(base, 0755); err != nil {
		return "", nil, err
	}

	runtimeDir, err := os.MkdirTemp(base, "")
	if err != nil {
		return "", nil, err
	}

	fs, err := tree.Restore(cfg.Repo)
	if err != nil {
		os.RemoveAll(runtimeDir)
		return "", nil, err
	}

	server, _, err := mountfs.MountFS(fs, runtimeDir, false)
	if err != nil {
		os.RemoveAll(runtimeDir)
		return "", nil, err
	}

	cleanup := func() error {
		server.Unmount()
		return os.RemoveAll(runtimeDir)
	}

	return runtimeDir, cleanup, nil
}
