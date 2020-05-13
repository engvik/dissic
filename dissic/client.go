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
	"github.com/engvik/dissic/reddit"
	"github.com/engvik/dissic/spotify"
)

type Client struct {
	Config  *config.Config
	Spotify *spotify.Client
	Reddit  *reddit.Client
	HTTP    *http.Server
}

func (c *Client) Run(ctx context.Context) {
	go func(s *http.Server) {
		if err := s.ListenAndServe(); err != http.ErrServerClosed {
			log.Fatalf("error starting http server: %s", err.Error())
		}
	}(c.HTTP)

	// Authenticate spotify
	c.Spotify.Log("awaiting authentication...")
	c.Spotify.Authenticate()
	<-c.Spotify.AuthChan
	c.Spotify.Log("authenticated!")

	// HTTP server no longer needed
	if err := c.HTTP.Shutdown(ctx); err != nil {
		fmt.Printf("error shutting down http server: %s", err.Error())
	}

	// Get Spotify playlists
	if err := c.Spotify.GetPlaylists(c.Config); err != nil {
		log.Fatalf("error preparing playlists: %s", err.Error())
	}

	// Prepare the reddit scanner
	if err := c.Reddit.PrepareScanner(); err != nil {
		log.Fatalf("error preparing reddit/graw scanner: %s", err.Error())
	}

	// Start listening and block until shutdown signal receieved
	func(ctx context.Context, c *Client) {
		shutdown := make(chan os.Signal, 1)
		signal.Notify(shutdown, os.Interrupt, syscall.SIGTERM)

		go c.Spotify.Listen()
		c.Spotify.Log("worker ready")
		go c.Reddit.Listen(shutdown)
		c.Reddit.Log("worker ready")

		<-shutdown

		c.Spotify.Close()
		c.Reddit.Close()

		if c.Config.Verbose {
			log.Println("bye, bye!")
		}
	}(ctx, c)
}