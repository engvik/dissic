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
			"should parse separator '-' correctly",
			"Something - Something",
			"-",
			"Something Something",
		},
		{
			"should parse separator '~' correctly",
			"Something ~ Something",
			"~",
			"Something Something",
		},
		{
			"should parse separator '|' correctly",
			"Something | Something",
			"|",
			"Something Something",
		},
		{
			"should parse separator 'by' correctly",
			"Something by Something",
			"by",
			"Something Something",
		},
		{
			"should parse separator '--' correctly",
			"Something -- Something",
			"--",
			"Something Something",
		},
		{
			"should parse separator 'ー' correctly",
			"Something ー Something",
			"ー",
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
