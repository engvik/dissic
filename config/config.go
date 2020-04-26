package config

import (
	"errors"
	"flag"
	"fmt"
	"strings"
)

type Config struct {
	Reddit  Reddit
	Spotify Spotify
}

type Reddit struct {
	Username string
	Subs     []string
}

type Spotify struct {
	ClientID     string
	ClientSecret string
	Playlist     string
}

func Parse() (*Config, error) {
	var redditUsername string
	var redditSubs string
	var spotifyClientID string
	var spotifyClientSecret string
	var spotifyPlaylist string

	flag.StringVar(&redditUsername, "reddit-username", "", "Reddit username")
	flag.StringVar(&redditSubs, "subreddits", "", "list of subreddits to listen to")
	flag.StringVar(&spotifyClientID, "spotify-client-id", "", "Spotify client ID")
	flag.StringVar(&spotifyClientSecret, "spotify-client-secret", "", "Spotify client secret")
	flag.StringVar(&spotifyPlaylist, "spotify-playlist", "", "Spotify playlist to add music to")

	flag.Parse()

	cfg := Config{
		Reddit: Reddit{
			Username: redditUsername,
			Subs:     strings.Split(redditSubs, ","),
		},
		Spotify: Spotify{
			ClientID:     spotifyClientID,
			ClientSecret: spotifyClientSecret,
			Playlist:     spotifyPlaylist,
		},
	}

	err := cfg.validate()
	if err != nil {
		return nil, err
	}

	return &cfg, nil
}

func (c *Config) GetRedditUserAgent() string {
	return fmt.Sprintf("graw:reddify:0.0.1 by /u/%s", c.Reddit.Username)
}

func (c *Config) validate() error {
	if c.Reddit.Username == "" {
		return errors.New("Reddit username is missing (--reddit-username)")
	}

	if len(c.Reddit.Subs) <= 0 {
		return errors.New("no subreddits passed (--subreddits=a,b,c)")
	}

	if c.Spotify.ClientID == "" {
		return errors.New("Spotify client ID is missing (--spotify-client-id)")
	}

	if c.Spotify.ClientSecret == "" {
		return errors.New("Spotify client secret is missing (--spotify-client-secret)")
	}

	if c.Spotify.Playlist == "" {
		return errors.New("no Spotify playlist specified (--spotify-playlist")
	}

	return nil
}
