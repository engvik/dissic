package spotify

import "testing"

func TestTitleStringSlice(t *testing.T) {
	tests := []struct {
		n   string
		m   Music
		exp []string
	}{
		{
			"should correctly format empty struct",
			Music{},
			[]string{"", "", ""},
		},
		{
			"should correctly format one passed field",
			Music{
				PostTitle: "post title",
			},
			[]string{"post title", "", ""},
		},
		{
			"should correctly format two passed fields",
			Music{
				PostTitle:  "post title",
				MediaTitle: "media title",
			},
			[]string{"post title", "media title", ""},
		},
		{
			"should correctly format three passed fields",
			Music{
				PostTitle:        "post title",
				MediaTitle:       "media title",
				SecureMediaTitle: "secure media title",
			},
			[]string{"post title", "media title", "secure media title"},
		},
	}

	for _, tc := range tests {
		t.Run(tc.n, func(t *testing.T) {
			titles := tc.m.titleStringSlice()

			if len(titles) != 3 {
				t.Errorf("unexpected slice length: got %d, exp %d", len(titles), 3)
			}

			for i, title := range titles {
				if title != tc.exp[i] {
					t.Errorf("unexpected value: got %s, exp %s, pos %d", title, tc.exp[i], i)
				}
			}
		})
	}
}
