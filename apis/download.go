package apis

import (
	"fmt"
	"os"
)

var (
	downConf DownConfig
)

func Download(link string, backend BaseBackend) {
	if MuteMode {
		NoBarMode = true
		os.Stdout, _ = os.Open(os.DevNull)
	}
	downConf.DebugMode = DebugMode
	err := backend.DoDownload(link, downConf)
	if err != nil {
		fmt.Printf("Download %s Error: %s", link, err)
	}
}
