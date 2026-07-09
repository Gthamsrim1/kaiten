// Copyright (c) 2026 Gautham Sriram All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package runtime

import (
	"io/fs"
	"os"
	"path/filepath"

	"github.com/Gthamsrim1/kaiten/internal/content"
	"github.com/Gthamsrim1/kaiten/internal/persist"
	"github.com/Gthamsrim1/kaiten/internal/tree"
)

func Import(repo, root string) error {
	kfs := tree.New()

	dirs := map[string]*tree.Directory{
		".": kfs.Root,
	}

	err := filepath.WalkDir(root, func(path string, d fs.DirEntry, err error) error {
		if err != nil {
			return err
		}

		rel, err := filepath.Rel(root, path)
		if err != nil {
			return err
		}

		if rel == "." {
			return nil
		}

		parentRel := filepath.Dir(rel)
		parent := dirs[parentRel]

		info, err := d.Info()
		if err != nil {
			return err
		}

		if d.IsDir() {
			dir, err := parent.CreateDirectory(filepath.Base(rel), uint32(info.Mode().Perm()))
			if err != nil {
				return err
			}

			dirs[rel] = dir
			return nil
		}

		if d.Type().IsRegular() {
			data, err := os.ReadFile(path)
			if err != nil {
				return err
			}

			_, err = parent.CreateFile(filepath.Base(rel), content.Memory(data), uint32(info.Mode().Perm()))
			if err != nil {
				return err
			}

			return nil
		}

		if d.Type()&fs.ModeSymlink != 0 {
			target, err := os.Readlink(path)
			if err != nil {
				return err
			}

			_, err = parent.CreateSymlink(filepath.Base(rel), target)
			return err
		}

		return nil
	})
	if err != nil {
		return err
	}

	return persist.Commit(repo, kfs)
}
