package persist

import (
	"encoding/hex"
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
		NextID:  snapshot.NextID,
		Nodes:   snapshot.Nodes,
		Objects: make([]ObjectRef, len(snapshot.Objects)),
	}

	for i, obj := range snapshot.Objects {
		meta.Objects[i] = ObjectRef{ID: obj.ID}
	}

	data, err := json.MarshalIndent(meta, "", "  ")
	if err != nil {
		return err
	}

	if err := writeAtomic(filepath.Join(path, "metadata.json"), data, 0644); err != nil {
		return err
	}

	for _, object := range snapshot.Objects {
		objectPath := filepath.Join(objectDir, hex.EncodeToString(object.ID[:]))

		if _, err := os.Stat(objectPath); err == nil {
			continue
		} else if !os.IsNotExist(err) {
			return err
		}

		if err := writeAtomic(objectPath, object.Data, 0644); err != nil {
			return err
		}
	}

	return nil
}

func syncDir(path string) error {
	dir, err := os.Open(path)
	if err != nil {
		return err
	}
	defer dir.Close()

	return dir.Sync()
}

func writeAtomic(path string, data []byte, perm os.FileMode) error {
	tmp := path + ".tmp"

	f, err := os.OpenFile(tmp, os.O_CREATE|os.O_TRUNC|os.O_WRONLY, perm)
	if err != nil {
		return err
	}

	success := false
	defer func() {
		if !success {
			_ = os.Remove(tmp)
		}
	}()

	if _, err := f.Write(data); err != nil {
		f.Close()
		return err
	}

	if err := f.Sync(); err != nil {
		f.Close()
		return err
	}

	if err := f.Close(); err != nil {
		return err
	}

	if err := os.Rename(tmp, path); err != nil {
		return err
	}

	if err := syncDir(filepath.Dir(path)); err != nil {
		return err
	}

	success = true
	return nil
}
