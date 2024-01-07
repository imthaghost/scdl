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

// AudioLink struct for unmarshalling data
type AudioLink struct {
	URL string `json:"url"`
}

// ExtractSong queries the SoundCloud api and receives a m3u8 file, then binds the segments received into a .mp3 file
func ExtractSong(url string) {
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

	//body, _ := io.ReadAll(resp.Body)
	//log.Println(string(body))

	// parse html
	doc, err := htmlquery.Parse(resp.Body)
	if err != nil {
		log.Println(err)
	}

	streamURL, err := soundcloud.ConstructStreamURL(doc)
	if err != nil {
		log.Println(err)
	}

	log.Println(streamURL)

	songName, err := soundcloud.GetTitle(doc)
	if err != nil {
		log.Println(err)
	}

	log.Println(songName)

	artwork, err := soundcloud.GetArtwork(doc)
	if err != nil {
		log.Println(err)
	}

	log.Println(artwork)

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
	// TODO: put this code somewhere so that the image gets set at the same time as the song data is being written for smoother transition
	mp3.SetCoverImage(songName+".mp3", image)
}
