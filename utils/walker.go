package utils

import (
	"fmt"
	"net/url"
	"regexp"
	"strings"
)

var urlRegex = regexp.MustCompile("(http|https)://")

func UploadWalker(items []string) []string {
	var uploadList []string
	for _, v := range items {
		if IsExist(v) {
			uploadList = append(uploadList, v)
		} else {
			fmt.Printf("error: \"%s\" is not found.\n", v)
		}
	}
	return uploadList
}

func DownloadWalker(items []string) []string {
	var downloadList []string

	for _, v := range items {
		if _, err := url.Parse(v); err == nil && urlRegex.MatchString(v) {
			downloadList = append(downloadList, strings.TrimSpace(v))
		} else {
			fmt.Printf("error: \"%s\" is incorrect.\n", v)
		}
	}
	return downloadList
}
