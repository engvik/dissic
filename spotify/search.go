package spotify

import (
	"errors"
	"fmt"
	"regexp"
	"strings"

	"github.com/zmb3/spotify"
)

func (c *Client) getTrack(title string) (spotify.FullTrack, error) {
	var track spotify.FullTrack

	sq, err := c.createSearchQuery(title)
	if err != nil {
		return track, fmt.Errorf("search query: %w", err)
	}

	c.log(fmt.Sprintf("\tsearch query: %s from title: %s", sq, title))

	res, err := c.search(sq)
	if err != nil {
		return track, fmt.Errorf("error searching: %w", err)
	}

	cmprTitle := strings.ToLower(title)

	for _, t := range res.Tracks.Tracks {
		if strings.Contains(cmprTitle, strings.ToLower(t.Name)) {
			for _, artist := range t.Artists {
				if strings.Contains(cmprTitle, strings.ToLower(artist.Name)) { // TODO attempt replacing & with and in title
					c.log(fmt.Sprintf("\ttrack found: %s (%s)", title, t.ID))
					return t, nil
				}
			}
		}
	}

	return track, errors.New(fmt.Sprintf("no track found: %s", title))
}

func (c *Client) createSearchQuery(t string) (string, error) {
	re := regexp.MustCompile(`\(([^)]+)\)|\[([^)]+)\]`)
	replacedTitle := re.ReplaceAll([]byte(t), []byte(""))
	title := string(replacedTitle)
	title = strings.ReplaceAll(title, "'", "")
	title = strings.ReplaceAll(title, "\"", "")

	var splitTitle []string
	separators := []string{"-", "~", "|", "by"}

	for _, s := range separators {
		splitTitle = strings.Split(title, fmt.Sprintf(" %s ", s))

		if len(splitTitle) >= 1 {
			break
		}
	}

	if len(splitTitle) <= 1 {
		return "", errors.New(fmt.Sprintf("not able to find title and/or artist: %s", title))
	}

	title = strings.Join(splitTitle, " ")
	title = strings.TrimSpace(title)

	return title, nil
}

func (c *Client) search(q string) (*spotify.SearchResult, error) {
	res, err := c.C.Search(q, spotify.SearchTypeAlbum|spotify.SearchTypeArtist|spotify.SearchTypeTrack)
	if err != nil {
		return nil, fmt.Errorf("error searching: %w", err)
	}

	return res, nil
}
