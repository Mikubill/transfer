package apis

import (
	"fmt"
	"os"
)

var (
	DownloadConfig DownConfig
)

func Download(link string, backend BaseBackend) {
	if MuteMode {
		transferConfig.NoBarMode = true
		os.Stdout, _ = os.Open(os.DevNull)
	}
	DownloadConfig.TransferConfig = transferConfig
	err := backend.DoDownload(link, DownloadConfig)
	if err != nil {
		fmt.Printf("Download %s Error: %s", link, err)
	}
}
