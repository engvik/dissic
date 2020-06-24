package spotify

import (
	"fmt"
)

// SetUser fetches the authenticated users and
// sets it on the client.
func (c *Client) SetUser() error {
	user, err := c.Spotify.CurrentUser()
	if err != nil {
		return fmt.Errorf("getting user: %w", err)
	}

	c.Log(fmt.Sprintf("retrived user: %s", user.ID))

	c.User = user

	return nil
}
