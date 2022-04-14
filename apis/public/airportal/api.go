package airportal

import (
	"fmt"
	"github.com/Mikubill/transfer/apis"
	"github.com/Mikubill/transfer/utils"
	"github.com/spf13/cobra"
)

var (
	Backend = new(airPortal)
)

type airPortal struct {
	apis.Backend
	token    uploadTicket
	Config   arpOptions
	Commands [][]string
}

func (b *airPortal) SetArgs(cmd *cobra.Command) {
	cmd.Flags().StringVarP(&b.Config.token, "token", "t", "", "Set your user token (optional)")
	cmd.Flags().StringVarP(&b.Config.username, "username", "u", "", "Set your user username (optional)")
	cmd.Flags().IntVarP(&b.Config.downloads, "downloads", "d", 2, "Set downloadable count")
	cmd.Flags().IntVarP(&b.Config.hours, "hours", "", 24, "Set expire hours")
	cmd.Long = fmt.Sprintf("AirPortal - https://airportal.cn/\n\n" +
		utils.Spacer("  Size Limit: 1M(Anonymous), ~1G(Login), ~10G(Premium)\n") +
		utils.Spacer("  Upload Service: multi-server, Global\n") +
		utils.Spacer("  Download Service: multi-server, Global\n") +
		"\n  Note: when Login by token, both username and token are required\n")
}
