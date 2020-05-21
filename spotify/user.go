package spotify

import (
	"fmt"
)

func (c *Client) SetUser() error {
	user, err := c.C.CurrentUser()
	if err != nil {
		return fmt.Errorf("getting user: %w", err)
	}

	c.Log(fmt.Sprintf("retrived user: %s", user.ID))

	c.User = user

	return nil
}
