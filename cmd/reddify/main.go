package main

import (
	"context"
	"fmt"
	"log"
	"net/http"
	"os"
	"os/signal"
	"syscall"

	"github.com/engvik/reddify/config"
	"github.com/engvik/reddify/reddit"
	"github.com/engvik/reddify/spotify"
)

func main() {
	ctx := context.Background()

	// Load config from config file and environment
	cfg, err := config.Load()
	if err != nil {
		log.Fatalf("error parsing config: %s", err.Error())
	}

	if cfg.Verbose {
		log.Printf("reddify %s", cfg.Version)
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
	rdfy := reddify{
		config:  cfg,
		spotify: s,
		reddit:  r,
		http: &http.Server{
			Addr:    fmt.Sprintf(":%d", cfg.HTTPPort),
			Handler: mux,
		},
	}

	rdfy.start(ctx)
}

type reddify struct {
	config  *config.Config
	spotify *spotify.Client
	reddit  *reddit.Client
	http    *http.Server
}

func (r *reddify) start(ctx context.Context) {
	go func(s *http.Server) {
		if err := s.ListenAndServe(); err != nil {
			log.Fatalf("error starting http server: %s", err.Error())
		}
	}(r.http)

	// Authenticate spotify
	r.spotify.Log("awaiting authentication...")
	r.spotify.Authenticate()
	<-r.spotify.AuthChan
	r.spotify.Log("authenticated!")

	// Get Spotify playlists
	if err := r.spotify.GetPlaylists(r.config); err != nil {
		log.Fatalf("error preparing playlists: %s", err.Error())
	}

	// Prepare the reddit scanner
	if err := r.reddit.PrepareScanner(); err != nil {
		log.Fatalf("error preparing reddit/graw scanner: %s", err.Error())
	}

	// Start listening and block until shutdown signal receieved
	func(ctx context.Context, r *reddify) {
		shutdown := make(chan os.Signal, 1)
		signal.Notify(shutdown, os.Interrupt, syscall.SIGTERM)

		go r.spotify.Listen()
		r.spotify.Log("worker ready")
		go r.reddit.Listen(shutdown)
		r.reddit.Log("worker ready")

		<-shutdown

		r.spotify.Close()
		r.reddit.Close()

		if err := r.http.Shutdown(ctx); err != nil {
			fmt.Printf("error shutting down http server: %s", err.Error())
		}

		if r.config.Verbose {
			log.Println("bye, bye!")
		}
	}(ctx, r)
}
