package filelink

import (
	"fmt"
	"github.com/spf13/cobra"
	"transfer/apis"
	"transfer/utils"
)

var (
	Backend = new(fileLink)
)

type fileLink struct {
	apis.Backend
	//baseConf *main.MainConfig
	Config   cbOptions
	Commands [][]string
}

func (b *fileLink) SetArgs(cmd *cobra.Command) {
	cmd.Flags().BoolVarP(&b.Config.DebugMode, "verbose", "", false, "Verbose mode to debug")
	cmd.Long = fmt.Sprintf("filelink - https://filelink.io/\n\n" +
		utils.Spacer("  Size Limit: 100M\n") +
		utils.Spacer("  Upload Service: Google Cloud Taiwan/Singapore\n") +
		utils.Spacer("  Download Service: Google Cloud Taiwan/Singapore\n"))
}
