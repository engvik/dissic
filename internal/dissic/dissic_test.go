package dissic

import (
	"fmt"
	"net/http"
	"os"
	"testing"

	"github.com/engvik/dissic/internal/config"
	"github.com/turnage/graw/reddit"
)

type spotifyTestService struct{}

func (s *spotifyTestService) Authenticate(openBrowser bool) error       { return nil }
func (s *spotifyTestService) Listen()                                   {}
func (s *spotifyTestService) Close()                                    {}
func (s *spotifyTestService) PreparePlaylists(cfg *config.Config) error { return nil }
func (s *spotifyTestService) AuthHandler() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {}
}
func (s *spotifyTestService) SetUser() error { return nil }

type redditTestService struct {
}

func (r *redditTestService) PrepareScanner() error            { return nil }
func (r *redditTestService) Listen(shutdown chan<- os.Signal) {}
func (r *redditTestService) Close()                           {}
func (r *redditTestService) Post(post *reddit.Post) error     { return nil }

func TestNew(t *testing.T) {
	cfg := &config.Config{}
	s := &spotifyTestService{}
	r := &redditTestService{}

	mux := http.NewServeMux()
	mux.HandleFunc("/spotifyAuth", s.AuthHandler())

	t.Run("should create dissic service", func(t *testing.T) {
		d := New(cfg, s, r, mux)

		if d.Config == nil {
			fmt.Errorf("dissic service missing config")
		}

		if d.Spotify == nil {
			fmt.Errorf("dissic service missing spotify service")
		}

		if d.Reddit == nil {
			fmt.Errorf("dissic service missing reddit service")
		}

		if d.HTTP == nil {
			fmt.Errorf("dissic service missing http server")
		}
	})
}
