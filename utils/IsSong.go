package utils

import (
	"bytes"
	"fmt"
	"io/ioutil"
	"log"
	"net/http"

	"github.com/PuerkitoBio/goquery"
)

// IsSong makes a request to the given URL and determines if the URL is a song or not
// TODO: implement tests
func IsSong(url string) bool {
	/*
		>>> https://soundcloud.com/uiceheidd
		<<< fasle

		>>> https://soundcloud.com/uiceheidd/tell-me-you-love-me
		<<< true
	*/
	if !ValidateURL(url) {
		return false
	}
	// request to soundcloud url
	resp, err := http.Get(url)
	if err != nil {
		log.Fatalln(err)
	}
	// response body
	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		fmt.Println(err)
	}
	// turn byte array into reader
	data := bytes.NewReader(body)
	// create new document from reader
	doc, err := goquery.NewDocumentFromReader(data)
	if err != nil {
		panic(err)
	}
	// find first instance of meta tag with property soundcloug:like_count
	x := doc.Find("meta[property='soundcloud:like_count']").First()
	// tag's data
	_, exists := x.Attr("content")
	if exists {
		return true
	}
	return false
}
