package spotify

import "net/http"

// AuthHandler is a simple http handler for the authentication callback.
func (c *Client) AuthHandler() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		token, err := c.Auth.Token(c.Session, r)
		if err != nil {
			http.Error(w, "Couldn't get token", http.StatusNotFound)
			return
		}

		c.Spotify = c.Auth.NewClient(token)
		c.AuthChan <- true
		w.Write([]byte("All good - you can close this window now"))
	}
}
