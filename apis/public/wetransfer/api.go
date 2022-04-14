package wetransfer

import (
	"fmt"
	"github.com/Mikubill/transfer/apis"
	"github.com/Mikubill/transfer/utils"
	"github.com/spf13/cobra"
)

var (
	Backend = new(weTransfer)
)

type weTransfer struct {
	apis.Backend

	baseConf *configBlock
	Config   wetOptions
	Commands [][]string
}

func (b *weTransfer) SetArgs(cmd *cobra.Command) {
	cmd.Flags().IntVarP(&b.Config.Parallel, "parallel", "p", 4, "Set the number of upload threads")
	cmd.Flags().IntVarP(&b.Config.interval, "timeout", "t", 30, "Request retry/timeout limit in second")
	cmd.Flags().BoolVarP(&b.Config.singleMode, "single", "s", false, "Single Upload Mode")
	cmd.Long = fmt.Sprintf("Wetransfer - https://wetransfer.com/\n\n" +
		utils.Spacer("  Size Limit: 2G(Anonymous), ~20G(Login)\n") +
		utils.Spacer("  Upload Service: Amazon S3, Ashburn Virginia, USA\n") +
		utils.Spacer("  Download Service: Amazon CloudFront CDN, Global\n"))
}
