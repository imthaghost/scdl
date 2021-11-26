package scdl

import (
	"github.com/imthaghost/scdl/pkg/soundcloud"
	"log"
	"os"
	"os/user"
	"path/filepath"
	"strings"
)

func scdl(args []string) {
	url := args[0]

	if Find == true {
		soundcloud.Search(url)
		// exit
		return
	}

	if len(args) > 1 {
		path := args[1]
		// download song to specified dir
		usr, _ := user.Current()
		dir := usr.HomeDir
		// we do not process Widnows shortcut because nobody uses it (%userprofile%) =)
		// but if user works via gitbash (or some other linux-like terminal) - it should work
		if path == "~" {
			// In case of "~", which won't be caught by the "else if"
			path = dir
		} else {
			if strings.HasPrefix(path, "~/") {
				// Use strings.HasPrefix so we don't match paths like
				// "/something/~/something/"
				path = filepath.Join(dir, path[2:])
			}
			// check if specified path exists
			exists, err := exists(path)
			if exists && err == nil {
				soundcloud.ExtractSong(url, path)
			} else {
				log.Fatalf("Provided path %s does not exist: %s", path)
			}
		}
	} else {
		// download song to base dir
		dir, err := os.Getwd()
		if err != nil {
			log.Fatal(err)
		}
		soundcloud.ExtractSong(url, dir)
	}
}

func exists(path string) (bool, error) {
	_, err := os.Stat(path)
	if err == nil {
		return true, nil
	}
	if os.IsNotExist(err) {
		return false, nil
	}
	return false, err
}
