package fileio

import (
	"regexp"
)

var matcher = regexp.MustCompile("(https://)?file\\.io/\\w+")

func (b fileio) LinkMatcher(v string) bool {
	return matcher.MatchString(v)
}

// func (b fileio) DoDownload(link string, config apis.DownConfig) error {
// 	return apis.DownloadFile(&apis.DownloaderConfig{
// 		Link:        link,
// 		Config:      config,
// 		Modifier:    addHeaders,
// 		RespHandler: respHandler,
// 	})
// }

// func addHeaders(req *http.Request) {}
// func respHandler(resp *http.Response) bool {
// 	if strings.Contains(resp.Request.URL.String(), "deleted") {
// 		return false
// 	}
// 	return true
// }
