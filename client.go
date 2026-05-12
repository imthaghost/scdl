package main

import (
	"fmt"
	"net/http"
	"net/url"
	"strings"
)

const (
	userAgent       = "Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/58.0.3029.110 Safari/537.3"
	apiV2Host       = "api-v2.soundcloud.com"
	authHeader      = "Authorization"
	authHeaderValue = "OAuth "
)

// Options bundles user-facing configuration for a Soundcloud client.
type Options struct {
	// Token is an optional SoundCloud OAuth token. When set, it's attached as
	// "Authorization: OAuth <token>" on api-v2 requests and as the oauth_token
	// cookie on page requests so the scraped track_authorization JWT carries
	// the user's sub (required for Go+ HQ and private tracks).
	Token string

	// OutputDir is the directory where downloaded mp3s are written. Empty
	// means the current working directory.
	OutputDir string

	// ProxyURL, when set, is parsed and used as the HTTP/HTTPS proxy for all
	// requests. Empty falls back to Go's default ProxyFromEnvironment behavior
	// (HTTPS_PROXY / HTTP_PROXY / NO_PROXY env vars are honored).
	ProxyURL string
}

// Soundcloud is an HTTP client for scraping and downloading SoundCloud tracks.
type Soundcloud struct {
	Client    *http.Client
	UserAgent string
	Token     string
	OutputDir string
}

// NewClient builds a Soundcloud configured from opts. Returns an error if the
// proxy URL is malformed.
func NewClient(opts Options) (*Soundcloud, error) {
	transport, err := buildTransport(opts.ProxyURL)
	if err != nil {
		return nil, err
	}
	return &Soundcloud{
		Client:    &http.Client{Transport: transport},
		UserAgent: userAgent,
		Token:     opts.Token,
		OutputDir: opts.OutputDir,
	}, nil
}

// buildTransport returns an http.Transport that uses proxyURL when non-empty,
// or Go's default proxy-from-environment behavior otherwise.
func buildTransport(proxyURL string) (*http.Transport, error) {
	t := http.DefaultTransport.(*http.Transport).Clone()
	if proxyURL == "" {
		t.Proxy = http.ProxyFromEnvironment
		return t, nil
	}
	parsed, err := url.Parse(proxyURL)
	if err != nil {
		return nil, fmt.Errorf("invalid proxy URL %q: %w", proxyURL, err)
	}
	t.Proxy = http.ProxyURL(parsed)
	return t, nil
}

// authedGet performs a GET request, attaching the OAuth header only when the
// URL targets api-v2.soundcloud.com and a token is configured. Other hosts
// (the page itself, CDN segments, asset bundles) don't accept the header.
func (s *Soundcloud) authedGet(rawURL string) (*http.Response, error) {
	req, err := http.NewRequest("GET", rawURL, nil)
	if err != nil {
		return nil, err
	}
	req.Header.Set("User-Agent", s.UserAgent)
	if s.Token != "" && strings.Contains(rawURL, apiV2Host) {
		req.Header.Set(authHeader, authHeaderValue+s.Token)
	}
	return s.Client.Do(req)
}
