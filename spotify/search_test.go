package spotify

import "testing"

func TestCreateSearchQuery(t *testing.T) {
	tests := []struct {
		name      string
		title     string
		separator string
		exp       string
	}{
		{
			"test",
			"Something - Something",
			"-",
			"Something Something",
		},
	}

	var c Client

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			sq, err := c.createSearchQuery(tc.title, tc.separator)
			if err != nil {
				t.Errorf("unpexected error: %s", err)
			}

			if sq != tc.exp {
				t.Errorf("unexpected search query: got %s, exp %s", sq, tc.exp)
			}
		})
	}
}
