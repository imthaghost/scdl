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
	assetURLRe   = regexp.MustCompile(`src="(https:\/\/a-v2\.sndcdn\.com\/assets\/[^\s"]+)"`)
	clientIDRe   = regexp.MustCompile(`client_id:\"([^\"]+)\"`)
	hydrationRe  = regexp.MustCompile(`window\.__sc_hydration\s*=\s*(\[{.*?\}]);`)
	secretPathRe = regexp.MustCompile(`/s-[A-Za-z0-9]+$`)

	errSoundNotFound    = errors.New("__sc_hydration sound entry not found")
	errPlaylistNotFound = errors.New("__sc_hydration playlist entry not found")
)

// hlsTranscoding is the chosen audio source for a track: its API URL plus the
// quality label SoundCloud advertises so we can report it back to the user.
type hlsTranscoding struct {
	URL     string
	Quality string // "hq" (Go+), "sq" (standard), or "lq"
}

// playlistInfo is the subset of playlist hydration we care about: a title for
// the output subdirectory and a list of constituent tracks.
type playlistInfo struct {
	Title  string
	Tracks []playlistTrack
}

// playlistTrack is one entry in a playlist. PermalinkURL may be empty when the
// hydration only included the track ID (a "stub"); call resolveStubs to fill
// those in via api-v2.
type playlistTrack struct {
	ID           int64
	PermalinkURL string
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
	sound, err := findHydration(doc, "//script[contains(text(), 'media') and contains(text(), 'transcodings')]", "sound")
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
	sound, err := findHydration(doc, "//script[contains(text(), 'track_authorization')]", "sound")
	if err != nil {
		return "", err
	}
	if ta, ok := sound["track_authorization"].(string); ok {
		return ta, nil
	}
	return "", errors.New("track_authorization not found")
}

// GetPlaylist extracts the playlist title and constituent track URLs from the
// page's hydration data. Returns errPlaylistNotFound when the page isn't a
// playlist (e.g., a single track URL).
func (s *Soundcloud) GetPlaylist(doc *html.Node) (*playlistInfo, error) {
	data, err := findHydration(doc, "//script[contains(text(), '__sc_hydration')]", "playlist")
	if err != nil {
		return nil, err
	}
	title, _ := data["title"].(string)
	rawTracks, _ := data["tracks"].([]interface{})
	tracks := make([]playlistTrack, 0, len(rawTracks))
	for _, t := range rawTracks {
		tm, ok := t.(map[string]interface{})
		if !ok {
			continue
		}
		var id int64
		switch v := tm["id"].(type) {
		case float64:
			id = int64(v)
		case json.Number:
			id, _ = v.Int64()
		}
		permalink, _ := tm["permalink_url"].(string)
		tracks = append(tracks, playlistTrack{ID: id, PermalinkURL: permalink})
	}
	return &playlistInfo{Title: title, Tracks: tracks}, nil
}

// stubBatchSize caps how many track IDs we send per api-v2 /tracks request.
// Empirically SoundCloud rejects much larger batches with 400.
const stubBatchSize = 50

// resolveStubs replaces any tracks in p whose PermalinkURL is empty with
// fully-resolved entries from api-v2 /tracks?ids=. Stubs occur when the
// playlist hydration only included track IDs. Requests are batched to stay
// under URL-length limits.
func (s *Soundcloud) resolveStubs(p *playlistInfo, clientID string) error {
	var stubIDs []string
	for _, t := range p.Tracks {
		if t.PermalinkURL == "" && t.ID != 0 {
			stubIDs = append(stubIDs, fmt.Sprintf("%d", t.ID))
		}
	}
	if len(stubIDs) == 0 {
		return nil
	}

	byID := make(map[int64]string)
	for i := 0; i < len(stubIDs); i += stubBatchSize {
		end := i + stubBatchSize
		if end > len(stubIDs) {
			end = len(stubIDs)
		}
		if err := s.fetchTracksBatch(stubIDs[i:end], clientID, byID); err != nil {
			return err
		}
	}
	for i := range p.Tracks {
		if p.Tracks[i].PermalinkURL == "" {
			p.Tracks[i].PermalinkURL = byID[p.Tracks[i].ID]
		}
	}
	return nil
}

func (s *Soundcloud) fetchTracksBatch(ids []string, clientID string, out map[int64]string) error {
	apiURL := fmt.Sprintf("https://api-v2.soundcloud.com/tracks?ids=%s&client_id=%s",
		strings.Join(ids, ","), url.QueryEscape(clientID))
	resp, err := s.authedGet(apiURL)
	if err != nil {
		return err
	}
	defer resp.Body.Close()
	if resp.StatusCode/100 != 2 {
		return fmt.Errorf("api-v2 /tracks returned %d (batch of %d ids)", resp.StatusCode, len(ids))
	}
	var resolved []struct {
		ID           int64  `json:"id"`
		PermalinkURL string `json:"permalink_url"`
	}
	if err := json.NewDecoder(resp.Body).Decode(&resolved); err != nil {
		return fmt.Errorf("decode /tracks response: %w", err)
	}
	for _, r := range resolved {
		out[r.ID] = r.PermalinkURL
	}
	return nil
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

// isPlaylistURL is a fast check used to route Download. SoundCloud playlist
// URLs always contain /sets/ in the path (both user playlists and the
// algorithmically-curated /discover/sets/...).
func isPlaylistURL(rawURL string) bool {
	u, err := url.Parse(rawURL)
	if err != nil {
		return false
	}
	return strings.Contains(u.Path, "/sets/")
}

// findHydration scans <script> nodes matching xpath for the
// window.__sc_hydration array and returns the data map of the first entry
// whose "hydratable" equals the given type.
func findHydration(doc *html.Node, xpath, hydratable string) (map[string]interface{}, error) {
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
			if item["hydratable"] != hydratable {
				continue
			}
			if d, ok := item["data"].(map[string]interface{}); ok {
				return d, nil
			}
		}
	}
	switch hydratable {
	case "sound":
		return nil, errSoundNotFound
	case "playlist":
		return nil, errPlaylistNotFound
	default:
		return nil, fmt.Errorf("__sc_hydration %q entry not found", hydratable)
	}
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
