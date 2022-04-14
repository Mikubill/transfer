package apis

import (
	"github.com/Mikubill/transfer/apis/methods"
	"github.com/spf13/cobra"
)

var (
	transferConfig methods.TransferConfig
	DebugMode      bool
	MuteMode       bool
)

func InitCmd(cmd *cobra.Command) {
	cmd.PersistentFlags().BoolVarP(&transferConfig.CryptoMode,
		"encrypt", "", false, "encrypt stream when upload")
	cmd.PersistentFlags().StringVarP(&transferConfig.CryptoKey,
		"encrypt-key", "", "", "specify the encrypt key")
	cmd.PersistentFlags().BoolVarP(&transferConfig.NoBarMode,
		"no-progress", "", false, "disable progress bar to reduce output")
	cmd.PersistentFlags().BoolVarP(&MuteMode,
		"silent", "", false, "enable silent mode to mute output")
	cmd.PersistentFlags().BoolVarP(&DebugMode,
		"verbose", "v", false, "enable verbose mode to debug")
	// workround
	transferConfig.DebugMode = &DebugMode

	cmd.Flags().StringVarP(&DownloadConfig.Prefix,
		"output", "o", ".", "download to another file/folder")
	cmd.Flags().BoolVarP(&DownloadConfig.ForceMode,
		"force", "f", false, "attempt to download file regardless error")
	cmd.Flags().IntVarP(&DownloadConfig.Parallel,
		"parallel", "p", 3, "set download task count")
	cmd.Flags().StringVarP(&DownloadConfig.Ticket,
		"ticket", "t", "", "set download ticket")
	cmd.HelpFunc()
}
