// Copyright (c) 2026 Gautham Sriram All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package runtime

func Run(cfg Config) error {
	rootfs, cleanup, err := mountSnapshot(cfg)
	if err != nil {
		return err
	}
	defer cleanup()

	return start(rootfs, cfg)
}
