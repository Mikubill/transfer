package cmd

import (
	"fmt"
	"github.com/spf13/cobra"
	"transfer/apis"
	"transfer/utils"
)

var (
	downloadCmd = &cobra.Command{
		Use:     "download",
		Aliases: []string{"down"},
		Short:   "Download a url or urls",
		Long: `
You can pass a url or urls to download.
If backend supports the url we will download it for you.

Example:
  transfer download https://c-t.work/abc

`,
		Run: func(cmd *cobra.Command, args []string) {
			links := utils.DownloadWalker(args)
			if len(links) != 0 {
				for _, item := range links {
					apis.Download(item, ParseLink(item))
				}
			} else {
				fmt.Println("Error: no url detected.")
				fmt.Println("Use \"transfer download --help\" for more information.")
			}

		},
	}
)

func init() {
	apis.InitDownloadCmd(downloadCmd)
	rootCmd.AddCommand(downloadCmd)
}
