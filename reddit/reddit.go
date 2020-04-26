package reddit

import (
	"fmt"
	"log"
	"time"

	"github.com/engvik/reddify/config"
	"github.com/engvik/reddify/spotify"
	"github.com/turnage/graw"
	"github.com/turnage/graw/reddit"
)

type Client struct {
	Config    graw.Config
	Script    reddit.Script
	MusicChan spotify.MusicChan
}

func New(cfg *config.Config, m spotify.MusicChan) (*Client, error) {
	s, err := reddit.NewScript(cfg.GetRedditUserAgent(), 5*time.Second)
	if err != nil {
		return nil, fmt.Errorf("new script: %w", err)
	}

	gCfg := graw.Config{Subreddits: cfg.Reddit.Subs}

	return &Client{
		Config:    gCfg,
		Script:    s,
		MusicChan: m,
	}, nil
}

func (c *Client) Listen() error {
	stop, wait, err := graw.Scan(&Client{}, c.Script, c.Config)
	if err != nil {
		return fmt.Errorf("graw scan failed: %w", err)
	}

	defer stop()

	log.Println("Streaming from:\n")

	for _, sub := range c.Config.Subreddits {
		log.Println("r/" + sub)
	}

	if err := wait(); err != nil {
		return fmt.Errorf("graw run encountered an error: %w", err)
	}

	return nil
}

func (c *Client) Post(post *reddit.Post) error {
	c.MusicChan <- spotify.Music{
		Sub:              post.Subreddit,
		PostTitle:        post.Title,
		MediaTitle:       post.Media.OEmbed.Title,
		SecureMediaTitle: post.SecureMedia.OEmbed.Title,
	}

	return nil
}
