package persist

import "os"

type Snapshotter interface {
	Snapshot(id string, parentID *string) (*Snapshot, error)
}

func Commit(repo string, s Snapshotter) error {
	id, err := NewSnapshotID()
	if err != nil {
		return err
	}

	parentID, err := CurrentSnapshotID(repo)
	if err != nil && !os.IsNotExist(err) {
		return err
	}

	var parent *string
	if err == nil {
		parent = &parentID
	}

	snap, err := s.Snapshot(id, parent)
	if err != nil {
		return err
	}

	return Save(repo, snap)
}