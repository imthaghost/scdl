package zhttp

import (
	"io/ioutil"
	"log"
	"net/http"
	"net/url"
	"time"
)

// Zhttp client
type Zhttp struct {
	client *http.Client
}

// New ...
func New(timeout time.Duration, proxy string) (*Zhttp, error) {
	zhttp := &Zhttp{
		client: http.DefaultClient,
	}

	if timeout > 0 {
		zhttp.client.Timeout = timeout
	}

	if proxy != "" {
		p, err := url.Parse(proxy)
		if err != nil {
			return nil, err
		}

		t := http.DefaultTransport.(*http.Transport)
		t.Proxy = func(*http.Request) (*url.URL, error) {
			return p, nil
		}
		zhttp.client.Transport = t
	}

	return zhttp, nil
}

// Get ...
func (zhttp *Zhttp) Get(url string) (int, []byte, error) {
	var code int
	req, err := http.Get(url)
	if err != nil {
		return 0, nil, err
	}
	// response body
	body, err := ioutil.ReadAll(req.Body)
	req.Body.Close()
	if err != nil {
		log.Fatalln(err)
	}

	code = req.StatusCode

	return code, body, nil
}
