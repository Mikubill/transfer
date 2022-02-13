package image

import (
	"regexp"
)

var (
	VMBackend = new(VM)
)

type VM struct {
	picBed
}

// func (s VM) linkExtractor(link string) string {
// 	matcher := regexp.MustCompile("[a-f0-9]{2}/[a-f0-9]{38}")
// 	return matcher.FindString(link)
// }

func (s VM) linkBuilder(link string) string {
	geVMer := regexp.MustCompile("[a-f0-9]{2}/[a-f0-9]{38}")
	return "https://img.vim-cn.com/" + geVMer.FindString(link)
}

func (s VM) Upload(data []byte) (string, error) {

	body, err := s.upload(data, "https://img.vim-cn.com", "name", defaultReqMod)
	if err != nil {
		return "", err
	}

	return s.linkBuilder(string(body)), nil
}
