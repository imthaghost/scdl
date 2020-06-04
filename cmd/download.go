package cmd

import (
	"fmt"

	"github.com/imthaghost/scdl/soundcloud"
)

func downloadSong(args []string) {
	url := args[0]

	if Artwork == true {
		// album art image
		fmt.Println("lmao")
	}
	// song name
	soundcloud.ExtractSong(url)

}
