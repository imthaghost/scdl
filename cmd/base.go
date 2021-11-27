package cmd

import "github.com/imthaghost/scdl/pkg/soundcloud"

func scdlSaveSong(url string, path string) {
	soundcloud.ExtractSong(url, path)
}

func scdlSavePlaylist(url string, path string) {
	// ToDo: implement request for song list
}

func scdlSaveUserLikes(url string, path string) {
	// ToDo: implement request for song list
}
