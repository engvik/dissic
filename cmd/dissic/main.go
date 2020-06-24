package main

import (
	"context"
	"log"
	"net/http"

	"github.com/engvik/dissic/internal/config"
	"github.com/engvik/dissic/internal/reddit"
	"github.com/engvik/dissic/internal/spotify"
	"github.com/engvik/dissic/pkg/dissic"
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

	// Set up spotify service
	s, err := spotify.New(cfg)
	if err != nil {
		log.Fatalf("error creating spotify client: %s", err.Error())
	}

	// Set up reddit service
	r, err := reddit.New(cfg, s.MusicChan)
	if err != nil {
		log.Fatalf("error creating reddit client: %s", err.Error())
	}

	// Set up http server
	mux := http.NewServeMux()
	mux.HandleFunc("/spotifyAuth", s.AuthHandler())

	// Set up dissic service
	d := dissic.New(cfg, s, r, mux)

	// Run dissic
	d.Run(ctx)
}
