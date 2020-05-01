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
	cfg, err := config.Parse()
	if err != nil {
		log.Fatalf("error parsing config: %s", err.Error())
	}

	if cfg.Verbose {
		log.Printf("reddify %s", cfg.Version)
	}

	s, err := spotify.New(cfg)
	if err != nil {
		log.Fatalf("error creating reddit client: %s", err.Error())
	}

	r, err := reddit.New(cfg, s.MusicChan)
	if err != nil {
		log.Fatalf("error creating reddit client: %s", err.Error())
	}

	http.HandleFunc("/spotifyAuth", s.AuthHandler())
	go func() {
		if err := http.ListenAndServe(fmt.Sprintf(":%s", cfg.HTTPPort), nil); err != nil {
			log.Fatalf("error starting http server: %s", err.Error())
		}
	}()

	s.Log("awaiting authentication...")

	s.Authenticate()
	<-s.AuthChan

	s.Log("authenticated!")

	if err := s.PreparePlaylist(cfg); err != nil {
		log.Fatalf("error preparing playlist: %s", err.Error())
	}

	s.Log(fmt.Sprintf("spotify playlist ready: %s (%s)", s.Playlist.Name, s.Playlist.ID))

	if err := r.PrepareScanner(); err != nil {
		log.Fatalf("error preparing reddit/graw scanner: %s", err.Error())
	}

	func(s *spotify.Client, r *reddit.Client) {
		shutdown := make(chan os.Signal, 1)
		signal.Notify(shutdown, os.Interrupt, syscall.SIGTERM)

		go s.Listen()
		s.Log("worker ready")
		go r.Listen()
		r.Log("worker ready")

		<-shutdown

		s.Close()
		r.Close()

		if cfg.Verbose {
			log.Println("bye, bye!")
		}
	}(s, r)
}
