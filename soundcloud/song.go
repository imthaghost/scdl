package soundcloud

import (
	"encoding/json"
	"io/ioutil"
	"log"
	"net/http"
	"regexp"

	"github.com/bogem/id3v2"
	"github.com/imthaghost/scdl/mp3"
)

type audioLink struct {
	URL string `json:"url"`
}

// var clientID = "iY8sfHHuO2UsXy1QOlxthZoMJEY9v0eI" // anonymous user clientID will be static in v1

// ExtractSong queries the SoundCloud api and receives a file with urls
func ExtractSong(url string) {

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
	// song name
	songname := GetTitle(body)
	// artwork url
	artworkURL := GetArtwork(body)

	clientID := GetClientID(body)

	// request to soundcloud url
	artworkresp, err := http.Get(artworkURL)
	if err != nil {
		log.Fatalln(err)
	}
	image, err := ioutil.ReadAll(artworkresp.Body)
	if err != nil {
		log.Fatalln(err)
	}
	// TODO improve pattern for finding encrypted string ID
	var re = regexp.MustCompile(`https:\/\/api-v2.*\/stream\/hls`) // pattern for finding encrypted string ID
	// TODO not needed if encrypted string ID regex pattern is improved
	var ree = regexp.MustCompile(`.+?(stream)`) // pattern for finding stream URL

	streamURL := re.FindString(string(body)) // stream URL

	baseURL := ree.FindString(streamURL) // baseURL ex: https://api-v2.soundcloud.com/media/soundcloud:tracks:816595765/0ad937d5-a278-4b36-b128-220ac89aec04/stream

	requestURL := baseURL + "/hls?client_id=" + clientID // API query string ex: https://api-v2.soundcloud.com/media/soundcloud:tracks:805856467/ddfb7463-50f1-476c-9010-729235958822/stream/hls?client_id=iY8sfHHuO2UsXy1QOlxthZoMJEY9v0eI

	// query API
	r, e := http.Get(requestURL)
	if err != nil {
		log.Fatalln(e)
	}

	// API response returns a m3u8 URL
	m3u8Reponse, er := ioutil.ReadAll(r.Body)
	if er != nil {
		log.Fatalln(er)
	}

	var a audioLink

	audioerr := json.Unmarshal(m3u8Reponse, &a)
	if er != nil {
		panic(audioerr)
	}

	// merege segments
	mp3.Merge(a.URL, songname)

	tag, err := id3v2.Open(songname+".mp3", id3v2.Options{Parse: true})
	if tag == nil || err != nil {
		log.Fatal("Error while opening mp3 file: ", err)
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
