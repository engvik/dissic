package spotify

import (
	"errors"
	"fmt"
	"net/url"
	"regexp"
	"strings"

	"github.com/zmb3/spotify"
)

func (c *Client) getTrackByURL(URL string) (*spotify.FullTrack, error) {
	parsedURL, err := url.Parse(URL)
	if err != nil {
		return nil, fmt.Errorf("parse url: %w", err)
	}

	if parsedURL.Host != "open.spotify.com" {
		return nil, errors.New(fmt.Sprintf("not a spotify url: %s", URL))
	}

	splitURL := strings.Split(parsedURL.Path, "/")

	if len(splitURL) != 3 {
		return nil, errors.New(fmt.Sprintf("unexpected path length: %s", parsedURL.Path))
	}

	if splitURL[1] != "track" {
		return nil, errors.New(fmt.Sprintf("not a track path: %s", parsedURL.Path))
	}

	return c.Spotify.GetTrack(spotify.ID(splitURL[2]))
}

func (c *Client) getTrackByTitles(m Music) (spotify.FullTrack, error) {
	var track spotify.FullTrack

	// loop through possible titles
	for _, title := range m.titleStringSlice() {
		if title == "" {
			continue
		}

		separators := []string{"-", "~", "|", "by", "--", "ー"}

		// attempt finding search query for different track separators
		for _, s := range separators {
			// create search query
			searchQuery, err := c.createSearchQuery(title, s)
			if err != nil {
				c.Logger.Infof("\tsearch query: %s, separator: %s", err, s)
				continue
			}

			// search by query
			res, err := c.Spotify.Search(searchQuery, spotify.SearchTypeAlbum|spotify.SearchTypeArtist|spotify.SearchTypeTrack)
			if err != nil {
				c.Logger.Infof("search: %s", err)
				continue
			}

			track, found := c.findMatchFromSearchResult(title, res)
			if found {
				return track, nil
			}
		}
	}

	return track, errors.New("no track found")
}

func (c *Client) findMatchFromSearchResult(title string, res *spotify.SearchResult) (spotify.FullTrack, bool) {
	cmprTitle := strings.ToLower(title)

	// figure out if the search result is a match
	for _, t := range res.Tracks.Tracks {
		if strings.Contains(cmprTitle, strings.ToLower(t.Name)) {
			for _, artist := range t.Artists {
				if strings.Contains(cmprTitle, strings.ToLower(artist.Name)) {
					c.Logger.Infof("\ttrack found: %s (%s)", title, t.ID)
					return t, true
				}
			}
		}
	}

	return spotify.FullTrack{}, false
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

	c.Logger.Info("\tsearch query: \"%s\" from title: %s", searchQuery, title)
	return searchQuery, nil
}
