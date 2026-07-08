package main

import (
	"flag"
	"fmt"
	"log"
	"os"
	"os/signal"
	"syscall"

	"github.com/Gthamsrim1/kaiten/internal/mountfs"
	"github.com/Gthamsrim1/kaiten/internal/persist"
)

func main() {
	debug := flag.Bool("debug", false, "enable FUSE debug logging")
	flag.Parse()

	args := flag.Args()

	var (
		mountPoint string
		repo       = "./kaiten-data"
	)

	switch len(args) {
	case 1:
		mountPoint = args[0]
	case 2:
		repo = args[0]
		mountPoint = args[1]
	default:
		fmt.Println("Usage:")
		fmt.Println("  kaitenfs [-debug] <mountpoint>")
		fmt.Println("  kaitenfs [-debug] <repository> <mountpoint>")
		os.Exit(1)
	}

	fs, server, createdMountPoint, err := mountfs.Mount(repo, mountPoint, *debug)
	if err != nil {
		log.Fatal(err)
	}

	sigChan := make(chan os.Signal, 1)
	signal.Notify(
		sigChan,
		syscall.SIGINT,
		syscall.SIGTERM,
		syscall.SIGQUIT,
	)

	go func() {
		<-sigChan
		log.Println("received shutdown signal, unmounting...")

		if err := server.Unmount(); err != nil {
			log.Printf("unmount failed: %v (try: fusermount -u %s)", err, mountPoint)
		}
	}()

	server.Wait()

	snapshot, err := fs.Snapshot()
	if err != nil {
		if createdMountPoint {
			_ = os.Remove(mountPoint)
		}
		log.Fatal(err)
	}

	saveErr := persist.Save(repo, snapshot)

	if createdMountPoint {
		_ = os.Remove(mountPoint)
	}

	if saveErr != nil {
		log.Fatal(saveErr)
	}
}