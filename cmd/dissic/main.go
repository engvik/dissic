package main

import (
	"context"
	"fmt"
	"log"
	"net/http"

	"github.com/engvik/dissic/config"
	"github.com/engvik/dissic/dissic"
	"github.com/engvik/dissic/reddit"
	"github.com/engvik/dissic/spotify"
)

func main() {
	ctx := context.Background()

	// Load config from config file and environment
	cfg, err := config.Load()
	if err != nil {
		log.Fatalf("error parsing config: %s", err.Error())
	}

	if cfg.Verbose {
		log.Printf("dissic %s", cfg.Version)
	}

	// Set up spotify client
	s, err := spotify.New(cfg)
	if err != nil {
		log.Fatalf("error creating spotify client: %s", err.Error())
	}

	// Set up reddit client
	r, err := reddit.New(cfg, s.MusicChan)
	if err != nil {
		log.Fatalf("error creating reddit client: %s", err.Error())
	}

	// Set up http server
	mux := http.NewServeMux()
	mux.HandleFunc("/spotifyAuth", s.AuthHandler())

	// Set up dissic client
	d := dissic.Client{
		Config:  cfg,
		Spotify: s,
		Reddit:  r,
		HTTP: &http.Server{
			Addr:    fmt.Sprintf(":%d", cfg.HTTPPort),
			Handler: mux,
		},
	}

	// Run dissic client
	d.Run(ctx)
}
