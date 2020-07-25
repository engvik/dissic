// Package dissic are responsible for setting up and starting dissic.
// It holds the services needed and sets up and tears down everything.
package dissic

import (
	"context"
	"fmt"
	"net/http"
	"os"
	"os/signal"
	"syscall"

	"github.com/engvik/dissic/internal/config"
	log "github.com/sirupsen/logrus"
	"github.com/turnage/graw/reddit"
)

type spotifyService interface {
	Authenticate(openBrowser bool) error
	Listen()
	Close()
	PreparePlaylists(cfg *config.Config) error
	AuthHandler() http.HandlerFunc
	SetUser() error
}

type redditService interface {
	PrepareScanner() error
	Listen(shutdown chan<- os.Signal)
	Close()
	Post(post *reddit.Post) error
}

// Service is the dissic service. It holds the config and all other services.
type Service struct {
	Config  *config.Config
	Spotify spotifyService
	Reddit  redditService
	HTTP    *http.Server
}

// New returns a new dissic service.
func New(cfg *config.Config, s spotifyService, r redditService, mux *http.ServeMux) *Service {
	d := &Service{
		Config:  cfg,
		Spotify: s,
		Reddit:  r,
		HTTP: &http.Server{
			Addr:    fmt.Sprintf(":%d", cfg.HTTPPort),
			Handler: mux,
		},
	}

	return d
}

// Start starts the dissic service. It takes care of authentication, sets up
// listeneres and are responisble for properly tearing everything down.
func (s *Service) Start(ctx context.Context) {
	go func(s *http.Server) {
		if err := s.ListenAndServe(); err != http.ErrServerClosed {
			log.Fatalf("error starting http server: %s", err)
		}
	}(s.HTTP)

	// Authenticate spotify
	log.WithFields(log.Fields{"service": "spotify"}).Infoln("awaiting authentication...")
	s.Spotify.Authenticate(s.Config.AuthOpenBrowser)
	log.WithFields(log.Fields{"service": "spotify"}).Infoln("authenticated!")

	// HTTP server no longer needed
	if err := s.HTTP.Shutdown(ctx); err != nil {
		fmt.Printf("error shutting down http server: %s", err)
	}

	// Get and set Spotify user
	if err := s.Spotify.SetUser(); err != nil {
		log.Fatalf("error setting user ID: %s", err)
	}

	// Get Spotify playlists
	if err := s.Spotify.PreparePlaylists(s.Config); err != nil {
		log.Fatalf("error preparing playlists: %s", err)
	}

	// Prepare the reddit scanner
	if err := s.Reddit.PrepareScanner(); err != nil {
		log.Fatalf("error preparing reddit/graw scanner: %s", err)
	}

	// Start listening and block until shutdown signal receieved
	func(ctx context.Context, s *Service) {
		shutdown := make(chan os.Signal, 1)
		signal.Notify(shutdown, os.Interrupt, syscall.SIGTERM)

		go s.Spotify.Listen()
		log.WithFields(log.Fields{"service": "spotify"}).Infoln("helper ready")
		go s.Reddit.Listen(shutdown)
		log.WithFields(log.Fields{"service": "reddit"}).Infoln("helper ready")

		<-shutdown

		s.Spotify.Close()
		s.Reddit.Close()

		log.WithFields(log.Fields{"service": "dissic"}).Infoln("bye, bye!")
	}(ctx, s)
}
