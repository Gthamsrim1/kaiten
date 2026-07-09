package tree

import (
	"fmt"
	"sort"

	"github.com/Gthamsrim1/kaiten/internal/chunk"
	"github.com/Gthamsrim1/kaiten/internal/node"
	"github.com/Gthamsrim1/kaiten/internal/persist"
	"github.com/Gthamsrim1/kaiten/internal/store"
)

func (fs *KaitenFS) snapshotNode(n node.FSNode, parentID uint64, snap *persist.Snapshot, objects map[[32]byte]struct{}) error {
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
		data, err := v.Content.Bytes()
		if err != nil {
			return err
		}

		chunks, err := chunk.Split(data, chunk.Default)
		if err != nil {
			return err
		}

		for _, c := range chunks {
			record.Chunks = append(record.Chunks, store.ChunkRef{
				Hash:   c.Hash,
				Length: uint32(len(c.Data)),
			})

			if _, ok := objects[c.Hash]; ok {
				continue
			}

			objects[c.Hash] = struct{}{}

			snap.Objects = append(snap.Objects, persist.Object{
				ID:   c.Hash,
				Data: c.Data,
			})
		}

		snap.Nodes = append(snap.Nodes, record)
	
	case *Symlink:
		record.Type = persist.TypeSymlink
		record.Target = v.Target

		snap.Nodes = append(snap.Nodes, record)

	default:
		return fmt.Errorf("unknown node type %T", n)
	}

	return nil
}

func (fs *KaitenFS) Snapshot(ID string, parentID *string) (*persist.Snapshot, error) {
	snap := &persist.Snapshot{
		ID:       ID,
		ParentID: parentID,
		NextID:   fs.CurrentID(),
	}

	objects := make(map[[32]byte]struct{})

	if err := fs.snapshotNode(fs.Root, 0, snap, objects); err != nil {
		return nil, err
	}

	return snap, nil
}
