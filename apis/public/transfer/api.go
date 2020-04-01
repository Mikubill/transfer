package transfer

import (
	"fmt"
	"github.com/spf13/cobra"
	"transfer/apis"
	"transfer/utils"
)

var (
	Backend = new(transfer)
)

type transfer struct {
	apis.Backend
	resp     string
	Config   cbOptions
	Commands [][]string
}

func (b *transfer) SetArgs(cmd *cobra.Command) {
	cmd.Flags().BoolVarP(&b.Config.DebugMode, "verbose", "", false, "Verbose mode to debug")
	cmd.Long = fmt.Sprintf("transfer.sh - https://transfer.sh/\n\n" +
		utils.Spacer("  Size Limit: -\n") +
		utils.Spacer("  Upload Service: Hetzner Online GmbH, Germany\n") +
		utils.Spacer("  Download Service: Hetzner Online GmbH, Germany\n"))
}
