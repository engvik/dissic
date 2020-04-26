package main

import (
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

	if cfg.Verbose {
		log.Printf("reddify %s\n", cfg.Version)
	}

	spotifyClient, err := spotify.New(cfg)
	if err != nil {
		log.Fatalf("error creating reddit client: %s", err.Error())
	}

	spotGotAuth := make(chan bool)

	if cfg.Verbose {
		log.Println("Awaiting Spotify authentication...")
	}

	http.HandleFunc("/spotifyAuth", spotifyClient.AuthHandler(spotGotAuth))
	go func() {
		if err := http.ListenAndServe(":1337", nil); err != nil {
			log.Fatalf("error starting http server: %s", err.Error())
		}
	}()

	<-spotGotAuth

	if cfg.Verbose {
		log.Println("Spotify client authenticated!")
	}

	if err := spotifyClient.PreparePlaylist(cfg.Spotify.Playlist); err != nil {
		log.Fatalf("error preparing playlist: %s", err.Error())
	}

	if cfg.Verbose {
		log.Printf("Spotify playlist ready: %s (%s)\n", spotifyClient.Playlist.Name, spotifyClient.Playlist.ID)
	}

	go spotifyClient.Listen()

	if cfg.Verbose {
		log.Println("Spotify worker ready")
	}

	redditClient, err := reddit.New(cfg, spotifyClient.MusicChan)
	if err != nil {
		log.Fatalf("error creating reddit client: %s", err.Error())
	}

	if cfg.Verbose {
		log.Println("Reddit worker ready")
	}

	if err := redditClient.Listen(); err != nil {
		log.Fatalf("reddit listen error: %s", err.Error())
	}
}
