package tree

import (
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"sort"

	"github.com/Gthamsrim1/kaiten/internal/node"
	"github.com/Gthamsrim1/kaiten/internal/persist"
)

func (fs *KaitenFS) snapshotNode(n node.FSNode, parentID uint64, snap *persist.Filesystem, objects map[string]struct{}) error {
	meta := n.GetNode()

	record := persist.Node{
		ID:       meta.ID,
		ParentID: parentID,

		Name: meta.Name,

		Mode:  meta.Mode,
		UID:   meta.UID,
		GID:   meta.GID,
		Nlink: meta.Nlink,

		Atime: meta.Atime,
		Mtime: meta.Mtime,
		Ctime: meta.Ctime,

		ObjectID: meta.ObjectID,
	}

	switch v := n.(type) {
	case *Directory:
		record.Type = persist.TypeDirectory
		snap.Nodes = append(snap.Nodes, record)

		names := make([]string, 0, len(v.Children))
		for name := range v.Children {
			names = append(names, name)
		}
		sort.Strings(names)

		for _, name := range names {
			if err := fs.snapshotNode(v.Children[name], meta.ID, snap, objects); err != nil {
				return err
			}
		}

	case *File:
		record.Type = persist.TypeFile
		data := v.Content.Bytes()

		sum := sha256.Sum256(data)
		hash := hex.EncodeToString(sum[:])

		record.ObjectID = &hash
		snap.Nodes = append(snap.Nodes, record)

		if _, ok := objects[hash]; !ok {
			objects[hash] = struct{}{}

			snap.Objects = append(snap.Objects, persist.Object{
				ID:   hash,
				Data: data,
			})
		}

	default:
		return fmt.Errorf("unknown node type %T", n)
	}

	return nil
}

func (fs *KaitenFS) Snapshot() (*persist.Filesystem, error) {
	snap := &persist.Filesystem{
		NextID: fs.CurrentID(),
	}

	objects := make(map[string]struct{})

	if err := fs.snapshotNode(fs.Root, 0, snap, objects); err != nil {
		return nil, err
	}

	return snap, nil
}