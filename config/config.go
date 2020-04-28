package config

import (
	"errors"
	"flag"
	"fmt"
	"strings"
)

const version = "0.0.1"

type Config struct {
	Reddit  Reddit
	Spotify Spotify
	Version string
	Verbose bool
}

type Reddit struct {
	Username string
	Subs     []string
}

type Spotify struct {
	ClientID     string
	ClientSecret string
	PlaylistName string
	PlaylistID   string
}

func Parse() (*Config, error) {
	var redditUsername string
	var redditSubs string
	var spotifyClientID string
	var spotifyClientSecret string
	var spotifyPlaylistName string
	var spotifyPlaylistID string
	var verbose bool

	flag.StringVar(&redditUsername, "reddit-username", "", "Reddit username")
	flag.StringVar(&redditSubs, "subreddits", "", "list of subreddits to listen to")
	flag.StringVar(&spotifyClientID, "spotify-client-id", "", "Spotify client ID")
	flag.StringVar(&spotifyClientSecret, "spotify-client-secret", "", "Spotify client secret")
	flag.StringVar(&spotifyPlaylistName, "spotify-playlist-name", "", "Spotify playlist name to add music to")
	flag.StringVar(&spotifyPlaylistID, "spotify-playlist-id", "", "Spotify playlist id to add music to")
	flag.BoolVar(&verbose, "verbose", false, "Verbose log output")

	flag.Parse()

	cfg := Config{
		Reddit: Reddit{
			Username: redditUsername,
			Subs:     strings.Split(redditSubs, ","),
		},
		Spotify: Spotify{
			ClientID:     spotifyClientID,
			ClientSecret: spotifyClientSecret,
			PlaylistName: spotifyPlaylistName,
			PlaylistID:   spotifyPlaylistID,
		},
		Version: version,
		Verbose: verbose,
	}

	err := cfg.validate()
	if err != nil {
		return nil, err
	}

	return &cfg, nil
}

func (c *Config) GetRedditUserAgent() string {
	return fmt.Sprintf("graw:reddify:%s by /u/%s", c.Version, c.Reddit.Username)
}

func (c *Config) validate() error {
	if c.Reddit.Username == "" {
		return errors.New("reddit username is missing (--reddit-username)")
	}

	if len(c.Reddit.Subs) <= 0 {
		return errors.New("no subreddits passed (--subreddits=a,b,c)")
	}

	if c.Spotify.ClientID == "" {
		return errors.New("spotify client id is missing (--spotify-client-id)")
	}

	if c.Spotify.ClientSecret == "" {
		return errors.New("spotify client secret is missing (--spotify-client-secret)")
	}

	if c.Spotify.PlaylistName == "" && c.Spotify.PlaylistID == "" {
		return errors.New("no spotify playlist specified (--spotify-playlist-name or --spotify-playlist-id)")
	}

	return nil
}
