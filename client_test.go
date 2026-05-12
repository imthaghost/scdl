package main

import (
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
)

// roundTripFunc lets us intercept requests without spinning up a real server,
// which is the only way to assert the request URL hits api-v2.soundcloud.com.
type roundTripFunc func(*http.Request) (*http.Response, error)

func (f roundTripFunc) RoundTrip(r *http.Request) (*http.Response, error) { return f(r) }

func TestAuthedGetAttachesOAuthHeaderOnlyForAPIV2(t *testing.T) {
	cases := []struct {
		name      string
		url       string
		token     string
		wantAuth  string
		wantAgent string
	}{
		{
			name:     "api-v2 with token: header attached",
			url:      "https://api-v2.soundcloud.com/media/foo",
			token:    "TEST_TOKEN",
			wantAuth: "OAuth TEST_TOKEN",
		},
		{
			name:     "api-v2 without token: no auth header",
			url:      "https://api-v2.soundcloud.com/media/foo",
			token:    "",
			wantAuth: "",
		},
		{
			name:     "non-api-v2 with token: token NOT leaked to other hosts",
			url:      "https://cdn.example.com/segment.ts",
			token:    "TEST_TOKEN",
			wantAuth: "",
		},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			var seen *http.Request
			httpClient := &http.Client{Transport: roundTripFunc(func(req *http.Request) (*http.Response, error) {
				seen = req
				return &http.Response{StatusCode: 200, Body: http.NoBody, Header: make(http.Header)}, nil
			})}
			s := NewClient(httpClient, tc.token)
			resp, err := s.authedGet(tc.url)
			if err != nil {
				t.Fatalf("authedGet: %v", err)
			}
			resp.Body.Close()

			if got := seen.Header.Get("Authorization"); got != tc.wantAuth {
				t.Errorf("Authorization = %q, want %q", got, tc.wantAuth)
			}
			if got := seen.Header.Get("User-Agent"); got == "" || !strings.HasPrefix(got, "Mozilla/") {
				t.Errorf("User-Agent = %q, want a Mozilla-style UA", got)
			}
		})
	}
}

// TestAuthedGetEndToEnd uses a real httptest server to make sure the request
// actually goes through (no transport oddities) and the body is readable.
func TestAuthedGetEndToEnd(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Write([]byte("ok"))
	}))
	defer srv.Close()

	resp, err := NewClient(http.DefaultClient, "").authedGet(srv.URL)
	if err != nil {
		t.Fatal(err)
	}
	defer resp.Body.Close()
	if resp.StatusCode != 200 {
		t.Errorf("status = %d, want 200", resp.StatusCode)
	}
}
