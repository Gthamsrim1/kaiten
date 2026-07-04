package store

import (
	"encoding/hex"
	"fmt"
	"os"
	"path/filepath"
)

// Sweep deletes chunk bodies with no live reference and returns the count
// freed. It is the only operation that deletes bodies.
func (s *Store) Sweep() (freed int, err error) {
	s.mu.Lock()
	defer s.mu.Unlock()

	// Flush first so a crash mid-sweep leaves an accurate table and a re-run
	// simply finishes the job.
	if err := s.flushLocked(); err != nil {
		return 0, err
	}

	dir := filepath.Join(s.root, blobsDir)
	entries, err := os.ReadDir(dir)
	if err != nil {
		return 0, fmt.Errorf("reading blobs dir: %w", err)
	}

	for _, e := range entries {
		if e.IsDir() {
			continue
		}
		name := e.Name()
		if len(name) < 2 || name[0] == '.' {
			continue
		}
		raw, decErr := hex.DecodeString(name)
		if decErr != nil || len(raw) != len(Hash{}) {
			continue
		}
		var h Hash
		copy(h[:], raw)

		if s.refs[h] == 0 {
			if rmErr := os.Remove(filepath.Join(dir, name)); rmErr != nil {
				return freed, fmt.Errorf("removing chunk %s: %w", name, rmErr)
			}
			freed++
		}
	}
	return freed, nil
}