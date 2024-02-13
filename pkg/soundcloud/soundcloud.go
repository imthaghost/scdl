package soundcloud

import "net/http"

const (
	userAgent = "Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/58.0.3029.110 Safari/537.3"
	version   = "2.3.8"
)

// Soundcloud struct represents a http client to make requests to SoundCloud's API
type Soundcloud struct {
	// Client is the http client used to make requests
	Client *http.Client

	// UserAgent is the User-Agent header used for requests
	UserAgent string
	// AuthToken is the token used for authenticated requests
	AuthToken string

	// reuse a single struct instead of allocating one for each service on the heap.
	common service

	// Services used for talking to different parts of the Soundcloud

	// Tracks is used for talking to the tracks endpoints
	Tracks *TracksService
	// Artwork is used for talking to the artwork endpoints
	Artwork *ArtworkService
	// User is used for talking to the user endpoints
	// User *UsersService
}

type service struct {
	client *Soundcloud
}

// NewClient returns a new Soundcloud client
func NewClient(authToken string, httpClient *http.Client) *Soundcloud {
	if httpClient == nil {
		httpClient = &http.Client{}
	}

	// TODO: Add a version header

	// TODO: consume authToken for authenticated requests

	return &Soundcloud{
		Client: &http.Client{},

		UserAgent: userAgent,
	}
}
