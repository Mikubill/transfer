package cmd

import (
	"fmt"
	"github.com/spf13/cobra"
	"transfer/apis/image"
)

var (
	picCmd = &cobra.Command{
		Use:   "image",
		Short: "Upload a image to imageBed",
		Long: `
Upload a image to imageBed.
Default backend is ali.backend, you can modify it by -b flag.

Backend support:
  alibaba(ali), baidu(bd), ccupload(cc), juejin(jj), 
  netease(nt), prntscr(pr), smms(sm), sogou(sg), 
  toutiao(tt), xiaomi(xm), vimcn(vm), suning(sn)

Example:
  # simply upload
  transfer image your-image

  # specify backend to upload
  transfer image -b sn your-image

Note: Image bed backend may have strict size or format limit.
`,
		Run: func(cmd *cobra.Command, args []string) {
			files := uploadWalker(args)
			if len(files) != 0 {
				image.Upload(files)
			} else {
				fmt.Println("Error: no file detected.")
				fmt.Println("Use \"transfer image --help\" for more information.")
			}

		},
	}
)

func init() {
	rootCmd.AddCommand(picCmd)
}
