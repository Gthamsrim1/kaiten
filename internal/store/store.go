package store

import (
	"encoding/hex"
	"fmt"
	"os"
	"path/filepath"
	"sync"
)

// Hash matches the chunk package's Chunk.Hash (sha256).
type Hash = [32]byte

const (
	blobsDir = "blobs"
	metaFile = "refs.json"
)

// Store is safe for concurrent use.
type Store struct {
	root     string
	metaPath string

	mu       sync.Mutex
	refs     map[Hash]int
	dirtyCnt int
}

// Open opens or creates a chunk store rooted at dir, loading any existing
// refcount table.
func Open(root string) (*Store, error) {
	if err := os.MkdirAll(filepath.Join(root, blobsDir), 0o755); err != nil {
		return nil, fmt.Errorf("creating store dirs: %w", err)
	}

	s := &Store{
		root:     root,
		metaPath: filepath.Join(root, metaFile),
		refs:     make(map[Hash]int),
	}
	if err := s.loadMeta(); err != nil {
		return nil, err
	}
	return s, nil
}

func (s *Store) blobPath(h Hash) string {
	return filepath.Join(s.root, blobsDir, hex.EncodeToString(h[:]))
}

// Has checks the filesystem, not the refcount table -- callers use it to
// decide whether bytes still need fetching.
func (s *Store) Has(h Hash) bool {
	_, err := os.Stat(s.blobPath(h))
	return err == nil
}

func (s *Store) RefCount(h Hash) int {
	s.mu.Lock()
	defer s.mu.Unlock()
	return s.refs[h]
}