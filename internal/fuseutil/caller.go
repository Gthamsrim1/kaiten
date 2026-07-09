// Copyright (c) 2026 Gautham Sriram All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package fuseutil

import (
	"context"
	"syscall"

	"github.com/hanwen/go-fuse/v2/fuse"
)

func Caller(ctx context.Context) (uid, gid uint32, errno syscall.Errno) {
	caller, ok := fuse.FromContext(ctx)
	if !ok {
		return 0, 0, syscall.EIO
	}

	return caller.Uid, caller.Gid, 0
}
