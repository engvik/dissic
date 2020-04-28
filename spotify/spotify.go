package spotify

import (
	"errors"
	"fmt"
	"log"
	"net/http"

	"github.com/engvik/reddify/config"
	"github.com/pkg/browser"
	"github.com/zmb3/spotify"
)

type Music struct {
	Sub              string
	PostTitle        string
	MediaTitle       string
	SecureMediaTitle string
}

type MusicChan chan Music

type Client struct {
	Auth      spotify.Authenticator
	AuthURL   string
	Session   string
	MusicChan MusicChan
	Playlist  *spotify.FullPlaylist
	C         spotify.Client
	Verbose   bool
}

const ErrInvalidID = "Invalid playlist Id"

func New(cfg *config.Config) (*Client, error) {
	auth := spotify.NewAuthenticator("http://localhost:1337/spotifyAuth", spotify.ScopePlaylistReadPrivate, spotify.ScopePlaylistReadPrivate, spotify.ScopePlaylistModifyPrivate, spotify.ScopePlaylistModifyPublic)

	c := Client{
		Auth:      auth,
		Session:   "asdasdff", // TODO: Generate
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

func (c *Client) AuthHandler(authenticated chan bool) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		token, err := c.Auth.Token(c.Session, r)
		if err != nil {
			http.Error(w, "Couldn't get token", http.StatusNotFound)
			return
		}

		c.C = c.Auth.NewClient(token)
		authenticated <- true
		w.Write([]byte("All good - you can close this window now"))
	}
}

func (c *Client) PreparePlaylist(cfg *config.Config) error {
	var playlist *spotify.FullPlaylist
	var err error

	if cfg.Spotify.PlaylistID != "" {
		playlist, err = c.preparePlaylistByID(cfg.Spotify.PlaylistID)
	} else if cfg.Spotify.PlaylistName != "" {
		playlist, err = c.preparePlaylistByName(cfg.Spotify.PlaylistName)
	}

	if err != nil {
		return fmt.Errorf("error getting playlist: %w", err)
	}

	if playlist == nil {
		return errors.New("unable to get spotify playlist")
	}

	c.Playlist = playlist

	return nil
}

func (c *Client) preparePlaylistByID(ID string) (*spotify.FullPlaylist, error) {
	playlist, err := c.C.GetPlaylist(spotify.ID(ID))
	if err != nil {
		return nil, fmt.Errorf("error getting playlist %s: %w", ID, err)
	}

	c.log(fmt.Sprintf("found playlist: %s (%s)", playlist.Name, playlist.ID))

	return playlist, nil
}

func (c *Client) preparePlaylistByName(name string) (*spotify.FullPlaylist, error) {
	user, err := c.C.CurrentUser()
	if err != nil {
		return nil, fmt.Errorf("error getting current user: %w", err)
	}

	c.log(fmt.Sprintf("retrived user: %s", user.ID))

	res, err := c.C.GetPlaylistsForUser(user.ID)
	if err != nil {
		return nil, fmt.Errorf("error getting playlists for %s: %w", user.ID, err)
	}

	var playlist *spotify.FullPlaylist

	for _, p := range res.Playlists {
		if p.Name == name {
			playlist, err = c.C.GetPlaylist(p.ID)
			if err != nil {
				return nil, fmt.Errorf("error getting playlist %s (%s): %w", p.Name, p.ID, err)
			}

			c.log(fmt.Sprintf("found playlist: %s (%s)", p.Name, p.ID))

			break
		}
	}

	if playlist == nil {
		playlist, err = c.C.CreatePlaylistForUser(user.ID, name, "reddify", false)
		if err != nil {
			return nil, fmt.Errorf("error creating playlist %s: %w", name, err)
		}

		c.log(fmt.Sprintf("created playlist: %s", name))
	}

	return playlist, nil
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
	c.log(fmt.Sprintf("%+v\n", m))

	// TODO:
	// - Pick best title
	// - try to retrive artist + song
	// - look for year
	// - maybe even grab tags?

	if m.PostTitle != "" {
		res, err := c.Search(m.PostTitle)
		if err != nil {
			log.Println(err.Error())
		}

		log.Printf("search: %+v", *res)

		c.log(fmt.Sprintf("tracks: %d", len(res.Tracks.Tracks)))
		c.log(fmt.Sprintf("albums: %d", len(res.Albums.Albums)))
		c.log(fmt.Sprintf("artists: %d", len(res.Artists.Artists)))
		for _, t := range res.Tracks.Tracks {
			c.log("track: " + t.String())
		}

		for _, a := range res.Albums.Albums {
			c.log("album: " + a.Name)
		}
		for _, a := range res.Artists.Artists {
			c.log("artist: " + a.Name)
		}

		// TODO: Determine if ok
	}

	/*	if m.SecureMediaTitle != "" {
			res, err := c.Search(m.SecureMediaTitle)
			if err != nil {
				log.Println(err.Error())
			}

			log.Printf("%+v", res)
			// TODO: Determine if ok
		}

		if m.MediaTitle != "" {
			res, err := c.Search(m.SecureMediaTitle)
			if err != nil {
				log.Println(err.Error())
			}

			log.Printf("%+v", res)
			// TODO: Determine if ok
		} */
}

func (c *Client) Search(q string) (*spotify.SearchResult, error) {
	res, err := c.C.Search(q, spotify.SearchTypeAlbum|spotify.SearchTypeArtist|spotify.SearchTypeTrack)
	if err != nil {
		return nil, fmt.Errorf("error searching: %w", err)
	}

	return res, nil
}

func (c *Client) log(s string) {
	if c.Verbose {
		log.Printf("spotify:\t%s\n", s)
	}
}
