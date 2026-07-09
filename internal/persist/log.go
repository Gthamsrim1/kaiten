// Copyright (c) 2026 Gautham Sriram All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package persist

import (
	"path/filepath"
)

func Log(repo string) ([]SnapshotInfo, error) {
	head, err := CurrentSnapshotID(repo)
	if err != nil {
		return nil, err
	}

	var history []SnapshotInfo

	for {
		info, err := ReadSnapshotInfo(
			filepath.Join(repo, "snapshots", head+".json"),
		)
		if err != nil {
			return nil, err
		}

		info.IsHEAD = len(history) == 0
		history = append(history, *info)

		if info.ParentID == nil {
			break
		}

		head = *info.ParentID
	}

	return history, nil
}
