package utils

import (
	strutil "github.com/torden/go-strutil"
)

// ValidateURL checks for a valid url
// TODO: implement tests
func ValidateURL(url string) bool {
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

// ValidateDomain checks for a valid domain
// TODO: implement tests
func ValidateDomain(domain string) bool {
	/*
		>>> google.com
		<<< true

		>>> google
		<<< false
	*/
	if !strutil.NewStringValidator().IsValidDomain(domain) {
		return false
	}

	return true
}
