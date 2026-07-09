// Copyright (c) 2026 Gautham Sriram All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package node

import (
	"syscall"
	"testing"

	"golang.org/x/sys/unix"
)

type fakeFSNode struct {
	n Node
}

func (f *fakeFSNode) GetNode() *Node {
	return &f.n
}

func TestFSNodeInterface(t *testing.T) {
	f := &fakeFSNode{n: Node{ID: 1, Name: "test"}}

	var fs FSNode = f
	if fs.GetNode().Name != "test" {
		t.Fatalf("expected %q, got %q", "test", fs.GetNode().Name)
	}
}

func TestCheckAccessOwner(t *testing.T) {
	n := Node{
		UID:  1000,
		GID:  1000,
		Mode: syscall.S_IFREG | 0644,
	}

	if errno := n.CheckAccess(1000, 1000, unix.R_OK); errno != 0 {
		t.Fatalf("owner should have read permission, got %v", errno)
	}

	if errno := n.CheckAccess(1000, 1000, unix.W_OK); errno != 0 {
		t.Fatalf("owner should have write permission, got %v", errno)
	}

	if errno := n.CheckAccess(1000, 1000, unix.X_OK); errno != syscall.EACCES {
		t.Fatalf("expected EACCES, got %v", errno)
	}
}

func TestCheckAccessGroup(t *testing.T) {
	n := Node{
		UID:  1000,
		GID:  2000,
		Mode: syscall.S_IFREG | 0640,
	}

	if errno := n.CheckAccess(3000, 2000, unix.R_OK); errno != 0 {
		t.Fatalf("group should have read permission, got %v", errno)
	}

	if errno := n.CheckAccess(3000, 2000, unix.W_OK); errno != syscall.EACCES {
		t.Fatalf("expected EACCES, got %v", errno)
	}
}

func TestCheckAccessOthers(t *testing.T) {
	n := Node{
		UID:  1000,
		GID:  1000,
		Mode: syscall.S_IFREG | 0644,
	}

	if errno := n.CheckAccess(2000, 2000, unix.R_OK); errno != 0 {
		t.Fatalf("others should have read permission, got %v", errno)
	}

	if errno := n.CheckAccess(2000, 2000, unix.W_OK); errno != syscall.EACCES {
		t.Fatalf("expected EACCES, got %v", errno)
	}
}

func TestCheckAccessReadOnly(t *testing.T) {
	n := Node{
		UID:  1000,
		GID:  1000,
		Mode: syscall.S_IFREG | 0444,
	}

	if errno := n.CheckAccess(1000, 1000, unix.R_OK); errno != 0 {
		t.Fatalf("owner should have read permission, got %v", errno)
	}

	if errno := n.CheckAccess(1000, 1000, unix.W_OK); errno != syscall.EACCES {
		t.Fatalf("expected EACCES, got %v", errno)
	}
}

func TestCheckAccessExecute(t *testing.T) {
	n := Node{
		UID:  1000,
		GID:  1000,
		Mode: syscall.S_IFREG | 0755,
	}

	if errno := n.CheckAccess(2000, 2000, unix.X_OK); errno != 0 {
		t.Fatalf("others should have execute permission, got %v", errno)
	}

	if errno := n.CheckAccess(1000, 1000, unix.X_OK); errno != 0 {
		t.Fatalf("owner should have execute permission, got %v", errno)
	}
}

func TestCheckAccessNone(t *testing.T) {
	n := Node{
		UID:  1000,
		GID:  1000,
		Mode: syscall.S_IFREG,
	}

	if errno := n.CheckAccess(1000, 1000, unix.R_OK); errno != syscall.EACCES {
		t.Fatalf("expected EACCES, got %v", errno)
	}

	if errno := n.CheckAccess(1000, 1000, unix.W_OK); errno != syscall.EACCES {
		t.Fatalf("expected EACCES, got %v", errno)
	}

	if errno := n.CheckAccess(1000, 1000, unix.X_OK); errno != syscall.EACCES {
		t.Fatalf("expected EACCES, got %v", errno)
	}
}

func TestCheckAccessRoot(t *testing.T) {
	n := Node{
		UID:  1000,
		GID:  1000,
		Mode: syscall.S_IFREG,
	}

	if errno := n.CheckAccess(0, 0, unix.R_OK|unix.W_OK|unix.X_OK); errno != 0 {
		t.Fatalf("root should bypass permission checks, got %v", errno)
	}
}

func TestCheckAccessReadWrite(t *testing.T) {
	n := Node{
		UID:  1000,
		GID:  1000,
		Mode: syscall.S_IFREG | 0644,
	}

	if errno := n.CheckAccess(1000, 1000, unix.R_OK|unix.W_OK); errno != 0 {
		t.Fatalf("owner should have read/write permission, got %v", errno)
	}

	if errno := n.CheckAccess(2000, 2000, unix.R_OK|unix.W_OK); errno != syscall.EACCES {
		t.Fatalf("others should not have read/write permission, got %v", errno)
	}
}
