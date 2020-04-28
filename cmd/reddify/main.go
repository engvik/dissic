package main

import (
	"fmt"
	"log"
	"net/http"

	"github.com/engvik/reddify/config"
	"github.com/engvik/reddify/reddit"
	"github.com/engvik/reddify/spotify"
)

func main() {
	cfg, err := config.Parse()
	if err != nil {
		log.Fatalf("error parsing config: %s", err.Error())
	}

	logger(cfg, fmt.Sprintf("version: %s", cfg.Version))

	spotifyClient, err := spotify.New(cfg)
	if err != nil {
		log.Fatalf("error creating reddit client: %s", err.Error())
	}

	spotGotAuth := make(chan bool)

	logger(cfg, "awaiting spotify authentication...")

	http.HandleFunc("/spotifyAuth", spotifyClient.AuthHandler(spotGotAuth))
	go func() {
		if err := http.ListenAndServe(":1337", nil); err != nil {
			log.Fatalf("error starting http server: %s", err.Error())
		}
	}()

	<-spotGotAuth

	logger(cfg, "spotify client authenticated!")

	if err := spotifyClient.PreparePlaylist(cfg); err != nil {
		log.Fatalf("error preparing playlist: %s", err.Error())
	}

	logger(cfg, fmt.Sprintf("spotify playlist ready: %s (%s)", spotifyClient.Playlist.Name, spotifyClient.Playlist.ID))

	go spotifyClient.Listen()

	logger(cfg, "spotify worker ready")

	redditClient, err := reddit.New(cfg, spotifyClient.MusicChan)
	if err != nil {
		log.Fatalf("error creating reddit client: %s", err.Error())
	}

	logger(cfg, "reddit worker ready")

	if err := redditClient.Listen(); err != nil {
		log.Fatalf("reddit listen error: %s", err.Error())
	}
}

func logger(cfg *config.Config, s string) {
	if cfg.Verbose {
		log.Printf("reddify:\t%s\n", s)
	}
}
