package main

import (
	"context"
	"net/http"

	"github.com/engvik/dissic/internal/config"
	"github.com/engvik/dissic/internal/dissic"
	"github.com/engvik/dissic/internal/reddit"
	"github.com/engvik/dissic/internal/spotify"
	log "github.com/sirupsen/logrus"
)

func main() {
	ctx := context.Background()

	log.SetFormatter(&log.TextFormatter{
		FullTimestamp: true,
	})

	// Load config from config file and environment
	cfg, err := config.Load()
	if err != nil {
		log.Fatalf("error parsing config: %s", err)
	}

	// Set up spotify service
	s, err := spotify.New(cfg)
	if err != nil {
		log.Fatalf("error creating spotify client: %s", err)
	}

	// Set up reddit service
	r, err := reddit.New(cfg, s.MusicChan)
	if err != nil {
		log.Fatalf("error creating reddit client: %s", err)
	}

	// Set up http server
	mux := http.NewServeMux()
	mux.HandleFunc("/spotifyAuth", s.AuthHandler())

	// Set up dissic service
	d := dissic.New(cfg, s, r, mux)

	// Run dissic
	d.Run(ctx)
}
