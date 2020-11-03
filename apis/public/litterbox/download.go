package litterbox

import (
	"regexp"
)

var (
	matcher = regexp.MustCompile("(https://)litter\\.catbox\\.moe/.*")
)

func (b litterbox) LinkMatcher(v string) bool {
	return matcher.MatchString(v)
}
