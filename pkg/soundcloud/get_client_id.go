package soundcloud

// Unfortunately, SoundCloud does not inject the client ID into the page source unless the request is made from a browser. ( Javascript is enabled )
// This is a workaround to get the client ID from the JS file that is injected into the page source.

// GetClientID returns a new generated client_id when a request is made to SoundCloud's API
func (s *Soundcloud) GetClientID() (string, error) {

	// this is the JS file that is injected into the page source
	// this can always change at some point, so we have to keep an eye on it
	_, err := s.Client.Get("https://a-v2.sndcdn.com/assets/2-1475fa5a.js")
	if err != nil {
		return "", err
	}

	// return hardcoded client ID for now
	return "nUB9ZvnjRiqKF43CkKf3iu69D8bboyKY", nil
}
