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

	parent.mu.RLock()
	_, exists := parent.Children[name]
	parent.mu.RUnlock()

	if exists {
		return ErrAlreadyExists
	}
	return nil
}

func (k *KaitenFS) validateExistingChild(name string, parent *Directory) (FSNode, error) {
	if err := ValidateName(name); err != nil {
		return nil, err
	}
	if parent == nil {
		return nil, ErrNilParent
	}

	parent.mu.RLock()
	node, exists := parent.Children[name]
	parent.mu.RUnlock()

	if !exists {
		return nil, ErrNotFound
	}
	return node, nil
}

func (k *KaitenFS) createFile(name string, parent *Directory, content Content) (*File, error) {
	if err := k.validateNewChild(name, parent); err != nil {
		return nil, err
	}

	file := &File{
		Node:    newNode(k, name, parent, syscall.S_IFREG),
		Content: content,
	}

	parent.mu.Lock()
	parent.Children[name] = file
	parent.mu.Unlock()

	return file, nil
}

func (k *KaitenFS) createDirectory(name string, parent *Directory) (*Directory, error) {
	if err := k.validateNewChild(name, parent); err != nil {
		return nil, err
	}

	directory := &Directory{
		Node:     newNode(k, name, parent, syscall.S_IFDIR), // fixed
		FS:       k,
		Children: make(map[string]FSNode),
	}

	parent.mu.Lock()
	parent.Children[name] = directory
	parent.mu.Unlock()

	return directory, nil
}

func (k *KaitenFS) deleteFile(name string, parent *Directory) error {
	file, err := k.validateExistingChild(name, parent)
	if err != nil {
		return err
	}
	if _, ok := file.(*File); !ok {
		return ErrNotFile
	}

	parent.mu.Lock()
	delete(parent.Children, name)
	parent.mu.Unlock()

	return nil
}

func (k *KaitenFS) deleteDirectory(name string, parent *Directory) error {
	dir, err := k.validateExistingChild(name, parent)
	if err != nil {
		return err
	}
	d, ok := dir.(*Directory)
	if !ok {
		return ErrNotDirectory
	}

	d.mu.RLock()
	empty := len(d.Children) == 0
	d.mu.RUnlock()
	if !empty {
		return ErrNotEmpty
	}

	parent.mu.Lock()
	delete(parent.Children, name)
	parent.mu.Unlock()

	return nil
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

func (k *KaitenFS) rename(oldParent *Directory, newParent *Directory, oldName string, newName string) error {
	node, err := k.validateExistingChild(oldName, oldParent)
	if err != nil {
		return err
	}

	if err := ValidateName(newName); err != nil {
		return err
	}

	newParent.mu.Lock()
	defer newParent.mu.Unlock()

	if existing, exists := newParent.Children[newName]; exists {
		switch existing.(type) {
		case *Directory:
			d := existing.(*Directory)
			d.mu.RLock()
			empty := len(d.Children) == 0
			d.mu.RUnlock()
			if !empty {
				return ErrNotEmpty
			}
		case *File:
			if _, isDir := node.(*Directory); isDir {
				return ErrNotDirectory
			}
		}
	}

	oldParent.mu.Lock()
	delete(oldParent.Children, oldName)
	oldParent.mu.Unlock()

	node.GetNode().Name = newName
	node.GetNode().Parent = newParent

	newParent.Children[newName] = node
	return nil
}