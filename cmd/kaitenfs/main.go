package main

import (
	"flag"
	"fmt"
	"log"
	"os"
	"os/signal"
	"syscall"

	"github.com/Gthamsrim1/kaiten/internal/mountfs"
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

	server, createdMountPoint, err := mountfs.Mount(mountPoint, *debug)
	if err != nil {
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