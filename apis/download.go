package apis

import (
	"fmt"
)

var (
	downConf DownConfig
)

func Download(link string, backend BaseBackend) {
	err := backend.DoDownload(link, downConf)
	if err != nil {
		fmt.Printf("Download %s Error: %s", link, err)
	}
}
