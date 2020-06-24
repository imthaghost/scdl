package soundcloud

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	"regexp"

	"github.com/bogem/id3v2"
	"github.com/fatih/color"
	"github.com/imthaghost/scdl/mp3"
)

type audioLink struct {
	URL string `json:"url"`
}

// ExtractSong queries the SoundCloud api and receives a m3u8 file, then binds the segments received into a .mp3 file
// TODO: implement tests
func ExtractSong(url string) {

	// request to user inputed SoundCloud URL
	resp, err := http.Get(url)
	if err != nil {
		log.Fatalln(err)
	}
	// response
	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		log.Fatalln(err)
	}

	// parse the response data to grab the song name
	songname := GetTitle(body)

	// parse the response data to grab the artwork image and url
	_, image := GetArtwork(body)

	// parse the response data and make a reqeust to receive clien_id embedded in the javascript
	clientID := GetClientID(body)

	// TODO improve pattern for finding encrypted string ID
	var re = regexp.MustCompile(`https:\/\/api-v2.*\/stream\/hls`) // pattern for finding encrypted string ID
	// TODO not needed if encrypted string ID regex pattern is improved
	var ree = regexp.MustCompile(`.+?(stream)`) // pattern for finding stream URL

	streamURL := re.FindString(string(body)) // stream URL

	baseURL := ree.FindString(streamURL) // baseURL ex: https://api-v2.soundcloud.com/media/soundcloud:tracks:816595765/0ad937d5-a278-4b36-b128-220ac89aec04/stream

	// TODO: replace with format string instead of concatenation
	requestURL := baseURL + "/hls?client_id=" + clientID // API query string ex: https://api-v2.soundcloud.com/media/soundcloud:tracks:805856467/ddfb7463-50f1-476c-9010-729235958822/stream/hls?client_id=iY8sfHHuO2UsXy1QOlxthZoMJEY9v0eI

	// query API
	r, err := http.Get(requestURL)
	if err != nil {
		red := color.New(color.FgRed).SprintFunc()
		fmt.Printf("%s Error making request to API %s\n", red("[-]"), err)
	}

	// API response returns a m3u8 file embedded in URL
	m3u8Reponse, err := ioutil.ReadAll(r.Body)
	if err != nil {
		red := color.New(color.FgRed).SprintFunc()
		fmt.Printf("%s Error creating reader from api response %s\n", red("[-]"), err)
	}

	var a audioLink

	// unmarshal json data from response
	audioerr := json.Unmarshal(m3u8Reponse, &a)
	if audioerr != nil {
		red := color.New(color.FgRed).SprintFunc()
		fmt.Printf("%s Error unmarshalling API response: %s\n", red("[-]"), audioerr)
	}

	// merege segments
	mp3.Merge(a.URL, songname)

	// replace empty cover image with SoundCloud artwork
	tag, err := id3v2.Open(songname+".mp3", id3v2.Options{Parse: true})
	if tag == nil || err != nil {
		red := color.New(color.FgRed).SprintFunc()
		fmt.Printf("%s Error while opening mp3 file: %s\n", red("[-]"), err)
	}
	pic := id3v2.PictureFrame{
		Encoding:    id3v2.EncodingUTF8,
		MimeType:    "image/jpeg",
		PictureType: id3v2.PTFrontCover,
		Description: "Front cover",
		Picture:     image,
	}
	tag.AddAttachedPicture(pic)
	tag.Save()
}
