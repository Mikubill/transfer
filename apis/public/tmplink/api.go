package tmplink

import (
	"fmt"

	"github.com/Mikubill/transfer/apis"
	"github.com/Mikubill/transfer/utils"

	"github.com/spf13/cobra"
)

var (
	Backend = new(tmpLink)
)

type tmpLink struct {
	apis.Backend
	resp     string
	Config   tmpOptions
	Commands [][]string
}

func (b *tmpLink) SetArgs(cmd *cobra.Command) {
	cmd.Flags().StringVarP(&b.Config.token, "token", "t", "", "cli upload token. (leave blank for anonymous upload)")
	cmd.Long = fmt.Sprintf("tmp.link - https://tmp.link/\n\n" +
		utils.Spacer("  Size Limit: 50G(Anonymous), Unlimited(Login)\n") +
		utils.Spacer("  Upload Service: multi-server, Global\n") +
		utils.Spacer("  Download Service: multi-server, Global\n"))
}
