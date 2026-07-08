package persist

import (
	"encoding/hex"
	"encoding/json"
	"os"
	"path/filepath"
)

func Load(path string) (*Filesystem, error) {
	data, err := os.ReadFile(filepath.Join(path, "metadata.json"))
	if err != nil {
		return nil, err
	}

	var meta Metadata
	if err := json.Unmarshal(data, &meta); err != nil {
		return nil, err
	}

	fs := &Filesystem{
		NextID:  meta.NextID,
		Nodes:   meta.Nodes,
		Objects: make([]Object, 0, len(meta.Objects)),
	}

	objectDir := filepath.Join(path, "objects")

	for _, ref := range meta.Objects {
		data, err := os.ReadFile(filepath.Join(objectDir, hex.EncodeToString(ref.ID[:])))
		if err != nil {
			return nil, err
		}

		fs.Objects = append(fs.Objects, Object{
			ID:   ref.ID,
			Data: data,
		})
	}

	return fs, nil
}
