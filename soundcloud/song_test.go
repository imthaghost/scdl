package soundcloud

import (
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	"testing"

	"github.com/fatih/color"
)

func testExtractSong(t *testing.T) {
	tables := []struct {
		url		 string
		expected string
	}{
		{"https://soundcloud.com/uiceheidd/tell-me-you-love-me", ""}
		{"https://soundcloud.com/uiceheidd/righteous", ""}
	}
	for _, table := range tables {
		// request to user inputed SoundCloud URL
		resp, err := http.Get(table.url)
		if err != nil {
			log.Fatalln(err)
		}
		// response
		body, err := ioutil.ReadAll(resp.Body)
		if err != nil {
			log.Fatalln(err)
		}
	}
}
	