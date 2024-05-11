package utils

import (
	"net/url"
	"regexp"
	"strings"
)

// hand write regex, not tested well.
var regexpValidEmail = regexp.MustCompile(`^[+-_.a-zA-Z0-9]+@[[:alnum:]]+(\.[[:alnum:]]+)+$`)

func IsEmail(email string) bool {
	return regexpValidEmail.MatchString(email)
}

func IsURL(Url string, addScheme bool) bool {
	if !strings.Contains(Url, `://`) && addScheme {
		Url = "http://" + Url
	}
	u, err := url.Parse(Url)
	if err != nil {
		return false
	}
	if !strings.Contains(u.Host, ".") {
		return false
	}
	return true
}
