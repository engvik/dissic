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

	spotifyClient, err := spotify.New(cfg)
	if err != nil {
		log.Fatalf("error creating reddit client: %s", err.Error())
	}

	spotGotAuth := make(chan bool)

	http.HandleFunc("/spotifyAuth", spotifyClient.AuthHandler(spotGotAuth))
	go func() {
		if err := http.ListenAndServe(":1337", nil); err != nil {
			log.Fatalf("error starting http server: %s", err.Error())
		}
	}()

	<-spotGotAuth

	if err := spotifyClient.PreparePlaylist(cfg.Spotify.Playlist); err != nil {
		log.Fatalf("error preparing playlist: %s", err.Error())
	}

	go spotifyClient.Listen()

	redditClient, err := reddit.New(cfg, spotifyClient.MusicChan)
	if err != nil {
		log.Fatalf("error creating reddit client: %s", err.Error())
	}

	if err := redditClient.Listen(); err != nil {
		log.Fatalf("reddit listen error: %s", err.Error())
	}
}
