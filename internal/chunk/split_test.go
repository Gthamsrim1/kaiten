// Copyright (c) 2026 Gautham Sriram All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package chunk

import (
	"bytes"
	"crypto/rand"
	"testing"
)

func testParams() Params {
	return Params{
		Window:  64,
		MinSize: 256,
		AvgSize: 512,
		MaxSize: 1024,
	}
}

func TestSplitEmpty(t *testing.T) {
	chunks, err := Split(nil, testParams())
	if err != nil {
		t.Fatal(err)
	}

	if len(chunks) != 0 {
		t.Fatalf("expected 0 chunks, got %d", len(chunks))
	}
}

func TestSplitSmallInput(t *testing.T) {
	data := []byte("Madoka Kaname")

	chunks, err := Split(data, testParams())
	if err != nil {
		t.Fatal(err)
	}

	if len(chunks) != 1 {
		t.Fatalf("expected 1 chunk, got %d", len(chunks))
	}

	if !bytes.Equal(chunks[0].Data, data) {
		t.Fatal("chunk contents differ")
	}
}

func TestSplitDeterministic(t *testing.T) {
	data := make([]byte, 8192)

	if _, err := rand.Read(data); err != nil {
		t.Fatal(err)
	}

	a, err := Split(data, testParams())
	if err != nil {
		t.Fatal(err)
	}

	b, err := Split(data, testParams())
	if err != nil {
		t.Fatal(err)
	}

	if len(a) != len(b) {
		t.Fatalf("chunk count differs: %d vs %d", len(a), len(b))
	}

	for i := range a {
		if a[i].Offset != b[i].Offset {
			t.Fatalf("offset mismatch at chunk %d", i)
		}

		if !bytes.Equal(a[i].Hash[:], b[i].Hash[:]) {
			t.Fatalf("hash mismatch at chunk %d", i)
		}

		if !bytes.Equal(a[i].Data, b[i].Data) {
			t.Fatalf("data mismatch at chunk %d", i)
		}
	}
}

func TestSplitReconstruct(t *testing.T) {
	data := make([]byte, 32*1024)

	if _, err := rand.Read(data); err != nil {
		t.Fatal(err)
	}

	chunks, err := Split(data, testParams())
	if err != nil {
		t.Fatal(err)
	}

	var rebuilt []byte

	for _, c := range chunks {
		rebuilt = append(rebuilt, c.Data...)
	}

	if !bytes.Equal(rebuilt, data) {
		t.Fatal("reconstructed data differs from original")
	}
}

func TestSplitChunkSizeLimits(t *testing.T) {
	p := testParams()

	data := make([]byte, 64*1024)

	if _, err := rand.Read(data); err != nil {
		t.Fatal(err)
	}

	chunks, err := Split(data, p)
	if err != nil {
		t.Fatal(err)
	}

	for i, c := range chunks {
		if len(c.Data) > p.MaxSize {
			t.Fatalf("chunk %d exceeds max size: %d", i, len(c.Data))
		}
	}

	if len(chunks) == 0 {
		t.Fatal("expected at least one chunk")
	}
}

func TestSplitOffsets(t *testing.T) {
	p := testParams()

	data := make([]byte, 16*1024)

	if _, err := rand.Read(data); err != nil {
		t.Fatal(err)
	}

	chunks, err := Split(data, p)
	if err != nil {
		t.Fatal(err)
	}

	offset := int64(0)

	for i, c := range chunks {
		if c.Offset != offset {
			t.Fatalf("chunk %d offset = %d, expected %d",
				i, c.Offset, offset)
		}

		offset += int64(len(c.Data))
	}

	if offset != int64(len(data)) {
		t.Fatalf("offsets end at %d, expected %d",
			offset, len(data))
	}
}

func TestSplitDifferentInputs(t *testing.T) {
	a := bytes.Repeat([]byte("A"), 8192)
	b := bytes.Repeat([]byte("B"), 8192)

	ca, err := Split(a, testParams())
	if err != nil {
		t.Fatal(err)
	}

	cb, err := Split(b, testParams())
	if err != nil {
		t.Fatal(err)
	}

	if len(ca) == 0 || len(cb) == 0 {
		t.Fatal("expected chunks")
	}

	if bytes.Equal(ca[0].Hash[:], cb[0].Hash[:]) {
		t.Fatal("different inputs produced identical first hash")
	}
}
