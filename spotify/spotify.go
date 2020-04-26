package spotify

import (
	"fmt"
	"net/http"

	"github.com/engvik/reddify/config"
	"github.com/pkg/browser"
	"github.com/zmb3/spotify"
)

type Client struct {
	Auth    spotify.Authenticator
	AuthURL string
	Session string
	C       spotify.Client
}

func New(cfg *config.Config) (*Client, error) {
	auth := spotify.NewAuthenticator("http://localhost:1337/spotifyAuth", spotify.ScopePlaylistModifyPrivate, spotify.ScopeUserReadEmail)

	c := Client{
		Auth:    auth,
		Session: "asdasdff", // TODO: Generate
	}

	c.Auth.SetAuthInfo(cfg.Spotify.ClientID, cfg.Spotify.ClientSecret)
	c.AuthURL = c.Auth.AuthURL(c.Session)

	if err := browser.OpenURL(c.AuthURL); err != nil {
		return nil, fmt.Errorf("error opening url: %w", err)
	}

	return &c, nil
}

func (c *Client) AuthHandler() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		token, err := c.Auth.Token(c.Session, r)
		if err != nil {
			http.Error(w, "Couldn't get token", http.StatusNotFound)
			return
		}

		c.C = c.Auth.NewClient(token)
	}
}
