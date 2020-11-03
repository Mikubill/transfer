package null

import (
	"regexp"
)

var matcher = regexp.MustCompile("(https://)?0x0.st/.*")

func (b null) LinkMatcher(v string) bool {
	return matcher.MatchString(v)
}
