package store

import (
	"crypto/sha256"
	"os"
	"path/filepath"
	"sync"
	"testing"
)

func hashOf(b []byte) Hash {
	return sha256.Sum256(b)
}

func TestPutGetRoundTrip(t *testing.T) {
	s, err := Open(t.TempDir())
	if err != nil {
		t.Fatal(err)
	}

	body := []byte("hello kaiten")
	h := hashOf(body)

	if s.Has(h) {
		t.Fatal("store reports chunk present before Put")
	}
	if err := s.Put(h, body); err != nil {
		t.Fatal(err)
	}
	if !s.Has(h) {
		t.Fatal("store reports chunk absent after Put")
	}

	got, err := s.Get(h)
	if err != nil {
		t.Fatal(err)
	}
	if string(got) != string(body) {
		t.Fatalf("got %q, want %q", got, body)
	}
}

func TestGetMissing(t *testing.T) {
	s, err := Open(t.TempDir())
	if err != nil {
		t.Fatal(err)
	}
	_, err = s.Get(hashOf([]byte("nope")))
	if err == nil {
		t.Fatal("expected error getting missing chunk")
	}
}

func TestPutDedups(t *testing.T) {
	root := t.TempDir()
	s, err := Open(root)
	if err != nil {
		t.Fatal(err)
	}

	body := []byte("a chunk that appears in two images")
	h := hashOf(body)

	if err := s.Put(h, body); err != nil {
		t.Fatal(err)
	}
	if err := s.Put(h, body); err != nil {
		t.Fatal(err)
	}

	entries, err := os.ReadDir(filepath.Join(root, blobsDir))
	if err != nil {
		t.Fatal(err)
	}
	count := 0
	for _, e := range entries {
		if !e.IsDir() && e.Name()[0] != '.' {
			count++
		}
	}
	if count != 1 {
		t.Fatalf("expected exactly 1 stored blob after storing identical content twice, got %d", count)
	}
}

func TestRefCountLifecycle(t *testing.T) {
	s, err := Open(t.TempDir())
	if err != nil {
		t.Fatal(err)
	}

	body := []byte("refcounted chunk")
	h := hashOf(body)
	if err := s.Put(h, body); err != nil {
		t.Fatal(err)
	}

	if got := s.RefCount(h); got != 0 {
		t.Fatalf("fresh chunk should have refcount 0, got %d", got)
	}

	n, err := s.IncRef(h)
	if err != nil {
		t.Fatal(err)
	}
	if n != 1 {
		t.Fatalf("expected refcount 1 after IncRef, got %d", n)
	}

	n, _ = s.IncRef(h)
	if n != 2 {
		t.Fatalf("expected refcount 2, got %d", n)
	}

	n, err = s.DecRef(h)
	if err != nil {
		t.Fatal(err)
	}
	if n != 1 {
		t.Fatalf("expected refcount 1 after DecRef, got %d", n)
	}

	n, _ = s.DecRef(h)
	if n != 0 {
		t.Fatalf("expected refcount 0, got %d", n)
	}
}

func TestDecRefUnderflow(t *testing.T) {
	s, err := Open(t.TempDir())
	if err != nil {
		t.Fatal(err)
	}
	h := hashOf([]byte("x"))
	if _, err := s.DecRef(h); err == nil {
		t.Fatal("expected underflow error decrementing an untracked chunk")
	}
}

func TestSweepDeletesUnreferenced(t *testing.T) {
	s, err := Open(t.TempDir())
	if err != nil {
		t.Fatal(err)
	}

	live := []byte("still referenced")
	dead := []byte("no longer referenced")
	lh, dh := hashOf(live), hashOf(dead)

	if err := s.Put(lh, live); err != nil {
		t.Fatal(err)
	}
	if err := s.Put(dh, dead); err != nil {
		t.Fatal(err)
	}

	if _, err := s.IncRef(lh); err != nil {
		t.Fatal(err)
	}
	if _, err := s.IncRef(dh); err != nil {
		t.Fatal(err)
	}
	if _, err := s.DecRef(dh); err != nil {
		t.Fatal(err)
	}

	freed, err := s.Sweep()
	if err != nil {
		t.Fatal(err)
	}
	if freed != 1 {
		t.Fatalf("expected to free 1 chunk, freed %d", freed)
	}

	if !s.Has(lh) {
		t.Fatal("referenced chunk was incorrectly swept")
	}
	if s.Has(dh) {
		t.Fatal("unreferenced chunk was not swept")
	}
}

func TestPersistenceAcrossReopen(t *testing.T) {
	root := t.TempDir()

	s1, err := Open(root)
	if err != nil {
		t.Fatal(err)
	}
	body := []byte("survives a restart")
	h := hashOf(body)
	if err := s1.Put(h, body); err != nil {
		t.Fatal(err)
	}
	if _, err := s1.IncRef(h); err != nil {
		t.Fatal(err)
	}
	if _, err := s1.IncRef(h); err != nil {
		t.Fatal(err)
	}
	if err := s1.Flush(); err != nil {
		t.Fatal(err)
	}

	s2, err := Open(root)
	if err != nil {
		t.Fatal(err)
	}
	if got := s2.RefCount(h); got != 2 {
		t.Fatalf("refcount did not survive reopen: got %d, want 2", got)
	}
	if !s2.Has(h) {
		t.Fatal("chunk body did not survive reopen")
	}
}

func TestConcurrentRefCounting(t *testing.T) {
	s, err := Open(t.TempDir())
	if err != nil {
		t.Fatal(err)
	}

	const chunks = 50
	const workers = 16
	const opsPerWorker = 200

	hashes := make([]Hash, chunks)
	for i := range hashes {
		body := []byte{byte(i), byte(i >> 8), 'c', 'h', 'u', 'n', 'k'}
		hashes[i] = hashOf(body)
		if err := s.Put(hashes[i], body); err != nil {
			t.Fatal(err)
		}
	}

	var wg sync.WaitGroup
	for w := 0; w < workers; w++ {
		wg.Add(1)
		go func(seed int) {
			defer wg.Done()
			for i := 0; i < opsPerWorker; i++ {
				h := hashes[(seed+i)%chunks]
				if _, err := s.IncRef(h); err != nil {
					t.Errorf("IncRef: %v", err)
					return
				}
				if _, err := s.DecRef(h); err != nil {
					t.Errorf("DecRef: %v", err)
					return
				}
			}
		}(w)
	}
	wg.Wait()

	for i, h := range hashes {
		if got := s.RefCount(h); got != 0 {
			t.Fatalf("chunk %d has nonzero refcount %d after balanced inc/dec -- lost update", i, got)
		}
	}
}

func TestSweepIgnoresTempFiles(t *testing.T) {
	root := t.TempDir()
	s, err := Open(root)
	if err != nil {
		t.Fatal(err)
	}

	tmp := filepath.Join(root, blobsDir, ".tmp-orphan")
	if err := os.WriteFile(tmp, []byte("garbage"), 0o644); err != nil {
		t.Fatal(err)
	}

	if _, err := s.Sweep(); err != nil {
		t.Fatal(err)
	}

	if _, err := os.Stat(tmp); err != nil {
		t.Fatalf("temp file was unexpectedly removed or errored: %v", err)
	}
}