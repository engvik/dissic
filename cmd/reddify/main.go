package main

import (
	"log"
	"net/http"

	"github.com/engvik/reddify/config"
	"github.com/engvik/reddify/reddit"
	"github.com/engvik/reddify/spotify"
)

func main() {

	// TODO:
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

	spotifyClient, err := spotify.New(cfg)
	if err != nil {
		log.Fatalf("error creating reddit client: %s", err.Error())
	}

	http.HandleFunc("/spotifyAuth", spotifyClient.AuthHandler())
	go http.ListenAndServe(":1337", nil)

	if err := redditClient.Listen(); err != nil {
		log.Fatalf("reddit listen error: %s", err.Error())
	}
}
