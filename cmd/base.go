package cmd

import (
	"github.com/imthaghost/scdl/pkg/soundcloud"
)

func scdl(args []string) {
	url := args[0]

	// download song
	soundcloud.Download(url)

}
