package transfer

import (
	"regexp"
)

var matcher = regexp.MustCompile("https://transfer\\.sh/[0-9a-zA-Z]+/.*")

func (b transfer) LinkMatcher(v string) bool {
	return matcher.MatchString(v)
}
