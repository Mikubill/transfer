package apis

import (
	"github.com/spf13/cobra"
)

var (
	Crypto bool
	Key    string
)

func InitUploadCmd(cmd *cobra.Command) {
	cmd.PersistentFlags().BoolVarP(&Crypto,
		"encrypt", "", false, "Encrypt stream when upload")
	cmd.PersistentFlags().StringVarP(&Key,
		"encrypt-key", "", "", "Specify the encrypt key")
}

func InitDownloadCmd(cmd *cobra.Command) {
	cmd.Flags().StringVarP(&downConf.Prefix,
		"output", "o", ".", "Write to another file/dictionary")
	cmd.Flags().BoolVarP(&downConf.ForceMode,
		"force", "f", false, "Attempt to download file regardless error")
	cmd.Flags().IntVarP(&downConf.Parallel,
		"parallel", "p", 3, "Set download task count")
	cmd.Flags().StringVarP(&downConf.Ticket,
		"ticket", "t", "", "Set download ticket")
	cmd.Flags().BoolVarP(&downConf.DebugMode,
		"verbose", "v", false, "Enable verbose mode to debug")
	cmd.PersistentFlags().BoolVarP(&Crypto,
		"decrypt", "", false, "Decrypt stream when download")
	cmd.PersistentFlags().StringVarP(&Key,
		"encrypt-key", "", "", "Specify the encrypt key")
}
