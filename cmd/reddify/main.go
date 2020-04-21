package main

import (
	"log"

	"github.com/engvik/reddify/config"
	"github.com/engvik/reddify/reddit"
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

	redditClient, err := reddit.New(cfg)
	if err != nil {
		log.Fatalf("error creating reddit client: %s", err.Error())
	}

	if err := redditClient.Listen(); err != nil {
		log.Fatalf("reddit listen error: %s", err.Error())
	}
}
