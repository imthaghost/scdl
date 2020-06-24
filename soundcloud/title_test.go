package soundcloud

import (
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	"testing"

	"github.com/fatih/color"
)

// Todo: Edge Cases
func TestGetTitle(t *testing.T) {
	tables := []struct {
		url      string
		expected string
	}{
		// song url
		{"https://soundcloud.com/uiceheidd/tell-me-you-love-me", "Tell Me U Luv Me (with Trippie Redd)"},
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
		result := GetTitle(body)
		expectedresult := table.expected
		if result != expectedresult {
			t.Error()
			red := color.New(color.FgRed).SprintFunc()
			fmt.Printf("%s GetTitle Failed: %s , expected %s got %s \n", red("[-]"), table.url, expectedresult, result)
		} else {
			green := color.New(color.FgGreen).SprintFunc()
			fmt.Printf("%s Passing: %s \n", green("[+]"), table.url)
		}
	}
}
