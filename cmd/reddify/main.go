package main

import (
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
	// Parse command line arguments into a config struct
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

	// Set up a very basic http server to handle auth callbacks
	http.HandleFunc("/spotifyAuth", s.AuthHandler())
	go func() {
		if err := http.ListenAndServe(fmt.Sprintf(":%d", cfg.HTTPPort), nil); err != nil {
			log.Fatalf("error starting http server: %s", err.Error())
		}
	}()

	// Authenticate spotify
	s.Log("awaiting authentication...")
	s.Authenticate()
	<-s.AuthChan
	s.Log("authenticated!")

	// Get Spotify playlists
	if err := s.GetPlaylists(cfg); err != nil {
		log.Fatalf("error preparing playlists: %s", err.Error())
	}

	// Prepare the reddit scanner
	if err := r.PrepareScanner(); err != nil {
		log.Fatalf("error preparing reddit/graw scanner: %s", err.Error())
	}

	// Start listening and block until shutdown signal receieved
	func(s *spotify.Client, r *reddit.Client) {
		shutdown := make(chan os.Signal, 1)
		signal.Notify(shutdown, os.Interrupt, syscall.SIGTERM)

		go s.Listen()
		s.Log("worker ready")
		go r.Listen(shutdown)
		r.Log("worker ready")

		<-shutdown

		s.Close()
		r.Close()

		if cfg.Verbose {
			log.Println("bye, bye!")
		}
	}(s, r)
}
