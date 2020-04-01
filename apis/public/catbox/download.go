package catbox

import (
	"regexp"
)

var (
	matcher = regexp.MustCompile("(https://)?files\\.catbox\\.moe/[0-9a-z]{6}(\\.\\w+)?")
)

func (b catBox) LinkMatcher(v string) bool {
	return matcher.MatchString(v)
}
