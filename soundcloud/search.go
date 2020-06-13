package soundcloud

import (
	"bytes"
	"fmt"

	"github.com/PuerkitoBio/goquery"
)

// Search returns the first URL on SoundCloud most similar to the user input
func Search(query string) string {
	base := "https://soundcloud.com/search?q=%s"
	searchQueryString := fmt.Sprintf(base, query)
	fmt.Println(searchQueryString)
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

	searchFlg := flag.String("searchFlg", "", "flag for the search function")
	flag.Parse()

	fmt.Println(searchFlg)

	return searchQueryString
}

// isSong looks at a given SoundCloud URL and determine if the URL is a song or not
func isSong(url string) bool {
	// TODO: implement
	base := "https://soundcloud.com/search?q=%s"
	
	resp, err := http.Get(base)
	if err != nil {
		panic(err)
	}
	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		log.Fatalln(err)
	}
	var re []string
	r := bytes.NewReader(body)
	doc, err := goquery.NewDocumentFromReader(r)
	if err != nil {
		panic(err)
	}
}
