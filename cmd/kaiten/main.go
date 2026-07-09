package main

import (
	"flag"
	"fmt"
	"log"
	"os"

	"github.com/Gthamsrim1/kaiten/internal/persist"
)

func main() {
	repo := flag.String("repo", "./kaiten-data", "repository path")
	flag.Parse()

	args := flag.Args()

	if len(args) == 0 {
		usage()
	}

	switch args[0] {

	case "snapshots":
		snapshots, err := persist.ListSnapshots(*repo)
		if err != nil {
			log.Fatal(err)
		}

		for _, s := range snapshots {
			prefix := " "
			if s.IsHEAD {
				prefix = "*"
			}

			if s.ParentID == nil {
				fmt.Printf("%s %s\n", prefix, s.ID)
			} else {
				fmt.Printf("%s %s <- %s\n", prefix, s.ID, *s.ParentID)
			}
		}

	case "log":
		history, err := persist.Log(*repo)
		if err != nil {
			log.Fatal(err)
		}

		for i, s := range history {
			prefix := " "

			if s.IsHEAD {
				prefix = "*"
			}

			fmt.Printf("%s %s\n", prefix, s.ID)

			if i != len(history)-1 {
				fmt.Println("│")
			}
		}

	case "checkout":
		if len(args) != 2 {
			log.Fatal("usage: kaiten checkout <snapshot-id>")
		}

		if err := persist.Checkout(*repo, args[1]); err != nil {
			log.Fatal(err)
		}

	case "gc":
		if err := persist.GC(*repo); err != nil {
			log.Fatal(err)
		}

	default:
		usage()
	}
}

func usage() {
	fmt.Println("Usage:")
	fmt.Println("  kaiten snapshots")
	fmt.Println("  kaiten <repo> snapshots")
	fmt.Println("  kaiten checkout <snapshot-id>")
	fmt.Println("  kaiten <repo> checkout <snapshot-id>")
	os.Exit(1)
}
