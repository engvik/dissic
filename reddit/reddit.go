package reddit

import (
	"fmt"
	"time"

	"github.com/engvik/reddify/config"
	"github.com/turnage/graw"
	"github.com/turnage/graw/reddit"
)

type Client struct {
	Config graw.Config
	Script reddit.Script
}

func New(cfg *config.Config) (*Client, error) {
	s, err := reddit.NewScript(cfg.GetRedditUserAgent(), 5*time.Second)
	if err != nil {
		return nil, fmt.Errorf("new script: %w", err)
	}

	gCfg := graw.Config{Subreddits: cfg.Reddit.Subs}

	return &Client{
		Config: gCfg,
		Script: s,
	}, nil
}

func (c *Client) Listen() error {
	stop, wait, err := graw.Scan(&Client{}, c.Script, c.Config)
	if err != nil {
		return fmt.Errorf("graw scan failed: %w", err)
	}
	defer stop()

	if err := wait(); err != nil {
		return fmt.Errorf("graw run encountered an error: %w", err)
	}

	return nil
}

func (c *Client) Post(post *reddit.Post) error {
	fmt.Printf("%s posted %s\n", post.Author, post.Title)
	return nil
}
