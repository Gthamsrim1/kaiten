package persist

import (
	"crypto/rand"
	"encoding/hex"
	"fmt"
	"os"
	"path/filepath"
	"strings"
)

func NewSnapshotID() (string, error) {
	var id [16]byte

	if _, err := rand.Read(id[:]); err != nil {
		return "", err
	}

	return hex.EncodeToString(id[:]), nil
}

func CurrentSnapshotID(repo string) (string, error) {
	data, err := os.ReadFile(filepath.Join(repo, "HEAD"))
	if err != nil {
		return "", err
	}

	return strings.TrimSpace(string(data)), nil
}

func Checkout(repo, id string) error {
	snapshotPath := filepath.Join(repo, "snapshots", id+".json")

	info, err := os.Stat(snapshotPath)
	switch {
	case os.IsNotExist(err):
		return fmt.Errorf("snapshot %q not found", id)

	case err != nil:
		return err

	case info.IsDir():
		return fmt.Errorf("snapshot %q is a directory", id)
	}

	return writeAtomic(
		filepath.Join(repo, "HEAD"),
		[]byte(id+"\n"),
		0644,
	)
}