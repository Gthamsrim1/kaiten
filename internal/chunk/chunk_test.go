package chunk

import (
	"bytes"
	"math/rand"
	"testing"
)

func TestSplitEmpty(t *testing.T) {
	chunks := SplitBytes(nil)
	if chunks != nil {
		t.Fatalf("expected nil chunks for empty input, got %d chunks", len(chunks))
	}
}

func TestSplitDeterministic(t *testing.T) {
	data := randomData(500 * 1024)

	a := SplitBytes(data)
	b := SplitBytes(data)

	if len(a) != len(b) {
		t.Fatalf("chunk count differs across runs: %d vs %d", len(a), len(b))
	}
	for i := range a {
		if a[i] != b[i] {
			t.Fatalf("chunk %d differs across runs: %+v vs %+v", i, a[i], b[i])
		}
	}
}

func TestSplitCoversWholeInput(t *testing.T) {
	data := randomData(300 * 1024)
	chunks := SplitBytes(data)

	if len(chunks) == 0 {
		t.Fatal("expected at least one chunk")
	}

	var offset int64
	for i, c := range chunks {
		if c.Offset != offset {
			t.Fatalf("chunk %d: expected offset %d, got %d", i, offset, c.Offset)
		}
		if c.Length <= 0 {
			t.Fatalf("chunk %d: non-positive length %d", i, c.Length)
		}
		offset += c.Length
	}

	if offset != int64(len(data)) {
		t.Fatalf("chunks cover %d bytes, expected %d", offset, len(data))
	}
}

func TestSplitRespectsSizeBounds(t *testing.T) {
	data := randomData(1024 * 1024)
	chunks := SplitBytes(data)

	for i, c := range chunks {
		last := i == len(chunks)-1
		if c.Length > MaxSize {
			t.Fatalf("chunk %d exceeds MaxSize: %d > %d", i, c.Length, MaxSize)
		}
		if !last && c.Length < MinSize {
			t.Fatalf("chunk %d is below MinSize: %d < %d", i, c.Length, MinSize)
		}
	}
}

func TestSplitHashMatchesContent(t *testing.T) {
	data := randomData(200 * 1024)
	chunks := SplitBytes(data)

	for i, c := range chunks {
		body := data[c.Offset : c.Offset+c.Length]
		want := SplitBytes(body)
		if len(want) != 1 {
			continue
		}
		if want[0].Hash != c.Hash {
			t.Fatalf("chunk %d hash does not match sha256 of its own bytes", i)
		}
	}
}

func TestSplitEditLocality(t *testing.T) {
	original := randomData(2 * 1024 * 1024)

	before := SplitBytes(original)
	if len(before) < 10 {
		t.Fatalf("test needs a multi-chunk input to be meaningful, got %d chunks", len(before))
	}

	// Insert 37 bytes roughly in the middle of the data.
	insertAt := len(original) / 2
	insertion := randomData(37)
	edited := make([]byte, 0, len(original)+len(insertion))
	edited = append(edited, original[:insertAt]...)
	edited = append(edited, insertion...)
	edited = append(edited, original[insertAt:]...)

	after := SplitBytes(edited)

	beforeHashes := make(map[[32]byte]bool, len(before))
	for _, c := range before {
		beforeHashes[c.Hash] = true
	}

	unchanged := 0
	for _, c := range after {
		if beforeHashes[c.Hash] {
			unchanged++
		}
	}

	preservedRatio := float64(unchanged) / float64(len(before))
	if preservedRatio < 0.8 {
		t.Fatalf("only %.1f%% of chunks survived a 37-byte insertion (wanted >=80%%); "+
			"got %d/%d unchanged -- content-defined chunking is not working as intended",
			preservedRatio*100, unchanged, len(before))
	}
	t.Logf("preserved %d/%d chunks (%.1f%%) after a 37-byte mid-file insertion",
		unchanged, len(before), preservedRatio*100)
}

// TestSplitDuplicateContentDedups proves the actual storage payoff
func TestSplitDuplicateContentDedups(t *testing.T) {
	shared := randomData(500 * 1024)

	fileA := append(randomData(50*1024), shared...)
	fileB := append(shared, randomData(50*1024)...)

	chunksA := SplitBytes(fileA)
	chunksB := SplitBytes(fileB)

	hashesB := make(map[[32]byte]bool, len(chunksB))
	for _, c := range chunksB {
		hashesB[c.Hash] = true
	}

	shared_count := 0
	for _, c := range chunksA {
		if hashesB[c.Hash] {
			shared_count++
		}
	}

	if shared_count == 0 {
		t.Fatal("expected at least some chunks to be shared between files with a large common substring")
	}
	t.Logf("%d/%d chunks in fileA were found verbatim in fileB", shared_count, len(chunksA))
}

func randomData(n int) []byte {
	buf := make([]byte, n)
	// Fixed seed: deterministic across test runs.
	rng := rand.New(rand.NewSource(42))
	_, _ = rng.Read(buf)
	return buf
}

func BenchmarkSplit(b *testing.B) {
	data := randomData(4 * 1024 * 1024) // 4 MiB
	b.ResetTimer()
	b.ReportAllocs()
	b.SetBytes(int64(len(data)))
	for i := 0; i < b.N; i++ {
		_ = SplitBytes(data)
	}
}

func TestNoOverlappingChunks(t *testing.T) {
	data := randomData(400 * 1024)
	chunks := SplitBytes(data)

	for i := 1; i < len(chunks); i++ {
		prevEnd := chunks[i-1].Offset + chunks[i-1].Length
		if chunks[i].Offset != prevEnd {
			t.Fatalf("gap or overlap between chunk %d (ends at %d) and chunk %d (starts at %d)",
				i-1, prevEnd, i, chunks[i].Offset)
		}
	}
}

func TestSplitBytesVsSplitReader(t *testing.T) {
	data := randomData(100 * 1024)

	viaBytes := SplitBytes(data)
	viaReader, err := Split(bytes.NewReader(data))
	if err != nil {
		t.Fatal(err)
	}

	if len(viaBytes) != len(viaReader) {
		t.Fatalf("Split and SplitBytes disagree on chunk count: %d vs %d", len(viaReader), len(viaBytes))
	}
	for i := range viaBytes {
		if viaBytes[i] != viaReader[i] {
			t.Fatalf("chunk %d differs between Split and SplitBytes", i)
		}
	}
}