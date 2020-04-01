package utils

import (
	"fmt"
	"net/url"
	"strings"
)

// Walker walk through paths, uploads for local files and download for urls
func Walker(files []string, backend Backend) {
	var uploadList []string
	var downloadList []string
	for _, v := range files {
		if strings.HasPrefix(v, "-") {
			continue
		}
		if IsExist(v) {
			uploadList = append(uploadList, v)
		} else if _, err := url.Parse(v); err == nil {
			downloadList = append(downloadList, v)
		} else {
			fmt.Printf(" %s not found or incorrect.\n", v)
		}
	}
	if len(uploadList) > 0 {
		backend.Upload(uploadList)
	}
	if len(downloadList) > 0 {
		backend.Download(downloadList)
	}
}
