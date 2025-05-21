package soundcloud

import (
	"fmt"
	"github.com/antchfx/htmlquery"
	"golang.org/x/net/html"
)

// GetTitle will return the title of the song
func (s *Soundcloud) GetTitle(doc *html.Node) (string, error) {
	// XPath query
	titlePath := "//meta[@property='og:title']/@content"

	// Query the document for the title node
	nodes, err := htmlquery.QueryAll(doc, titlePath)
	if err != nil {
		fmt.Println("Error executing XPath query:", err)
		return "", err
	}

	// Check if any nodes were found
	if len(nodes) > 0 {
		// Extract the content from the first node
		title := htmlquery.InnerText(nodes[0])
		return title, nil
	}

	return "", nil
}
