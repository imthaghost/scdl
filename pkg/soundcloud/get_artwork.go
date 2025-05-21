package soundcloud

import (
	"fmt"
	"github.com/antchfx/htmlquery"
	"golang.org/x/net/html"
)

// GetArtwork returns the artwork for the song
func (s *Soundcloud) GetArtwork(doc *html.Node) (string, error) {
	// XPath query
	artworkPath := "//meta[@property='og:image']/@content"

	// Query the document for the artwork node
	nodes, err := htmlquery.QueryAll(doc, artworkPath)
	if err != nil {
		fmt.Println("Error executing XPath query:", err)
		return "", err
	}

	// Check if any nodes were found
	if len(nodes) > 0 {
		// Extract the content from the first node
		artwork := htmlquery.InnerText(nodes[0])

		return artwork, nil
	}

	return "", nil
}
