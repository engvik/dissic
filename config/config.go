package config

import (
	"errors"
	"flag"
)

type Config struct {
	Reddit Reddit
}

type Reddit struct {
	ClientID     string
	ClientSecret string
	Username     string
	Password     string
}

func Parse() (*Config, error) {
	var redditClientID string
	var redditClientSecret string
	var redditUsername string
	var redditPassword string

	flag.StringVar(&redditClientID, "reddit-client-id", "", "Reddit client ID")
	flag.StringVar(&redditClientSecret, "reddit-client-secret", "", "Reddit client secret")
	flag.StringVar(&redditUsername, "reddit-username", "", "Reddit username")
	flag.StringVar(&redditPassword, "reddit-password", "", "Reddit password")

	flag.Parse()

	cfg := Config{
		Reddit: Reddit{
			ClientID:     redditClientID,
			ClientSecret: redditClientSecret,
			Username:     redditUsername,
			Password:     redditPassword,
		},
	}

	err := cfg.validate()
	if err != nil {
		return nil, err
	}

	return &cfg, nil
}

func (c *Config) validate() error {
	if c.Reddit.ClientID == "" {
		return errors.New("Reddit client ID is missing (--reddit-client-id)")
	}

	if c.Reddit.ClientSecret == "" {
		return errors.New("Reddit client secret is missing (--reddit-client-secret)")
	}

	if c.Reddit.Username == "" {
		return errors.New("Reddit username is missing (--reddit-username)")
	}

	if c.Reddit.Password == "" {
		return errors.New("Reddit password is missing (--reddit-password)")
	}

	return nil
}
