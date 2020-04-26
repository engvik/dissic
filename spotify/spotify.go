package spotify

import (
	"fmt"
	"net/http"

	"github.com/engvik/reddify/config"
	"github.com/pkg/browser"
	"github.com/zmb3/spotify"
)

type Client struct {
	Auth     spotify.Authenticator
	AuthURL  string
	Session  string
	Playlist *spotify.FullPlaylist
	C        spotify.Client
}

const ErrInvalidID = "Invalid playlist Id"

func New(cfg *config.Config) (*Client, error) {
	auth := spotify.NewAuthenticator("http://localhost:1337/spotifyAuth", spotify.ScopePlaylistReadPrivate, spotify.ScopePlaylistReadPrivate, spotify.ScopePlaylistModifyPrivate, spotify.ScopePlaylistModifyPublic)

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

func (c *Client) AuthHandler(authenticated chan bool) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		token, err := c.Auth.Token(c.Session, r)
		if err != nil {
			http.Error(w, "Couldn't get token", http.StatusNotFound)
			return
		}

		c.C = c.Auth.NewClient(token)
		authenticated <- true
	}
}

func (c *Client) PreparePlaylist(name string) error {
	user, err := c.C.CurrentUser()
	if err != nil {
		return fmt.Errorf("error getting current user: %w", err)
	}

	res, err := c.C.GetPlaylistsForUser(user.ID)
	if err != nil {
		return fmt.Errorf("error getting playlists for %s: %w", user.ID, err)
	}

	var playlist *spotify.FullPlaylist

	for _, p := range res.Playlists {
		if p.Name == name {
			playlist, err = c.C.GetPlaylist(p.ID)
			if err != nil {
				return fmt.Errorf("error getting playlist %s (%s): %w", p.Name, p.ID, err)
			}

			break
		}
	}

	if playlist == nil {
		playlist, err = c.C.CreatePlaylistForUser(user.ID, name, "reddify", false)
		if err != nil {
			return fmt.Errorf("error creating playlist %s: %w", name, err)
		}
	}

	c.Playlist = playlist

	return nil
}
