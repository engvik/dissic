// package dissic are responsible for setting up and starting dissic.
// It holds the services needed and sets up and tears down everything.
package dissic

import (
	"context"
	"fmt"
	"log"
	"net/http"
	"os"
	"os/signal"
	"syscall"

	"github.com/engvik/dissic/config"
	"github.com/turnage/graw/reddit"
)

type spotifyService interface {
	Authenticate(openBrowser bool) error
	Listen()
	Log(s string)
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
	Log(s string)
}

// Service is the dissic service. It holds the config and all other services.
type Service struct {
	Config  *config.Config
	Spotify spotifyService
	Reddit  redditService
	HTTP    *http.Server
}

// Run starts the dissic service. It takes care of authentication, sets up
// listeneres and are responisble for properly tearing everything down.
func (s *Service) Run(ctx context.Context) {
	go func(s *http.Server) {
		if err := s.ListenAndServe(); err != http.ErrServerClosed {
			log.Fatalf("error starting http server: %s", err.Error())
		}
	}(s.HTTP)

	// Authenticate spotify
	s.Spotify.Log("awaiting authentication...")
	s.Spotify.Authenticate(s.Config.AuthOpenBrowser)
	s.Spotify.Log("authenticated!")

	// HTTP server no longer needed
	if err := s.HTTP.Shutdown(ctx); err != nil {
		fmt.Printf("error shutting down http server: %s", err.Error())
	}

	// Get and set Spotify user
	if err := s.Spotify.SetUser(); err != nil {
		log.Fatalf("error setting user ID: %s", err.Error())
	}

	// Get Spotify playlists
	if err := s.Spotify.PreparePlaylists(s.Config); err != nil {
		log.Fatalf("error preparing playlists: %s", err.Error())
	}

	// Prepare the reddit scanner
	if err := s.Reddit.PrepareScanner(); err != nil {
		log.Fatalf("error preparing reddit/graw scanner: %s", err.Error())
	}

	// Start listening and block until shutdown signal receieved
	func(ctx context.Context, s *Service) {
		shutdown := make(chan os.Signal, 1)
		signal.Notify(shutdown, os.Interrupt, syscall.SIGTERM)

		go s.Spotify.Listen()
		s.Spotify.Log("worker ready")
		go s.Reddit.Listen(shutdown)
		s.Reddit.Log("worker ready")

		<-shutdown

		s.Spotify.Close()
		s.Reddit.Close()

		if s.Config.Verbose {
			log.Println("bye, bye!")
		}
	}(ctx, s)
}
