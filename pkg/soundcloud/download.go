package soundcloud

import (
	"encoding/json"
	"fmt"
	"github.com/antchfx/htmlquery"
	"github.com/imthaghost/scdl/pkg/mp3"
	"io/ioutil"
	"log"
	"net/http"
)

// Download queries the SoundCloud api and receives a m3u8 file, then binds the segments received into a .mp3 file
func Download(url string) {
	// TODO: This client should be created higher up in the stack
	soundcloud := NewClient()

	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		log.Println(err)
	}

	// set Non Hacker User Agent
	req.Header.Set("Accept", soundcloud.userAgent)

	resp, err := soundcloud.Client.Do(req)
	if err != nil {
		log.Println(err)
	}

	// parse html
	doc, err := htmlquery.Parse(resp.Body)
	if err != nil {
		log.Println(err)
	}

	streamURL, err := soundcloud.ConstructStreamURL(doc)
	if err != nil {
		log.Println(err)
	}

	songName, err := soundcloud.GetTitle(doc)
	if err != nil {
		log.Println(err)
	}

	artwork, err := soundcloud.GetArtwork(doc)
	if err != nil {
		log.Println(err)
	}

	// Get the response from the URL
	streamResp, err := http.Get(streamURL)
	if err != nil {
		log.Println(err)
		return
	}
	defer streamResp.Body.Close()

	// Read the body of the response
	body, err := ioutil.ReadAll(streamResp.Body)
	if err != nil {
		log.Println("Error reading response body:", err)
		return
	}

	// Unmarshal the JSON into the struct
	var audioResp AudioLink
	err = json.Unmarshal(body, &audioResp)
	if err != nil {
		fmt.Println("Error unmarshalling JSON:", err)
		return
	}

	// merge segments
	mp3.Merge(audioResp.URL, songName)

	artworkResp, err := http.Get(artwork)
	image, err := ioutil.ReadAll(artworkResp.Body)
	if err != nil {
		log.Println(err)
	}

	// set cover image for mp3 file
	mp3.SetCoverImage(songName+".mp3", image)
}
