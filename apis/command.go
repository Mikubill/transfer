package apis

import (
	"github.com/spf13/cobra"
)

var (
	Crypto     bool
	Key        string
	DebugMode  bool
	SilentMode bool
)

func InitCmd(cmd *cobra.Command) {
	cmd.PersistentFlags().BoolVarP(&Crypto,
		"encrypt", "", false, "encrypt stream when upload")
	cmd.PersistentFlags().StringVarP(&Key,
		"encrypt-key", "", "", "specify the encrypt key")
	cmd.PersistentFlags().BoolVarP(&SilentMode,
		"silent", "", false, "enable silent mode to reduce output")
	cmd.PersistentFlags().BoolVarP(&DebugMode,
		"verbose", "", false, "enable verbose mode to debug")
	cmd.Flags().StringVarP(&downConf.Prefix,
		"output", "o", ".", "download to another file/folder")
	cmd.Flags().BoolVarP(&downConf.ForceMode,
		"force", "f", false, "attempt to download file regardless error")
	cmd.Flags().IntVarP(&downConf.Parallel,
		"parallel", "p", 3, "set download task count")
	cmd.Flags().StringVarP(&downConf.Ticket,
		"ticket", "t", "", "set download ticket")
	cmd.HelpFunc()
}
