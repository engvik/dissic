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

	s, err := spotify.New(cfg)
	if err != nil {
		log.Fatalf("error creating reddit client: %s", err.Error())
	}

	http.HandleFunc("/spotifyAuth", s.AuthHandler())
	go func() {
		if err := http.ListenAndServe(fmt.Sprintf(":%s", cfg.HTTPPort), nil); err != nil {
			log.Fatalf("error starting http server: %s", err.Error())
		}
	}()

	logger(cfg, "awaiting spotify authentication...")

	<-s.AuthChan

	logger(cfg, "spotify client authenticated!")

	if err := s.PreparePlaylist(cfg); err != nil {
		log.Fatalf("error preparing playlist: %s", err.Error())
	}

	logger(cfg, fmt.Sprintf("spotify playlist ready: %s (%s)", s.Playlist.Name, s.Playlist.ID))

	go s.Listen()

	logger(cfg, "spotify worker ready")

	r, err := reddit.New(cfg, s.MusicChan)
	if err != nil {
		log.Fatalf("error creating reddit client: %s", err.Error())
	}

	logger(cfg, "reddit worker ready")

	if err := r.Listen(); err != nil {
		log.Fatalf("reddit listen error: %s", err.Error())
	}
}

func logger(cfg *config.Config, s string) {
	if cfg.Verbose {
		log.Printf("reddify:\t%s\n", s)
	}
}
