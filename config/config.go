package config

import (
	"errors"
	"fmt"
	"io/ioutil"
	"os"

	"github.com/kelseyhightower/envconfig"
	"gopkg.in/yaml.v2"
)

const version = "0.0.1"

type environment struct {
	RedditUsername      string `envconfig:"REDDIT_USERNAME"`
	SpotifyClientID     string `envconfig:"SPOTIFY_CLIENT_ID"`
	SpotifyClientSecret string `envconfig:"SPOTIFY_CLIENT_SECRET"`
}

type Config struct {
	Reddit   Reddit  `yaml:"reddit"`
	Spotify  Spotify `yaml:"spotify"`
	HTTPPort int     `yaml:"http-port"`
	Verbose  bool    `yaml:"verbose"`
	Version  string
}

type Reddit struct {
	Username string   `yaml:"username"`
	Subs     []string `yaml:"subreddits"`
}

type Spotify struct {
	ClientID     string `yaml:"client-id"`
	ClientSecret string `yaml:"client-secret"`
	PlaylistName string `yaml:"playlist-name"`
	PlaylistID   string `yaml:"playlist-id"`
}

func Load() (*Config, error) {
	cf, err := readConfigFile()
	if err != nil {
		return nil, fmt.Errorf("reading config file: %w", err)
	}

	var cfg Config
	if err := yaml.Unmarshal(cf, &cfg); err != nil {
		return nil, fmt.Errorf("unmarshal yaml file: %w", err)
	}

	var env environment
	if err := envconfig.Process("reddify", &env); err != nil {
		return nil, fmt.Errorf("parsing environment variables: %w", err)
	}

	cfg.addEnvironment(env)

	cfg.Version = version

	if err := cfg.validate(); err != nil {
		return nil, err
	}

	return &cfg, nil
}

func (c *Config) GetRedditUserAgent() string {
	return fmt.Sprintf("graw:reddify:%s by /u/%s", c.Version, c.Reddit.Username)
}

func (c *Config) addEnvironment(e environment) {
	if c.Reddit.Username == "" {
		c.Reddit.Username = e.RedditUsername
	}

	if c.Spotify.ClientID == "" {
		c.Spotify.ClientID = e.SpotifyClientID
	}

	if c.Spotify.ClientSecret == "" {
		c.Spotify.ClientSecret = e.SpotifyClientSecret
	}
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

func readConfigFile() ([]byte, error) {
	path, err := os.Getwd()
	if err != nil {
		return nil, fmt.Errorf("error getting path: %w", err)
	}

	data, err := ioutil.ReadFile(fmt.Sprintf("%s/config.yaml", path))
	if err != nil {
		return nil, fmt.Errorf("error reading file: %w", err)
	}

	return data, nil
}
