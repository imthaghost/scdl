package soundcloud

import (
	"fmt"
	"golang.org/x/net/html"
	"strings"
)

// ConstructStreamURL will construct a stream url from a SoundCloud song url
func (s *Soundcloud) ConstructStreamURL(doc *html.Node) (string, error) {

	// get client id
	clientID, err := s.GetClientID()
	if err != nil {
		return "", err
	}

	// get track authorization
	trackAuth, err := s.GetTrackAuthorization(doc)
	if err != nil {
		return "", err
	}

	// get hls stream url
	hlsStreamURL, err := s.GetHLSURL(doc)
	if err != nil {
		return "", err
	}
	
	trackID, streamToken, err := getTrackInfo(hlsStreamURL)

	// construct stream url
	baseURL := "https://api-v2.soundcloud.com/media/soundcloud:tracks:%s/%s/stream/hls?client_id=%s&track_authorization=%s"

	streamURL := fmt.Sprintf(baseURL, trackID, streamToken, clientID, trackAuth)

	return streamURL, nil
}

func getTrackInfo(url string) (string, string, error) {
	// Split the URL by '/'
	parts := strings.Split(url, "/")

	// Check if the parts length is as expected
	if len(parts) < 8 {
		return "", "", fmt.Errorf("invalid URL format")
	}

	// Extract the part that contains 'soundcloud:tracks:TRACK_ID'
	trackIDPart := parts[4] // "soundcloud:tracks:373180994"

	// Further split the track ID part to get the actual track ID
	trackIDParts := strings.Split(trackIDPart, ":")
	if len(trackIDParts) < 3 {
		return "", "", fmt.Errorf("invalid track ID format in URL")
	}
	trackID := trackIDParts[2]

	// The stream token is the next part of the URL
	streamToken := parts[5] // "ca04c4eb-a299-4f4b-9ff1-ac20ed014fe5"

	return trackID, streamToken, nil
}
