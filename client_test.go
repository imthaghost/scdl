package main

import (
	"context"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"
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
			s, err := NewClient(Options{Token: tc.token})
			if err != nil {
				t.Fatal(err)
			}
			s.Client.Transport = roundTripFunc(func(req *http.Request) (*http.Response, error) {
				seen = req
				return &http.Response{StatusCode: 200, Body: http.NoBody, Header: make(http.Header)}, nil
			})
			resp, err := s.authedGet(context.Background(), tc.url)
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

func TestNewClientProxyURL(t *testing.T) {
	t.Run("valid proxy URL", func(t *testing.T) {
		s, err := NewClient(Options{ProxyURL: "http://proxy.example.com:3128"})
		if err != nil {
			t.Fatal(err)
		}
		tr, ok := s.Client.Transport.(*http.Transport)
		if !ok {
			t.Fatalf("transport = %T, want *http.Transport", s.Client.Transport)
		}
		if tr.Proxy == nil {
			t.Fatal("transport.Proxy is nil; expected ProxyURL set")
		}
		// Probe the proxy func with a dummy request.
		req, _ := http.NewRequest("GET", "https://soundcloud.com", nil)
		u, err := tr.Proxy(req)
		if err != nil {
			t.Fatal(err)
		}
		if u == nil || u.Host != "proxy.example.com:3128" {
			t.Errorf("proxy host = %v, want proxy.example.com:3128", u)
		}
	})

	t.Run("empty proxy URL falls back to ProxyFromEnvironment", func(t *testing.T) {
		s, err := NewClient(Options{})
		if err != nil {
			t.Fatal(err)
		}
		tr := s.Client.Transport.(*http.Transport)
		if tr.Proxy == nil {
			t.Error("transport.Proxy should default to ProxyFromEnvironment, not nil")
		}
	})

	t.Run("malformed proxy URL surfaces error", func(t *testing.T) {
		if _, err := NewClient(Options{ProxyURL: "://not a url"}); err == nil {
			t.Error("expected error for malformed proxy, got nil")
		}
	})
}

func TestNewClientStoresOutputDir(t *testing.T) {
	s, err := NewClient(Options{OutputDir: "/tmp/scdl-out"})
	if err != nil {
		t.Fatal(err)
	}
	if s.OutputDir != "/tmp/scdl-out" {
		t.Errorf("OutputDir = %q, want %q", s.OutputDir, "/tmp/scdl-out")
	}
}

// TestAuthedGetEndToEnd uses a real httptest server to make sure the request
// actually goes through (no transport oddities) and the body is readable.
func TestAuthedGetEndToEnd(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		_, _ = w.Write([]byte("ok"))
	}))
	defer srv.Close()

	sc, err := NewClient(Options{})
	if err != nil {
		t.Fatal(err)
	}
	resp, err := sc.authedGet(context.Background(), srv.URL)
	if err != nil {
		t.Fatal(err)
	}
	defer resp.Body.Close()
	if resp.StatusCode != 200 {
		t.Errorf("status = %d, want 200", resp.StatusCode)
	}
}

// TestAuthedGetCancellation verifies that cancelling the context aborts an
// in-flight request rather than hanging.
func TestAuthedGetCancellation(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// Block until the client disconnects (i.e. ctx is cancelled).
		<-r.Context().Done()
	}))
	defer srv.Close()

	sc, err := NewClient(Options{})
	if err != nil {
		t.Fatal(err)
	}

	ctx, cancel := context.WithCancel(context.Background())
	errCh := make(chan error, 1)
	go func() {
		_, err := sc.authedGet(ctx, srv.URL)
		errCh <- err
	}()

	time.Sleep(20 * time.Millisecond)
	cancel()

	select {
	case err := <-errCh:
		if err == nil {
			t.Fatal("expected error after cancel, got nil")
		}
		if !strings.Contains(err.Error(), "context canceled") {
			t.Errorf("error = %q, want one mentioning context canceled", err)
		}
	case <-time.After(2 * time.Second):
		t.Fatal("authedGet did not return after cancel; request did not respect ctx")
	}
}
