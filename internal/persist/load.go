package persist

import (
	"encoding/json"
	"os"
	"path/filepath"
)

func Load(path string) (*Filesystem, *Repository, error) {
	data, err := os.ReadFile(filepath.Join(path, "metadata.json"))
	if err != nil {
		return nil, nil, err
	}

	var meta Metadata
	if err := json.Unmarshal(data, &meta); err != nil {
		return nil, nil, err
	}

	fs := &Filesystem{
		NextID:  meta.NextID,
		Nodes:   meta.Nodes,
	}

	repo := &Repository{
        Path: path,
    }

	return fs, repo, nil
}
