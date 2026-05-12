package main

import (
	"net/http"
	"strings"
)

const (
	userAgent       = "Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/58.0.3029.110 Safari/537.3"
	apiV2Host       = "api-v2.soundcloud.com"
	authHeader      = "Authorization"
	authHeaderValue = "OAuth "
)

// Soundcloud is an HTTP client for scraping and downloading SoundCloud tracks.
//
// Token is an optional SoundCloud OAuth token. When set, it's attached as
// "Authorization: OAuth <token>" on requests to api-v2.soundcloud.com, which
// unlocks Go+ high-quality transcodings and private tracks the account has
// access to.
type Soundcloud struct {
	Client    *http.Client
	UserAgent string
	Token     string
}

func NewClient(httpClient *http.Client, token string) *Soundcloud {
	if httpClient == nil {
		httpClient = &http.Client{}
	}
	return &Soundcloud{Client: httpClient, UserAgent: userAgent, Token: token}
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
