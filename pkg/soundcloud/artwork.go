package soundcloud

import (
	"bytes"
	"io/ioutil"
	"log"
	"net/http"

	"github.com/PuerkitoBio/goquery"
)

// GetArtwork returns the song artwork url and image
func GetArtwork(data []byte) (string, []byte) {
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
	artworkresp, err := http.Get(url)
	if err != nil {
		log.Fatalln(err)
	}
	// image data
	image, err := ioutil.ReadAll(artworkresp.Body)
	if err != nil {
		log.Fatalln(err)
	}
	return url, image
}
