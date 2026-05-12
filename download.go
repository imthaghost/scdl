package main

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"os"
	"path/filepath"
	"strings"

	"github.com/antchfx/htmlquery"
	"golang.org/x/net/html"
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

// Download fetches a SoundCloud URL and routes to the right downloader based
// on whether the URL is a track or a playlist/set.
func (s *Soundcloud) Download(ctx context.Context, rawURL string) error {
	if isPlaylistURL(rawURL) {
		return s.downloadPlaylist(ctx, rawURL, s.OutputDir)
	}
	return s.downloadTrack(ctx, rawURL, s.OutputDir)
}

// downloadTrack downloads a single track URL into baseDir. Private share links
// (with ?secret_token=... or /s-XXX path suffix) are supported.
func (s *Soundcloud) downloadTrack(ctx context.Context, rawURL, baseDir string) error {
	pageURL, secretToken, err := extractSecretToken(rawURL)
	if err != nil {
		return err
	}

	doc, err := s.fetchPage(ctx, pageURL)
	if err != nil {
		return err
	}

	streamURL, transcoding, err := s.ConstructStreamURL(ctx, doc, secretToken)
	if err != nil {
		return err
	}

	title, err := s.GetTitle(doc)
	if err != nil {
		return err
	}
	if title == "" {
		return errors.New("track title not found on page")
	}

	artworkURL, _ := s.GetArtwork(doc)
	playlistURL, err := s.fetchPlaylistURL(ctx, streamURL)
	if err != nil {
		return err
	}

	outPath := filepath.Join(baseDir, filenameSanitizer.Replace(title)+".mp3")
	if err := os.MkdirAll(filepath.Dir(outPath), 0o755); err != nil {
		return fmt.Errorf("create output dir: %w", err)
	}

	fmt.Printf("%s Quality: %s\n", green("[+]"), qualityLabel(transcoding.Quality))
	if err := s.downloadHLS(ctx, playlistURL, outPath); err != nil {
		return err
	}

	if artworkURL != "" {
		image, err := s.fetchBytes(ctx, artworkURL)
		if err != nil {
			fmt.Printf("%s Cover fetch failed (track downloaded without art): %s\n", red("[-]"), err)
			return nil
		}
		if err := SetCoverImage(outPath, image); err != nil {
			fmt.Printf("%s Embed cover failed (track downloaded without art): %s\n", red("[-]"), err)
		}
	}
	return nil
}

// downloadPlaylist iterates the tracks of a /sets/ URL, downloading each into
// a subdirectory under baseDir named after the playlist. Per-track errors are
// logged and collected; the playlist completes whatever it can, then returns a
// single error summarizing failures. Context cancellation aborts the loop
// early.
func (s *Soundcloud) downloadPlaylist(ctx context.Context, rawURL, baseDir string) error {
	doc, err := s.fetchPage(ctx, rawURL)
	if err != nil {
		return err
	}

	p, err := s.GetPlaylist(doc)
	if err != nil {
		return err
	}
	if len(p.Tracks) == 0 {
		return errors.New("playlist contains no tracks")
	}

	if err := s.resolveStubs(ctx, p); err != nil {
		fmt.Printf("%s Resolve stub tracks failed (some tracks may be skipped): %s\n", red("[-]"), err)
	}

	playlistDir := filepath.Join(baseDir, filenameSanitizer.Replace(p.Title))
	if err := os.MkdirAll(playlistDir, 0o755); err != nil {
		return fmt.Errorf("create playlist dir: %w", err)
	}

	fmt.Printf("%s Playlist %q: %d tracks → %s\n", green("[+]"), p.Title, len(p.Tracks), playlistDir)

	var failed []string
	for i, t := range p.Tracks {
		if err := ctx.Err(); err != nil {
			return err
		}
		if t.PermalinkURL == "" {
			fmt.Printf("%s [%d/%d] Skipping unresolvable track (id=%d)\n", red("[-]"), i+1, len(p.Tracks), t.ID)
			failed = append(failed, fmt.Sprintf("id=%d", t.ID))
			continue
		}
		fmt.Printf("%s [%d/%d] %s\n", green("[+]"), i+1, len(p.Tracks), t.PermalinkURL)
		if err := s.downloadTrack(ctx, t.PermalinkURL, playlistDir); err != nil {
			if errors.Is(err, context.Canceled) || errors.Is(err, context.DeadlineExceeded) {
				return err
			}
			fmt.Printf("%s [%d/%d] failed: %s\n", red("[-]"), i+1, len(p.Tracks), err)
			failed = append(failed, t.PermalinkURL)
		}
	}
	if len(failed) > 0 {
		return fmt.Errorf("%d/%d tracks failed: %s", len(failed), len(p.Tracks), strings.Join(failed, ", "))
	}
	return nil
}

// fetchPage GETs the track or playlist page and parses it into an html.Node.
// When a token is configured it's sent as the oauth_token cookie so the
// scraped track_authorization JWT carries the user's sub.
func (s *Soundcloud) fetchPage(ctx context.Context, pageURL string) (*html.Node, error) {
	req, err := http.NewRequestWithContext(ctx, "GET", pageURL, nil)
	if err != nil {
		return nil, err
	}
	req.Header.Set("User-Agent", s.UserAgent)
	if s.Token != "" {
		req.AddCookie(&http.Cookie{Name: "oauth_token", Value: s.Token})
	}
	resp, err := s.Client.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()
	if resp.StatusCode/100 != 2 {
		return nil, fmt.Errorf("page request returned status %d", resp.StatusCode)
	}
	return htmlquery.Parse(resp.Body)
}

// fetchPlaylistURL hits the stream endpoint and unwraps the {"url": "..."}
// response that points at the actual HLS playlist. The request is authed
// because the endpoint lives on api-v2.soundcloud.com.
func (s *Soundcloud) fetchPlaylistURL(ctx context.Context, streamURL string) (string, error) {
	resp, err := s.authedGet(ctx, streamURL)
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
		return "", errors.New("stream response had empty URL")
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
	case "lq":
		return "lq (preview)"
	case "":
		return "unknown"
	default:
		return q
	}
}

func (s *Soundcloud) fetchBytes(ctx context.Context, rawURL string) ([]byte, error) {
	req, err := http.NewRequestWithContext(ctx, "GET", rawURL, nil)
	if err != nil {
		return nil, err
	}
	resp, err := s.Client.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()
	if resp.StatusCode/100 != 2 {
		return nil, fmt.Errorf("status %d fetching %s", resp.StatusCode, rawURL)
	}
	return io.ReadAll(resp.Body)
}
