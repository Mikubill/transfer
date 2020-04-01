package vimcn

import (
	"fmt"
	"regexp"
	"transfer/utils"
)

var (
	regex = regexp.MustCompile("https://img\\.vim-cn\\.com/[0-9a-f]{2}/[0-9a-f]{38}(\\.\\w+)?")
)

func (b vimcn) Download(files []string) {
	for _, v := range files {
		err := b.download(v)
		if err != nil {
			fmt.Printf("download failed on %s, returns %s\n", v, err)
		}
	}
}

func (b vimcn) download(v string) error {
	if !regex.MatchString(v) {
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
