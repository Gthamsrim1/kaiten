package main

import (
	"log"
	"os"

	"github.com/Gthamsrim1/kaiten/runtime"
)

func main() {
	rt := runtime.New()

	if err := rt.Run(os.Args[1:]); err != nil {
		log.Fatal(err)
	}
}
