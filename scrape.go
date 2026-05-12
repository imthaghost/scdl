package main

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
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

// hydrationEntry is one item in the window.__sc_hydration array. Data is
// kept as RawMessage so we can decode it into a type-specific struct once we
// know which "hydratable" we want.
type hydrationEntry struct {
	Hydratable string          `json:"hydratable"`
	Data       json.RawMessage `json:"data"`
}

// soundData is the relevant subset of a "hydratable=sound" entry's data.
type soundData struct {
	Title              string     `json:"title"`
	TrackAuthorization string     `json:"track_authorization"`
	Media              soundMedia `json:"media"`
}

type soundMedia struct {
	Transcodings []transcoding `json:"transcodings"`
}

type transcoding struct {
	URL     string            `json:"url"`
	Quality string            `json:"quality"`
	Format  transcodingFormat `json:"format"`
}

type transcodingFormat struct {
	MimeType string `json:"mime_type"`
	Protocol string `json:"protocol"`
}

// hlsTranscoding is the chosen audio source for a track: its API URL plus the
// quality label SoundCloud advertises so we can report it back to the user.
type hlsTranscoding struct {
	URL     string
	Quality string // "hq" (Go+), "sq" (standard), or "lq"
}

// playlistInfo is the relevant subset of a "hydratable=playlist" entry's data.
type playlistInfo struct {
	Title  string          `json:"title"`
	Tracks []playlistTrack `json:"tracks"`
}

// playlistTrack is one entry in a playlist. PermalinkURL may be empty when the
// hydration only included the track ID (a "stub"); resolveStubs fills those
// in via api-v2.
type playlistTrack struct {
	ID           int64  `json:"id"`
	PermalinkURL string `json:"permalink_url"`
}

// GetClientID returns a cached SoundCloud client_id, scraping it on first use
// by following the JS asset bundles linked from the homepage.
func (s *Soundcloud) GetClientID(ctx context.Context) (string, error) {
	s.clientIDOnce.Do(func() {
		s.clientIDVal, s.clientIDErr = s.fetchClientID(ctx)
	})
	return s.clientIDVal, s.clientIDErr
}

func (s *Soundcloud) fetchClientID(ctx context.Context) (string, error) {
	homepageBody, err := s.getBody(ctx, "https://soundcloud.com")
	if err != nil {
		return "", fmt.Errorf("fetch homepage: %w", err)
	}
	matches := assetURLRe.FindAllSubmatch(homepageBody, -1)
	if len(matches) == 0 {
		return "", errors.New("no SoundCloud asset URLs found on homepage")
	}
	for _, m := range matches {
		assetBody, err := s.getBody(ctx, string(m[1]))
		if err != nil {
			if ctx.Err() != nil {
				return "", ctx.Err()
			}
			continue
		}
		if cid := clientIDRe.FindSubmatch(assetBody); len(cid) > 1 {
			return string(cid[1]), nil
		}
	}
	return "", errors.New("client_id not found in any asset bundle")
}

