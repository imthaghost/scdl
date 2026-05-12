package main

import (
	"context"
	"strings"
	"testing"

	"github.com/antchfx/htmlquery"
	"golang.org/x/net/html"
)

const trackHTMLFixture = `<!DOCTYPE html>
<html>
<head>
  <meta property="og:title" content="My Test Song">
  <meta property="og:image" content="https://example.com/cover.jpg">
</head>
<body>
  <script>window.__sc_hydration = [{"hydratable":"user","data":{"id":1}},{"hydratable":"sound","data":{"track_authorization":"AUTH_TOKEN_123","media":{"transcodings":[{"format":{"mime_type":"audio/ogg; codecs=\"opus\"","protocol":"hls"},"quality":"sq","url":"https://example.com/opus/playlist"},{"format":{"mime_type":"audio/mpeg","protocol":"hls"},"quality":"sq","url":"https://api-v2.soundcloud.com/media/soundcloud:tracks:111/abc/stream/hls"}]}}}];</script>
</body>
</html>`

// realisticAuthedHTMLFixture mirrors what SoundCloud actually returns for a
// major-label authed track: a mix of mp4/AAC, audio/mpegurl (encrypted), and a
// single plain-HLS audio/mpeg entry. The picker should ignore everything but
// the last. Single-line JSON because the live page's hydration is single-line
// and the regex doesn't span newlines.
const realisticAuthedHTMLFixture = `<!DOCTYPE html>
<html>
<head><meta property="og:title" content="Real Track"></head>
<body>
  <script>window.__sc_hydration = [{"hydratable":"sound","data":{"track_authorization":"AUTH","media":{"transcodings":[{"format":{"mime_type":"audio/mp4; codecs=\"mp4a.40.2\"","protocol":"cbc-encrypted-hls"},"quality":"sq","url":"https://api-v2.soundcloud.com/media/x/aac-cbc/stream/cbc-encrypted-hls"},{"format":{"mime_type":"audio/mpegurl","protocol":"cbc-encrypted-hls"},"quality":"sq","url":"https://api-v2.soundcloud.com/media/x/mpegurl-cbc/stream/cbc-encrypted-hls"},{"format":{"mime_type":"audio/mpegurl","protocol":"ctr-encrypted-hls"},"quality":"sq","url":"https://api-v2.soundcloud.com/media/x/mpegurl-ctr/stream/ctr-encrypted-hls"},{"format":{"mime_type":"audio/mpeg","protocol":"hls"},"quality":"sq","url":"https://api-v2.soundcloud.com/media/x/mpeg-hls/stream/hls"},{"format":{"mime_type":"audio/mpeg","protocol":"progressive"},"quality":"sq","url":"https://api-v2.soundcloud.com/media/x/mpeg-prog/stream/progressive"}]}}}];</script>
</body>
</html>`

const goPlusTrackHTMLFixture = `<!DOCTYPE html>
<html>
<head><meta property="og:title" content="HQ Track"></head>
<body>
  <script>window.__sc_hydration = [{"hydratable":"sound","data":{"track_authorization":"AUTH","media":{"transcodings":[{"format":{"mime_type":"audio/mpeg","protocol":"hls"},"quality":"sq","url":"https://api-v2.soundcloud.com/media/soundcloud:tracks:222/sq/stream/hls"},{"format":{"mime_type":"audio/mpeg","protocol":"hls"},"quality":"hq","url":"https://api-v2.soundcloud.com/media/soundcloud:tracks:222/hq/stream/hls"}]}}}];</script>
</body>
</html>`

const playlistHTMLFixture = `<!DOCTYPE html>
<html>
<head><meta property="og:title" content="Some Playlist"></head>
<body>
  <script>window.__sc_hydration = [{"hydratable":"playlist","data":{"id":42,"tracks":[]}}];</script>
</body>
</html>`

func parseFixture(t *testing.T, source string) *html.Node {
	t.Helper()
	doc, err := htmlquery.Parse(strings.NewReader(source))
	if err != nil {
		t.Fatalf("parse fixture: %v", err)
	}
	return doc
}

func TestMetaContent(t *testing.T) {
	doc := parseFixture(t, trackHTMLFixture)

	got, err := metaContent(doc, "og:title")
	if err != nil {
		t.Fatal(err)
	}
	if got != "My Test Song" {
		t.Errorf("og:title = %q, want %q", got, "My Test Song")
	}

	missing, err := metaContent(doc, "og:does-not-exist")
	if err != nil {
		t.Fatal(err)
	}
	if missing != "" {
		t.Errorf("missing meta = %q, want empty", missing)
	}
}

func TestFindHydration(t *testing.T) {
	doc := parseFixture(t, trackHTMLFixture)
	var sd soundData
	if err := findHydration(doc, "//script[contains(text(), 'track_authorization')]", "sound", &sd); err != nil {
		t.Fatalf("findHydration sound: %v", err)
	}
	if sd.TrackAuthorization != "AUTH_TOKEN_123" {
		t.Errorf("track_authorization = %q, want %q", sd.TrackAuthorization, "AUTH_TOKEN_123")
	}

	var sd2 soundData
	if err := findHydration(parseFixture(t, playlistHTMLFixture), "//script[contains(text(), 'hydration')]", "sound", &sd2); err == nil {
		t.Error("expected error when no sound entry is present, got nil")
	}
}

