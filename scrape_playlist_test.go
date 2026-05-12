package main

import "testing"

const playlistFullTracksFixture = `<!DOCTYPE html>
<html><body>
<script>window.__sc_hydration = [{"hydratable":"user","data":{"id":1}},{"hydratable":"playlist","data":{"id":42,"title":"My Set","tracks":[{"id":1001,"permalink_url":"https://soundcloud.com/u/track-1"},{"id":1002,"permalink_url":"https://soundcloud.com/u/track-2"}]}}];</script>
</body></html>`

const playlistStubTracksFixture = `<!DOCTYPE html>
<html><body>
<script>window.__sc_hydration = [{"hydratable":"playlist","data":{"id":43,"title":"Stubby","tracks":[{"id":2001,"permalink_url":"https://soundcloud.com/u/track-a"},{"id":2002},{"id":2003}]}}];</script>
</body></html>`

func TestGetPlaylist(t *testing.T) {
	s := &Soundcloud{}

	t.Run("fully expanded tracks", func(t *testing.T) {
		p, err := s.GetPlaylist(parseFixture(t, playlistFullTracksFixture))
		if err != nil {
			t.Fatal(err)
		}
		if p.Title != "My Set" {
			t.Errorf("title = %q, want %q", p.Title, "My Set")
		}
		if len(p.Tracks) != 2 {
			t.Fatalf("len(tracks) = %d, want 2", len(p.Tracks))
		}
		if p.Tracks[0].PermalinkURL != "https://soundcloud.com/u/track-1" {
			t.Errorf("track[0] = %q", p.Tracks[0].PermalinkURL)
		}
		if p.Tracks[1].ID != 1002 {
			t.Errorf("track[1].id = %d, want 1002", p.Tracks[1].ID)
		}
	})

	t.Run("mix of expanded and stub tracks", func(t *testing.T) {
		p, err := s.GetPlaylist(parseFixture(t, playlistStubTracksFixture))
		if err != nil {
			t.Fatal(err)
		}
		if len(p.Tracks) != 3 {
			t.Fatalf("len(tracks) = %d, want 3", len(p.Tracks))
		}
		// First track has permalink, others are stubs (id-only).
		if p.Tracks[0].PermalinkURL == "" {
			t.Error("track[0] should have permalink")
		}
		if p.Tracks[1].PermalinkURL != "" {
			t.Errorf("track[1] should be a stub, got permalink %q", p.Tracks[1].PermalinkURL)
		}
		if p.Tracks[1].ID != 2002 {
			t.Errorf("track[1].id = %d, want 2002", p.Tracks[1].ID)
		}
	})

	t.Run("error when page isn't a playlist", func(t *testing.T) {
		if _, err := s.GetPlaylist(parseFixture(t, trackHTMLFixture)); err == nil {
			t.Error("expected errPlaylistNotFound on a track page, got nil")
		}
	})
}

func TestIsPlaylistURL(t *testing.T) {
	cases := []struct {
		url  string
		want bool
	}{
		{"https://soundcloud.com/user/track", false},
		{"https://soundcloud.com/user/sets/my-playlist", true},
		{"https://soundcloud.com/discover/sets/personalized-tracks::user-1:2", true},
		{"https://soundcloud.com/user/track?secret_token=s-abc", false},
		{"https://soundcloud.com/user/sets/my-playlist?some=query", true},
	}
	for _, tc := range cases {
		if got := isPlaylistURL(tc.url); got != tc.want {
			t.Errorf("isPlaylistURL(%q) = %v, want %v", tc.url, got, tc.want)
		}
	}
}
