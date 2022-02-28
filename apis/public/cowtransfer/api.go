package cowtransfer

import (
	"fmt"
	"github.com/Mikubill/transfer/apis"
	"github.com/Mikubill/transfer/utils"

	"github.com/spf13/cobra"
)

var (
	Backend = new(cowTransfer)
)

type cowTransfer struct {
	apis.Backend
	sendConf prepareSendResp
	Config   cowOptions
	Commands [][]string
}

func (b *cowTransfer) SetArgs(cmd *cobra.Command) {
	cmd.Flags().IntVarP(&b.Config.validDays, "valid", "", 0, "Set the valid days for transfer")

	cmd.Flags().IntVarP(&b.Config.Parallel, "parallel", "p", 2, "Set the number of upload threads")
	cmd.Flags().StringVarP(&b.Config.token, "cookie", "c", "", "Your user cookie (optional)")
	cmd.Flags().StringVarP(&b.Config.authCode, "auth", "a", "", "Your auth code (optional)")
	// cmd.Flags().StringVarP(&b.Config.header, "header-file", "", "", "Extra header you want to add (read from file, optional)")

	cmd.Flags().Int64VarP(&b.Config.blockSize, "block", "", 1200000, "Upload block size")
	cmd.Flags().IntVarP(&b.Config.interval, "timeout", "t", 10, "Request retry/timeout limit in second")
	cmd.Flags().BoolVarP(&b.Config.singleMode, "single", "s", false, "Upload multi files in a single link")
	cmd.Flags().BoolVarP(&b.Config.hashCheck, "hash", "", false, "Check hash after block upload")
	cmd.Flags().StringVarP(&b.Config.passCode, "password", "", "", "Set password")
	cmd.Long = fmt.Sprintf("cowTransfer - https://cowtransfer.com/\n\n" +
		utils.Spacer("  Size Limit: 2G(Anonymous), ~100G(Login)\n") +
		utils.Spacer("  Upload Service: qiniu object storage, East China\n") +
		utils.Spacer("  Download Service: qiniu cdn, Global\n"))
}