func TestGetHLSTranscoding(t *testing.T) {
	s := &Soundcloud{}

	t.Run("sq only picks the audio/mpeg sq", func(t *testing.T) {
		got, err := s.GetHLSTranscoding(parseFixture(t, trackHTMLFixture))
		if err != nil {
			t.Fatal(err)
		}
		want := "https://api-v2.soundcloud.com/media/soundcloud:tracks:111/abc/stream/hls"
		if got.URL != want {
			t.Errorf("url = %q, want %q", got.URL, want)
		}
		if got.Quality != "sq" {
			t.Errorf("quality = %q, want sq", got.Quality)
		}
	})

	t.Run("prefers hq when present", func(t *testing.T) {
		got, err := s.GetHLSTranscoding(parseFixture(t, goPlusTrackHTMLFixture))
		if err != nil {
			t.Fatal(err)
		}
		if !strings.HasSuffix(got.URL, "/hq/stream/hls") {
			t.Errorf("expected hq URL, got %q", got.URL)
		}
		if got.Quality != "hq" {
			t.Errorf("quality = %q, want hq", got.Quality)
		}
	})

	t.Run("error when no audio/mpeg transcoding", func(t *testing.T) {
		if _, err := s.GetHLSTranscoding(parseFixture(t, playlistHTMLFixture)); err == nil {
			t.Error("expected error on playlist fixture, got nil")
		}
	})

	t.Run("skips audio/mpegurl and progressive, picks only audio/mpeg+hls", func(t *testing.T) {
		got, err := s.GetHLSTranscoding(parseFixture(t, realisticAuthedHTMLFixture))
		if err != nil {
			t.Fatal(err)
		}
		want := "https://api-v2.soundcloud.com/media/x/mpeg-hls/stream/hls"
		if got.URL != want {
			t.Errorf("url = %q, want %q (must skip mpegurl and progressive)", got.URL, want)
		}
	})
}

func TestExtractSecretToken(t *testing.T) {
	tests := []struct {
		name        string
		in          string
		wantCleaned string
		wantSecret  string
	}{
		{
			name:        "public URL has no secret",
			in:          "https://soundcloud.com/user/track",
			wantCleaned: "https://soundcloud.com/user/track",
			wantSecret:  "",
		},
		{
			name:        "query string form",
			in:          "https://soundcloud.com/user/track?secret_token=s-abc123",
			wantCleaned: "https://soundcloud.com/user/track?secret_token=s-abc123",
			wantSecret:  "s-abc123",
		},
		{
			name:        "path suffix form is stripped from page URL",
			in:          "https://soundcloud.com/user/track/s-XYZ789",
			wantCleaned: "https://soundcloud.com/user/track",
			wantSecret:  "s-XYZ789",
		},
		{
			name:        "trailing s- look-alike without dash is ignored",
			in:          "https://soundcloud.com/user/some-song",
			wantCleaned: "https://soundcloud.com/user/some-song",
			wantSecret:  "",
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			cleaned, secret, err := extractSecretToken(tt.in)
			if err != nil {
				t.Fatal(err)
			}
			if cleaned != tt.wantCleaned {
				t.Errorf("cleaned = %q, want %q", cleaned, tt.wantCleaned)
			}
			if secret != tt.wantSecret {
				t.Errorf("secret = %q, want %q", secret, tt.wantSecret)
			}
		})
	}
}

func TestGetTrackAuthorization(t *testing.T) {
	s := &Soundcloud{}
	got, err := s.GetTrackAuthorization(parseFixture(t, trackHTMLFixture))
	if err != nil {
		t.Fatalf("GetTrackAuthorization: %v", err)
	}
	if got != "AUTH_TOKEN_123" {
		t.Errorf("track auth = %q, want %q", got, "AUTH_TOKEN_123")
	}

	if _, err := s.GetTrackAuthorization(parseFixture(t, playlistHTMLFixture)); err == nil {
		t.Error("expected error on playlist fixture, got nil")
	}
}

func TestFilenameSanitizer(t *testing.T) {
	tests := []struct {
		in, want string
	}{
		{"clean title", "clean title"},
		{`weird/title:with*chars?`, "weirdtitlewithchars"},
		{"Polo G - Flex (feat. Juice WRLD)", "Polo G - Flex feat Juice WRLD"},
		{"hello\\world", "helloworld"},
	}
	for _, tt := range tests {
		if got := filenameSanitizer.Replace(tt.in); got != tt.want {
			t.Errorf("Replace(%q) = %q, want %q", tt.in, got, tt.want)
		}
	}
}

// TestGetClientIDIsCached locks in the sync.Once contract: the underlying
// fetch runs at most once per Soundcloud value, no matter how many tracks
// call GetClientID. We can't easily mock the homepage URL (it's hardcoded),
// so we seed the cache via Once.Do directly and verify subsequent calls
// don't re-invoke.
func TestGetClientIDIsCached(t *testing.T) {
	sc, err := NewClient(Options{})
	if err != nil {
		t.Fatal(err)
	}

	const fakeClientID = "TEST_CID_12345"
	calls := 0
	sc.clientIDOnce.Do(func() {
		calls++
		sc.clientIDVal = fakeClientID
	})

	for i := 0; i < 5; i++ {
		got, err := sc.GetClientID(context.Background())
		if err != nil {
			t.Fatalf("call %d: %v", i, err)
		}
		if got != fakeClientID {
			t.Errorf("call %d: clientID = %q, want %q", i, got, fakeClientID)
		}
	}
	if calls != 1 {
		t.Errorf("inner fetcher invoked %d times, want 1", calls)
	}
}
