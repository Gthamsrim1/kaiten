package persist

import (
	"encoding/json"
	"os"
	"path/filepath"
)

func Save(path string, snapshot *Filesystem) error {
	if err := os.MkdirAll(path, 0755); err != nil {
		return err
	}

	objectDir := filepath.Join(path, "objects")
	if err := os.MkdirAll(objectDir, 0755); err != nil {
		return err
	}

	meta := Metadata{
		NextID: snapshot.NextID,
		Nodes:  snapshot.Nodes,
		Objects: make([]ObjectRef, len(snapshot.Objects)),
	}
	meta.Objects = make([]ObjectRef, len(snapshot.Objects))

	for i, obj := range snapshot.Objects {
		meta.Objects[i] = ObjectRef{ID: obj.ID}
	}

	data, err := json.MarshalIndent(meta, "", "  ")
	if err != nil {
		return err
	}

	if err := os.WriteFile(filepath.Join(path, "metadata.json"), data, 0644); err != nil {
		return err
	}

	for _, object := range snapshot.Objects {
		if err := os.WriteFile(filepath.Join(objectDir, object.ID), object.Data, 0644); err != nil {
			return err
		}
	}

	return nil
}