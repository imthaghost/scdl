package main

import (
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/url"
	"regexp"
	"strings"

	"github.com/antchfx/htmlquery"
	"golang.org/x/net/html"
)

var (
	assetURLRe    = regexp.MustCompile(`src="(https:\/\/a-v2\.sndcdn\.com\/assets\/[^\s"]+)"`)
	clientIDRe    = regexp.MustCompile(`client_id:\"([^\"]+)\"`)
	hydrationRe   = regexp.MustCompile(`window\.__sc_hydration\s*=\s*(\[{.*?\}]);`)
	secretPathRe  = regexp.MustCompile(`/s-[A-Za-z0-9]+$`)
	errHydrationN = errors.New("__sc_hydration sound entry not found")
)

// hlsTranscoding is the chosen audio source for a track: its API URL plus the
// quality label SoundCloud advertises so we can report it back to the user.
type hlsTranscoding struct {
	URL     string
	Quality string // "hq" (Go+) or "sq" (standard)
}

// GetClientID scrapes a fresh client_id by following the JS asset bundles
// linked from the SoundCloud homepage.
func (s *Soundcloud) GetClientID() (string, error) {
	resp, err := s.Client.Get("https://soundcloud.com")
	if err != nil {
		return "", err
	}
	body, err := io.ReadAll(resp.Body)
	resp.Body.Close()
	if err != nil {
		return "", fmt.Errorf("read homepage: %w", err)
	}

	matches := assetURLRe.FindAllSubmatch(body, -1)
	if len(matches) == 0 {
		return "", errors.New("no SoundCloud asset URLs found on homepage")
	}

	for _, m := range matches {
		assetResp, err := s.Client.Get(string(m[1]))
		if err != nil {
			continue
		}
		assetBody, err := io.ReadAll(assetResp.Body)
		assetResp.Body.Close()
		if err != nil {
			continue
		}
		if cid := clientIDRe.FindSubmatch(assetBody); len(cid) > 1 {
			return string(cid[1]), nil
		}
	}
	return "", errors.New("client_id not found in any asset bundle")
}

// GetTitle returns the song title from the og:title meta tag.
func (s *Soundcloud) GetTitle(doc *html.Node) (string, error) {
	return metaContent(doc, "og:title")
}

// GetArtwork returns the artwork URL from the og:image meta tag.
func (s *Soundcloud) GetArtwork(doc *html.Node) (string, error) {
	return metaContent(doc, "og:image")
}

// GetHLSTranscoding picks the best plain-HLS MP3 transcoding from the page's
// hydration data. Selection rules:
//
//   - mime_type must be exactly "audio/mpeg" (not "audio/mpegurl" or "audio/mp4")
//   - protocol must be "hls" (not "progressive" or "cbc-encrypted-hls" etc.)
//   - among matches, prefer quality "hq" over "sq" over "lq"
//
// The page only advertises hq when the request is authenticated with a Go+
// token (via the oauth_token cookie on the page fetch).
func (s *Soundcloud) GetHLSTranscoding(doc *html.Node) (hlsTranscoding, error) {
	sound, err := findSoundHydration(doc, "//script[contains(text(), 'media') and contains(text(), 'transcodings')]")
	if err != nil {
		return hlsTranscoding{}, err
	}
	media, _ := sound["media"].(map[string]interface{})
	transcodings, _ := media["transcodings"].([]interface{})

	best := hlsTranscoding{}
	bestScore := -1
	for _, t := range transcodings {
		tm, ok := t.(map[string]interface{})
		if !ok {
			continue
		}
		format, _ := tm["format"].(map[string]interface{})
		if mime, _ := format["mime_type"].(string); mime != "audio/mpeg" {
			continue
		}
		if proto, _ := format["protocol"].(string); proto != "hls" {
			continue
		}
		u, ok := tm["url"].(string)
		if !ok {
			continue
		}
		quality, _ := tm["quality"].(string)
		score := qualityScore(quality)
		if score > bestScore {
			best = hlsTranscoding{URL: u, Quality: quality}
			bestScore = score
		}
	}
	if best.URL == "" {
		return hlsTranscoding{}, errors.New("no plain-HLS audio/mpeg transcoding found")
	}
	return best, nil
}

