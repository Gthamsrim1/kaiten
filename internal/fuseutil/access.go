// Copyright (c) 2026 Gautham Sriram All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package fuseutil

import (
	"context"
	"syscall"

	"github.com/Gthamsrim1/kaiten/internal/node"
)

func RequireAccess(ctx context.Context, n *node.Node, mask uint32) syscall.Errno {
	uid, gid, errno := Caller(ctx)
	if errno != 0 {
		return errno
	}

	return n.CheckAccess(uid, gid, mask)
}
