package main

import (
	"context"
	"fmt"
	"log"
	"net/http"
	"os"
	"os/signal"
	"syscall"

	"github.com/engvik/dissic/config"
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
	d := dissic{
		config:  cfg,
		spotify: s,
		reddit:  r,
		http: &http.Server{
			Addr:    fmt.Sprintf(":%d", cfg.HTTPPort),
			Handler: mux,
		},
	}

	d.start(ctx)
}

type dissic struct {
	config  *config.Config
	spotify *spotify.Client
	reddit  *reddit.Client
	http    *http.Server
}

func (d *dissic) start(ctx context.Context) {
	go func(s *http.Server) {
		if err := s.ListenAndServe(); err != nil {
			log.Fatalf("error starting http server: %s", err.Error())
		}
	}(d.http)

	// Authenticate spotify
	d.spotify.Log("awaiting authentication...")
	d.spotify.Authenticate()
	<-d.spotify.AuthChan
	d.spotify.Log("authenticated!")

	// Get Spotify playlists
	if err := d.spotify.GetPlaylists(d.config); err != nil {
		log.Fatalf("error preparing playlists: %s", err.Error())
	}

	// Prepare the reddit scanner
	if err := d.reddit.PrepareScanner(); err != nil {
		log.Fatalf("error preparing reddit/graw scanner: %s", err.Error())
	}

	// Start listening and block until shutdown signal receieved
	func(ctx context.Context, d *dissic) {
		shutdown := make(chan os.Signal, 1)
		signal.Notify(shutdown, os.Interrupt, syscall.SIGTERM)

		go d.spotify.Listen()
		d.spotify.Log("worker ready")
		go d.reddit.Listen(shutdown)
		d.reddit.Log("worker ready")

		<-shutdown

		d.spotify.Close()
		d.reddit.Close()

		if err := d.http.Shutdown(ctx); err != nil {
			fmt.Printf("error shutting down http server: %s", err.Error())
		}

		if d.config.Verbose {
			log.Println("bye, bye!")
		}
	}(ctx, d)
}
