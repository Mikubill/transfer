package catbox

import (
	"fmt"
	"regexp"
	"transfer/utils"
)

var (
	regex = regexp.MustCompile("https://files\\.catbox\\.moe/[0-9a-z]{6}(\\.\\w+)?")
)

func (b catBox) Download(files []string) {
	for _, v := range files {
		err := b.download(v)
		if err != nil {
			fmt.Printf("download failed on %s, returns %s\n", v, err)
		}
	}
}

func (b catBox) download(v string) error {
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
