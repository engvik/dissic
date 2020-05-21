package spotify

// Music contains data about potential new music to add to
// a spotify list.
type Music struct {
	Subreddit        string
	PostTitle        string
	MediaTitle       string
	SecureMediaTitle string
}

func (m *Music) titleStringSlice() []string {
	return []string{m.PostTitle, m.MediaTitle, m.SecureMediaTitle}
}
