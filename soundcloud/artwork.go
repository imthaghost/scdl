package soundcloud

import (
	"bytes"

	"github.com/PuerkitoBio/goquery"
)

// TODO implement tests
// TODO return image instead of url
// GetArtwork returns the song artwork url
func GetArtwork(data []byte) string {
	var url string
	r := bytes.NewReader(data)
	doc, err := goquery.NewDocumentFromReader(r)
	if err != nil {
		panic(err)
	}
	doc.Find("meta[property='twitter:image']").Each(func(i int, s *goquery.Selection) {
		data, exists := s.Attr("content")
		if exists {
			url = data
		}
	})
	return url
}
