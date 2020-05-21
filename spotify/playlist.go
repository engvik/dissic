package spotify

import (
	"fmt"
	"strings"
	"time"

	"github.com/engvik/dissic/config"
	"github.com/zmb3/spotify"
)

type SubredditPlaylist map[string]spotify.ID

const ErrInvalidID = "Invalid playlist Id"

func (c *Client) PreparePlaylists(cfg *config.Config) error {
	subredditPlaylist := make(SubredditPlaylist, len(cfg.Reddit.Subreddits))

	for _, p := range cfg.Playlists {
		// get playlist
		playlist, err := c.getPlaylist(p)
		if err != nil {
			return fmt.Errorf("unable to get playlist: %w", err)
		}

		// not found, but name provided, create playlist
		if playlist == nil && p.Name != "" {
			playlist, err = c.createPlaylist(p.Name)
			if err != nil {
				return fmt.Errorf("error creating playlist: %w", err)
			}
		}

		// create subreddit playlist map
		for _, s := range p.Subreddits {
			subreddit := strings.ToLower(s)
			subredditPlaylist[subreddit] = playlist.ID
		}

		// Be nice to the Spotify API
		time.Sleep(1 * time.Second)
	}

	c.SubredditPlaylist = subredditPlaylist

	return nil
}

func (c *Client) getPlaylist(p config.Playlist) (*spotify.FullPlaylist, error) {
	// prefer getting by id
	if p.ID != "" {
		return c.getPlaylistByID(p.ID)
	}

	// fallback to getting by name
	return c.getPlaylistByName(p.Name)
}

func (c *Client) getPlaylistByID(ID string) (*spotify.FullPlaylist, error) {
	playlist, err := c.C.GetPlaylist(spotify.ID(ID))
	if err != nil {
		return nil, fmt.Errorf("getting playlist %s: %w", ID, err)
	}

	c.Log(fmt.Sprintf("found playlist: %s (%s)", playlist.Name, playlist.ID))

	return playlist, nil
}

func (c *Client) getPlaylistByName(name string) (*spotify.FullPlaylist, error) {
	res, err := c.C.GetPlaylistsForUser(c.User.ID)
	if err != nil {
		return nil, fmt.Errorf("getting playlists for %s: %w", c.User.ID, err)
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

	return playlist, nil
}

func (c *Client) createPlaylist(name string) (*spotify.FullPlaylist, error) {
	playlist, err := c.C.CreatePlaylistForUser(c.User.ID, name, "dissic", false)
	if err != nil {
		return nil, fmt.Errorf("creating playlist %s: %w", name, err)
	}

	c.Log(fmt.Sprintf("created playlist: %s", name))

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
