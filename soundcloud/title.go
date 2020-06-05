package soundcloud

import (
	"bytes"

	"github.com/PuerkitoBio/goquery"
)

// GetTitle returns title of the song
func GetTitle(data []byte) string {
	var title string
	r := bytes.NewReader(data)
	doc, err := goquery.NewDocumentFromReader(r)
	if err != nil {
		panic(err)
	}
	doc.Find("meta[property='twitter:title']").Each(func(i int, s *goquery.Selection) {
		data, exists := s.Attr("content")
		if exists {
			title = data

		}
	})
	return title
}
