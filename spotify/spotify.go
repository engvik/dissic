package spotify

import (
	"fmt"
	"log"
	"time"

	"github.com/engvik/dissic/config"
	"github.com/zmb3/spotify"
)

type Client struct {
	Auth              spotify.Authenticator
	AuthURL           string
	Session           string
	AuthChan          chan bool
	MusicChan         chan Music
	C                 spotify.Client
	SubredditPlaylist map[string]spotify.ID
	Verbose           bool
}

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

func (c *Client) Authenticate() error {
	// TODO: Config if auto-browser
	c.Log(fmt.Sprintf("open url to authenticate: %s", c.AuthURL))
	/*	if err := browser.OpenURL(c.AuthURL); err != nil {
		return fmt.Errorf("opening url (%s): %w", c.AuthURL, err)
	} */

	return nil
}

func (c *Client) Listen() {
	for {
		select {
		case m := <-c.MusicChan:
			go c.handle(m)
		}
	}
}

func (c *Client) handle(m Music) {
	for _, title := range m.titleStringArray() {
		if title == "" {
			continue
		}

		track, err := c.getTrack(title)
		if err != nil {
			c.Log(fmt.Sprintf("\ttrack: %s", err.Error()))
			continue
		}

		playlist, ok := c.SubredditPlaylist[m.Subreddit]
		if !ok {
			c.Log(fmt.Sprintf("no playlist found for subreddit: %s", m.Subreddit))
		}

		if err := c.addToPlaylist(playlist, track.ID); err != nil {
			c.Log(fmt.Sprintf("\teadding track to playlist: %s", err.Error()))
		}

		return
	}
}

func (c *Client) Log(s string) {
	if c.Verbose {
		log.Printf("spotify:\t%s\n", s)
	}
}

func (c *Client) Close() {
	c.Log("shutting down")
	close(c.AuthChan)
	close(c.MusicChan)
}
