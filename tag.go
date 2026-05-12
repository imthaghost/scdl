package main

import (
	"fmt"

	"github.com/bogem/id3v2"
)

// SetCoverImage writes the given JPEG bytes as the ID3v2 front-cover frame of
// the mp3 at path.
func SetCoverImage(path string, image []byte) error {
	tag, err := id3v2.Open(path, id3v2.Options{Parse: true})
	if err != nil {
		return fmt.Errorf("open mp3 for tagging: %w", err)
	}
	if tag == nil {
		return fmt.Errorf("open mp3 for tagging: no tag returned")
	}
	defer tag.Close()

	tag.AddAttachedPicture(id3v2.PictureFrame{
		Encoding:    id3v2.EncodingUTF8,
		MimeType:    "image/jpeg",
		PictureType: id3v2.PTFrontCover,
		Description: "Front cover",
		Picture:     image,
	})
	return tag.Save()
}
