package fuseutil

import (
	"github.com/Gthamsrim1/kaiten/internal/node"
	"github.com/hanwen/go-fuse/v2/fuse"
)

func UpdateAttributes(in *fuse.SetAttrIn) node.AttrUpdate {
	var update node.AttrUpdate

	if mode, ok := in.GetMode(); ok {
		update.Mode = &mode
	}

	if uid, ok := in.GetUID(); ok {
		update.UID = &uid
	}

	if gid, ok := in.GetGID(); ok {
		update.GID = &gid
	}

	if atime, ok := in.GetATime(); ok {
		update.ATime = &atime
	}

	if mtime, ok := in.GetMTime(); ok {
		update.MTime = &mtime
	}

	return update
}
