package spotify

import (
	"fmt"
	"strings"
	"time"

	"github.com/engvik/dissic/internal/config"
	"github.com/zmb3/spotify"
)

// ErrInvalidID is the error for an invalid playlist id
const ErrInvalidID = "Invalid playlist Id"

// PreparePlaylists checks the playlists defined in the config and fetches
// them from Spotify. If a playlist is passed by name, it's created if it
// doesn't exist. It also connects the subreddits to a corresponding playlist id.
func (c *Client) PreparePlaylists(cfg *config.Config) error {
	subredditPlaylist := make(map[string]spotify.ID, len(cfg.Reddit.Subreddits))

	for _, p := range cfg.Playlists {
		// get playlist
		playlist, err := c.getPlaylist(p)
		if err != nil {
			return fmt.Errorf("unable to get playlist: %w", err)
		}

		// not found, but name provided, create playlist
		if playlist == nil && p.Name != "" {
			playlist, err = c.Spotify.CreatePlaylistForUser(c.User.ID, p.Name, cfg.PlaylistDescription, false)
			if err != nil {
				return fmt.Errorf("error creating playlist: %w", err)
			}

			c.Logger.Infof("created playlist: %s", p.Name)
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
	playlist, err := c.Spotify.GetPlaylist(spotify.ID(ID))
	if err != nil {
		return nil, fmt.Errorf("getting playlist %s: %w", ID, err)
	}

	c.Logger.Infof("found playlist: %s (%s)", playlist.Name, playlist.ID)

	return playlist, nil
}

func (c *Client) getPlaylistByName(name string) (*spotify.FullPlaylist, error) {
	res, err := c.Spotify.GetPlaylistsForUser(c.User.ID)
	if err != nil {
		return nil, fmt.Errorf("getting playlists for %s: %w", c.User.ID, err)
	}

	var playlist *spotify.FullPlaylist

	for _, p := range res.Playlists {
		if p.Name == name {
			playlist, err = c.Spotify.GetPlaylist(p.ID)
			if err != nil {
				return nil, fmt.Errorf("getting playlist %s (%s): %w", p.Name, p.ID, err)
			}

			c.Logger.Infof("found playlist: %s (%s)", p.Name, p.ID)

			break
		}
	}

	return playlist, nil
}

func (c *Client) addToPlaylist(subreddit string, trackID spotify.ID) error {
	playlistID, ok := c.SubredditPlaylist[subreddit]
	if !ok {
		return fmt.Errorf("no playlist found for subreddit: %s", subreddit)
	}

	playlist, err := c.Spotify.GetPlaylistTracks(playlistID)
	if err != nil {
		return fmt.Errorf("error getting playlist tracks (%s): %w", playlistID, err)
	}

	for _, t := range playlist.Tracks {
		if t.Track.ID == trackID {
			return fmt.Errorf("track already in playlist: %s - %s (%s)", t.Track.Artists, t.Track.Name, trackID)
		}
	}

	snapshotID, err := c.Spotify.AddTracksToPlaylist(playlistID, trackID)
	if err != nil {
		return fmt.Errorf("adding track: playlist %s, track %s: %w", playlistID, trackID, err)
	}

	c.Logger.Infof("\tadded track to playlist %s, snapshot id: %s", playlistID, snapshotID)

	return nil
}
