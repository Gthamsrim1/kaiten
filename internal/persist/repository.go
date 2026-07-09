// Copyright (c) 2026 Gautham Sriram All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package persist

import (
	"encoding/hex"
	"os"
	"path/filepath"
)

type Repository struct {
	Path string
}

func (r *Repository) Load(hash [32]byte) ([]byte, error) {
	path := filepath.Join(r.Path, "objects", hex.EncodeToString(hash[:]))

	return os.ReadFile(path)
}
