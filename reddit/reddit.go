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
	Verbose   bool
}

func New(cfg *config.Config, m spotify.MusicChan) (*Client, error) {
	s, err := reddit.NewScript(cfg.GetRedditUserAgent(), 5*time.Second)
	if err != nil {
		return nil, fmt.Errorf("new script: %w", err)
	}

	gCfg := graw.Config{Subreddits: cleanSubNames(cfg.Reddit.Subs)}

	c := Client{
		Config:    gCfg,
		Script:    s,
		MusicChan: m,
		Verbose:   cfg.Verbose,
	}

	c.log("client setup ok")

	return &c, nil
}

func (c *Client) Listen() error {
	stop, wait, err := graw.Scan(c, c.Script, c.Config)
	if err != nil {
		return fmt.Errorf("graw scan failed: %w", err)
	}

	defer stop()

	c.log("watching subreddits:")

	for _, sub := range c.Config.Subreddits {
		c.log("\tr/" + sub)
	}

	if err := wait(); err != nil {
		return fmt.Errorf("graw run encountered an error: %w", err)
	}

	return nil
}

func (c *Client) Post(post *reddit.Post) error {
	c.log(fmt.Sprintf("r/%s: %s (%s)", post.Subreddit, post.Title, post.URL))
	c.MusicChan <- spotify.Music{
		Sub:              post.Subreddit,
		PostTitle:        post.Title,
		MediaTitle:       post.Media.OEmbed.Title,
		SecureMediaTitle: post.SecureMedia.OEmbed.Title,
	}

	return nil
}

func (c *Client) log(s string) {
	if c.Verbose {
		log.Printf("reddit:\t%s\n", s)
	}
}

func cleanSubNames(subs []string) []string {
	for i, sub := range subs {
		if sub[:2] == "r/" {
			subs[i] = sub[2:]
		}
	}

	return subs
}
