package store

import (
	"bufio"
	"encoding/hex"
	"encoding/json"
	"errors"
	"fmt"
	"os"
)

// persistedRefs is JSON for debuggability; the table is tiny next to chunk
// data, so hand-inspection beats a binary format's size win.
type persistedRefs struct {
	Refs map[string]int `json:"refs"`
}

// Flush should also be called at commit points (e.g. after ingesting an
// image), not just relied on via the automatic threshold.
func (s *Store) Flush() error {
	s.mu.Lock()
	defer s.mu.Unlock()
	return s.flushLocked()
}

// flushLocked requires s.mu held.
func (s *Store) flushLocked() error {
	out := persistedRefs{Refs: make(map[string]int, len(s.refs))}
	for h, c := range s.refs {
		out.Refs[hex.EncodeToString(h[:])] = c
	}

	tmp, err := os.CreateTemp(s.root, ".refs-*.tmp")
	if err != nil {
		return fmt.Errorf("creating temp meta: %w", err)
	}
	tmpName := tmp.Name()
	defer os.Remove(tmpName)

	w := bufio.NewWriter(tmp)
	if err := json.NewEncoder(w).Encode(&out); err != nil {
		tmp.Close()
		return fmt.Errorf("encoding meta: %w", err)
	}
	if err := w.Flush(); err != nil {
		tmp.Close()
		return fmt.Errorf("flushing meta buffer: %w", err)
	}
	if err := tmp.Sync(); err != nil {
		tmp.Close()
		return fmt.Errorf("syncing meta: %w", err)
	}
	if err := tmp.Close(); err != nil {
		return fmt.Errorf("closing meta: %w", err)
	}
	if err := os.Rename(tmpName, s.metaPath); err != nil {
		return fmt.Errorf("committing meta: %w", err)
	}

	s.dirtyCnt = 0
	return nil
}

// loadMeta runs only from Open, before the store is shared, so it takes no lock.
func (s *Store) loadMeta() error {
	data, err := os.ReadFile(s.metaPath)
	if err != nil {
		if errors.Is(err, os.ErrNotExist) {
			return nil
		}
		return fmt.Errorf("reading meta: %w", err)
	}

	var in persistedRefs
	if err := json.Unmarshal(data, &in); err != nil {
		return fmt.Errorf("parsing meta: %w", err)
	}
	for hexHash, count := range in.Refs {
		raw, decErr := hex.DecodeString(hexHash)
		if decErr != nil || len(raw) != len(Hash{}) {
			return fmt.Errorf("meta contains invalid hash %q", hexHash)
		}
		var h Hash
		copy(h[:], raw)
		s.refs[h] = count
	}
	return nil
}