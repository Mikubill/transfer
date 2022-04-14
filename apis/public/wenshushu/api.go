package wenshushu

import (
	"fmt"
	"github.com/Mikubill/transfer/apis"
	"github.com/Mikubill/transfer/utils"
	"github.com/spf13/cobra"
)

var (
	Backend = new(wssTransfer)
)

type wssTransfer struct {
	apis.Backend
	baseConf sendConfigBlock
	Config   wssOptions
	Commands [][]string
}

func (b *wssTransfer) SetArgs(cmd *cobra.Command) {
	cmd.Flags().StringVarP(&b.Config.token, "cookie", "c", "", "Your User cookie")
	cmd.Flags().IntVarP(&b.Config.blockSize, "block", "b", 1048576, "Upload Block Size")
	cmd.Flags().IntVarP(&b.Config.interval, "timeout", "t", 10, "Request retry/timeout limit")
	cmd.Flags().BoolVarP(&b.Config.singleMode, "single", "s", false, "Single Upload Mode")
	cmd.Flags().StringVarP(&b.Config.passCode, "password", "", "", "Set password")
	cmd.Flags().IntVarP(&b.Config.Parallel, "parallel", "p", 2, "Set the number of upload threads")
	cmd.Long = fmt.Sprintf("Wenshushu - https://wenshushu.cn/\n\n" +
		utils.Spacer("  Size Limit: 5G(Anonymous), ~20G(Login)\n") +
		utils.Spacer("  Upload Service: Qcloud object storage, Chengdu, China\n") +
		utils.Spacer("  Download Service: Qcloud CDN, China\n"))
}
