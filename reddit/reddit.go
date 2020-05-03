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
	Config           graw.Config
	Script           reddit.Script
	MusicChan        chan spotify.Music
	Retry            time.Duration
	MaxRetryAttempts int
	Stop             func()
	Wait             func() error
	Verbose          bool
}

func New(cfg *config.Config, m chan spotify.Music) (*Client, error) {
	s, err := reddit.NewScript(cfg.GetRedditUserAgent(), 5*time.Second)
	if err != nil {
		return nil, fmt.Errorf("new script: %w", err)
	}

	gCfg := graw.Config{Subreddits: cleanSubNames(cfg.Reddit.Subreddits)}

	c := Client{
		Config:           gCfg,
		Script:           s,
		MusicChan:        m,
		Retry:            time.Duration(10), // TODO: make configurable
		MaxRetryAttempts: 10,                // TODO: make configurable
		Verbose:          cfg.Verbose,
	}

	c.Log("client setup ok")

	return &c, nil
}

func (c *Client) PrepareScanner() error {
	stop, wait, err := graw.Scan(c, c.Script, c.Config)
	if err != nil {
		return fmt.Errorf("graw preparation failed: %w", err)
	}

	c.Stop = stop
	c.Wait = wait

	return nil
}

func (c *Client) Listen() {
	c.Log(fmt.Sprintf("watching %d subreddits:", len(c.Config.Subreddits)))

	for _, sub := range c.Config.Subreddits {
		c.Log("\tr/" + sub)
	}

	var retryAttempt int

	for {
		if retryAttempt == c.MaxRetryAttempts {
			return
		}

		if err := c.Wait(); err != nil {
			c.Log(fmt.Sprintf("reddit/graw error: %s", err.Error()))
		}

		c.Log(fmt.Sprintf("restarting reddit worker in %s seconds", c.Retry))
		time.Sleep(c.Retry * time.Second)

		if err := c.PrepareScanner(); err != nil {
			c.Log(fmt.Sprintf("error restarting reddit worker: %s", err.Error()))
			return
		}

		retryAttempt++
	}
}

func (c *Client) Close() {
	c.Log("shutting down")
	c.Stop()
}

func (c *Client) Post(post *reddit.Post) error {
	c.Log(fmt.Sprintf("r/%s: %s (https://reddit.com%s)", post.Subreddit, post.Title, post.Permalink))
	c.MusicChan <- spotify.Music{
		Sub:              post.Subreddit,
		PostTitle:        post.Title,
		MediaTitle:       post.Media.OEmbed.Title,
		SecureMediaTitle: post.SecureMedia.OEmbed.Title,
	}

	return nil
}

func (c *Client) Log(s string) {
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
