package soundcloud

import "regexp"

// GetSongName uses regex to grab the song name from the URL string
func GetSongName(url string) string {
	/*
		>>> https://soundcloud.com/chillem-637935049/1400-999-freestyle
		<<< 1400-999-freestyle

		>>> https://soundcloud.com/a-boogie-wit-da-hoodie/demons-and-angels
		<<< demons-and-angels
	*/

	var re = regexp.MustCompile(`([^\/]*)$`)

	name := re.FindString(url)

	return name
}
