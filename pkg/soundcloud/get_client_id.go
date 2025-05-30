package soundcloud

import (
	"fmt"
	"io"
	"log"
	"regexp"
)

// Unfortunately, SoundCloud does not inject the client ID into the page source unless the request is made from a browser. ( Javascript is enabled )
// This is a workaround to get the client ID from the JS file that is injected into the page source.

// GetClientID returns a new generated client_id when a request is made to SoundCloud's API
func (s *Soundcloud) GetClientID() (string, error) {
	// Make the initial request to SoundCloud
	resp, err := s.Client.Get("https://soundcloud.com")
	if err != nil {
		return "", err
	}
	defer resp.Body.Close()

	// Read the response body
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		log.Printf("failed to read response body: %v", err)
		return "", fmt.Errorf("failed to read response body: %v", err)
	}

	// Compile regex to find asset URLs
	assetre := regexp.MustCompile(`src="(https:\/\/a-v2\.sndcdn\.com\/assets\/[^\s"]+)"`)
	assetMatches := assetre.FindAllSubmatch(body, -1)

	// Check if any asset matches were found
	if len(assetMatches) == 0 {
		return "", fmt.Errorf("asset not found")
	}

	// Iterate over asset matches to find the client ID
	for _, match := range assetMatches {
		assetURL := string(match[1])

		// Make a request to the asset URL
		resp, err := s.Client.Get(assetURL)
		if err != nil {
			return "", err
		}
		defer resp.Body.Close()

		// Read the response body
		body, err := io.ReadAll(resp.Body)
		if err != nil {
			log.Printf("failed to read response body: %v\n", err)
			return "", fmt.Errorf("failed to read response body: %v", err)
		}

		// Compile regex to find the client ID
		re := regexp.MustCompile(`client_id:\"([^\"]+)\"`)
		matches := re.FindSubmatch(body)

		// Check if a client ID was found
		if len(matches) > 1 {
			return string(matches[1]), nil
		}
	}

	// If no client ID was found
	log.Println("client_id not found")
	return "", fmt.Errorf("client_id not found")
}
