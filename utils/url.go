package utils

import (
	strutil "github.com/torden/go-strutil"
)

// ValidateURL checks for a valid url
func ValidURL(url string) bool {
	/*
		>>> https://google.com
		<<< true

		>>> google.com
		<<< false
	*/
	if !strutil.NewStringValidator().IsValidURL(url) {
		return false
	}

	return true
}
