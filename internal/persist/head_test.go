package persist

import "testing"

func TestCheckout(t *testing.T) {
	repo := t.TempDir()

	s1 := &Snapshot{ID: "snap1"}
	s2 := &Snapshot{ID: "snap2"}

	if err := Save(repo, s1); err != nil {
		t.Fatal(err)
	}

	if err := Save(repo, s2); err != nil {
		t.Fatal(err)
	}

	if err := Checkout(repo, "snap1"); err != nil {
		t.Fatal(err)
	}

	head, err := CurrentSnapshotID(repo)
	if err != nil {
		t.Fatal(err)
	}

	if head != "snap1" {
		t.Fatalf("expected HEAD snap1, got %q", head)
	}
}

func TestCheckoutMissingSnapshot(t *testing.T) {
	repo := t.TempDir()

	if err := Checkout(repo, "does-not-exist"); err == nil {
		t.Fatal("expected error")
	}
}