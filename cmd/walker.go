package cmd

import (
	"github.com/Mikubill/transfer/utils"
	"net/url"
	"regexp"
	"strings"
)

var urlRegex = regexp.MustCompile("(http|https)://")

func uploadWalker(items []string) []string {
	var uploadList []string
	for _, v := range items {
		if utils.IsExist(v) {
			uploadList = append(uploadList, v)
		}
	}
	return uploadList
}

func downloadWalker(items []string) []string {
	var downloadList []string

	for _, v := range items {
		if _, err := url.Parse(v); err == nil && urlRegex.MatchString(v) {
			downloadList = append(downloadList, strings.TrimSpace(v))
		}
	}
	return downloadList
}
