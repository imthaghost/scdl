package soundcloud

// Service is the interface that wraps the basic methods to interact with SoundCloud's API
type Service interface {
	GetClientID() (string, error)
	Download(url string) error
}
