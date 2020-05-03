package spotify

import (
	"errors"
	"fmt"

	"github.com/engvik/reddify/config"
	"github.com/zmb3/spotify"
)

const ErrInvalidID = "Invalid playlist Id"

func (c *Client) PreparePlaylists(cfg *config.Config) error {
	spm := make(map[string]spotify.ID, len(cfg.Reddit.Subreddits))
	user, err := c.C.CurrentUser()
	if err != nil {
		return fmt.Errorf("error getting current user: %w", err)
	}

	for _, p := range cfg.Playlists {
		var playlist *spotify.FullPlaylist
		var err error

		if p.ID != "" {
			playlist, err = c.preparePlaylistByID(p.ID)
		} else if p.Name != "" {
			playlist, err = c.preparePlaylistByName(user, p.Name)
		}

		if err != nil {
			return fmt.Errorf("preparing playlist: %w", err)
		}

		if playlist == nil {
			return errors.New("unable to get playlist")
		}

		for _, s := range p.Subreddits {
			spm[s] = playlist.ID
		}
	}

	c.SubredditPlaylist = spm

	return nil
}

func (c *Client) preparePlaylistByID(ID string) (*spotify.FullPlaylist, error) {
	playlist, err := c.C.GetPlaylist(spotify.ID(ID))
	if err != nil {
		return nil, fmt.Errorf("error getting playlist %s: %w", ID, err)
	}

	c.Log(fmt.Sprintf("found playlist: %s (%s)", playlist.Name, playlist.ID))

	return playlist, nil
}

func (c *Client) preparePlaylistByName(user *spotify.PrivateUser, name string) (*spotify.FullPlaylist, error) {
	c.Log(fmt.Sprintf("retrived user: %s", user.ID))

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

			c.Log(fmt.Sprintf("found playlist: %s (%s)", p.Name, p.ID))

			break
		}
	}

	if playlist == nil {
		playlist, err = c.C.CreatePlaylistForUser(user.ID, name, "reddify", false)
		if err != nil {
			return nil, fmt.Errorf("error creating playlist %s: %w", name, err)
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

	c.Log(fmt.Sprintf("\tadded track to playlist, snapshot id: %s", snapshotID))
	return nil
}