// getBody is a small helper for ctx-aware GETs that consume the entire body.
func (s *Soundcloud) getBody(ctx context.Context, rawURL string) ([]byte, error) {
	req, err := http.NewRequestWithContext(ctx, "GET", rawURL, nil)
	if err != nil {
		return nil, err
	}
	req.Header.Set("User-Agent", s.UserAgent)
	resp, err := s.Client.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()
	if resp.StatusCode/100 != 2 {
		return nil, fmt.Errorf("status %d", resp.StatusCode)
	}
	return io.ReadAll(resp.Body)
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
	var sd soundData
	if err := findHydration(doc, "//script[contains(text(), 'media') and contains(text(), 'transcodings')]", "sound", &sd); err != nil {
		return hlsTranscoding{}, err
	}
	best := hlsTranscoding{}
	bestScore := -1
	for _, t := range sd.Media.Transcodings {
		if t.Format.MimeType != "audio/mpeg" || t.Format.Protocol != "hls" || t.URL == "" {
			continue
		}
		if score := qualityScore(t.Quality); score > bestScore {
			best = hlsTranscoding{URL: t.URL, Quality: t.Quality}
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
	var sd soundData
	if err := findHydration(doc, "//script[contains(text(), 'track_authorization')]", "sound", &sd); err != nil {
		return "", err
	}
	if sd.TrackAuthorization == "" {
		return "", errors.New("track_authorization not found")
	}
	return sd.TrackAuthorization, nil
}

// GetPlaylist extracts the playlist title and constituent track URLs from the
// page's hydration data. Returns errPlaylistNotFound when the page isn't a
// playlist (e.g., a single track URL).
func (s *Soundcloud) GetPlaylist(doc *html.Node) (*playlistInfo, error) {
	var p playlistInfo
	if err := findHydration(doc, "//script[contains(text(), '__sc_hydration')]", "playlist", &p); err != nil {
		return nil, err
	}
	return &p, nil
}

// stubBatchSize caps how many track IDs we send per api-v2 /tracks request.
// Empirically SoundCloud rejects much larger batches with 400.
const stubBatchSize = 50

// resolveStubs replaces any tracks in p whose PermalinkURL is empty with
// fully-resolved entries from api-v2 /tracks?ids=. Stubs occur when the
// playlist hydration only included track IDs. Requests are batched to stay
// under URL-length limits. Skips the api-v2 round-trip entirely when there
// are no stubs to resolve.
func (s *Soundcloud) resolveStubs(ctx context.Context, p *playlistInfo) error {
	var stubIDs []string
	for _, t := range p.Tracks {
		if t.PermalinkURL == "" && t.ID != 0 {
			stubIDs = append(stubIDs, fmt.Sprintf("%d", t.ID))
		}
	}
	if len(stubIDs) == 0 {
		return nil
	}

	clientID, err := s.GetClientID(ctx)
	if err != nil {
		return fmt.Errorf("get client_id for stub resolution: %w", err)
	}

	byID := make(map[int64]string)
	for i := 0; i < len(stubIDs); i += stubBatchSize {
		end := i + stubBatchSize
		if end > len(stubIDs) {
			end = len(stubIDs)
		}
		if err := s.fetchTracksBatch(ctx, stubIDs[i:end], clientID, byID); err != nil {
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

func (s *Soundcloud) fetchTracksBatch(ctx context.Context, ids []string, clientID string, out map[int64]string) error {
	apiURL := fmt.Sprintf("https://api-v2.soundcloud.com/tracks?ids=%s&client_id=%s",
		strings.Join(ids, ","), url.QueryEscape(clientID))
	resp, err := s.authedGet(ctx, apiURL)
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
func (s *Soundcloud) ConstructStreamURL(ctx context.Context, doc *html.Node, secretToken string) (string, hlsTranscoding, error) {
	clientID, err := s.GetClientID(ctx)
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
// window.__sc_hydration array, locates the first entry whose "hydratable"
// equals the given type, and unmarshals its "data" field into out.
func findHydration(doc *html.Node, xpath, hydratable string, out interface{}) error {
	nodes, err := htmlquery.QueryAll(doc, xpath)
	if err != nil {
		return fmt.Errorf("xpath %q: %w", xpath, err)
	}
	for _, node := range nodes {
		match := hydrationRe.FindStringSubmatch(htmlquery.InnerText(node))
		if len(match) <= 1 {
			continue
		}
		var entries []hydrationEntry
		if err := json.Unmarshal([]byte(match[1]), &entries); err != nil {
			return fmt.Errorf("parse hydration json: %w", err)
		}
		for _, e := range entries {
			if e.Hydratable != hydratable {
				continue
			}
			if err := json.Unmarshal(e.Data, out); err != nil {
				return fmt.Errorf("unmarshal %s data: %w", hydratable, err)
			}
			return nil
		}
	}
	switch hydratable {
	case "sound":
		return errSoundNotFound
	case "playlist":
		return errPlaylistNotFound
	default:
		return fmt.Errorf("__sc_hydration %q entry not found", hydratable)
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
