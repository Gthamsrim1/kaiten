package fs

import (
	"os"
	"syscall"
	"time"
)

func ValidateName(name string) error {
	if len(name) == 0 {
		return ErrEmptyName
	}

	if len(name) > 255 {
		return ErrNameTooLong
	}

	if name == "." || name == ".." {
		return ErrInvalidName
	}

	return nil
}

func (k *KaitenFS) validateNewChild(name string, parent *Directory) error {
	if err := ValidateName(name); err != nil {
		return err
	}

	if parent == nil {
		return ErrNilParent
	}

	if _, exists := parent.Children[name]; exists {
		return ErrAlreadyExists
	}

	return nil
}

func (k *KaitenFS) createFile(name string, parent *Directory, content Content) (*File, error) {
	if err := k.validateNewChild(name, parent); err != nil {
		return nil, err
	}

	file := &File{
		Node: newNode(k, name, parent, syscall.S_IFREG),
		Content: content,
	}

	parent.Children[name] = file

	return file, nil
}

func (k *KaitenFS) createDirectory(name string, parent *Directory) (*Directory, error) {
	if err := k.validateNewChild(name, parent); err != nil {
		return nil, err
	}

	directory := &Directory{
		Node: newNode(k, name, parent, syscall.S_IFREG),
		FS:       k,
		Children: make(map[string]FSNode),
	}

	parent.Children[name] = directory
	return directory, nil
}

func newNode(fs *KaitenFS, name string, parent *Directory, mode uint32) Node {
    now := time.Now()

    return Node{
        ID:     fs.nextID(),
        Name:   name,
        Parent: parent,
        Mode:   mode,
        UID:    uint32(os.Getuid()),
        GID:    uint32(os.Getgid()),
        Atime:  now,
        Mtime:  now,
        Ctime:  now,
    }
}
