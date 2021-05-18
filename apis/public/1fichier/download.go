package fichier

// var matcher = regexp.MustCompile("whitecats.dip.jp/up/download/([0-9]{10})")

func (b fichier) LinkMatcher(v string) bool {
	return false
	// return matcher.MatchString(v)
}
