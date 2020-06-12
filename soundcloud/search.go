package soundcloud

import (
	"fmt"
)

// Search returns the first URL on SoundCloud most similar to the user input
func Search(query string) string {
	base := "https://soundcloud.com/search?q=%s"
	searchQueryString := fmt.Sprintf(base, query)
	// TODO: implement return url
	return searchQueryString
}

// isSong looks at a given SoundCloud URL and determine if the URL is a song or not
func isSong(url string) bool {
	// TODO: implement
	return true
}
