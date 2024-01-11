package soundcloud

import (
	"encoding/json"
	"errors"
	"fmt"
	"github.com/antchfx/htmlquery"
	"golang.org/x/net/html"
	"regexp"
)

// GetHLSURL gets the HLS URL from the SoundCloud song URL
func (s *Soundcloud) GetHLSURL(doc *html.Node) (string, error) {
	// TODO: This function is hideous. Refactor it.
	// Declare the variable to hold the HLS URL
	var hlsURL string

	// XPath query to find the script node containing the relevant JSON
	hlsURLPath := "//script[contains(text(), 'media') and contains(text(), 'transcodings')]"

	// Query the document for the script node
	nodes, err := htmlquery.QueryAll(doc, hlsURLPath)
	if err != nil {
		fmt.Println("Error executing XPath query:", err)
		return "", err
	}

	found := false // Flag to indicate if HLS URL is found
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

			// Traverse the JSON data to find the HLS URL
			for _, item := range data {
				if val, ok := item["hydratable"]; ok && val == "sound" {
					if soundData, ok := item["data"].(map[string]interface{}); ok {
						if media, ok := soundData["media"].(map[string]interface{}); ok {
							if transcodings, ok := media["transcodings"].([]interface{}); ok {
								for _, transcoding := range transcodings {
									if t, ok := transcoding.(map[string]interface{}); ok {
										if format, ok := t["format"].(map[string]interface{}); ok {
											if mimeType, ok := format["mime_type"].(string); ok && mimeType == "audio/mpeg" {
												if url, ok := t["url"].(string); ok {
													hlsURL = url
													found = true
													break // Found the HLS URL
												}
											}
										}
									}
								}
							}
						}
					}
				}
				if found {
					break // Exit the loop as we found the HLS URL
				}
			}
		}
		if found {
			break // Exit the main loop as HLS URL is found
		}
	}

	if !found {
		return "", errors.New("HLS URL not found")
	}

	// Return the HLS URL
	return hlsURL, nil
}
