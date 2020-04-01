package gofile

import (
	"fmt"
	"strings"
	"transfer/utils"
)

func (b goFile) Download(files []string) {
	for _, v := range files {
		err := b.download(v)
		if err != nil {
			fmt.Printf("download failed on %s, returns %s\n", v, err)
		}
	}
}

func (b goFile) download(v string) error {
	if !strings.Contains(v, "gofile.io/download") {
		return fmt.Errorf("url format invalid: %v", v)
	}
	err := utils.DownloadFile(b.Config.prefix, v, utils.DownloadConfig{
		Force:    b.Config.forceMode,
		Debug:    b.Config.debugMode,
		Parallel: b.Config.parallel,
		Modifier: utils.DefaultModifier,
	})
	if err != nil {
		return fmt.Errorf("failed fetching %s: %v", v, err)
	}
	return nil
}
