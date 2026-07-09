package persist

import (
	"encoding/hex"
	"os"
	"path/filepath"
)

type Repository struct {
    Path string
}

func (r *Repository) Load(hash [32]byte) ([]byte, error) {
    path := filepath.Join(r.Path, "objects", hex.EncodeToString(hash[:]))

    return os.ReadFile(path)
}