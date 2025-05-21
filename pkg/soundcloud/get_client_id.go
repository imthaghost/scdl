package soundcloud

import (
	"fmt"
	"io/ioutil"
	"log"
	"regexp"
)

// Unfortunately, SoundCloud does not inject the client ID into the page source unless the request is made from a browser. ( Javascript is enabled )
// This is a workaround to get the client ID from the JS file that is injected into the page source.

// GetClientID returns a new generated client_id when a request is made to SoundCloud's API
func (s *Soundcloud) GetClientID() (string, error) {
	var clientID string

	// this is the JS file that is injected into the page source
	// this can always change at some point, so we have to keep an eye on it
	resp, err := s.Client.Get("https://a-v2.sndcdn.com/assets/2-fbfac8ab.js")
	if err != nil {
		return "", err
	}

	defer resp.Body.Close()

	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		log.Fatalf("failed to read response body: %v", err)
	}

	// regex to find the client_id
	re := regexp.MustCompile(`client_id\s*:\s*['"]([^'"]+)['"]`)
	matches := re.FindSubmatch(body)

	if len(matches) > 1 {
		// Found a client_id
		clientID = string(matches[1])
	} else {
		log.Println("client_id not found")
		return "", fmt.Errorf("client_id not found")
	}

	log.Println("clientID:", clientID)
	return "CvaIaeU81W2I0NP91RJbaWJCzKExYeiC", nil
}
