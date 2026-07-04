package main

import (
	"fmt"
	"log"
	"os"

	kfs "github.com/Gthamsrim1/kaiten/fs"
	gofuse "github.com/hanwen/go-fuse/v2/fs"
	"github.com/hanwen/go-fuse/v2/fuse"
)

func main() {
	if len(os.Args) != 2 {
		fmt.Println("Usage: kaitenfs <mountpoint>")
		os.Exit(1)
	}

	mountPoint := os.Args[1]

	if _, err := os.Stat(mountPoint); os.IsNotExist(err) {
		if err := os.MkdirAll(mountPoint, 0755); err != nil {
			log.Fatal(err)
		}
	}

	fs := kfs.New()
	fs.Seed()

	server, err := gofuse.Mount(
		mountPoint,
		fs.Root,
		&gofuse.Options{
			MountOptions: fuse.MountOptions{
				Debug: false,
			},
		},
	)

	if err != nil {
		log.Fatal(err)
	}

	server.Wait()
}
