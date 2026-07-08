package integration

import (
	"crypto/rand"
	"fmt"
	"testing"

	"github.com/Gthamsrim1/kaiten/internal/chunk"
	"github.com/Gthamsrim1/kaiten/internal/content"
)

func TestRoundTripEmpty(t *testing.T) {
	fs := newFS(t)

	restored := roundTrip(t, fs)

	compareFilesystem(t, fs, restored)
}

func TestRoundTripSingleFile(t *testing.T) {
	fs := newFS(t)

	_, err := fs.Root.CreateFile("hello", content.Memory([]byte("Hello Kaiten")), 0644)
	if err != nil {
		t.Fatal(err)
	}

	restored := roundTrip(t, fs)

	compareFilesystem(t, fs, restored)
}

func TestRoundTripEmptyFile(t *testing.T) {
	fs := newFS(t)

	_, err := fs.Root.CreateFile("empty", content.Memory(nil), 0644)
	if err != nil {
		t.Fatal(err)
	}

	restored := roundTrip(t, fs)

	compareFilesystem(t, fs, restored)
}

func TestRoundTripNestedDirectories(t *testing.T) {
	fs := newFS(t)

	usr, _ := fs.Root.CreateDirectory("usr", 0755)
	local, _ := usr.CreateDirectory("local", 0755)
	bin, _ := local.CreateDirectory("bin", 0755)

	_, err := bin.CreateFile("hello", content.Memory([]byte("Hello")), 0755)
	if err != nil {
		t.Fatal(err)
	}

	restored := roundTrip(t, fs)

	compareFilesystem(t, fs, restored)
}

func TestRoundTripManyFiles(t *testing.T) {
	fs := newFS(t)

	for i := 0; i < 100; i++ {
		_, err := fs.Root.CreateFile(fmt.Sprintf("file-%03d", i), content.Memory([]byte(fmt.Sprintf("data-%03d", i))), 0644)
		if err != nil {
			t.Fatal(err)
		}
	}

	restored := roundTrip(t, fs)

	compareFilesystem(t, fs, restored)
}

func TestRoundTripLargeFile(t *testing.T) {
	fs := newFS(t)

	data := make([]byte, chunk.Default.MaxSize*4)

	if _, err := rand.Read(data); err != nil {
		t.Fatal(err)
	}

	_, err := fs.Root.CreateFile(
		"large.bin",
		content.Memory(data),
		0644,
	)
	if err != nil {
		t.Fatal(err)
	}

	restored := roundTrip(t, fs)

	compareFilesystem(t, fs, restored)
}

func TestRoundTripDuplicateFiles(t *testing.T) {
	fs := newFS(t)

	data := []byte("Madoka Kaname")

	_, _ = fs.Root.CreateFile("a", content.Memory(data), 0644)
	_, _ = fs.Root.CreateFile("b", content.Memory(data), 0644)

	restored := roundTrip(t, fs)

	compareFilesystem(t, fs, restored)

	restored2 := roundTrip(t, restored)
	compareFilesystem(t, restored, restored2)
}
