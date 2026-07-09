package integration

import (
	"testing"

	"github.com/Gthamsrim1/kaiten/internal/persist"
	"github.com/Gthamsrim1/kaiten/internal/tree"
)

func newFS(t *testing.T) *tree.KaitenFS {
	fs := tree.New()

	return fs
}

func snapshotTestFS(t *testing.T, fs *tree.KaitenFS) *persist.Snapshot {
	t.Helper()

	id, err := persist.NewSnapshotID()
	if err != nil {
		t.Fatal(err)
	}

	snap, err := fs.Snapshot(id, nil)
	if err != nil {
		t.Fatal(err)
	}

	return snap
}

func roundTrip(t *testing.T, fs *tree.KaitenFS) *tree.KaitenFS {
	repo := t.TempDir()
	snap := snapshotTestFS(t, fs)

	if err := persist.Save(repo, snap); err != nil {
		t.Fatal(err)
	}

	newfs, err := tree.Restore(repo)
	if err != nil {
		t.Fatal(err)
	}

	return newfs
}

func compareFilesystem(t *testing.T, expected *tree.KaitenFS, actual *tree.KaitenFS) {
	if expected.ID != actual.ID {
		t.Fatal("Id mistach")
	}
	compareDirectory(t, expected.Root, actual.Root)
}

func compareFile(t *testing.T, expected, actual *tree.File) {
	t.Helper()

	if expected.Node.Name != actual.Node.Name {
		t.Fatalf("name mismatch: %q != %q",
			expected.Node.Name, actual.Node.Name)
	}

	if expected.Node.Mode != actual.Node.Mode {
		t.Fatalf("mode mismatch")
	}

	if expected.Node.UID != actual.Node.UID {
		t.Fatalf("uid mismatch")
	}

	if expected.Node.GID != actual.Node.GID {
		t.Fatalf("gid mismatch")
	}

	if expected.Node.Nlink != actual.Node.Nlink {
		t.Fatalf("nlink mismatch")
	}

	data1, err := expected.Content.Bytes()
	if err != nil {
		t.Fatal(err)
	}

	data2, err := actual.Content.Bytes()
	if err != nil {
		t.Fatal(err)
	}

	if string(data1) != string(data2) {
		t.Fatalf("content mismatch")
	}
}

func compareDirectory(t *testing.T, expected, actual *tree.Directory) {
	t.Helper()

	if expected.Name != actual.Name {
		t.Fatalf("directory name mismatch")
	}

	if expected.Node.Mode != actual.Node.Mode {
		t.Fatalf("mode mismatch")
	}

	expectedChildren := expected.ChildrenSnapshot()
	actualChildren := actual.ChildrenSnapshot()

	if len(expectedChildren) != len(actualChildren) {
		t.Fatalf("child count mismatch")
	}

	for name, expectedChild := range expectedChildren {
		actualChild, ok := actualChildren[name]
		if !ok {
			t.Fatalf("missing child %q", name)
		}

		switch e := expectedChild.(type) {
		case *tree.File:
			a, ok := actualChild.(*tree.File)
			if !ok {
				t.Fatalf("%q changed type", name)
			}
			compareFile(t, e, a)

		case *tree.Directory:
			a, ok := actualChild.(*tree.Directory)
			if !ok {
				t.Fatalf("%q changed type", name)
			}
			compareDirectory(t, e, a)

		default:
			t.Fatalf("unknown node type")
		}
	}
}
