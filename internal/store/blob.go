package store

import (
	"encoding/hex"
	"errors"
	"fmt"
	"os"
	"path/filepath"

	"github.com/Gthamsrim1/kaiten/internal/errs"
)

// Get does not verify bytes against the hash -- integrity is checked at
// ingest (before Put) to keep reads cheap.
func (s *Store) Get(h Hash) ([]byte, error) {
	data, err := os.ReadFile(s.blobPath(h))
	if err != nil {
		if errors.Is(err, os.ErrNotExist) {
			return nil, fmt.Errorf("chunk %s: %w", hex.EncodeToString(h[:]), errs.ErrObjectMissing)
		}
		return nil, err
	}
	return data, nil
}

// Put stores data under its content hash, which MUST be its sha256. It is
// idempotent and does not touch refcounts (see IncRef/DecRef).
func (s *Store) Put(h Hash, data []byte) error {
	path := s.blobPath(h)
	if _, err := os.Stat(path); err == nil {
		return nil
	}

	// Temp-file-then-rename so a concurrent reader never sees a partial body.
	tmp, err := os.CreateTemp(filepath.Join(s.root, blobsDir), ".tmp-*")
	if err != nil {
		return fmt.Errorf("creating temp blob: %w", err)
	}
	tmpName := tmp.Name()
	defer os.Remove(tmpName)

	if _, err := tmp.Write(data); err != nil {
		tmp.Close()
		return fmt.Errorf("writing blob: %w", err)
	}
	if err := tmp.Sync(); err != nil {
		tmp.Close()
		return fmt.Errorf("syncing blob: %w", err)
	}
	if err := tmp.Close(); err != nil {
		return fmt.Errorf("closing blob: %w", err)
	}
	if err := os.Rename(tmpName, path); err != nil {
		return fmt.Errorf("committing blob: %w", err)
	}
	return nil
}