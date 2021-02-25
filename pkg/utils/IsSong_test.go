package utils

import (
	"fmt"
	"testing"

	"github.com/fatih/color"
)

func TestIsSong(t *testing.T) {
	tables := []struct {
		url      string
		expected bool
	}{
		// user url
		{"https://soundcloud.com/uiceheidd", false},
		// song url
		{"https://soundcloud.com/uiceheidd/tell-me-you-love-me", true},
		// user query url
		{"https://soundcloud.com/search?q=girl%20with%20the%20blonde%20hair%20juice%20wrld&query_urn=soundcloud%3Asearch-autocomplete%3A0dc4d347dc0a47648649ca276f5d285a", false},
		// playlist url
		{"https://soundcloud.com/soundcloud-hustle/sets/rap-new-hot", false},
		// non-url
		{"few9078", false},
		// song url
		{"https://soundcloud.com/polo-g/polo-g-feat-lil-baby-be", true},
		// song url
		{"https://soundcloud.com/user-245594022/juice-wrld-girl-with-the-blonde-hair-unreleased", false},
		// playlist url
		{"https://soundcloud.com/soundcloud-the-peak/sets/on-the-up-the-peak-hot-new", false},
	}
	for _, table := range tables {
		result := IsSong(table.url)
		expectedresult := table.expected
		if result != expectedresult {
			t.Error()
			red := color.New(color.FgRed).SprintFunc()
			fmt.Printf("%s IsSong Failed: %s , expected %t got %t \n", red("[-]"), table.url, expectedresult, result)

		} else {
			green := color.New(color.FgGreen).SprintFunc()
			fmt.Printf("%s Passing: %s \n", green("[+]"), table.url)
		}
	}
}
