package tmplink

import (
	"fmt"
	"regexp"
	"strings"
	"transfer/utils"
)

var (
	regex = regexp.MustCompile("[0-9a-z]{13}")
)

func (b tmpLink) Download(files []string) {
	if b.Config.token == "" {
		fmt.Println("tmpLink: token is required.")
		return
	}
	for _, v := range files {
		err := b.download(v)
		if err != nil {
			fmt.Printf("download failed on %s, returns %s\n", v, err)
		}
	}
}

func (b tmpLink) download(v string) error {
	fileID := regex.FindString(v)
	if !strings.Contains(v, "tmp.link/f") || fileID == "" {
		return fmt.Errorf("url format invalid: %v", v)
	}
	link := fmt.Sprintf("https://send.tmp.link/connect-%s-%s", b.Config.token, fileID)
	err := utils.DownloadFile(b.Config.prefix, link, utils.DownloadConfig{
		Force:    b.Config.forceMode,
		Debug:    b.Config.debugMode,
		Parallel: b.Config.parallel,
		Modifier: utils.DefaultModifier,
	})
	if err != nil {
		return fmt.Errorf("failed fetching %s: %v", link, err)
	}
	return nil
}
