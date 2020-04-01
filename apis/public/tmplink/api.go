package tmplink

import (
	"fmt"
	"github.com/spf13/cobra"
	"transfer/apis"
	"transfer/utils"
)

var (
	Backend = new(tmpLink)
)

type tmpLink struct {
	apis.Backend
	resp     string
	Config   wssOptions
	Commands [][]string
}

func (b *tmpLink) SetArgs(cmd *cobra.Command) {
	cmd.Flags().BoolVarP(&b.Config.DebugMode, "verbose", "", false, "Verbose mode to debug")
	cmd.Flags().StringVarP(&b.Config.token, "token", "t", "", "Set your user token (required)")
	cmd.Long = fmt.Sprintf("tmp.link - https://tmp.link/\n\n" +
		utils.Spacer("  Size Limit: 1G(Anonymous), ~10G(Login)\n") +
		utils.Spacer("  Upload Service: multi-server, Global\n") +
		utils.Spacer("  Download Service: multi-server, Global\n") +
		"\n  Note: This backend only supports login users. (use -t to set token)\n")
}
