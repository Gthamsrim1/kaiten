package tree

import (
	"fmt"

	"github.com/Gthamsrim1/kaiten/internal/content"
	"github.com/Gthamsrim1/kaiten/internal/node"
	"github.com/Gthamsrim1/kaiten/internal/persist"
)

func Restore(repo string) (*KaitenFS, error) {
	pfs, loader, err := persist.Load(repo)
	if err != nil {
		return nil, err
	}

	fs := New()
	fs.ID.Store(pfs.NextID)

	nodes := make(map[uint64]node.FSNode, len(pfs.Nodes))

	for _, n := range pfs.Nodes {

		switch n.Type {
		case persist.TypeDirectory:
			nodes[n.ID] = &Directory{
				Node:     restoreNode(n),
				FS:       fs,
				Children: make(map[string]node.FSNode),
			}

		case persist.TypeFile:
			nodes[n.ID] = &File{
				Node: restoreNode(n),
				FS:   fs,
				Content: content.Lazy(
					n.Chunks,
					loader,
				),
			}
		}
	}

	for _, n := range pfs.Nodes {
		current := nodes[n.ID]

		if n.ParentID == 0 {
			root := current.(*Directory)
			fs.Root = root
			root.Node.Parent = nil
			continue
		}

		parentNode, ok := nodes[n.ParentID]
		if !ok {
			return nil, fmt.Errorf("parent %d not found", n.ParentID)
		}

		parent, ok := parentNode.(*Directory)
		if !ok {
			return nil, fmt.Errorf("parent %d is not a directory", n.ParentID)
		}

		current.GetNode().Parent = parent

		parent.mu.Lock()
		parent.Children[n.Name] = current
		parent.mu.Unlock()
	}

	return fs, nil
}

func restoreNode(n persist.Node) node.Node {
	return node.Node{
		ID:     n.ID,
		Name:   n.Name,
		Mode:   n.Mode,
		Chunks: n.Chunks,
		UID:    n.UID,
		GID:    n.GID,
		Nlink:  n.Nlink,
		Atime:  n.Atime,
		Mtime:  n.Mtime,
		Ctime:  n.Ctime,
	}
}
