package vimcn

import (
	"regexp"
)

var matcher = regexp.MustCompile("(https://)?img\\.vim-cn\\.com/[0-9a-f]{2}/[0-9a-f]{38}(\\.\\w+)?")

func (b vimcn) LinkMatcher(v string) bool {
	return matcher.MatchString(v)
}
