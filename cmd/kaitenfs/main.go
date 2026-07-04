package main

import (
	"flag"
	"fmt"
	"log"
	"os"
	"os/signal"
	"syscall"

	kfs "github.com/Gthamsrim1/kaiten/fs"
	gofuse "github.com/hanwen/go-fuse/v2/fs"
	"github.com/hanwen/go-fuse/v2/fuse"
)

func main() {
	debug := flag.Bool("debug", false, "enable FUSE debug logging")
	flag.Parse()

	args := flag.Args()
	if len(args) != 1 {
		fmt.Println("Usage: kaitenfs [-debug] <mountpoint>")
		os.Exit(1)
	}

	mountPoint := args[0]

	createdMountPoint, err := ensureMountPoint(mountPoint)
	if err != nil {
		log.Fatal(err)
	}

	kaitenFS := kfs.New()
	kaitenFS.Seed()

	server, err := gofuse.Mount(
		mountPoint,
		kaitenFS.Root,
		&gofuse.Options{
			MountOptions: fuse.MountOptions{
				Debug: *debug,
			},
		},
	)
	if err != nil {
		if createdMountPoint {
			os.Remove(mountPoint)
		}
		log.Fatal(err)
	}

	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, syscall.SIGINT, syscall.SIGTERM)
	go func() {
		<-sigChan
		log.Println("received shutdown signal, unmounting...")
		if err := server.Unmount(); err != nil {
			log.Printf("unmount failed: %v (try: fusermount -u %s)", err, mountPoint)
		}
	}()

	server.Wait()

	if createdMountPoint {
		os.Remove(mountPoint)
	}
}

func ensureMountPoint(path string) (created bool, err error) {
	info, err := os.Stat(path)
	switch {
	case os.IsNotExist(err):
		if err := os.MkdirAll(path, 0755); err != nil {
			return false, fmt.Errorf("creating mountpoint: %w", err)
		}
		return true, nil

	case err != nil:
		return false, fmt.Errorf("checking mountpoint: %w", err)

	case !info.IsDir():
		return false, fmt.Errorf("mountpoint %q exists and is not a directory", path)
	}

	entries, err := os.ReadDir(path)
	if err != nil {
		return false, fmt.Errorf("reading mountpoint: %w", err)
	}
	if len(entries) > 0 {
		return false, fmt.Errorf("mountpoint %q is not empty", path)
	}

	return false, nil
}