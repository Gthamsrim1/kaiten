package content

import (
	"bytes"
	"testing"
)

func TestMemoryRead(t *testing.T) {
	m := Memory([]byte("hello"))

	buf := make([]byte, 5)

	n, err := m.Read(0, buf)
	if err != nil {
		t.Fatal(err)
	}

	if n != 5 {
		t.Fatalf("expected 5 bytes, got %d", n)
	}

	if !bytes.Equal(buf, []byte("hello")) {
		t.Fatalf("expected %q, got %q", "hello", string(buf))
	}
}

func TestMemoryWrite(t *testing.T) {
	m := Memory(nil)

	n, err := m.Write(0, []byte("hello"))
	if err != nil {
		t.Fatal(err)
	}

	if n != 5 {
		t.Fatalf("expected to write 5 bytes, got %d", n)
	}

	if !bytes.Equal(m.data, []byte("hello")) {
		t.Fatalf("expected %q, got %q", "hello", string(m.data))
	}
}

func TestMemoryOverwrite(t *testing.T) {
	m := Memory([]byte("hello"))

	_, err := m.Write(0, []byte("HELLO"))
	if err != nil {
		t.Fatal(err)
	}

	if !bytes.Equal(m.data, []byte("HELLO")) {
		t.Fatalf("expected %q, got %q", "HELLO", string(m.data))
	}
}

func TestMemoryAppend(t *testing.T) {
	m := Memory([]byte("hello"))

	_, err := m.Write(5, []byte(" world"))
	if err != nil {
		t.Fatal(err)
	}

	expected := []byte("hello world")

	if !bytes.Equal(m.data, expected) {
		t.Fatalf("expected %q, got %q", expected, m.data)
	}
}

func TestMemoryCopiesInput(t *testing.T) {
	data := []byte("hello")

	m := Memory(data)

	data[0] = 'H'

	if bytes.Equal(m.data, data) {
		t.Fatal("Memory should copy the input slice")
	}
}

func TestMemoryReadEmpty(t *testing.T) {
	m := Memory(nil)

	buf := make([]byte, 10)

	n, err := m.Read(0, buf)
	if err != nil {
		t.Fatal(err)
	}

	if n != 0 {
		t.Fatalf("expected 0 bytes, got %d", n)
	}
}

func TestMemoryReadPastEOF(t *testing.T) {
	m := Memory([]byte("hello"))

	buf := make([]byte, 10)

	n, err := m.Read(100, buf)
	if err != nil {
		t.Fatal(err)
	}

	if n != 0 {
		t.Fatalf("expected 0 bytes, got %d", n)
	}
}

func TestMemoryWritePastEnd(t *testing.T) {
	m := Memory([]byte("hello"))

	_, err := m.Write(10, []byte("abc"))
	if err != nil {
		t.Fatal(err)
	}

	expected := []byte{
		'h', 'e', 'l', 'l', 'o',
		0, 0, 0, 0, 0,
		'a', 'b', 'c',
	}

	if !bytes.Equal(m.data, expected) {
		t.Fatalf("unexpected data: %v", m.data)
	}
}

func TestMemorySize(t *testing.T) {
	m := Memory([]byte("hello"))

	if m.Size() != 5 {
		t.Fatalf("expected size 5, got %d", m.Size())
	}

	_, _ = m.Write(5, []byte(" world"))

	if m.Size() != 11 {
		t.Fatalf("expected size 11, got %d", m.Size())
	}
}