package cmd

import (
	"github.com/imthaghost/scdl/soundcloud"
)

func downloadSong(args []string) {
	url := args[0]

	if Find == true {

		soundcloud.Search(url)
		// exit
		return
	}
	// song name
	soundcloud.ExtractSong(url)

}
