package soundcloud

import (
	"bytes"
	"io/ioutil"
	"log"
	"net/http"
	"regexp"

	"github.com/PuerkitoBio/goquery"
)

// GetClientID returns a the new generated client_id when a request is made to SoundCloud's API
// TODO implment tests
func GetClientID(data []byte) string {
	var url string
	r := bytes.NewReader(data)
	doc, err := goquery.NewDocumentFromReader(r)
	if err != nil {
		panic(err)
	}
	doc.Find("body > script:nth-child(15)").Each(func(i int, s *goquery.Selection) {
		data, exists := s.Attr("src")
		if exists {
			url = data
		}
	})
	// request to soundcloud url
	resp, err := http.Get(url)
	if err != nil {
		log.Fatalln(err)
	}
	// response body
	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		log.Fatalln(err)
	}
	var re = regexp.MustCompile(`client_id:"(.+)",env`) // pattern for finding encrypted string ID
	clientString := re.FindString(string(body))         // stream URL
	var ree = regexp.MustCompile(`"([^"].*?)"`)

	ID := ree.FindString(clientString)
	reg, err := regexp.Compile("[^a-zA-Z0-9]+")
	if err != nil {
		log.Fatal(err)
	}
	clientID := reg.ReplaceAllString(ID, "")

	return clientID
}
