package spotify

import (
	"errors"
	"fmt"
	"regexp"
	"strings"
	"time"

	"github.com/zmb3/spotify"
)

func (c *Client) getTrack(title string) (spotify.FullTrack, error) {
	var track spotify.FullTrack

	separators := []string{"-", "~", "|", "by", "--"}

	for _, s := range separators {
		searchQuery, err := c.createSearchQuery(title, s)
		if err != nil {
			return track, fmt.Errorf("search query: %w", err)
		}

		res, err := c.search(searchQuery)
		if err != nil {
			return track, fmt.Errorf("searching: %w", err)
		}

		cmprTitle := strings.ToLower(title)

		for _, t := range res.Tracks.Tracks {
			if strings.Contains(cmprTitle, strings.ToLower(t.Name)) {
				for _, artist := range t.Artists {
					if strings.Contains(cmprTitle, strings.ToLower(artist.Name)) { // TODO attempt replacing & with and in title
						c.Log(fmt.Sprintf("\ttrack found: %s (%s)", title, t.ID))
						return t, nil
					}
				}
			}
		}

		time.Sleep(1 * time.Second) // TODO: Handle better with workers for entire chan
	}

	return track, errors.New(fmt.Sprintf("no track found: %s", title))
}

func (c *Client) createSearchQuery(title string, separator string) (string, error) {
	re := regexp.MustCompile(`\(([^)]+)\)|\[([^)]+)\]`)
	replacedTitle := re.ReplaceAll([]byte(title), []byte(""))
	searchQuery := string(replacedTitle)
	searchQuery = strings.ReplaceAll(searchQuery, "'", "")
	searchQuery = strings.ReplaceAll(searchQuery, "\"", "")

	splitTitle := strings.Split(searchQuery, fmt.Sprintf(" %s ", separator))

	if len(splitTitle) <= 1 {
		return "", errors.New(fmt.Sprintf("not able to find title and/or artist: %s", searchQuery))
	}

	searchQuery = strings.Join(splitTitle, " ")
	searchQuery = strings.TrimSpace(searchQuery)

	c.Log(fmt.Sprintf("\tsearch query: \"%s\" from title: %s", searchQuery, title))
	return title, nil
}

func (c *Client) search(q string) (*spotify.SearchResult, error) {
	return c.C.Search(q, spotify.SearchTypeAlbum|spotify.SearchTypeArtist|spotify.SearchTypeTrack)
}
