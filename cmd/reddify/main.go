package main

import (
	"fmt"
	"log"

	"github.com/engvik/reddify/config"
)

func main() {

	// TODO:
	// * Load config
	// * Read subreddits
	// * Pick out artist / song
	// * Add to spotify playlist

	cfg, err := config.Parse()
	if err != nil {
		log.Fatalf("error parsing config: %s", err.Error())
	}

	fmt.Println("reddify - reddit to spotify playlist")
	fmt.Println(cfg)
}
