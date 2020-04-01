package airportal

import (
	"regexp"
)

var (
	matcher = regexp.MustCompile("https://airportal\\.cn/[0-9]+")
)

func (b airPortal) LinkMatcher(v string) bool {
	return matcher.MatchString(v)
}
