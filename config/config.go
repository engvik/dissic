package config

import (
	"errors"
	"flag"
	"fmt"
	"strings"
)

type Config struct {
	Reddit Reddit
}

type Reddit struct {
	Username string
	Subs     []string
}

func Parse() (*Config, error) {
	var redditUsername string
	var redditSubs string

	flag.StringVar(&redditUsername, "reddit-username", "", "Reddit username")
	flag.StringVar(&redditSubs, "subreddits", "", "list of subreddits to listen to")

	flag.Parse()

	cfg := Config{
		Reddit: Reddit{
			Username: redditUsername,
			Subs:     strings.Split(redditSubs, ","),
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

	return nil
}
