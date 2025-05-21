package cmd

import (
	"github.com/imthaghost/scdl/pkg/soundcloud"
)

func scdl(args []string) {
	url := args[0]

	// Create a new SoundCloud client
	sc := soundcloud.NewClient("", nil)
	// Download the song
	sc.Download(url)

}
