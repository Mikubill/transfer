package filelink

import (
	"regexp"
)

var matcher = regexp.MustCompile("https://fileLink\\.io/[0-9a-z]+/.*")

func (b fileLink) LinkMatcher(v string) bool {
	return matcher.MatchString(v)
}