// qualityScore ranks SoundCloud's quality labels so the picker can prefer hq.
func qualityScore(q string) int {
	switch q {
	case "hq":
		return 2
	case "sq":
		return 1
	case "lq":
		return 0
	default:
		return 0
	}
}

// GetTrackAuthorization extracts the per-track authorization token from the
// page's hydration data.
func (s *Soundcloud) GetTrackAuthorization(doc *html.Node) (string, error) {
	sound, err := findSoundHydration(doc, "//script[contains(text(), 'track_authorization')]")
	if err != nil {
		return "", err
	}
	if ta, ok := sound["track_authorization"].(string); ok {
		return ta, nil
	}
	return "", errors.New("track_authorization not found")
}

// ConstructStreamURL takes the picked transcoding URL from the page and
// appends the query params api-v2 requires to authorize the stream. Crucially,
// it doesn't dissect or rebuild the URL — different protocols use different
// endpoint suffixes (/stream/hls, /stream/progressive, /stream/cbc-encrypted-hls)
// and we preserve whatever SoundCloud advertised.
//
// secretToken, when non-empty, authorizes private share-link tracks.
func (s *Soundcloud) ConstructStreamURL(doc *html.Node, secretToken string) (string, hlsTranscoding, error) {
	clientID, err := s.GetClientID()
	if err != nil {
		return "", hlsTranscoding{}, err
	}
	trackAuth, err := s.GetTrackAuthorization(doc)
	if err != nil {
		return "", hlsTranscoding{}, err
	}
	tr, err := s.GetHLSTranscoding(doc)
	if err != nil {
		return "", hlsTranscoding{}, err
	}

	streamURL := tr.URL + querySep(tr.URL) +
		"client_id=" + url.QueryEscape(clientID) +
		"&track_authorization=" + url.QueryEscape(trackAuth)
	if secretToken != "" {
		streamURL += "&secret_token=" + url.QueryEscape(secretToken)
	}
	return streamURL, tr, nil
}

func querySep(u string) string {
	if strings.Contains(u, "?") {
		return "&"
	}
	return "?"
}

// extractSecretToken returns any share-link secret token embedded in a track
// URL, either as ?secret_token=s-XXX or as a trailing /s-XXX path segment.
// Returns the cleaned URL (with the path-style token stripped) and the token.
func extractSecretToken(rawURL string) (cleaned, secret string, err error) {
	u, err := url.Parse(rawURL)
	if err != nil {
		return "", "", err
	}
	if v := u.Query().Get("secret_token"); v != "" {
		return rawURL, v, nil
	}
	if m := secretPathRe.FindString(u.Path); m != "" {
		secret = strings.TrimPrefix(m, "/")
		u.Path = strings.TrimSuffix(u.Path, m)
		return u.String(), secret, nil
	}
	return rawURL, "", nil
}

// findSoundHydration runs the given XPath, scans matching <script> nodes for
// window.__sc_hydration, and returns the `data` map of the "sound" entry.
func findSoundHydration(doc *html.Node, xpath string) (map[string]interface{}, error) {
	nodes, err := htmlquery.QueryAll(doc, xpath)
	if err != nil {
		return nil, fmt.Errorf("xpath %q: %w", xpath, err)
	}
	for _, node := range nodes {
		match := hydrationRe.FindStringSubmatch(htmlquery.InnerText(node))
		if len(match) <= 1 {
			continue
		}
		var data []map[string]interface{}
		if err := json.Unmarshal([]byte(match[1]), &data); err != nil {
			return nil, fmt.Errorf("parse hydration json: %w", err)
		}
		for _, item := range data {
			if item["hydratable"] != "sound" {
				continue
			}
			if sound, ok := item["data"].(map[string]interface{}); ok {
				return sound, nil
			}
		}
	}
	return nil, errHydrationN
}

// metaContent reads <meta property="…" content="…"> from the parsed doc.
func metaContent(doc *html.Node, property string) (string, error) {
	xpath := fmt.Sprintf("//meta[@property='%s']/@content", property)
	nodes, err := htmlquery.QueryAll(doc, xpath)
	if err != nil {
		return "", err
	}
	if len(nodes) == 0 {
		return "", nil
	}
	return htmlquery.InnerText(nodes[0]), nil
}

