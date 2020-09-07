package config

import (
	"fmt"
	"testing"

	"gopkg.in/yaml.v2"
)

var testConfig = []byte(`
# Verbose log output
verbose: true

# HTTP port
http-port: 8080

# auto open browser for auth
auth-open-browser: true

reddit:
    username: "test"
    request-rate: 5
    max-retry-attempts: 10
    retry-attempt-wait-time: 10
  
spotify:
    client-id: "test1337"
    client-secret: "1337test"

playlists:
    -
        name: "dissic-test"
        subreddits:
            - music
`)

func parseConfig(t *testing.T) (*Config, error) {
	t.Helper()

	var cfg Config
	if err := yaml.Unmarshal(testConfig, &cfg); err != nil {
		return nil, fmt.Errorf("unmarshal yaml file: %w", err)
	}

	return &cfg, nil
}

func TestGetSubreddits(t *testing.T) {
	cfg, err := parseConfig(t)
	if err != nil {
		t.Fatalf("error setting up test: %v", err)
	}

	t.Run("should return one subreddit", func(t *testing.T) {
		subs := cfg.getSubreddits()
		exp := []string{"music"}

		if len(subs) != len(exp) {
			t.Errorf("unexpected slice length: got %d, exp %d", len(subs), len(exp))
		}

		for i, sub := range subs {
			if sub != exp[i] {
				t.Errorf("unexpected value: got %s, exp %s, pos %d", sub, exp[i], i)
			}
		}
	})
}

func TestAddEnvironment(t *testing.T) {
	cfg, err := parseConfig(t)
	if err != nil {
		t.Fatalf("error setting up test: %v", err)
	}

	cfg.Reddit.Username = ""
	cfg.Spotify.ClientID = ""
	cfg.Spotify.ClientSecret = ""

	env := environment{
		RedditUsername:      "newtestuser",
		SpotifyClientID:     "testclientid",
		SpotifyClientSecret: "testclientid",
	}

	t.Run("should set config from environment", func(t *testing.T) {
		cfg.addEnvironment(env)

		if cfg.Reddit.Username != env.RedditUsername {
			t.Errorf("unexpected value: got %s, exp %s", cfg.Reddit.Username, env.RedditUsername)
		}

		if cfg.Spotify.ClientID != env.SpotifyClientID {
			t.Errorf("unexpected value: got %s, exp %s", cfg.Spotify.ClientID, env.SpotifyClientID)
		}

		if cfg.Spotify.ClientSecret != env.SpotifyClientSecret {
			t.Errorf("unexpected value: got %s, exp %s", cfg.Spotify.ClientSecret, env.SpotifyClientSecret)
		}
	})
}

func TestValidate(t *testing.T) {
	cfg, err := parseConfig(t)
	if err != nil {
		t.Fatalf("error setting up test: %v", err)
	}

	tests := []struct {
		n   string
		cfg *Config
		exp string
	}{
		{
			"should validate",
			cfg,
			"",
		},
		{
			"should not validate reddit username",
			func(cfg Config) *Config {
				cfg.Reddit.Username = ""
				return &cfg
			}(*cfg),
			"reddit username is missing",
		},
		{
			"should not validate reddit request rate",
			func(cfg Config) *Config {
				cfg.Reddit.RequestRate = 1
				return &cfg
			}(*cfg),
			"reddit request rate must be 2 or higher",
		},
		{
			"should not validate spotify client id",
			func(cfg Config) *Config {
				cfg.Spotify.ClientID = ""
				return &cfg
			}(*cfg),
			"spotify client id is missing",
		},
		{
			"should not validate spotify client secret",
			func(cfg Config) *Config {
				cfg.Spotify.ClientSecret = ""
				return &cfg
			}(*cfg),
			"spotify client secret is missing",
		},
	}

	for _, tc := range tests {
		t.Run(tc.n, func(t *testing.T) {
			err := tc.cfg.validate()

			if err != nil && err.Error() != tc.exp {
				t.Errorf("unexpected result: got %s, exp %s", err, tc.exp)
			}
		})
	}
}

func TestSetDefaultValues(t *testing.T) {
	cfg, err := parseConfig(t)
	if err != nil {
		t.Fatalf("error setting up test: %v", err)
	}

	cfg.Reddit.RequestRate = 0
	cfg.Reddit.MaxRetryAttempts = 0
	cfg.Reddit.RetryAttemptWaitTime = 0

	expRequestRate := 5
	expMaxRetryAttempts := 10
	expRetryAttemptWaitTime := 10

	t.Run("should set default values", func(t *testing.T) {
		cfg.setDefaultValues()

		if cfg.Reddit.RequestRate != expRequestRate {
			t.Errorf("unexpected value: got %d, exp %d", cfg.Reddit.RequestRate, expRequestRate)
		}

		if cfg.Reddit.MaxRetryAttempts != expMaxRetryAttempts {
			t.Errorf("unexpected value: got %d, exp %d", cfg.Reddit.MaxRetryAttempts, expMaxRetryAttempts)
		}

		if cfg.Reddit.RetryAttemptWaitTime != expRetryAttemptWaitTime {
			t.Errorf("unexpected value: got %d, exp %d", cfg.Reddit.RetryAttemptWaitTime, expMaxRetryAttempts)
		}
	})

}
