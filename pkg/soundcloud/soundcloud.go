package soundcloud

import "net/http"

// Soundcloud struct represents a http client to make requests to SoundCloud's API
type Soundcloud struct {
	Client *http.Client // http client

	userAgent string // User-Agent header
}

// NewClient returns a new Soundcloud client
func NewClient() *Soundcloud {
	return &Soundcloud{
		Client: &http.Client{},

		userAgent: "Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/58.0.3029.110 Safari/537.3",
	}
}
