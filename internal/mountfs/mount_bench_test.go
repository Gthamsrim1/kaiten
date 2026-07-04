package mountfs

import (
	"fmt"
	"os"
	"path/filepath"
	"testing"
)

func setupBenchMount(b *testing.B) (mountDir string, cleanup func()) {
	b.Helper()

	mountDir, err := os.MkdirTemp("", "kaitenfs-bench-*")
	if err != nil {
		b.Fatal(err)
	}

	server, _, err := Mount(mountDir, false)
	if err != nil {
		os.RemoveAll(mountDir)
		b.Fatal(err)
	}

	cleanup = func() {
		if err := server.Unmount(); err != nil {
			b.Logf("unmount failed: %v", err)
		}
		os.RemoveAll(mountDir)
	}
	return mountDir, cleanup
}

func BenchmarkFUSE_Read(b *testing.B) {
	mountDir, cleanup := setupBenchMount(b)
	defer cleanup()

	path := filepath.Join(mountDir, "benchfile")
	if err := os.WriteFile(path, make([]byte, 4096), 0644); err != nil {
		b.Fatal(err)
	}

	f, err := os.Open(path)
	if err != nil {
		b.Fatal(err)
	}
	defer f.Close()

	buf := make([]byte, 4096)
	b.ResetTimer()
	b.ReportAllocs()
	for i := 0; i < b.N; i++ {
		if _, err := f.ReadAt(buf, 0); err != nil {
			b.Fatal(err)
		}
	}
}

func BenchmarkFUSE_Write(b *testing.B) {
	mountDir, cleanup := setupBenchMount(b)
	defer cleanup()

	path := filepath.Join(mountDir, "benchwrite")
	f, err := os.Create(path)
	if err != nil {
		b.Fatal(err)
	}
	defer f.Close()

	data := make([]byte, 4096)
	b.ResetTimer()
	b.ReportAllocs()
	for i := 0; i < b.N; i++ {
		if _, err := f.WriteAt(data, 0); err != nil {
			b.Fatal(err)
		}
	}
}

func BenchmarkFUSE_CreateFile(b *testing.B) {
	mountDir, cleanup := setupBenchMount(b)
	defer cleanup()

	b.ResetTimer()
	b.ReportAllocs()
	for i := 0; i < b.N; i++ {
		path := filepath.Join(mountDir, fmt.Sprintf("f%d", i))
		f, err := os.Create(path)
		if err != nil {
			b.Fatal(err)
		}
		f.Close()
	}
}

func BenchmarkFUSE_Readdir(b *testing.B) {
	mountDir, cleanup := setupBenchMount(b)
	defer cleanup()

	const n = 1000
	for i := 0; i < n; i++ {
		f, err := os.Create(filepath.Join(mountDir, fmt.Sprintf("f%d", i)))
		if err != nil {
			b.Fatal(err)
		}
		f.Close()
	}

	b.ResetTimer()
	b.ReportAllocs()
	for i := 0; i < b.N; i++ {
		if _, err := os.ReadDir(mountDir); err != nil {
			b.Fatal(err)
		}
	}
}