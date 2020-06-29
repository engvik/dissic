// Package config contains the dissic config and methods to add
// default values and validation.
package config

import (
	"errors"
	"flag"
	"fmt"
	"io/ioutil"
	"runtime"

	"github.com/kelseyhightower/envconfig"
	log "github.com/sirupsen/logrus"
	"gopkg.in/yaml.v2"
)

const version = "0.0.1"

type environment struct {
	RedditUsername      string `envconfig:"REDDIT_USERNAME"`
	SpotifyClientID     string `envconfig:"SPOTIFY_CLIENT_ID"`
	SpotifyClientSecret string `envconfig:"SPOTIFY_CLIENT_SECRET"`
	ConfigFile          string `envconfig:"DISSIC_CONFIG"`
}

// Config holds the entire dissic config (config.yaml and env vars)
type Config struct {
	Reddit              Reddit     `yaml:"reddit"`
	Spotify             Spotify    `yaml:"spotify"`
	Playlists           []Playlist `yaml:"playlists"`
	HTTPPort            int        `yaml:"http-port"`
	Verbose             bool       `yaml:"verbose"`
	AuthOpenBrowser     bool       `yaml:"auth-open-browser"`
	Version             string
	PlaylistDescription string
}

// Reddit holds the reddit related configuration.
type Reddit struct {
	Username             string `yaml:"username"`
	RequestRate          int    `yaml:"request-rate"`
	MaxRetryAttempts     int    `yaml:"max-retry-attempts"`
	RetryAttemptWaitTime int    `yaml:"retry-attempt-wait-time"`
	Subreddits           []string
}

// Spotify holds the reddit related configuration.
type Spotify struct {
	ClientID     string `yaml:"client-id"`
	ClientSecret string `yaml:"client-secret"`
}

// Playlist contains the playlist configuration
type Playlist struct {
	Name       string   `yaml:"name"`
	ID         string   `yaml:"id"`
	Subreddits []string `yaml:"subreddits"`
}

// Load reads config from file and environment variables. It also adds
// default values where applicable and validates the config before returning.
func Load() (*Config, error) {
	var configFile string
	flag.StringVar(&configFile, "config", "", "path to config file")
	flag.Parse()

	var env environment
	if err := envconfig.Process("dissic", &env); err != nil {
		return nil, fmt.Errorf("parsing environment variables: %w", err)
	}

	if configFile == "" {
		configFile = env.ConfigFile
	}

	cf, err := readConfigFile(configFile)
	if err != nil {
		return nil, fmt.Errorf("reading config file: %w", err)
	}

	var cfg Config
	if err := yaml.Unmarshal(cf, &cfg); err != nil {
		return nil, fmt.Errorf("unmarshal yaml file: %w", err)
	}

	cfg.addEnvironment(env)
	cfg.setDefaultValues()

	if err := cfg.validate(); err != nil {
		return nil, err
	}

	if cfg.Verbose {
		log.SetLevel(log.InfoLevel)
	} else {
		log.SetLevel(log.ErrorLevel)
	}

	return &cfg, nil
}

// GetRedditUserAgent creates a reddit user agent for graw
func (c *Config) GetRedditUserAgent() string {
	return fmt.Sprintf("%s:github.com/engvik/dissic:%s (by /u/%s)", runtime.GOOS, c.Version, c.Reddit.Username)
}

func (c *Config) getSubreddits() []string {
	var subs []string

	for _, p := range c.Playlists {
		for _, sub := range p.Subreddits {
			subs = append(subs, sub)
		}
	}

	return subs
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
		return errors.New("reddit username is missing")
	}

	if c.Reddit.RequestRate < 2 {
		return errors.New("reddit request rate must be 2 or higher")
	}

	if len(c.Reddit.Subreddits) <= 0 {
		return errors.New("no subreddits passed")
	}

	if c.Spotify.ClientID == "" {
		return errors.New("spotify client id is missing")
	}

	if c.Spotify.ClientSecret == "" {
		return errors.New("spotify client secret is missing")
	}

	for i, p := range c.Playlists {
		if p.ID == "" && p.Name == "" {
			return fmt.Errorf("playlist number %d is missing ID or name", i)
		}
	}

	return nil
}

func (c *Config) setDefaultValues() {
	c.Version = version
	c.PlaylistDescription = "Auto-generated playlist. Generate your own with dissic: https://github.com/engvik/dissic"
	c.Reddit.Subreddits = c.getSubreddits()

	if c.Reddit.RequestRate == 0 {
		c.Reddit.RequestRate = 5
	}

	if c.Reddit.MaxRetryAttempts == 0 {
		c.Reddit.MaxRetryAttempts = 10
	}

	if c.Reddit.RetryAttemptWaitTime == 0 {
		c.Reddit.RetryAttemptWaitTime = 10
	}
}

func readConfigFile(path string) ([]byte, error) {
	data, err := ioutil.ReadFile(path)
	if err != nil {
		return nil, fmt.Errorf("reading config file: %s, %w", path, err)
	}

	return data, nil
}
