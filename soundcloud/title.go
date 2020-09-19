package soundcloud

import (
	"bytes"
	"fmt"
	"strings"

	"github.com/PuerkitoBio/goquery"
)

// GetTitle returns title of the song
// TODO: implement tests
func GetTitle(data []byte) string {
	var title string
	r := bytes.NewReader(data)
	doc, err := goquery.NewDocumentFromReader(r)
	if err != nil {
		panic(err)
	}
	// the title of the song can be found in the meta tag from initial response
	doc.Find("meta[property='twitter:title']").Each(func(i int, s *goquery.Selection) {
		// get the data from found element's content attribute
		data, exists := s.Attr("content")
		if exists {
			title = data

		}
	})
	// sanitize
	title = strings.Replace(title, "/", "", -1)

	fmt.Println(title)

	return title
}
