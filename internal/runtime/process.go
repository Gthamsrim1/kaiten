// Copyright (c) 2026 Gautham Sriram All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package runtime

import (
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"syscall"

	"golang.org/x/sys/unix"
)

func Child() error {
	if len(os.Args) < 4 {
		return fmt.Errorf("invalid child arguments")
	}

	rootfs := os.Args[2]
	command := os.Args[3:]

	if err := unix.Mount("", "/", "", unix.MS_REC|unix.MS_PRIVATE, ""); err != nil {
		return fmt.Errorf("make mounts private: %w", err)
	}

	if err := unix.Mount(rootfs, rootfs, "", unix.MS_BIND|unix.MS_REC, ""); err != nil {
		return fmt.Errorf("bind mount rootfs: %w", err)
	}

	oldRoot := filepath.Join(rootfs, "oldroot")

	if err := os.MkdirAll(oldRoot, 0755); err != nil {
		return fmt.Errorf("create oldroot: %w", err)
	}

	if err := unix.PivotRoot(rootfs, oldRoot); err != nil {
		return fmt.Errorf("pivot_root: %w", err)
	}

	if err := os.Chdir("/"); err != nil {
		return fmt.Errorf("chdir: %w", err)
	}

	if err := unix.Unmount("/oldroot", unix.MNT_DETACH); err != nil {
		return fmt.Errorf("unmount oldroot: %w", err)
	}

	if err := os.Remove("/oldroot"); err != nil {
		return fmt.Errorf("remove oldroot: %w", err)
	}

	if err := os.MkdirAll("/proc", 0555); err != nil {
		return fmt.Errorf("create /proc: %w", err)
	}

	if err := unix.Mount("proc", "/proc", "proc", 0, ""); err != nil {
		return fmt.Errorf("mount proc: %w", err)
	}

	fmt.Printf("Exec: %q\n", command)

	path, err := exec.LookPath(command[0])
	if err != nil {
		return err
	}

	return unix.Exec(path, command, os.Environ())
}

func start(rootfs string, cfg Config) error {
	args := append([]string{ChildCommand, rootfs}, cfg.Command...)

	cmd := exec.Command("/proc/self/exe", args...)

	cmd.Stdin = os.Stdin
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr

	cmd.SysProcAttr = &syscall.SysProcAttr{
		Cloneflags: syscall.CLONE_NEWNS |
			syscall.CLONE_NEWPID |
			syscall.CLONE_NEWUTS |
			syscall.CLONE_NEWIPC,
	}

	return cmd.Run()
}
