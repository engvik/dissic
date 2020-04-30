package spotify

type Music struct {
	Sub              string
	PostTitle        string
	MediaTitle       string
	SecureMediaTitle string
}

func (m *Music) titleStringArray() []string {
	return []string{m.PostTitle, m.MediaTitle, m.SecureMediaTitle}
}
