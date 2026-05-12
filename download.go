package main

import (
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
	"strings"

	"github.com/antchfx/htmlquery"
)

// audioLink is SoundCloud's response wrapper around an HLS playlist URL.
type audioLink struct {
	URL string `json:"url"`
}

// filenameSanitizer strips characters that are invalid or annoying in cross-
// platform filenames. Kept narrow on purpose so it doesn't rewrite track
// titles into mush.
var filenameSanitizer = strings.NewReplacer(
	"\\", "", "/", "", ":", "", "*", "", "?", "", "\"", "",
	"<", "", ">", "", "|", "", "+", "", "=", "", ",", "",
	".", "", "!", "", "@", "", "#", "", "$", "", "%", "",
	"^", "", "&", "", "(", "", ")", "",
)

// Download fetches a SoundCloud track page, resolves its HLS stream, downloads
// every segment, assembles them into an mp3, and embeds the cover artwork.
//
// If the input URL is a private share link (?secret_token=... or /s-XXXX path
// suffix), the secret token is forwarded to the stream URL request.
func (s *Soundcloud) Download(trackURL string) {
	pageURL, secretToken, err := extractSecretToken(trackURL)
	if err != nil {
		log.Println(err)
		return
	}

	req, err := http.NewRequest("GET", pageURL, nil)
	if err != nil {
		log.Println(err)
		return
	}
	req.Header.Set("User-Agent", s.UserAgent)
	// SoundCloud binds the scraped track_authorization JWT to the requester's
	// identity. To get a JWT with the user's sub (needed for Go+ tracks and
	// private tracks), we have to fetch the page with the oauth_token cookie,
	// not just attach the bearer on api-v2 calls later.
	if s.Token != "" {
		req.AddCookie(&http.Cookie{Name: "oauth_token", Value: s.Token})
	}

	resp, err := s.Client.Do(req)
	if err != nil {
		log.Println(err)
		return
	}
	defer resp.Body.Close()

	doc, err := htmlquery.Parse(resp.Body)
	if err != nil {
		log.Println(err)
		return
	}

	streamURL, transcoding, err := s.ConstructStreamURL(doc, secretToken)
	if err != nil {
		log.Println(err)
		return
	}

	title, err := s.GetTitle(doc)
	if err != nil {
		log.Println(err)
		return
	}
	songName := filenameSanitizer.Replace(title)

	artworkURL, err := s.GetArtwork(doc)
	if err != nil {
		log.Println(err)
		return
	}

	playlistURL, err := s.fetchPlaylistURL(streamURL)
	if err != nil {
		log.Println(err)
		return
	}

	outPath := songName + ".mp3"
	fmt.Printf("%s Quality: %s\n", green("[+]"), qualityLabel(transcoding.Quality))
	if err := downloadHLS(playlistURL, outPath); err != nil {
		log.Println(err)
		return
	}

	if artworkURL != "" {
		image, err := fetchBytes(artworkURL)
		if err != nil {
			log.Println(err)
			return
		}
		if err := SetCoverImage(outPath, image); err != nil {
			log.Println(err)
		}
	}
}

// fetchPlaylistURL hits the stream endpoint and unwraps the {"url": "..."}
// response that points at the actual HLS playlist. The request is authed
// because the endpoint lives on api-v2.soundcloud.com.
func (s *Soundcloud) fetchPlaylistURL(streamURL string) (string, error) {
	resp, err := s.authedGet(streamURL)
	if err != nil {
		return "", err
	}
	defer resp.Body.Close()
	if resp.StatusCode/100 != 2 {
		return "", fmt.Errorf("stream URL request returned status %d", resp.StatusCode)
	}
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return "", err
	}
	var link audioLink
	if err := json.Unmarshal(body, &link); err != nil {
		return "", fmt.Errorf("unmarshal stream response: %w", err)
	}
	if link.URL == "" {
		return "", fmt.Errorf("stream response had empty URL")
	}
	return link.URL, nil
}

// qualityLabel turns SoundCloud's "hq"/"sq" into something readable.
func qualityLabel(q string) string {
	switch q {
	case "hq":
		return "hq (Go+ 256 kbps)"
	case "sq":
		return "sq (standard 128 kbps)"
	case "":
		return "unknown"
	default:
		return q
	}
}

func fetchBytes(u string) ([]byte, error) {
	resp, err := http.Get(u)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()
	return io.ReadAll(resp.Body)
}
