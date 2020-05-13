package spotify

import (
	"errors"
	"fmt"
	"strings"
	"time"

	"github.com/engvik/dissic/config"
	"github.com/zmb3/spotify"
)

const ErrInvalidID = "Invalid playlist Id"

func (c *Client) GetPlaylists(cfg *config.Config) error {
	spm := make(map[string]spotify.ID, len(cfg.Reddit.Subreddits))
	user, err := c.C.CurrentUser()
	if err != nil {
		return fmt.Errorf("getting current user: %w", err)
	}

	c.Log(fmt.Sprintf("retrived user: %s", user.ID))

	for _, p := range cfg.Playlists {
		var playlist *spotify.FullPlaylist
		var err error

		if p.ID != "" {
			playlist, err = c.getPlaylistByID(p.ID)
		} else if p.Name != "" {
			playlist, err = c.getPlaylistByName(user, p.Name)
		}

		if err != nil {
			return fmt.Errorf("preparing playlist: %w", err)
		}

		if playlist == nil {
			return errors.New("unable to get playlist")
		}

		for _, s := range p.Subreddits {
			subreddit := strings.ToLower(s)
			spm[subreddit] = playlist.ID
		}

		// Be nice to the Spotify API
		time.Sleep(1 * time.Second)
	}

	c.SubredditPlaylist = spm

	return nil
}

func (c *Client) getPlaylistByID(ID string) (*spotify.FullPlaylist, error) {
	playlist, err := c.C.GetPlaylist(spotify.ID(ID))
	if err != nil {
		return nil, fmt.Errorf("getting playlist %s: %w", ID, err)
	}

	c.Log(fmt.Sprintf("found playlist: %s (%s)", playlist.Name, playlist.ID))

	return playlist, nil
}

func (c *Client) getPlaylistByName(user *spotify.PrivateUser, name string) (*spotify.FullPlaylist, error) {
	res, err := c.C.GetPlaylistsForUser(user.ID)
	if err != nil {
		return nil, fmt.Errorf("getting playlists for %s: %w", user.ID, err)
	}

	var playlist *spotify.FullPlaylist

	for _, p := range res.Playlists {
		if p.Name == name {
			playlist, err = c.C.GetPlaylist(p.ID)
			if err != nil {
				return nil, fmt.Errorf("getting playlist %s (%s): %w", p.Name, p.ID, err)
			}

			c.Log(fmt.Sprintf("found playlist: %s (%s)", p.Name, p.ID))

			break
		}
	}

	if playlist == nil {
		playlist, err = c.C.CreatePlaylistForUser(user.ID, name, "dissic", false)
		if err != nil {
			return nil, fmt.Errorf("creating playlist %s: %w", name, err)
		}

		c.Log(fmt.Sprintf("created playlist: %s", name))
	}

	return playlist, nil
}

func (c *Client) addToPlaylist(playlistID spotify.ID, trackID spotify.ID) error {
	snapshotID, err := c.C.AddTracksToPlaylist(playlistID, trackID)
	if err != nil {
		return fmt.Errorf("adding track: playlist %s, track %s: %w", playlistID, trackID, err)
	}

	c.Log(fmt.Sprintf("\tadded track to playlist %s, snapshot id: %s", playlistID, snapshotID))
	return nil
}
