// package reddit contains an implementaion of the reddit service.
package reddit

import (
	"fmt"
	"os"
	"strings"
	"time"

	"github.com/engvik/dissic/internal/config"
	"github.com/engvik/dissic/internal/spotify"
	log "github.com/sirupsen/logrus"
	"github.com/turnage/graw"
	"github.com/turnage/graw/reddit"
)

// Client is the reddit client
type Client struct {
	Config               graw.Config
	Script               reddit.Script
	MusicChan            chan<- spotify.Music
	RetryAttemptWaitTime time.Duration
	MaxRetryAttempts     int
	ShouldRetry          bool
	Stop                 func()
	Wait                 func() error
	Logger               *log.Entry
}

// New sets up a new reddit client. It takes the configuration and the channel
// to publish new posts to for processing. Returns a client or an error.
func New(cfg *config.Config, m chan<- spotify.Music) (*Client, error) {
	s, err := reddit.NewScript(cfg.GetRedditUserAgent(), time.Duration(cfg.Reddit.RequestRate))
	if err != nil {
		return nil, fmt.Errorf("new script: %w", err)
	}

	gCfg := graw.Config{Subreddits: cleanSubNames(cfg.Reddit.Subreddits)}

	c := Client{
		Config:               gCfg,
		Script:               s,
		MusicChan:            m,
		RetryAttemptWaitTime: time.Duration(cfg.Reddit.MaxRetryAttempts),
		MaxRetryAttempts:     cfg.Reddit.MaxRetryAttempts,
		ShouldRetry:          true,
		Logger:               log.WithFields(log.Fields{"service": "reddit"}),
	}

	c.Logger.Infoln("client setup ok")

	return &c, nil
}

// PrepareScanner calls graw to set up the reddit post scanner.
// It also makes the stop and wait function returned by graw to
// the client struct.
func (c *Client) PrepareScanner() error {
	stop, wait, err := graw.Scan(c, c.Script, c.Config)
	if err != nil {
		return fmt.Errorf("graw preparation failed: %w", err)
	}

	c.Stop = stop
	c.Wait = wait

	return nil
}

// Listen starts listening for reddit posts. It also contains logic for
// reconnecting if an error occurs.
func (c *Client) Listen(shutdown chan<- os.Signal) {
	c.Logger.Infof("watching %d subreddits:", len(c.Config.Subreddits))

	for _, sub := range c.Config.Subreddits {
		c.Logger.Infoln("\tr/" + sub)
	}

	var retryAttempt int

	for {
		if retryAttempt == c.MaxRetryAttempts {
			c.Logger.Errorf("hit maximum retry attempts %d - quitting", retryAttempt)
			shutdown <- os.Interrupt
		}

		if err := c.Wait(); err != nil {
			retryAttempt = 0
			c.Logger.Errorf("reddit/graw error: %s", err)
		}

		if c.ShouldRetry {
			c.Logger.Infof("restarting reddit helper in %s seconds", c.RetryAttemptWaitTime)
			time.Sleep(c.RetryAttemptWaitTime * time.Second)

			if err := c.PrepareScanner(); err != nil {
				c.Logger.Errorf("error restarting reddit helper: %s", err)
			}

			retryAttempt++
		}
	}
}

// Close shuts down the reddit client.
func (c *Client) Close() {
	c.ShouldRetry = false
	c.Logger.Println("shutting down")
	c.Stop()
}

// Post receives incoming posts from reddit and passes them
// on to the spotify processor.
func (c *Client) Post(post *reddit.Post) error {
	c.Logger.Infof("r/%s: %s (https://reddit.com%s)", post.Subreddit, post.Title, post.Permalink)
	c.MusicChan <- spotify.Music{
		Subreddit:        strings.ToLower(post.Subreddit),
		PostTitle:        post.Title,
		MediaTitle:       post.Media.OEmbed.Title,
		SecureMediaTitle: post.SecureMedia.OEmbed.Title,
		URL:              post.URL,
	}

	return nil
}

func cleanSubNames(subs []string) []string {
	for i, sub := range subs {
		if sub[:2] == "r/" {
			subs[i] = sub[2:]
		}
	}

	return subs
}
