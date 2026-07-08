package persist

import (
	"os"
	"path/filepath"
)

func GC(repo string) error {
	fs, err := Load(repo)
	if err != nil {
		return err
	}

	live := make(map[string]struct{})

	for _, node := range fs.Nodes {
		if node.ObjectID != nil {
			live[*node.ObjectID] = struct{}{}
		}
	}

	objectDir := filepath.Join(repo, "objects")

	entries, err := os.ReadDir(objectDir)
	if err != nil {
		return err
	}

	for _, entry := range entries {
		if entry.IsDir() {
			continue
		}

		id := entry.Name()

		if _, ok := live[id]; ok {
			continue
		}

		if err := os.Remove(filepath.Join(objectDir, id)); err != nil {
			return err
		}
	}

	return nil
}