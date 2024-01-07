package soundcloud

import (
	"encoding/json"
	"errors"
	"fmt"
	"github.com/antchfx/htmlquery"
	"golang.org/x/net/html"
	"regexp"
)

// GetTrackAuthorization will get the track authorization URL from the SoundCloud song URL
func (s *Soundcloud) GetTrackAuthorization(doc *html.Node) (string, error) {
	// Declare the variable to hold the track authorization
	var trackAuth string

	// XPath query
	trackAuthorizationPath := "//script[contains(text(), 'track_authorization')]"

	// Query the document for the script node
	nodes, err := htmlquery.QueryAll(doc, trackAuthorizationPath)
	if err != nil {
		fmt.Println("Error executing XPath query:", err)
		return "", err
	}

	found := false // Flag to indicate if track authorization is found
	for _, node := range nodes {
		scriptContent := htmlquery.InnerText(node)

		// Use regex to extract the JSON part from the script content
		re := regexp.MustCompile(`window\.__sc_hydration\s*=\s*(\[{.*?\}]);`)
		matches := re.FindStringSubmatch(scriptContent)
		if len(matches) > 1 {
			jsonData := matches[1]

			// Parse the JSON
			var data []map[string]interface{}
			if err := json.Unmarshal([]byte(jsonData), &data); err != nil {
				fmt.Println("Error parsing JSON:", err)
				return "", err
			}

			// Traverse the JSON data to find the track_authorization
			for _, item := range data {
				if val, ok := item["hydratable"]; ok && val == "sound" {
					if soundData, ok := item["data"].(map[string]interface{}); ok {
						if ta, ok := soundData["track_authorization"].(string); ok {
							trackAuth = ta
							found = true
							break // Exit the loop as we found the track authorization
						}
					}
				}
			}
		}
		if found {
			break // Exit the main loop as track authorization is found
		}
	}

	if !found {
		return "", errors.New("track_authorization not found")
	}

	// Return the track authorization
	return trackAuth, nil
}
