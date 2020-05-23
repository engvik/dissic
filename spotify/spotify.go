// package spotify contains an implementaion of the spotify service.
package spotify

import (
	"fmt"
	"log"
	"time"

	"github.com/engvik/dissic/config"
	"github.com/pkg/browser"
	"github.com/zmb3/spotify"
)

// Client is the spotify client
type Client struct {
	Auth              spotify.Authenticator
	AuthURL           string
	Session           string
	AuthChan          chan bool
	MusicChan         chan Music
	C                 spotify.Client
	SubredditPlaylist map[string]spotify.ID
	Verbose           bool
	User              *spotify.PrivateUser
}

// New sets up a new spotify client. It takes the configuration and returns
// a client or an error.
func New(cfg *config.Config) (*Client, error) {
	callbackURL := fmt.Sprintf("http://localhost:%d/spotifyAuth", cfg.HTTPPort)
	auth := spotify.NewAuthenticator(callbackURL, spotify.ScopePlaylistReadPrivate, spotify.ScopePlaylistReadPrivate, spotify.ScopePlaylistModifyPrivate, spotify.ScopePlaylistModifyPublic)

	c := Client{
		Auth:      auth,
		Session:   fmt.Sprintf("dissic:%d", time.Now().Unix()),
		AuthChan:  make(chan bool),
		MusicChan: make(chan Music),
		Verbose:   cfg.Verbose,
	}

	c.Auth.SetAuthInfo(cfg.Spotify.ClientID, cfg.Spotify.ClientSecret)
	c.AuthURL = c.Auth.AuthURL(c.Session)

	c.Log("client setup ok")

	return &c, nil
}

// Authenticate handles the authentication against the Spotify API.
// It either opens the browser or tells the user to navigate to a URL.
// It will also block until authentication is done.
func (c *Client) Authenticate(openBrowser bool) error {
	if openBrowser {
		if err := browser.OpenURL(c.AuthURL); err != nil {
			return fmt.Errorf("opening url (%s): %w", c.AuthURL, err)
		}
	} else {
		c.Log(fmt.Sprintf("open url to authenticate: %s", c.AuthURL))
	}

	// Block until authenticated
	<-c.AuthChan

	return nil
}

// Listen listens for incoming data on the music channel.
func (c *Client) Listen() {
	for {
		select {
		case m := <-c.MusicChan:
			go c.handle(m)
		}
	}
}

func (c *Client) handle(m Music) {
	// TODO: Refactor and cleanup this
	if m.URL != "" {
		track, err := c.getTrackByURL(m.URL)
		if err != nil {
			c.Log(fmt.Sprintf("\ttrack by url: %s", err.Error()))
		}

		if track != nil {
			if err := c.addToPlaylist(m.Subreddit, track.ID); err != nil {
				c.Log(fmt.Sprintf("\tadding track to playlist: %s", err.Error()))
			}

			return
		}
	}

	track, err := c.getTrackByTitles(m)
	if err != nil {
		c.Log(fmt.Sprintf("\ttrack by title: %s", err.Error()))
		return
	}

	if err := c.addToPlaylist(m.Subreddit, track.ID); err != nil {
		c.Log(fmt.Sprintf("\tadding track to playlist: %s", err.Error()))
	}
}

// Log logs..
func (c *Client) Log(s string) {
	if c.Verbose {
		log.Printf("spotify:\t%s\n", s)
	}
}

// Closes properly closes the Spotify client
func (c *Client) Close() {
	c.Log("shutting down")
	close(c.AuthChan)
	close(c.MusicChan)
}
