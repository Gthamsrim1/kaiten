package store

import (
	"encoding/hex"
	"fmt"

	"github.com/Gthamsrim1/kaiten/internal/errs"
)

// flushThreshold batches ref changes so ingesting a large image doesn't
// rewrite the whole table per chunk. A lost batch can only under-collect,
// never delete a live chunk, so batching is safe.
const flushThreshold = 256

func (s *Store) IncRef(h Hash) (int, error) {
	s.mu.Lock()
	defer s.mu.Unlock()

	s.refs[h]++
	s.dirtyCnt++
	if err := s.maybeFlushLocked(); err != nil {
		return s.refs[h], err
	}
	return s.refs[h], nil
}

// DecRef does not delete at zero -- Sweep does, so a chunk that briefly hits
// zero and is re-referenced isn't needlessly dropped and re-fetched.
func (s *Store) DecRef(h Hash) (int, error) {
	s.mu.Lock()
	defer s.mu.Unlock()

	cur, ok := s.refs[h]
	if !ok || cur <= 0 {
		return 0, fmt.Errorf("chunk %s: %w", hex.EncodeToString(h[:]), errs.ErrRefUnderflow)
	}

	cur--
	if cur == 0 {
		delete(s.refs, h)
	} else {
		s.refs[h] = cur
	}
	s.dirtyCnt++
	if err := s.maybeFlushLocked(); err != nil {
		return cur, err
	}
	return cur, nil
}

// maybeFlushLocked requires s.mu held.
func (s *Store) maybeFlushLocked() error {
	if s.dirtyCnt >= flushThreshold {
		return s.flushLocked()
	}
	return nil
}