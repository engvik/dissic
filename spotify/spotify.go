package spotify

import (
	"fmt"
	"log"

	"github.com/engvik/reddify/config"
	"github.com/pkg/browser"
	"github.com/zmb3/spotify"
)

type Client struct {
	Auth      spotify.Authenticator
	AuthURL   string
	Session   string
	AuthChan  chan bool
	MusicChan chan Music
	Playlist  *spotify.FullPlaylist
	C         spotify.Client
	Verbose   bool
}

func New(cfg *config.Config) (*Client, error) {
	callbackURL := fmt.Sprintf("http://localhost:%s/spotifyAuth", cfg.HTTPPort)
	auth := spotify.NewAuthenticator(callbackURL, spotify.ScopePlaylistReadPrivate, spotify.ScopePlaylistReadPrivate, spotify.ScopePlaylistModifyPrivate, spotify.ScopePlaylistModifyPublic)

	c := Client{
		Auth:      auth,
		Session:   "asdasdff", // TODO: Generate
		AuthChan:  make(chan bool),
		MusicChan: make(chan Music),
		Verbose:   cfg.Verbose,
	}

	c.Auth.SetAuthInfo(cfg.Spotify.ClientID, cfg.Spotify.ClientSecret)
	c.AuthURL = c.Auth.AuthURL(c.Session)

	if err := browser.OpenURL(c.AuthURL); err != nil {
		return nil, fmt.Errorf("error opening url: %w", err)
	}

	c.log("client setup ok")

	return &c, nil
}

func (c *Client) Listen() {
	for {
		select {
		case m := <-c.MusicChan:
			go c.Handle(m)
		}
	}
}

func (c *Client) Handle(m Music) {
	for _, title := range m.titleStringArray() {
		if title == "" {
			continue
		}

		track, err := c.getTrack(title)
		if err != nil {
			c.log(fmt.Sprintf("\terror getting track: %s", err.Error()))
			continue
		}

		if err := c.addToPlaylist(track.ID); err != nil {
			c.log(fmt.Sprintf("\terror adding track to playlist: %s", err.Error()))
		}

		return
	}
}

func (c *Client) log(s string) {
	if c.Verbose {
		log.Printf("spotify:\t%s\n", s)
	}
}
