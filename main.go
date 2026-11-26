package main

import (
	"log"

	"exo_API/internal/web"
)

func main() {
	if err := web.Serve(); err != nil {
		log.Fatal(err)
	}
}
