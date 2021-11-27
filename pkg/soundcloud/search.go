package soundcloud

import (
	"bytes"
	"fmt"
	"io/ioutil"
	"log"
	"net/http"

	"github.com/PuerkitoBio/goquery"
)

// Search returns the first URL on SoundCloud most similar to the user input
// TODO implment tests
func Search(query string) string {
	base := "https://soundcloud.com/search?q=%s"
	searchQueryString := fmt.Sprintf(base, query)
	//fmt.Println(searchQueryString)
	// request to soundcloud url
	resp, err := http.Get(searchQueryString)
	if err != nil {
		log.Fatalln(err)
	}
	// response body
	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		log.Fatalln(err)
	}
	var urls []string
	r := bytes.NewReader(body)
	doc, err := goquery.NewDocumentFromReader(r)
	if err != nil {
		panic(err)
	}
	// TODO: fix
	// Find the search items
	doc.Find("noscript").Each(func(i int, s *goquery.Selection) {
		fmt.Println(s.Text())
		x := s.Find("ul").Text()
		fmt.Print(x)
		// For each item found, get a tag
		// link, exists := x.Attr("href")
		// fmt.Println(link)
		// if exists {
		// 	urls = append(urls, link)
		// }

	})
	fmt.Println(urls)
	// TODO: implement return url
	return searchQueryString
}
