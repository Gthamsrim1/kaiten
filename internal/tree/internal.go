// Copyright (c) 2026 Gautham Sriram All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package tree

import (
	"fmt"
	"os"
	"syscall"
	"time"

	"github.com/Gthamsrim1/kaiten/internal/content"
	"github.com/Gthamsrim1/kaiten/internal/errs"
	"github.com/Gthamsrim1/kaiten/internal/node"
	"github.com/Gthamsrim1/kaiten/internal/store"
)

// Validation
func ValidateName(name string) error {
	if len(name) == 0 {
		return errs.ErrEmptyName
	}

	if len(name) > 255 {
		return errs.ErrNameTooLong
	}

	if name == "." || name == ".." {
		return errs.ErrInvalidName
	}

	return nil
}

func (k *KaitenFS) validateNewChild(name string, parent *Directory) error {
	if err := ValidateName(name); err != nil {
		return err
	}
	if parent == nil {
		return errs.ErrNilParent
	}

	parent.mu.RLock()
	_, exists := parent.Children[name]
	parent.mu.RUnlock()

	if exists {
		return errs.ErrAlreadyExists
	}
	return nil
}

func (k *KaitenFS) validateExistingChild(name string, parent *Directory) (node.FSNode, error) {
	if err := ValidateName(name); err != nil {
		return nil, err
	}
	if parent == nil {
		return nil, errs.ErrNilParent
	}

	parent.mu.RLock()
	node, exists := parent.Children[name]
	parent.mu.RUnlock()

	if !exists {
		return nil, errs.ErrNotFound
	}
	return node, nil
}

// Deletion
func (k *KaitenFS) deleteFile(name string, parent *Directory) error {
	node, err := k.validateExistingChild(name, parent)
	if err != nil {
		return err
	}
	file, ok := node.(*File)
	if !ok {
		return errs.ErrNotFile
	}

	file.Content.Backing().Release()

	parent.mu.Lock()
	delete(parent.Children, name)
	parent.mu.Unlock()

	return nil
}

func (k *KaitenFS) deleteDirectory(name string, parent *Directory) error {
	n, err := k.validateExistingChild(name, parent)
	if err != nil {
		return err
	}

	dir, ok := n.(*Directory)
	if !ok {
		return errs.ErrNotDirectory
	}

	k.releaseNode(dir)

	parent.mu.Lock()
	delete(parent.Children, name)
	parent.mu.Unlock()

	return nil
}

func (k *KaitenFS) releaseNode(n node.FSNode) {
	switch v := n.(type) {
	case *File:
		v.Content.Backing().Release()

	case *Directory:
		v.mu.RLock()

		children := make([]node.FSNode, 0, len(v.Children))
		for _, child := range v.Children {
			children = append(children, child)
		}

		v.mu.RUnlock()

		for _, child := range children {
			k.releaseNode(child)
		}
	}
}

// Constructors
func newNode(k *KaitenFS, name string, parent *Directory, fileType uint32, perm uint32) node.Node {
	now := time.Now()

	return node.Node{
		ID:     k.nextID(),
		Name:   name,
		Parent: parent,
		Mode:   fileType | perm,
		UID:    uint32(os.Getuid()),
		GID:    uint32(os.Getgid()),
		Nlink:  1,
		Atime:  now,
		Mtime:  now,
		Ctime:  now,
	}
}

func (k *KaitenFS) createFile(name string, parent *Directory, content content.Content, perm uint32) (*File, error) {
	if err := k.validateNewChild(name, parent); err != nil {
		return nil, err
	}

	file := &File{
		Node:    newNode(k, name, parent, syscall.S_IFREG, perm),
		FS:      k,
		Content: content,
	}

	parent.mu.Lock()
	parent.Children[name] = file
	parent.mu.Unlock()

	return file, nil
}

func (k *KaitenFS) createDirectory(name string, parent *Directory, perm uint32) (*Directory, error) {
	if err := k.validateNewChild(name, parent); err != nil {
		return nil, err
	}

	directory := &Directory{
		Node:     newNode(k, name, parent, syscall.S_IFDIR, perm),
		FS:       k,
		Children: make(map[string]node.FSNode),
	}

	parent.mu.Lock()
	parent.Children[name] = directory
	parent.mu.Unlock()

	return directory, nil
}

func (k *KaitenFS) createSymlink(name string, parent *Directory, target string) (*Symlink, error) {
	if err := k.validateNewChild(name, parent); err != nil {
		return nil, err
	}

	symlink := &Symlink{
		Node:   newNode(k, name, parent, syscall.S_IFLNK, 0777),
		FS:     k,
		Target: target,
	}

	parent.mu.Lock()
	parent.Children[name] = symlink
	parent.mu.Unlock()

	return symlink, nil
}

// Renaming
func (k *KaitenFS) rename(oldParent *Directory, newParent *Directory, oldName string, newName string) error {
	node, err := k.validateExistingChild(oldName, oldParent)
	if err != nil {
		return err
	}

	if err := ValidateName(newName); err != nil {
		return err
	}

	if oldParent == newParent {
		oldParent.mu.Lock()
		defer oldParent.mu.Unlock()
	} else {
		// Lock in a consistent order to avoid ABBA deadlocks between
		// concurrent renames going in opposite directions.
		first, second := oldParent, newParent
		if fmt.Sprintf("%p", first) > fmt.Sprintf("%p", second) {
			first, second = second, first
		}
		first.mu.Lock()
		defer first.mu.Unlock()
		second.mu.Lock()
		defer second.mu.Unlock()
	}

	if existing, exists := newParent.Children[newName]; exists {
		switch existing.(type) {
		case *Directory:
			d := existing.(*Directory)
			d.mu.RLock()
			empty := len(d.Children) == 0
			d.mu.RUnlock()
			if !empty {
				return errs.ErrNotEmpty
			}
		case *File:
			if _, isDir := node.(*Directory); isDir {
				return errs.ErrNotDirectory
			}
		}
	}

	delete(oldParent.Children, oldName)

	node.GetNode().Name = newName
	node.GetNode().Parent = newParent

	newParent.Children[newName] = node
	return nil
}

// Dirty
func (k *KaitenFS) MarkDirty() {
	k.dirty.Store(true)
}

func (k *KaitenFS) ClearDirty() {
	k.dirty.Store(false)
}

func (k *KaitenFS) IsDirty() bool {
	return k.dirty.Load()
}

// Cloning
func (k *KaitenFS) Clone(src node.FSNode, dst *Directory, name string) (node.FSNode, error) {
	dst.mu.RLock()
	_, exists := dst.Children[name]
	dst.mu.RUnlock()

	if exists {
		return nil, errs.ErrAlreadyExists
	}

	var (
		clone node.FSNode
		err   error
	)

	switch n := src.(type) {
	case *File:
		clone, err = k.cloneFile(n, dst, name)

	case *Directory:
		clone, err = k.cloneDirectory(n, dst, name)

	default:
		return nil, errs.ErrInvalidInode
	}

	if err != nil {
		return nil, err
	}

	dst.mu.Lock()
	dst.Children[name] = clone
	dst.mu.Unlock()

	k.MarkDirty()

	return clone, nil
}

func cloneNode(k *KaitenFS, src *node.Node, name string, parent *Directory) node.Node {
	return node.Node{
		ID:     k.nextID(),
		Name:   name,
		Parent: parent,

		Mode:  src.Mode,
		UID:   src.UID,
		GID:   src.GID,
		Nlink: 1,

		Atime: src.Atime,
		Mtime: src.Mtime,
		Ctime: src.Ctime,

		Chunks: append([]store.ChunkRef(nil), src.Chunks...),
	}
}

func (k *KaitenFS) cloneFile(src *File, dst *Directory, name string) (*File, error) {
	backing := src.Content.Backing()
	backing.Acquire()

	return &File{
		Node:    cloneNode(k, &src.Node, name, dst),
		FS:      k,
		Content: content.FromBacking(backing),
	}, nil
}

func (k *KaitenFS) cloneDirectory(src *Directory, dst *Directory, name string) (*Directory, error) {
	dir := &Directory{
		Node:     cloneNode(k, &src.Node, name, dst),
		FS:       k,
		Children: make(map[string]node.FSNode),
	}

	src.mu.RLock()
	defer src.mu.RUnlock()

	for childName, child := range src.Children {
		if _, err := k.Clone(child, dir, childName); err != nil {
			return nil, err
		}
	}

	return dir, nil
}
